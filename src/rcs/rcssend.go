package rcs

import (
	//"bytes"
	"bytes"
	"database/sql"
	"io"
	"net"
	"net/http"
	"webagent/src/common"
	"webagent/src/config"
	"webagent/src/databasepool"

	"encoding/json"
	"fmt"

	//"io/ioutil"

	//"net/http"
	c "strconv"
	s "strings"

	"sync"
	"time"

	"github.com/lib/pq"
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

		reqsql := "SELECT * FROM RCS_MESSAGE ORDER BY rcs_id LIMIT 500"

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

		rcsResValues := []common.RcsMsgRes{}

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

			currentTime := time.Now().Format("2006-01-02 15:04:05")

			rcsResValue := common.RcsMsgRes{}

			rcsResValue.Rcs_id = result["rcs_id"]
			rcsResValue.Msg_id = result["msg_id"]
			rcsResValue.User_contact = result["user_contact"]
			rcsResValue.Schedule_type = result["schdule_type"]
			rcsResValue.Msg_group_id = result["msg_group_id"]
			rcsResValue.Msg_service_type = result["msg_service_type"]
			rcsResValue.Chatbot_id = result["chatbot_id"]
			rcsResValue.Agency_id = result["agency_id"]
			rcsResValue.Messagebase_id = result["messagebase_id"]
			rcsResValue.Service_type = result["service_type"]
			rcsResValue.Expiry_option = result["expiry_option"]
			rcsResValue.Header = result["header"]
			rcsResValue.Footer = result["footer"]
			rcsResValue.Copy_allowed = result["copy_allowed"]
			rcsResValue.Body = result["body"]
			rcsResValue.Buttons = result["buttons"]
			rcsResValue.Status = status
			rcsResValue.SentTime = currentTime
			rcsResValue.Timestamp = currentTime
			rcsResValue.Error = rcsResp.Error.Message
			rcsResValue.Proc = proc

			rcsResValues = append(rcsResValues, rcsResValue)

			delrcsids = append(delrcsids, result["rcs_id"])

			procCount++
		}

		if len(rcsResValues) > 0 {
			insertRcsMsgRes(rcsResValues)
		}

		if len(delrcsids) > 0 {

			var commastr = "delete from RCS_MESSAGE where rcs_id in ("

			for i := 0; i < len(delrcsids); i++ {
				if i == 0 {
					commastr += "$1"
				} else {
					commastr += fmt.Sprintf(", $%d", i+1)
				}
			}

			commastr += ")"

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
	var authStr RcsAuth

	authStr.RcsId = config.RCSID
	authStr.RcsSecret = config.RCSPW
	authStr.GrantType = "clientCredentials"

	authBytes, _ := json.Marshal(authStr)
	req, err := http.NewRequest("POST", config.Conf.RCSSENDURL+"/corp/v1/token", bytes.NewBuffer(authBytes))
	if err != nil {
		config.Stdlog.Println("RCS Token Request 생성 실패:", err)
		return ""
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := config.GoClient.Do(req)
	if err != nil {
		// 에러가 발생한 경우 처리
		if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
			// 타임아웃 오류 처리
			config.Stdlog.Println("RCS Token 발급 타임아웃 response : ", resp, " / error : ", err.Error())
		} else {
			// 기타 오류 처리
			config.Stdlog.Println("RCS Token 발급 실패 response : ", resp, " / error : ", err.Error())
		}
		return ""
	}

	var authResp RcsAuthResp
	bodyData, _ := io.ReadAll(resp.Body)
	json.Unmarshal(bodyData, &authResp)

	defer resp.Body.Close()

	return authResp.Data.TokenInfo.AccessToken

}

func sendRcs(reswg *sync.WaitGroup, c chan<- resultStr, msg MessageInfo, temp resultStr) {
	defer reswg.Done()
	jsonData, _ := json.Marshal(msg)
	req, err := http.NewRequest("POST", config.Conf.RCSRESULTURL+"/corp/v1/message", bytes.NewBuffer(jsonData))
	if err != nil {
		config.Stdlog.Println("RCS 발송 에러 request 만들기 실패 ", err.Error())
		return
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+Token)
	resp, err := config.GoClient.Do(req)
	if err != nil {
		if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
			// 타임아웃 오류 처리
			config.Stdlog.Println("RCS 발송 타임아웃 error : ", err.Error())
		} else {
			// 기타 오류 처리
			config.Stdlog.Println("RCS 발송 실패 error : ", err.Error())
		}
		return
	} else {
		bodyData, _ := io.ReadAll(resp.Body)
		temp.Statuscode = resp.StatusCode
		temp.BodyData = bodyData
	}

	resp.Body.Close()

	c <- temp

}

func insertRcsMsgRes(RcsMsgResValue []common.RcsMsgRes) {
	tx, err := databasepool.DB.Begin()
	if err != nil {
		config.Stdlog.Println("rcssend.go / insertRcsMsgRes / rcs_message_result / 트랜잭션 초기화 실패 ", err)
	}
	defer tx.Rollback()
	rcsMsgStmt, err := tx.Prepare(pq.CopyIn("rcs_message_result", common.GetRcsColumnPq(common.RcsMsgRes{})...))
	if err != nil {
		config.Stdlog.Println("rcssend.go / insertRcsMsgRes / rcs_message_result / rcsMsgStmt 초기화 실패 ", err)
		return
	}
	for _, data := range RcsMsgResValue {
		_, err := rcsMsgStmt.Exec(data.Rcs_id, data.Msg_id, data.User_contact, data.Schedule_type, data.Msg_group_id, data.Msg_service_type, data.Chatbot_id, data.Agency_id, data.Messagebase_id, data.Service_type, data.Expiry_option, data.Header, data.Footer, data.Copy_allowed, data.Body, data.Buttons, data.Status, data.SentTime, data.Timestamp, data.Error, data.Proc)
		if err != nil {
			config.Stdlog.Println("rcssend.go / insertRcsMsgRes / rcs_message_result / rcsMsgStmt personal Exec ", err)
		}
	}

	_, err = rcsMsgStmt.Exec()
	if err != nil {
		rcsMsgStmt.Close()
		config.Stdlog.Println("rcssend.go / insertRcsMsgRes / rcs_message_result / rcsMsgStmt Exec ", err)
	}
	rcsMsgStmt.Close()
	err = tx.Commit()
	if err != nil {
		config.Stdlog.Println("rcssend.go / insertRcsMsgRes / rcs_message_result / rcsMsgStmt commit ", err)
	}
}
