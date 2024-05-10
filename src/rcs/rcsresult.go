package rcs

import (
	//"bytes"
	"bytes"
	"database/sql"
	"io"
	"net"
	"webagent/src/common"
	"webagent/src/config"
	"webagent/src/databasepool"

	"encoding/json"
	"fmt"
	"webagent/src/baseprice"

	//"io/ioutil"

	"net/http"
	s "strings"

	//c "strconv"

	"sync"
	"time"

	"github.com/lib/pq"
)

var RToken string
var RToken2 string
var Interval int32 = 1000
var Interval2 int32 = 60000

func ResultProcess() {
	var wg sync.WaitGroup
	for {
		wg.Add(1)
		go resultProcess(&wg)
		wg.Wait()
	}
}

func resultProcess(wg *sync.WaitGroup) {
	//config.Stdlog.Println("자도 작업 처리 실행!!")
	defer wg.Done()

	var db = databasepool.DB
	var stdlog = config.Stdlog

	var resultReq RcsResultReq

	resultReq.QueryType = "rcsId"
	resultReq.RcsId = config.RCSID
	//resultReq.MsgId = "3_41744"

	var startNow = time.Now()
	var startTime = fmt.Sprintf("%02d:%02d:%02d.%09d", startNow.Hour(), startNow.Minute(), startNow.Second(), startNow.Nanosecond())

	resultReq.QueryId = "DHN_" + startTime

	resAfter6 := `UPDATE rcs_message_result 
		SET result_status = 'success',
			status = 'success',
			proc = 'P'
		WHERE TO_TIMESTAMP(senttime, 'YYYY-MM-DD HH24:MI:SS') + INTERVAL '6 hours' < CURRENT_TIMESTAMP
		AND proc = 'N'`

	databasepool.DB.Exec(resAfter6)

	RToken = getTokenInfo()

	resultReqJson, _ := json.Marshal(resultReq)
	requestBody := []byte(resultReqJson)

	// 요청 생성
	req, err := http.NewRequest("POST", config.Conf.RCSRESULTURL+"/corp/v1/querymsgstatus", bytes.NewBuffer(requestBody))
	if err != nil {
		config.Stdlog.Println("요청 생성 실패:", err)
		return
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+RToken)

	// HTTP 클라이언트 생성 및 요청 보내기
	resp, err := config.GoClient.Do(req)
	if err != nil {
		// 에러가 발생한 경우 처리
		if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
			// 타임아웃 오류 처리
			config.Stdlog.Println("RCS querymsgstatus 타임아웃 error : ", err.Error())
		} else {
			// 기타 오류 처리
			config.Stdlog.Println("RCS querymsgstatus 실패 error : ", err.Error())
		}
		return
	} else {

		var resultInfo RcsResultInfo
		// 응답 바디 읽기
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			config.Stdlog.Println("rcsresult.go / querymsgstatus 응답 바디 읽기 실패 ", err)
			return
		}

		// 응답 바디를 맵으로 매핑
		err = json.Unmarshal(body, &resultInfo)
		if err != nil {
			config.Stdlog.Println("rcsresult.go / querymsgstatus JSON매핑 실패 ", err)
			return
		}

		rcsStatusValues := []common.RcsStatusColumn{}

		for i := 0; i < len(resultInfo.StatusInfo); i++ {
			si := resultInfo.StatusInfo[i]

			rcsStatusValue := common.RcsStatusColumn{}

			rcsStatusValue.Rcs_id = si.RcsId
			rcsStatusValue.Msg_id = si.MsgId
			rcsStatusValue.User_contact = si.UserContact
			rcsStatusValue.Status = si.Status
			rcsStatusValue.Service_type = si.ServiceType
			rcsStatusValue.Mno_info = si.MnoInfo
			rcsStatusValue.Sent_time = si.SentTime
			rcsStatusValue.Error = si.Error.Message
			rcsStatusValue.Timestamp = si.Timestamp

			rcsStatusValues = append(rcsStatusValues, rcsStatusValue)

			if len(rcsStatusValues) >= 500 {
				insertRcsStatus(rcsStatusValues)
				rcsStatusValues = []common.RcsStatusColumn{}
			}
			resUpdatestr := ` update rcs_message_result
				set result_status = '` + si.Status + `' 
				,result_error = '` + si.Error.Message + `'
				,proc = 'P'
				where proc = 'N' 
				and msg_id='` + si.MsgId + `' 
				and user_contact = '` + si.UserContact + `'`

			//stdlog.Println("RCS_MESSAGE_STATUS Table 수정 Query : " + resUpdatestr)
			databasepool.DB.Exec(resUpdatestr)
		}

		if len(rcsStatusValues) > 0 {
			insertRcsStatus(rcsStatusValues)
			rcsStatusValues = []common.RcsStatusColumn{}
		}
		/*
						resUpdatestr := `update RCS_MESSAGE_RESULT rmr
			set rmr.result_status = (select r.status from RCS_MESSAGE_STATUS   r where r.msg_id = rmr.msg_id and r.user_contact = rmr.user_contact limit 1)
			   ,rmr.result_error = (select r.error from RCS_MESSAGE_STATUS  r where r.msg_id = rmr.msg_id and r.user_contact = rmr.user_contact limit 1)
			   ,rmr.proc = 'P'
			 where rmr.proc = 'N'
			   and exists
			 (select 1 from RCS_MESSAGE_STATUS  r where r.msg_id = rmr.msg_id and r.user_contact = rmr.user_contact and r.status is not null)`

						db.Exec(resUpdatestr)
		*/

		// postgresql에서는 mariadb와는 달리 자동 형변환이 없다. 기존 mariadb에서 msg_id는 bigint, msg_group_id는 char(20) 타입이다.
		groupsql := `SELECT DISTINCT
				rmr.msg_group_id AS mst_id,
				wms.send_mem_id AS usermem_id
			FROM
				RCS_MESSAGE_RESULT rmr
			INNER JOIN
				cb_wt_msg_sent wms
				ON rmr.msg_group_id::BIGINT = wms.mst_id 
			WHERE
				rmr.proc = 'P'`

		grows, err := db.Query(groupsql)
		if err != nil {
			stdlog.Println("RCS_MESSAGE_RESULT select 오류 ", err)
		} else {
			defer grows.Close()

			var mst_id, usermem_id sql.NullString

			for grows.Next() {

				oSmsValues := []common.OSmsColumn{}
				oMmsValues := []common.OMmsColumn{}

				var scnt int = 0
				var ecnt int = 0

				grows.Scan(&mst_id, &usermem_id)

				ressql := `SELECT rmr.msg_id as msgid
					,rmr.user_contact as phnstr
					,rmr.chatbot_id as  sms_sender
					,rmr.body
					,rmr.msg_group_id as remark4
					,'00000000000000' as reserve_dt
					,(select mem_userid from cb_member cm where cm.mem_id = wms.mst_mem_id) as userid
					,COALESCE(rmr.result_status, rmr.status)  as res_status
					,COALESCE(rmr.result_error,rmr.error) as res_error
					,rmr.service_type
					,(SELECT mi.origin1_path FROM cb_mms_images mi where wms.mst_mms_content = mi.mms_id and length(mst_mms_content ) > 5 ) as mms_file1
					,(SELECT mi.origin2_path FROM cb_mms_images mi where wms.mst_mms_content = mi.mms_id and length(mst_mms_content ) > 5 ) as mms_file2
					,(SELECT mi.origin3_path FROM cb_mms_images mi where wms.mst_mms_content = mi.mms_id and length(mst_mms_content ) > 5 ) as mms_file3
					,wms.mst_sent_voucher  
					,wms.mst_mem_id as send_mem_id
					,wms.mst_type2
					,wms.mst_type3
					,(select ( case when rcs.msr_exptime < NOW() then 'N' WHEn rcs.msr_exptime IS NULL then 'N' else 'Y' END) 
						from cb_wt_msg_rcs rcs 
						where rmr.msg_group_id::BIGINT =rcs.msr_mst_id) as msr_exptime
					,wms.mst_lms_content 
				FROM RCS_MESSAGE_RESULT rmr 
					 inner join cb_wt_msg_sent wms on rmr.msg_group_id::BIGINT = wms.mst_id 
				where rmr.proc = 'P'
					and rmr.msg_group_id = $1
				`

				resrows, err := db.Query(ressql, mst_id.String)
				if err != nil {
					stdlog.Println("결과 처리 Select 오류 ", err)
				} else {
					defer resrows.Close()

					var msgid, phnstr, sms_sender, body, remark4, reserve_dt, userid, res_status, res_error, service_type, mms_file1, mms_file2, mms_file3, mst_sent_voucher, send_mem_id, mst_type2, mst_type3, msr_exptime, mst_lms_content sql.NullString

					amtValues := []common.AmtColumn{}

					for resrows.Next() {

						var amount float64
						var memo string
						var payback float64
						var admin_amt float64

						resrows.Scan(&msgid, &phnstr, &sms_sender, &body, &remark4, &reserve_dt, &userid, &res_status, &res_error, &service_type, &mms_file1, &mms_file2, &mms_file3, &mst_sent_voucher, &send_mem_id, &mst_type2, &mst_type3, &msr_exptime, &mst_lms_content)

						cprice := baseprice.GetPrice(db, send_mem_id.String, stdlog)

						switch s.ToLower(res_status.String) {
						case "success":
							db.Exec("update cb_msg_"+userid.String+" set CODE = 'RCS', MESSAGE_TYPE='rc', MESSAGE = $1, RESULT = $2 where remark4=$3 and msgid = $4", "RCS 성공", "Y", remark4.String, msgid.String)
							scnt++
						case "fail":
							amtValue := common.AmtColumn{}
							oSmsValue := common.OSmsColumn{}
							oMmsValue := common.OMmsColumn{}
							amtValue.Amt_datetime = "now()"
							amtValue.Amt_kind = "3"

							if s.EqualFold(mst_sent_voucher.String, "V") {
								switch service_type.String {
								case "RCSSMS":
									amount = cprice.V_price_rcs_sms.Float64
									payback = cprice.V_price_rcs_sms.Float64 - cprice.P_price_rcs_sms.Float64
									admin_amt = cprice.B_price_rcs_sms.Float64
									memo = "RCS SMS 발송실패 환불,바우처"
								case "RCSLMS":
									amount = cprice.V_price_rcs.Float64
									payback = cprice.V_price_rcs.Float64 - cprice.P_price_rcs.Float64
									admin_amt = cprice.B_price_rcs.Float64
									memo = "RCS LMS 발송실패 환불,바우처"
								case "RCSMMS":
									amount = cprice.V_price_rcs_mms.Float64
									payback = cprice.V_price_rcs_mms.Float64 - cprice.P_price_rcs_mms.Float64
									admin_amt = cprice.B_price_rcs_mms.Float64
									memo = "RCS MMS 발송실패 환불,바우처"
								case "RCSTMPL":
									amount = cprice.V_price_rcs_tem.Float64
									payback = cprice.V_price_rcs_tem.Float64 - cprice.P_price_rcs_tem.Float64
									admin_amt = cprice.B_price_rcs_tem.Float64
									memo = "RCS TMPL 발송실패 환불,바우처"
								}

							} else {
								switch service_type.String {
								case "RCSSMS":
									amount = cprice.C_price_rcs_sms.Float64
									payback = cprice.C_price_rcs_sms.Float64 - cprice.P_price_rcs_sms.Float64
									admin_amt = cprice.B_price_rcs_sms.Float64
									if s.EqualFold(mst_sent_voucher.String, "B") {
										memo = "RCS SMS 발송실패 환불,보너스"
									} else {
										memo = "RCS SMS 발송실패 환불"
									}
								case "RCSLMS":
									amount = cprice.C_price_rcs.Float64
									payback = cprice.C_price_rcs.Float64 - cprice.P_price_rcs.Float64
									admin_amt = cprice.B_price_rcs.Float64
									if s.EqualFold(mst_sent_voucher.String, "B") {
										memo = "RCS LMS 발송실패 환불,보너스"
									} else {
										memo = "RCS LMS 발송실패 환불"
									}
								case "RCSMMS":
									amount = cprice.C_price_rcs_mms.Float64
									payback = cprice.C_price_rcs_mms.Float64 - cprice.P_price_rcs_mms.Float64
									admin_amt = cprice.B_price_rcs_mms.Float64
									if s.EqualFold(mst_sent_voucher.String, "B") {
										memo = "RCS MMS 발송실패 환불,보너스"
									} else {
										memo = "RCS MMS 발송실패 환불"
									}
								case "RCSTMPL":
									amount = cprice.C_price_rcs_tem.Float64
									payback = cprice.C_price_rcs_tem.Float64 - cprice.P_price_rcs_tem.Float64
									admin_amt = cprice.B_price_rcs_tem.Float64
									if s.EqualFold(mst_sent_voucher.String, "B") {
										memo = "RCS TMPL 발송실패 환불,보너스"
									} else {
										memo = "RCS TMPL 발송실패 환불"
									}
								}
							}

							amtValue.Amt_amount = amount
							amtValue.Amt_memo = memo
							amtValue.Amt_reason = msgid.String + "/" + phnstr.String
							amtValue.Amt_payback = payback * -1
							amtValue.Amt_admin = admin_amt * -1

							amtValues = append(amtValues, amtValue)

							var rcsBody RcsBody
							json.Unmarshal([]byte(body.String), &rcsBody)
							if s.Contains(mst_type3.String, "wc") && s.EqualFold(msr_exptime.String, "Y") {

								stdlog.Println("RCS 실패 -> WEB(C) 발송 처리 ", mst_type3.String, msr_exptime.String, msgid.String)

								db.Exec("update cb_msg_"+userid.String+" set CODE = 'SMT', MESSAGE_TYPE='sm' where remark4 = $1  and msgid = $2", remark4.String, msgid.String)

								if s.EqualFold(mst_type3.String, "wcs") {

									oSmsValue.Sender = sms_sender.String
									oSmsValue.Receiver = phnstr.String
									oSmsValue.Msg = mst_lms_content.String
									oSmsValue.URL = ""
									if s.EqualFold(reserve_dt.String, "00000000000000") {
										oSmsValue.ReserveDT = sql.NullString{}
									} else {
										oSmsValue.ReserveDT = reserve_dt.String
									}
									oSmsValue.TimeoutDT = "null"
									oSmsValue.SendResult = "0"
									oSmsValue.Mst_id = remark4.String
									oSmsValue.Cb_msg_id = msgid.String

									oSmsValues = append(oSmsValues, oSmsValue)

									if s.EqualFold(mst_sent_voucher.String, "V") {
										amount = cprice.V_price_smt_sms.Float64
										payback = cprice.V_price_smt_sms.Float64 - cprice.P_price_smt_sms.Float64
										admin_amt = cprice.B_price_smt_sms.Float64
										memo = "웹(C) SMS,바우처"
									} else {
										amount = cprice.C_price_smt_sms.Float64
										payback = cprice.C_price_smt_sms.Float64 - cprice.P_price_smt_sms.Float64
										admin_amt = cprice.B_price_smt_sms.Float64
										if s.EqualFold(mst_sent_voucher.String, "B") {
											memo = "웹(C) SMS,보너스"
										} else {
											memo = "웹(C) SMS"
										}
									}

									amtValue.Amt_datetime = "now()"
									amtValue.Amt_kind = "P"
									amtValue.Amt_amount = amount
									amtValue.Amt_memo = memo
									amtValue.Amt_reason = msgid.String + "/" + phnstr.String
									amtValue.Amt_payback = payback
									amtValue.Amt_admin = admin_amt

									amtValues = append(amtValues, amtValue)

								} else if s.EqualFold(mst_type3.String, "wc") {

									oMmsValue.MsgGroupID = remark4.String
									oMmsValue.Sender = sms_sender.String
									oMmsValue.Receiver = phnstr.String
									oMmsValue.Subject = rcsBody.Title
									oMmsValue.Msg = mst_lms_content.String
									if s.EqualFold(reserve_dt.String, "00000000000000") {
										oMmsValue.ReserveDT = sql.NullString{}
									} else {
										oMmsValue.ReserveDT = reserve_dt.String
									}
									oMmsValue.TimeoutDT = "null"
									oMmsValue.SendResult = "0"
									oMmsValue.File_Path1 = sql.NullString{}
									oMmsValue.File_Path2 = sql.NullString{}
									oMmsValue.File_Path3 = sql.NullString{}
									oMmsValue.Mst_id = remark4.String
									oMmsValue.Cb_msg_id = msgid.String

									oMmsValues = append(oMmsValues, oMmsValue)

									if len(mms_file1.String) <= 0 {
										if s.EqualFold(mst_sent_voucher.String, "V") {
											amount = cprice.V_price_smt.Float64
											payback = cprice.V_price_smt.Float64 - cprice.P_price_smt.Float64
											admin_amt = cprice.B_price_smt.Float64
											memo = "웹(C) LMS,바우처"
										} else {
											amount = cprice.C_price_smt.Float64
											payback = cprice.C_price_smt.Float64 - cprice.P_price_smt.Float64
											admin_amt = cprice.B_price_smt.Float64
											if s.EqualFold(mst_sent_voucher.String, "B") {
												memo = "웹(C) LMS,보너스"
											} else {
												memo = "웹(C) LMS"
											}
										}
									} else {
										if s.EqualFold(mst_sent_voucher.String, "V") {
											amount = cprice.V_price_smt_mms.Float64
											payback = cprice.V_price_smt_mms.Float64 - cprice.P_price_smt_mms.Float64
											admin_amt = cprice.B_price_smt_mms.Float64
											memo = "웹(C) MMS,바우처"
										} else {
											amount = cprice.C_price_smt_mms.Float64
											payback = cprice.C_price_smt_mms.Float64 - cprice.P_price_smt_mms.Float64
											admin_amt = cprice.B_price_smt_mms.Float64
											if s.EqualFold(mst_sent_voucher.String, "B") {
												memo = "웹(C) MMS,보너스"
											} else {
												memo = "웹(C) MMS"
											}
										}

									}

									amtValue.Amt_datetime = "now()"
									amtValue.Amt_kind = "P"
									amtValue.Amt_amount = amount
									amtValue.Amt_memo = memo
									amtValue.Amt_reason = msgid.String + "/" + phnstr.String
									amtValue.Amt_payback = payback
									amtValue.Amt_admin = admin_amt

									amtValues = append(amtValues, amtValue)
								}
							} else {
								ecnt++
								db.Exec("update cb_msg_"+userid.String+" set CODE = 'RCS', MESSAGE_TYPE='rc', MESSAGE = $1, RESULT = $2 where remark4 = $3 and msgid = $4", res_error.String, "N", remark4.String, msgid.String)
							}
						}

						db.Exec("update RCS_MESSAGE_RESULT set proc='Y' where msg_group_id = $1 and proc='P' and msg_id = $2", mst_id.String, msgid.String)

					}

					if len(oSmsValues) > 0 {
						insertOSms(oSmsValues)
						oSmsValues = []common.OSmsColumn{}
					}

					if len(oMmsValues) > 0 {
						insertOMms(oMmsValues)
						oMmsValues = []common.OMmsColumn{}
					}

					if len(amtValues) > 0 {
						insertAmt(amtValues, userid.String)
						amtValues = []common.AmtColumn{}
					}

					db.Exec("update cb_wt_msg_sent set mst_rcs = COALESCE(mst_rcs,0) + $1,mst_err_rcs = COALESCE(mst_err_rcs,0) + $2, mst_wait = mst_wait - $3  where mst_id=$4", scnt, ecnt, (ecnt + scnt), mst_id.String)

					stdlog.Println("RCS 처리 : (", mst_id.String, " ) 성공 : ", scnt, " / 실패 : ", ecnt)
				}
			}

			if s.EqualFold(resultInfo.MoreToSend, "1") {
				Interval = 1
			} else {
				Interval = 1000
			}
		}
		//} else {
		//	Interval = 1000
		//}
	}

	defer resp.Body.Close()

	time.Sleep(time.Millisecond * time.Duration(Interval))

}

func RetryProcess() {

	var wg sync.WaitGroup
	for {
		wg.Add(1)
		go retryProc(&wg)
		wg.Wait()
	}
}

func retryProc(wg *sync.WaitGroup) {
	defer wg.Done()
	//config.Stdlog.Println("수작업 처리 실행!!")
	var db = databasepool.DB
	var stdlog = config.Stdlog

	var resultReq RcsResultReq

	sqlStr := "select msg_id from RCS_MESSAGE_RESULT where proc = 'T' and result_status is null"

	retryrows, err := db.Query(sqlStr)
	if err != nil {
		stdlog.Println("RCS_MESSAGE_RESULT 수작업 select 오류", err)
		return
	}
	defer retryrows.Close()

	for retryrows.Next() {
		var msg_id sql.NullString

		retryrows.Scan(&msg_id)

		resultReq.QueryType = "msgId"
		//resultReq.RcsId = config.RCSID
		resultReq.MsgId = msg_id.String

		var startNow = time.Now()
		var startTime = fmt.Sprintf("%02d:%02d:%02d.%09d", startNow.Hour(), startNow.Minute(), startNow.Second(), startNow.Nanosecond())

		resultReq.QueryId = "DHN_" + startTime

		RToken2 = getTokenInfo()

		/*
			resp, err := config.Client.R().
				SetHeaders(map[string]string{"Content-Type": "application/json", "Authorization": "Bearer " + RToken2}).
				SetBody(resultReq).
				Post(config.RCSRESULTURL + "/corp/v1/querymsgstatus")
		*/
		resultReqJson, _ := json.Marshal(resultReq)
		requestBody := []byte(resultReqJson)

		// 요청 생성
		req, err := http.NewRequest("POST", config.Conf.RCSRESULTURL+"/corp/v1/querymsgstatus", bytes.NewBuffer(requestBody))
		if err != nil {
			config.Stdlog.Println("요청 생성 실패:", err)
			return
		}
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+RToken2)

		// HTTP 클라이언트 생성 및 요청 보내기
		resp, err := config.GoClient.Do(req)
		if err != nil {
			config.Stdlog.Println("POST 요청 실패:", err)
			return
		}
		defer resp.Body.Close()

		//fmt.Println(resp, resultReq)

		if err != nil {
			config.Stdlog.Println("RCS 메시지 결과 서버 호출 오류 : ", err)
			//	return nil
		} else {

			var resultInfo RcsResultInfo
			// 응답 바디 읽기
			body, err := io.ReadAll(resp.Body)
			if err != nil {
				config.Stdlog.Println("응답 바디 읽기 실패:", err)
				return
			}

			// 응답 바디를 맵으로 매핑
			err = json.Unmarshal(body, &resultInfo)
			if err != nil {
				config.Stdlog.Println("JSON 매핑 실패:", err)
				return
			}

			// 매핑된 데이터 출력
			//fmt.Println("매핑된 데이터:", resultInfo)

			//json.Unmarshal(resp.Body(), &resultInfo)

			//fmt.Println(len(resultInfo.StatusInfo), resultInfo.StatusInfo)
			//if len(resultInfo.StatusInfo) > 0 {
			for i := 0; i < len(resultInfo.StatusInfo); i++ {
				si := resultInfo.StatusInfo[i]

				//fmt.Println(si, si.Status, si.MsgId)
				switch s.ToLower(si.Status) {
				case "success":
					db.Exec("update RCS_MESSAGE_RESULT set result_status = '" + si.Status + "', proc='P' where msg_id = '" + si.MsgId + "'")
				case "fail":
					db.Exec("update RCS_MESSAGE_RESULT set result_status = '" + si.Status + "', result_error = '" + si.Error.Message + "', proc='P' where msg_id = '" + si.MsgId + "'")
				}
			}

		}
	}
	time.Sleep(time.Millisecond * time.Duration(Interval2))

}

func insertRcsStatus(rcsStatusValues []common.RcsStatusColumn) {

	tx, err := databasepool.DB.Begin()
	if err != nil {
		config.Stdlog.Println("rcsresult.go / insertRcsStatus / rcs_message_status / 트랜잭션 초기화 실패 ", err)
	}
	defer tx.Rollback()
	rcsStmt, err := tx.Prepare(pq.CopyIn("rcs_message_status", common.GetRcsColumnPq(common.RcsStatusColumn{})...))
	if err != nil {
		config.Stdlog.Println("rcsresult.go / insertRcsStatus / rcs_message_status / rcsStmt 초기화 실패 ", err)
		return
	}
	for _, data := range rcsStatusValues {
		_, err := rcsStmt.Exec(data.Rcs_id, data.Msg_id, data.User_contact, data.Status, data.Service_type, data.Mno_info, data.Sent_time, data.Error, data.Timestamp)
		if err != nil {
			config.Stdlog.Println("rcsresult.go / insertRcsStatus / rcs_message_status / rcsStmt personal Exec ", err)
		}
	}

	_, err = rcsStmt.Exec()
	if err != nil {
		rcsStmt.Close()
		config.Stdlog.Println("rcsresult.go / insertRcsStatus / rcs_message_status / rcsStmt Exec ", err)
	}
	rcsStmt.Close()
	err = tx.Commit()
	if err != nil {
		config.Stdlog.Println("rcsresult.go / insertRcsStatus / rcs_message_status / rcsStmt commit ", err)
	}

}

func insertAmt(rcsAmtValue []common.AmtColumn, userid string) {

	tx, err := databasepool.DB.Begin()
	if err != nil {
		config.Stdlog.Println("rcsresult.go / insertAmt / cb_amt_"+userid+" / 트랜잭션 초기화 실패 ", err)
	}
	defer tx.Rollback()
	amtStmt, err := tx.Prepare(pq.CopyIn("cb_amt_"+userid, common.GetRcsColumnPq(common.AmtColumn{})...))
	if err != nil {
		config.Stdlog.Println("rcsresult.go / insertAmt / cb_amt_"+userid+" / amtStmt 초기화 실패 ", err)
		return
	}
	for _, data := range rcsAmtValue {
		_, err := amtStmt.Exec(data.Amt_datetime, data.Amt_kind, data.Amt_amount, data.Amt_memo, data.Amt_reason, data.Amt_payback, data.Amt_admin)
		if err != nil {
			config.Stdlog.Println("rcsresult.go / insertAmt / cb_amt_"+userid+" / amtStmt personal Exec ", err)
		}
	}

	_, err = amtStmt.Exec()
	if err != nil {
		amtStmt.Close()
		config.Stdlog.Println("rcsresult.go / insertAmt / cb_amt_"+userid+" / amtStmt Exec ", err)
	}
	amtStmt.Close()
	err = tx.Commit()
	if err != nil {
		config.Stdlog.Println("rcsresult.go / insertAmt / cb_amt_"+userid+" / amtStmt commit ", err)
	}

}

func insertOSms(oSmsValue []common.OSmsColumn) {

	tx, err := databasepool.DB.Begin()
	if err != nil {
		config.Stdlog.Println("rcsresult.go / insertOSms / OShotSMS / 트랜잭션 초기화 실패 ", err)
	}
	defer tx.Rollback()
	oSmsStmt, err := tx.Prepare(pq.CopyIn("OShotSMS", common.GetRcsColumnPq(common.OSmsColumn{})...))
	if err != nil {
		config.Stdlog.Println("rcsresult.go / insertOSms / OShotSMS / oSmsStmt 초기화 실패 ", err)
		return
	}
	for _, data := range oSmsValue {
		_, err := oSmsStmt.Exec(data.Sender, data.Receiver, data.Msg, data.URL, data.ReserveDT, data.TimeoutDT, data.SendResult, data.Mst_id, data.Cb_msg_id)
		if err != nil {
			config.Stdlog.Println("rcsresult.go / insertOSms / OShotSMS / oSmsStmt personal Exec ", err)
		}
	}

	_, err = oSmsStmt.Exec()
	if err != nil {
		oSmsStmt.Close()
		config.Stdlog.Println("rcsresult.go / insertOSms / OShotSMS / oSmsStmt Exec ", err)
	}
	oSmsStmt.Close()
	err = tx.Commit()
	if err != nil {
		config.Stdlog.Println("rcsresult.go / insertOSms / OShotSMS / oSmsStmt commit ", err)
	}

}

func insertOMms(oMmsValue []common.OMmsColumn) {

	tx, err := databasepool.DB.Begin()
	if err != nil {
		config.Stdlog.Println("rcsresult.go / insertOMms / OShotMMS / 트랜잭션 초기화 실패 ", err)
	}
	defer tx.Rollback()
	oMmsStmt, err := tx.Prepare(pq.CopyIn("OShotMMS", common.GetRcsColumnPq(common.OMmsColumn{})...))
	if err != nil {
		config.Stdlog.Println("rcsresult.go / insertOMms / OShotMMS / oMmsStmt 초기화 실패 ", err)
		return
	}
	for _, data := range oMmsValue {
		_, err := oMmsStmt.Exec(data.MsgGroupID, data.Sender, data.Receiver, data.Subject, data.Msg, data.ReserveDT, data.TimeoutDT, data.SendResult, data.File_Path1, data.File_Path2, data.File_Path3, data.Mst_id, data.Cb_msg_id)
		if err != nil {
			config.Stdlog.Println("rcsresult.go / insertOMms / OShotMMS / oMmsStmt personal Exec ", err)
		}
	}

	_, err = oMmsStmt.Exec()
	if err != nil {
		oMmsStmt.Close()
		config.Stdlog.Println("rcsresult.go / insertOMms / OShotMMS / oMmsStmt Exec ", err)
	}
	oMmsStmt.Close()
	err = tx.Commit()
	if err != nil {
		config.Stdlog.Println("rcsresult.go / insertOSms / OShotSMS / oMmsStmt commit ", err)
	}

}
