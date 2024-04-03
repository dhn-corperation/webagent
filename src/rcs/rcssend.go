package rcs

import (
	//"bytes"
	"bytes"
	"webagent/src/config"
	"database/sql"
	"webagent/src/databasepool"
	"io"
	"net/http"

	"encoding/json"
	"fmt"

	//"io/ioutil"

	//"net/http"
	c "strconv"
	s "strings"

	"sync"
	"time"
)

var Token string
var SendInterval int32 = 500

type resultStr struct {
	Statuscode int
	BodyData   []byte
	Result     map[string]string
}

func Process() {
	for {
		var db = databasepool.DB
		var stdlog = config.Stdlog

		reqsql := "SELECT * FROM RCS_MESSAGE ORDER BY rcs_id LIMIT 0, 500"

		reqrows, err := db.Query(reqsql)
		if err != nil {
			stdlog.Fatal(err)
		}
		defer reqrows.Close()

		columnTypes, err := reqrows.ColumnTypes()
		if err != nil {
			stdlog.Fatal(err)
		}
		count := len(columnTypes)

		var procCount int
		procCount = 0

		resultChan := make(chan resultStr, 500)
		var reswg sync.WaitGroup

		resinsStrs := []string{}
		resinsValues := []interface{}{}
		resinsquery := `insert IGNORE into RCS_MESSAGE_RESULT(
	rcs_id ,
	msg_id ,
	user_contact ,
	schedule_type ,
	msg_group_id ,
	msg_service_type ,
	chatbot_id ,
	agency_id ,
	messagebase_id ,
	service_type ,
	expiry_option ,
	header ,
	footer ,
	copy_allowed ,
	body ,
	buttons ,
	status ,
	sentTime ,
	timestamp ,
	error,
	proc  ) values %s`

		delrcsids := []interface{}{}

		for reqrows.Next() {

			if procCount == 0 {
				var startNow = time.Now()
				var startTime = fmt.Sprintf("%02d:%02d:%02d", startNow.Hour(), startNow.Minute(), startNow.Second())
				stdlog.Printf(" ( %s ) 처리 시작 ", startTime)
			}

			if len(Token) < 10 {
				Token = getTokenInfo()
			}

			scanArgs := make([]interface{}, count)

			var msgInfo MessageInfo
			var cmnInfo CommonInfo
			var rcsInfo RcsInfo
			var body RcsBody
			var btn []RcsButton
			result := map[string]string{}

			for i, v := range columnTypes {

				switch v.DatabaseTypeName() {
				case "VARCHAR", "TEXT", "UUID", "TIMESTAMP":
					scanArgs[i] = new(sql.NullString)
					break
				case "BOOL":
					scanArgs[i] = new(sql.NullBool)
					break
				case "INT4":
					scanArgs[i] = new(sql.NullInt64)
					break
				default:
					scanArgs[i] = new(sql.NullString)
				}
			}

			err := reqrows.Scan(scanArgs...)
			if err != nil {
				stdlog.Fatal(err)
			}

			//masterData := map[string]interface{}{}

			for i, v := range columnTypes {

				switch s.ToLower(v.Name()) {
				case "rcs_id":
					if z, ok := (scanArgs[i]).(*sql.NullString); ok {
						cmnInfo.Rcs_id, _ = c.Atoi(z.String)
					}
				case "msg_id":
					if z, ok := (scanArgs[i]).(*sql.NullString); ok {
						cmnInfo.Msg_id = z.String
					}
				case "user_contact":
					if z, ok := (scanArgs[i]).(*sql.NullString); ok {
						cmnInfo.User_contact = z.String
					}
				case "schedule_type":
					if z, ok := (scanArgs[i]).(*sql.NullString); ok {
						cmnInfo.Schedule_type, _ = c.Atoi(z.String)
					}
				case "msg_group_id":
					if z, ok := (scanArgs[i]).(*sql.NullString); ok {
						cmnInfo.Msg_group_id = z.String
					}
				case "msg_service_type":
					if z, ok := (scanArgs[i]).(*sql.NullString); ok {
						cmnInfo.Msg_service_type = z.String
					}
				case "chatbot_id":
					if z, ok := (scanArgs[i]).(*sql.NullString); ok {
						rcsInfo.Chatbot_id = z.String
					}
				case "agency_id":
					if z, ok := (scanArgs[i]).(*sql.NullString); ok {
						rcsInfo.Agency_id = z.String
					}
				case "agency_key":
					if z, ok := (scanArgs[i]).(*sql.NullString); ok {
						rcsInfo.AgencyKey = z.String
					}
				case "brand_key":
					if z, ok := (scanArgs[i]).(*sql.NullString); ok {
						rcsInfo.BrandKey = z.String
					}
				case "messagebase_id":
					if z, ok := (scanArgs[i]).(*sql.NullString); ok {
						rcsInfo.Messagebase_id = z.String
					}
				case "service_type":
					if z, ok := (scanArgs[i]).(*sql.NullString); ok {
						rcsInfo.Service_type = z.String
					}
				case "expiry_option":
					if z, ok := (scanArgs[i]).(*sql.NullString); ok {
						rcsInfo.Expiry_option, _ = c.Atoi(z.String)
					}
				case "header":
					if z, ok := (scanArgs[i]).(*sql.NullString); ok {
						rcsInfo.Header = z.String
					}
				case "footer":
					if z, ok := (scanArgs[i]).(*sql.NullString); ok {
						rcsInfo.Footer = z.String
					}
				case "copy_allowed":
					if z, ok := (scanArgs[i]).(*sql.NullString); ok {
						if s.EqualFold(z.String, "1") {
							rcsInfo.Copy_allowed = true
						} else {
							rcsInfo.Copy_allowed = false
						}
					}
				case "body":
					if z, ok := (scanArgs[i]).(*sql.NullString); ok {
						// b := s.Replace(z.String, "(광고)", "", -1)
						json.Unmarshal([]byte(z.String), &body)
						rcsInfo.Body = body
					}
				case "buttons":
					if z, ok := (scanArgs[i]).(*sql.NullString); ok {
						json.Unmarshal([]byte(z.String), &btn)
						//stdlog.Println("Buttons", z.String, btn)
						if len(btn) > 0 {
							rcsInfo.Buttons = btn
						}
					}
				}

				if z, ok := (scanArgs[i]).(*sql.NullString); ok {
					result[s.ToLower(v.Name())] = z.String
				}

				if z, ok := (scanArgs[i]).(*sql.NullInt32); ok {
					result[s.ToLower(v.Name())] = string(z.Int32)
				}

				if z, ok := (scanArgs[i]).(*sql.NullInt64); ok {
					result[s.ToLower(v.Name())] = string(z.Int64)
				}

			}

			msgInfo.Common = cmnInfo
			msgInfo.Rcs = rcsInfo

			//miT, _ := json.Marshal(msgInfo)
			//miTJ := string(miT)

			//stdlog.Println("RCS msg Infor ", miTJ)
			var temp resultStr
			temp.Result = result
			reswg.Add(1)

			go sendRcs(&reswg, resultChan, msgInfo, temp)
			procCount++
		}

		if procCount > 0 {
			stdlog.Println("RCS 발송 요청 : ", procCount, " 건 처리 완료 ")
		}
		procCount = 0
		reswg.Wait()
		chanCnt := len(resultChan)

		for i := 0; i < chanCnt; i++ {

			resChan := <-resultChan
			result := resChan.Result

			var rcsResp RcsSendResp
			status := "200"

			if resChan.Statuscode == 200 {
				rcsResp.Status = "200"
			} else {
				json.Unmarshal(resChan.BodyData, &rcsResp)
			}

			//fmt.Println(rcsResp, result)
			proc := "N"

			if !s.EqualFold(rcsResp.Status, "200") {
				proc = "P"
				status = "fail"
			}

			resinsStrs = append(resinsStrs, "(?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,now(),now(),?,?)")
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
			resinsValues = append(resinsValues, rcsResp.Error.Message)
			resinsValues = append(resinsValues, proc)

			delrcsids = append(delrcsids, result["rcs_id"])
			procCount++
		}

		if len(resinsStrs) > 0 {
			stmt := fmt.Sprintf(resinsquery, s.Join(resinsStrs, ","))
			_, err := db.Exec(stmt, resinsValues...)

			if err != nil {
				stdlog.Println("RCS Result Table Insert 처리 중 오류 발생 " + err.Error())
			}
		}

		if len(delrcsids) > 0 {

			var commastr = "delete from RCS_MESSAGE where rcs_id in ("

			for i := 1; i < len(delrcsids); i++ {
				commastr = commastr + "?,"
			}

			commastr = commastr + "?)"

			_, err1 := db.Exec(commastr, delrcsids...)

			if err1 != nil {
				stdlog.Fatal(err)
			}
		}
		if procCount > 0 {
			stdlog.Println("RCS 발송 : ", procCount, " 건 처리 완료 ")
			SendInterval = 1
		} else {
			SendInterval = 1000
		}
		Token = ""
		time.Sleep(time.Millisecond * time.Duration(SendInterval))
	}
}

func getTokenInfo() string {
	config.Stdlog.Println("RCS 토큰 가져오기 시작 =============")

	var authStr RcsAuth

	authStr.RcsId = config.RCSID
	authStr.RcsSecret = config.RCSPW
	authStr.GrantType = "clientCredentials"
	/*
		resp, err := config.Client.R().
			SetHeaders(map[string]string{"Content-Type": "application/json"}).
			SetBody(authStr).
			Post(config.RCSSENDURL + "/corp/v1/token")
	*/

	//resultReqJson, _ := json.Marshal(resultReq)
	//auth := []byte(authStr)
	authBytes, err := json.Marshal(authStr)
	if err != nil {
		config.Stdlog.Println("JSON 직렬화 실패:", err)
		return ""
	}

	config.Stdlog.Println("JSON 직렬화 성공:", authBytes)

	// 요청 생성
	req, err := http.NewRequest("POST", config.Config.RCSSENDURL+"/corp/v1/token", bytes.NewBuffer(authBytes))
	if err != nil {
		config.Stdlog.Println("http.NewRequest 요청 생성 실패:", err)
		return ""
	}
	req.Header.Set("Content-Type", "application/json")

	config.Stdlog.Println("http.NewRequest 요청 생성 성공:", req)

	// HTTP 클라이언트 생성 및 요청 보내기
	resp, err2 := RCSClient.Do(req)

	//fmt.Println("SEND :", resp, err)
	//config.Stdlog.Println("SEND :", resp, err)

	config.Stdlog.Println("http.NewRequest 요청 보내기:", resp)

	if err2 == nil {
		var authResp RcsAuthResp
		// 응답 바디 읽기
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			config.Stdlog.Println("응답 바디 읽기 실패:", err)
			return ""
		}
		config.Stdlog.Println("응답 바디 읽기 성공:", body)

		// 응답 바디를 맵으로 매핑
		err = json.Unmarshal(body, &authResp)
		if err != nil {
			config.Stdlog.Println("JSON 매핑 실패:", authResp)
			return ""
		}
		config.Stdlog.Println("JSON 매핑 성공:", err)
		config.Stdlog.Println("RCS 토큰 가져오기 끝 =============")
		//var authResp RcsAuthResp
		//json.Unmarshal(resp.Body(), &authResp)
		return authResp.Data.TokenInfo.AccessToken
	} else {
		config.Stdlog.Println("Token receipt fail. - ", resp, err)
	}
	defer resp.Body.Close()

	return ""
}

func sendRcs(reswg *sync.WaitGroup, c chan<- resultStr, msg MessageInfo, temp resultStr) {
	defer reswg.Done()

	//mapB, _ := json.Marshal(msg)
	//fmt.Println(string(mapB))

	resp, err := config.Client.R().
		SetHeaders(map[string]string{"Content-Type": "application/json", "Authorization": "Bearer " + Token}).
		SetBody(msg).
		Post(config.Config.RCSSENDURL + "/corp/v1/message")

	//fmt.Println("SEND :", resp, err)

	if err != nil {
		config.Stdlog.Println("RCS 메시지 서버 호출 오류 : ", err)
		temp.Statuscode = 499
		temp.BodyData = []byte("{\"status\": \"499\", \"error\": { \"code\": \"99999\", \"message\": \"Send Server Error\" } }")
	} else {
		//config.Stdlog.Println(resp.StatusCode(), resp.Body())
		temp.Statuscode = resp.StatusCode()
		temp.BodyData = resp.Body()
	}
	c <- temp

}
