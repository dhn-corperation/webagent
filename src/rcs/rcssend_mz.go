package rcs

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"
	"sync"
	"time"

	"webagent/src/config"
	"webagent/src/databasepool"
)

// 메가버드 API 키 (설정에서 관리)
const MegabirdSendUrl = "https://api.megabird.com/v1/openapi/rcs/send"

func ProcessMZ(ctx context.Context) {
	config.Stdlog.Println("Rcs MZ(Megabird) - 발송 프로세스 시작")
	var wg sync.WaitGroup
	for {
		select {
		case <-ctx.Done():
			config.Stdlog.Println("Rcs MZ(Megabird) - process가 15초 후에 종료")
			time.Sleep(15 * time.Second)
			config.Stdlog.Println("Rcs MZ(Megabird) - process 종료 완료")
			return
		default:
			wg.Add(1)
			go rcsProcessMZ(&wg)
			wg.Wait()
		}
	}
}

func rcsProcessMZ(wg *sync.WaitGroup) {
	defer wg.Done()
	defer func() {
		if r := recover(); r != nil {
			config.Stdlog.Println("Rcs MZ - rcssend panic 발생 원인 : ", r)
			if err, ok := r.(error); ok {
				if strings.Contains(err.Error(), "connection refused") {
					for {
						config.Stdlog.Println("Rcs MZ - rcssend send ping to DB")
						if dbErr := databasepool.DB.Ping(); dbErr == nil {
							break
						}
						time.Sleep(10 * time.Second)
					}
				}
			}
		}
	}()
	var db = databasepool.DB
	var stdlog = config.Stdlog

	// 🚀 platform = 'MZ' 인 데이터 500건씩 조회
	reqsql := "SELECT rcs_id, msg_id, user_contact, schedule_type, msg_group_id, msg_service_type, chatbot_id, agency_id, messagebase_id, service_type, expiry_option, header, footer, copy_allowed, body, buttons, agency_key, (select distinct brd_brand from cb_wt_brand b where a.brand_key = b.brd_brand_key) as brandId FROM RCS_MESSAGE a WHERE platform = 'MZ' ORDER BY rcs_id LIMIT 0, 500"

	reqrows, err := db.Query(reqsql)
	if err != nil {
		panic(err)
	}
	defer reqrows.Close()

	var procCount int = 0
	resultChan := make(chan resultStr, 500)
	var reswg sync.WaitGroup

	for reqrows.Next() {
		var rcsID, expiryOpt, header string
		var msgID, userContact, msgGroupID, msgServiceType, chatbotID, agencyID, messageBaseID, serviceType, footer, body, buttons, agencyKey, brandId sql.NullString
		var scheduleType int
		var copyAllowed int

		err := reqrows.Scan(&rcsID, &msgID, &userContact, &scheduleType, &msgGroupID, &msgServiceType, &chatbotID, &agencyID, &messageBaseID, &serviceType, &expiryOpt, &header, &footer, &copyAllowed, &body, &buttons, &agencyKey, &brandId)
		if err != nil {
			continue
		}

		// 1. KT DB 데이터를 Megabird 규격으로 매핑
		reqData := MegabirdRcsRequest{
			MessageBaseID: messageBaseID.String,
			BrandID:       brandId.String, // 메가버드는 brandId 필수
			SnPhnum:       chatbotID.String, // 발신번호 (chatbot_id 매핑)
			MessageReceiverList: []MegabirdReceiver{
				{Mbnum: userContact.String},
			},
			ExpiryOpt:   expiryOpt,
			SubTxtUseYn: "N",
			MsgMetaData: map[string]interface{}{
				"rcs_id": rcsID,
				"msg_id": msgID.String,
			},
		}

		// RcsMsgType 변환 (DB의 RCSLMS -> 메가버드의 RCL)
		svcType := strings.ToUpper(serviceType.String)
		if strings.Contains(svcType, "LMS") {
			reqData.RcsMsgType = "RCL"
		} else if strings.Contains(svcType, "MMS") {
			reqData.RcsMsgType = "RCM"
		} else {
			reqData.RcsMsgType = "RCS"
		}

		if header == "1" {
			reqData.AdIncluYn = "Y"
		} else {
			reqData.AdIncluYn = "N"
		}

		reqData.UseDefault080 = "N"

		if copyAllowed == 1 {
			reqData.CopyYn = "Y"
		} else {
			reqData.CopyYn = "N"
		}

		// Body (JSON) 파싱
		var parsedBody KtBodyStruct
		if body.Valid {
			json.Unmarshal([]byte(body.String), &parsedBody)
		}

		// 💡 SMS(RCS 타입)인 경우 제목(Title)을 빈 문자열로 처리하여 JSON에서 완전 제거(omitempty 발동)
		if reqData.RcsMsgType == "RCS" {
			parsedBody.Title = ""
		}

		// 💡 버튼 파싱 (형님이 주신 KT 구조 -> 메가버드 양식 변환)
		var mzButtons []map[string]interface{}
		if buttons.Valid && len(buttons.String) > 5 {
			mzButtons = parseKTButtonsToMegabird(buttons.String) 
		}

		reqData.RcsContent = MegabirdRcsContent{
			MsgData: []MegabirdMsgData{
				{
					Title:       parsedBody.Title,
					Description: parsedBody.Description,
					Buttons:     mzButtons,
				},
			},
		}

		// 2. 비동기 발송 스레드 출발
		resultMap := map[string]string{
			"rcs_id":           rcsID,
			"msg_id":           msgID.String,
			"user_contact":     userContact.String,
			"schedule_type":    fmt.Sprintf("%d", scheduleType),
			"msg_group_id":     msgGroupID.String,
			"msg_service_type": msgServiceType.String,
			"chatbot_id":       chatbotID.String,
			"agency_id":        agencyID.String,
			"messagebase_id":   messageBaseID.String,
			"service_type":     serviceType.String,
			"expiry_option":    expiryOpt,
			"header":           header,
			"footer":           footer.String,
			"copy_allowed":     fmt.Sprintf("%d", copyAllowed),
			"body":             body.String,
			"buttons":          buttons.String,
			"platform":         "MZ",
		}

		var temp resultStr
		temp.Result = resultMap
		reswg.Add(1)

		go sendRcsMZ(&reswg, resultChan, reqData, temp)
		procCount++
	}

	if procCount > 0 {
		stdlog.Println("Rcs MZ - 발송 요청 : ", procCount, " 건 처리 완료 ")
	}
	reswg.Wait()
	
	chanCnt := len(resultChan)

	// 💡 메가버드 발송 응답 처리를 위한 Bulk Insert 변수 세팅
	resinsStrs := []string{}
	resinsValues := []interface{}{}
	resinsquery := `insert IGNORE into RCS_MESSAGE_RESULT(
rcs_id , msg_id , user_contact , schedule_type , msg_group_id , msg_service_type , 
chatbot_id , agency_id , messagebase_id , service_type , expiry_option , header , 
footer , copy_allowed , body , buttons , status , sentTime , timestamp , error, proc, platform, megabird_msg_grp_no) 
values %s`

	delrcsids := []interface{}{}

	// 🚀 큐(Channel)에 담긴 발송 결과 까보기
	for i := 0; i < chanCnt; i++ {
		resChan := <-resultChan
		result := resChan.Result

		var rcsResp MegabirdRcsResponse
		
		status := "200"
		proc := "N"
		errMsg := ""
		var msgGrpNo int64 = 0 // 💡 그룹번호 초기화

		json.Unmarshal(resChan.BodyData, &rcsResp)

		if resChan.Statuscode == 200 && rcsResp.Success {
			status = "200"
			msgGrpNo = rcsResp.Data.MsgGrpNo // 💡 성공 시 메가버드 번호 추출
		} else {
			proc = "P"
			status = "fail"
			if len(rcsResp.Description) > 0 {
				errMsg = rcsResp.Description[0].Message
			} else {
				errMsg = rcsResp.Message 
			}
		}

		// 💡 2. ?(물음표) 갯수 맨 뒤에 1개 더 추가
		resinsStrs = append(resinsStrs, "(?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,now(),now(),?,?,?,?)")
		resinsValues = append(resinsValues, result["rcs_id"])
		resinsValues = append(resinsValues, result["msg_id"])
		resinsValues = append(resinsValues, result["user_contact"])
		resinsValues = append(resinsValues, result["schedule_type"])
		resinsValues = append(resinsValues, result["msg_group_id"])
		resinsValues = append(resinsValues, result["msg_service_type"])
		resinsValues = append(resinsValues, result["chatbot_id"])
		resinsValues = append(resinsValues, result["agency_id"])
		resinsValues = append(resinsValues, result["messagebase_id"])
		resinsValues = append(resinsValues, result["service_type"])
		resinsValues = append(resinsValues, result["expiry_option"])
		resinsValues = append(resinsValues, result["header"])
		resinsValues = append(resinsValues, result["footer"])
		resinsValues = append(resinsValues, result["copy_allowed"])
		resinsValues = append(resinsValues, result["body"])
		resinsValues = append(resinsValues, result["buttons"])
		resinsValues = append(resinsValues, status)
		resinsValues = append(resinsValues, errMsg) // 파싱한 에러 메시지 삽입
		resinsValues = append(resinsValues, proc)
		resinsValues = append(resinsValues, "MZ") // 메가버드 플랫폼 코드 MZ
		resinsValues = append(resinsValues, msgGrpNo) // 💡 3. 마지막에 메가버드 그룹번호 바인딩

		delrcsids = append(delrcsids, result["rcs_id"])
	}

	// 💡 결과 테이블 Bulk Insert
	if len(resinsStrs) > 0 {
		stmt := fmt.Sprintf(resinsquery, strings.Join(resinsStrs, ","))
		_, err := db.Exec(stmt, resinsValues...)

		if err != nil {
			config.Stdlog.Println("Rcs MZ - Result Table Insert 처리 중 오류 발생 " + err.Error())
		}
	}

	// 💡 기존 대기열 큐 테이블에서 발송 완료된 건 Bulk Delete
	if len(delrcsids) > 0 {
		var commastr = "delete from RCS_MESSAGE where rcs_id in ("
		for i := 1; i < len(delrcsids); i++ {
			commastr = commastr + "?,"
		}
		commastr = commastr + "?)"

		_, err1 := db.Exec(commastr, delrcsids...)
		if err1 != nil {
			config.Stdlog.Println("Rcs MZ - Delete 오류 발생", err1)
		}
	}

	if procCount > 0 {
		config.Stdlog.Println("Rcs MZ - 발송 및 결과 처리 : ", procCount, " 건 완료 ")
	}
}

func sendRcsMZ(reswg *sync.WaitGroup, c chan<- resultStr, reqData MegabirdRcsRequest, temp resultStr) {
	defer reswg.Done()

	resp, err := config.Client.R().
		SetHeaders(map[string]string{
			"Content-Type":  "application/json",
			"Authorization": config.Conf.MZRCSKEY, 
		}).
		SetBody(reqData).
		Post(MegabirdSendUrl)

	if err != nil {
		config.Stdlog.Println("Rcs MZ - 메시지 서버 호출 오류 : ", err)
		temp.Statuscode = 499
		temp.BodyData = []byte(`{"code": "99999", "message": "Send Server Error"}`)
	} else {
		temp.Statuscode = resp.StatusCode()
		temp.BodyData = resp.Body()
	}
	c <- temp
}

// 💡 3. 버튼 파서 함수 추가 (KT 샘플 데이터 -> 메가버드 양식 변환)
func parseKTButtonsToMegabird(ktBtnJson string) []map[string]interface{} {
	// 형님이 주셨던 KT 샘플 JSON 구조체 맵핑
	var ktButtons []struct {
		Suggestions []struct {
			Action struct {
				UrlAction *struct {
					OpenUrl struct {
						Url string `json:"url"`
					} `json:"openUrl"`
				} `json:"urlAction,omitempty"`
				DisplayText string `json:"displayText"`
			} `json:"action"`
		} `json:"suggestions"`
	}

	var mzButtons []map[string]interface{}

	if err := json.Unmarshal([]byte(ktBtnJson), &ktButtons); err != nil {
		config.Stdlog.Println("❌ 버튼 JSON 파싱 에러:", err)
		return mzButtons
	}

	for _, btnGrp := range ktButtons {
		for _, sugg := range btnGrp.Suggestions {
			mzBtn := make(map[string]interface{})
			mzBtn["buttonName"] = sugg.Action.DisplayText

			if sugg.Action.UrlAction != nil {
				mzBtn["type"] = "URL"
				mzBtn["url"] = sugg.Action.UrlAction.OpenUrl.Url
			} else {
				// 기본값 또는 다른 버튼 액션 타입 대비
				mzBtn["type"] = "URL" 
			}

			mzButtons = append(mzButtons, mzBtn)
		}
	}

	return mzButtons
}