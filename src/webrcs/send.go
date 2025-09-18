package webrcs

import (
	"fmt"
	"sync"
	"time"
	"context"
	"strconv"
	s "strings"
	"database/sql"
	"encoding/json"

	"webagent/src/config"
	"webagent/src/databasepool"
)

var token string

var tickSql string
var uSql string
var reqSql string
var resSucQuery string
var resTokenQuery string
var resRetryQuery string
var updFailSql string

func RcsProc(ctx context.Context) {
	config.Stdlog.Println("(신) Rcs - 발송 프로세스 시작")
	rcsSendInit()
	procCnt := 0

	for {
		if procCnt < 1 {
			select {
				case <- ctx.Done():
				    config.Stdlog.Println("(신) Rcs - process가 15초 후에 종료")
				    time.Sleep(15 * time.Second)
				    config.Stdlog.Println("(신) Rcs - process 종료 완료")
				    return
				default:
					var count sql.NullInt64

					cnterr := databasepool.DB.QueryRowContext(ctx, tickSql).Scan(&count)

					if cnterr != nil && cnterr != sql.ErrNoRows {
						config.Stdlog.Println("(신) Rcs - process DHN_REQUEST_RCS 테이블 조회 에러 : " + cnterr.Error())
						time.Sleep(10 * time.Second)
					} else {
						if count.Valid && count.Int64 > 0 {
							tx, err := databasepool.DB.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelReadCommitted})

							if err != nil {
								config.Stdlog.Println("(신) Rcs - tx 초기화 에러 : ", err)
								continue
							}

							var startNow = time.Now()
							var groupNo = fmt.Sprintf("%02d%02d%02d%09d", startNow.Hour(), startNow.Minute(), startNow.Second(), startNow.Nanosecond()) + strconv.Itoa(procCnt)

							
							updateRows, err := tx.Exec(uSql, groupNo)

							if err != nil {
								config.Stdlog.Println("(신) Rcs - send_group 갱신 에러 : ", err)
								tx.Rollback()
								continue
							}

							rowCount, err := updateRows.RowsAffected()

							if err != nil {
								config.Stdlog.Println("(신) Rcs - RowsAffected 에러 : ", err)
								tx.Rollback()
								continue
							}

							if rowCount == 0 {
								tx.Rollback()
								time.Sleep(500 * time.Millisecond)
								continue
							}

							if err := tx.Commit(); err != nil {
								config.Stdlog.Println("(신) Rcs - 트랜잭션 커밋 에러 : ", err)
								tx.Rollback()
								continue
							}

							procCnt++
							config.Stdlog.Println("(신) Rcs - 발송 처리 시작 ( ", groupNo, " ) : ", rowCount, " 건  ( Proc Cnt :", procCnt, ") - START")

							go func() {
								defer func() {
									procCnt--
								}()
								sendProcess(groupNo, procCnt)
							}()
						} else {
							time.Sleep(500 * time.Millisecond)
						}
					}
			}
		}
	}
}

func rcsSendInit(){
	tickSql = `
		select
			count(1) as cnt
		from
			` + config.Conf.RCSTABLE + `
		where
			status = '1'
			and send_group is null
		limit 1`

	uSql = `
		update
			` + config.Conf.RCSTABLE + `
		set
			send_group = ?
		where
			status = '1'
			and send_group is null
		limit 500`

	reqSql = `
		select
			a.rr_id,
			a.msg_id,
			a.user_contract,
			b.schedule_type,
			b.msg_group_id,
			b.msg_service_type,
			b.chatbot_id,
			b.agency_id,
			b.agency_key,
			b.brand_key,
			b.messagebase_id,
			b.service_type,
			b.expire_option,
			a.header,
			a.footer,
			b.cdr_id,
			b.copy_allowed,
			a.body,
			a.buttons,
			a.chip_list,
			a.reply_id
		from
			` + config.Conf.RCSTABLE + ` a
		left join
			cb_wt_msg_sent_rcs b on a.mstr_id = b.mstr_id
		where
			a.status = 1
			and a.send_group = ?
		limit 500`

	resSucQuery = `
		update
			` + config.Conf.RCSTABLE + `
		set
			status = 2,
			send_dt = now()
		where
			rr_id in (%s)`

	resTokenQuery = `
		update
			` + config.Conf.RCSTABLE + `
		set
			send_group = null,
			result_message = if(result_message = '' or result_message is null, '토큰 발급 실패', ',토큰 발급 실패')
		where
			rr_id in (%s)`

	resRetryQuery = `
		update
			` + config.Conf.RCSTABLE + `
		set
			send_group = null,
			result_message = if(result_message = '' or result_message is null, '메시지 서버 호출 오류', ',메시지 서버 호출 오류')
		where
			rr_id in (%s)`

	updFailSql = `
		update
			` + config.Conf.RCSTABLE + `
		set
			status = 3,
			result_code = ?,
			result_message = if(result_message = '' or result_message is null, ?, concat(',', ?))
		where
			rr_id = ?`
}

func sendProcess(groupNo string, procCnt int) {
	defer func() {
		if r := recover(); r != nil {
			config.Stdlog.Println("(신) Rcs - panic 발생 원인 : ", r)
			if err, ok := r.(error); ok {
				if s.Contains(err.Error(), "connection refused") {
					for {
						config.Stdlog.Println("(신) Rcs - send ping to DB")
						err := databasepool.DB.Ping()
						if err == nil {
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

	reqRows, err := db.Query(reqSql, groupNo)
	if err != nil {
		panic(err)
	}
	defer reqRows.Close()

	columnTypes, err := reqRows.ColumnTypes()
	if err != nil {
		panic(err)
	}
	count := len(columnTypes)

	resultChan := make(chan DHNResultStr, 500)
	var reswg sync.WaitGroup

	resSucId := []interface{}{}

	resTokenId := []interface{}{}

	resRetryId := []interface{}{}

	failCount := 0
	sendCount := 0

	for reqRows.Next() {
		var id sql.NullString
		reqRows.Scan(&id)

		if len(token) < 10 {
			rcsAuthResponse, err := rcsAuthRequest.getTokenInfo()
			if err != nil {
				config.Stdlog.Println("(신) Rcs - 토큰 발급 통신 실패 err : ", err)
				time.Sleep(1 * time.Second)
				resTokenId = append(resTokenId, id.String)
				continue
			} else {
				if rcsAuthResponse.Status == "200" {
					token = rcsAuthResponse.Data.TokenInfo.AccessToken
				} else {
					config.Stdlog.Println("(신) Rcs - 토큰 발급 실패 err : ", rcsAuthResponse.Error.Message)
					time.Sleep(1 * time.Second)
					resTokenId = append(resTokenId, id.String)
					continue
				}
			}
		}

		sendCount++

		scanArgs := make([]interface{}, count)

		var rcsSendRequest RcsSendRequest
		var rcsCommonInfo RcsCommonInfo
		var rcsInfo RcsInfo
		var rcsBody RcsBody
		var rcsButtons RcsButtons
		var rcsChipList []RcsChipList
		// var rcsLegacyInfo *RcsLegacyInfo

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

		reqRows.Scan(scanArgs...)

		for i, v := range columnTypes {

			switch s.ToLower(v.Name()) {
				case "msg_id":
					if z, ok := (scanArgs[i]).(*sql.NullString); ok {
						rcsCommonInfo.MsgId = z.String
					}
				case "user_contact":
					if z, ok := (scanArgs[i]).(*sql.NullString); ok {
						rcsCommonInfo.UserContact = z.String
					}
				case "schedule_type":
					if z, ok := (scanArgs[i]).(*sql.NullString); ok {
						rcsCommonInfo.ScheduleType, _ = strconv.Atoi(z.String)
					}
				case "msg_group_id":
					if z, ok := (scanArgs[i]).(*sql.NullString); ok {
						rcsCommonInfo.MsgGroupId = z.String
					}
				case "msg_service_type":
					if z, ok := (scanArgs[i]).(*sql.NullString); ok {
						rcsCommonInfo.MsgServiceType = z.String
					}
				case "chatbot_id":
					if z, ok := (scanArgs[i]).(*sql.NullString); ok {
						rcsInfo.ChatbotId = z.String
					}
				case "agency_id":
					if z, ok := (scanArgs[i]).(*sql.NullString); ok {
						rcsInfo.AgencyId = z.String
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
						rcsInfo.MessagebaseId = z.String
					}
				case "service_type":
					if z, ok := (scanArgs[i]).(*sql.NullString); ok {
						rcsInfo.ServiceType = z.String
					}
				case "expiry_option":
					if z, ok := (scanArgs[i]).(*sql.NullString); ok {
						rcsInfo.ExpiryOption, _ = strconv.Atoi(z.String)
					}
				case "header":
					if z, ok := (scanArgs[i]).(*sql.NullString); ok {
						rcsInfo.Header = z.String
					}
				case "footer":
					if z, ok := (scanArgs[i]).(*sql.NullString); ok {
						rcsInfo.Footer = z.String
					}
				case "cdr_id":
					if z, ok := (scanArgs[i]).(*sql.NullString); ok {
						rcsInfo.CdrId = z.String
					}
				case "copy_allowed":
					if z, ok := (scanArgs[i]).(*sql.NullString); ok {
						if z.String == "1" {
							rcsInfo.CopyAllowed = true
						} else {
							rcsInfo.CopyAllowed = false
						}
					}
				case "body":
					if z, ok := (scanArgs[i]).(*sql.NullString); ok {
						json.Unmarshal([]byte(z.String), &rcsBody)
						rcsInfo.Body = rcsBody
					}
				case "buttons":
					if z, ok := (scanArgs[i]).(*sql.NullString); ok {
						json.Unmarshal([]byte(z.String), &rcsButtons)
						rcsInfo.Buttons = rcsButtons
					}
				case "chip_list":
					if z, ok := (scanArgs[i]).(*sql.NullString); ok {
						json.Unmarshal([]byte(z.String), &rcsChipList)
						rcsInfo.ChipList = rcsChipList
					}
				case "reply_id":
					if z, ok := (scanArgs[i]).(*sql.NullString); ok {
						rcsInfo.ReplyId = z.String
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

		rcsSendRequest.Common = rcsCommonInfo
		rcsSendRequest.Rcs = rcsInfo

		// if rcsLegacyInfo != nil {
		// 	rcsSendRequest.Legacy = rcsLegacyInfo
		// }

		var temp DHNResultStr
		temp.Result = result
		reswg.Add(1)

		go sendRcs(&reswg, resultChan, rcsSendRequest, temp)
	}

	reswg.Wait()
	
	chanCnt := len(resultChan)

	for i := 0; i < chanCnt; i++ {

		resChan := <-resultChan
		result := resChan.Result

		var rcsResponseErrorInfo RcsResponseErrorInfo

		if resChan.Statuscode == 200 {
			resSucId = append(resSucId, result["rr_id"])
		} else if resChan.Statuscode != 499 {
			json.Unmarshal(resChan.BodyData, &rcsResponseErrorInfo)
			
			db.Exec(updFailSql, rcsResponseErrorInfo.Error.Code, rcsResponseErrorInfo.Error.Message, rcsResponseErrorInfo.Error.Message, result["rr_id"])
			failCount++
		} else {
			resRetryId = append(resRetryId, result["rr_id"])
		}

	}

	if len(resSucId) > 0 {
		stmt := fmt.Sprintf(resSucQuery, getQuestionMark(resSucId))
		_, err := db.Exec(stmt, resSucId...)

		if err != nil {
			stdlog.Println("(신) Rcs - 발송 성공 상태 값 Update 처리 중 오류 발생 err : " + err.Error())
			stdlog.Println(resSucId)
		}
	}

	if len(resTokenId) > 0 {
		stmt := fmt.Sprintf(resTokenQuery, getQuestionMark(resTokenId))
		_, err := db.Exec(stmt, resTokenId...)

		if err != nil {
			stdlog.Println("(신) Rcs - 토큰 습득 처리 중 오류 데이터 Update 처리 중 오류 발생 err : " + err.Error())
			stdlog.Println(resTokenId)
		}
	}

	if len(resRetryId) > 0 {
		stmt := fmt.Sprintf(resRetryQuery, getQuestionMark(resRetryId))
		_, err := db.Exec(stmt, resRetryId...)

		if err != nil {
			stdlog.Println("(신) Rcs - 메시지 서버 호출 오류 데이터 Update 처리 중 오류 발생 err : " + err.Error())
			stdlog.Println(resRetryId)
		}
	}

	token = ""

	stdlog.Println("(신) Rcs - 발송 처리 완료 ( ", groupNo, " ) : ", len(resSucId), " / ", failCount," 건 ( Proc Cnt :", procCnt, ") - END")
}

func sendRcs(reswg *sync.WaitGroup, c chan<- DHNResultStr, rcsSendRequest RcsSendRequest, temp DHNResultStr) {
	defer reswg.Done()

	resp, err := config.Client.R().
		SetHeaders(map[string]string{"Content-Type": "application/json", "Authorization": "Bearer " + token}).
		SetBody(rcsSendRequest).
		Post(config.Conf.RCSSENDURL + "/corp/v1/message")

	if err != nil {
		config.Stdlog.Println("(신) RCS - 메시지 서버 호출 오류 : ", err)
		temp.Statuscode = 499
		temp.BodyData = []byte("{\"status\": \"499\", \"error\": { \"code\": \"99999\", \"message\": \"Send Server Error\" } }")
	} else {
		//config.Stdlog.Println(resp.StatusCode(), resp.Body())
		temp.Statuscode = resp.StatusCode()
		temp.BodyData = resp.Body()
	}
	c <- temp

}