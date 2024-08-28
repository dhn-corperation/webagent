package rcs

import (
	//"bytes"
	"database/sql"
	"webagent/src/config"
	"webagent/src/databasepool"

	"encoding/json"
	"fmt"
	"webagent/src/baseprice"

	//"io/ioutil"

	//"net/http"
	s "strings"
	//c "strconv"

	"sync"
	"time"
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
	defer func() {
		if r := recover(); r != nil {
			for {
				err := databasepool.DB.Ping()
				if err == nil {
					break
				}
				time.Sleep(10 * time.Second)
			}
		}
	}()
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

	resAfter6 := `update RCS_MESSAGE_RESULT rmr 
set rmr.result_status = 'success' 
   ,rmr.status = 'success'
   ,rmr.proc = 'P'
 where date_add(STR_TO_DATE(rmr.sentTime, '%Y-%m-%d %H:%i:%s'), interval 6 HOUR) < NOW() AND rmr.proc = 'N'`

	databasepool.DB.Exec(resAfter6)

	RToken = getTokenInfo()

	resp, err := config.Client.R().
		SetHeaders(map[string]string{"Content-Type": "application/json", "Authorization": "Bearer " + RToken}).
		SetBody(resultReq).
		Post(config.Conf.RCSRESULTURL + "/corp/v1/querymsgstatus")

	//fmt.Println(resp)

	if err != nil {
		config.Stdlog.Println("RCS 메시지 결과 서버 호출 오류 : ", err)
		//	return nil
	} else {
		var resultInfo RcsResultInfo
		json.Unmarshal(resp.Body(), &resultInfo)

		//fmt.Println(len(resultInfo.StatusInfo), resultInfo.StatusInfo)
		//if len(resultInfo.StatusInfo) > 0 {
		resinsStrs := []string{}
		resinsValues := []interface{}{}
		resinsquery := `insert IGNORE into RCS_MESSAGE_STATUS(
rcs_id ,
msg_id ,
user_contact ,
status ,
service_type ,
mno_info ,
sent_time ,
error ,
timestamp ) values %s`

		for i := 0; i < len(resultInfo.StatusInfo); i++ {
			si := resultInfo.StatusInfo[i]

			resinsStrs = append(resinsStrs, "(?,?,?,?,?,?,?,?,?)")
			resinsValues = append(resinsValues, si.RcsId)
			resinsValues = append(resinsValues, si.MsgId)
			resinsValues = append(resinsValues, si.UserContact)
			resinsValues = append(resinsValues, si.Status)
			resinsValues = append(resinsValues, si.ServiceType)
			resinsValues = append(resinsValues, si.MnoInfo)
			resinsValues = append(resinsValues, si.SentTime)
			resinsValues = append(resinsValues, si.Error.Message)
			resinsValues = append(resinsValues, si.Timestamp)

			if len(resinsStrs) >= 500 {
				stmt := fmt.Sprintf(resinsquery, s.Join(resinsStrs, ","))

				_, err := databasepool.DB.Exec(stmt, resinsValues...)

				if err != nil {
					stdlog.Println("RCS_MESSAGE_STATUS Table Insert 처리 중 오류 발생 " + err.Error())
				}

				resinsStrs = nil
				resinsValues = nil
			}
			resUpdatestr := `update RCS_MESSAGE_RESULT rmr 
set rmr.result_status = '` + si.Status + `' 
   ,rmr.result_error = '` + si.Error.Message + `'
   ,rmr.proc = 'P'
 where rmr.proc = 'N' and rmr.msg_id='` + si.MsgId + `' and rmr.user_contact = '` + si.UserContact + `'`

			//stdlog.Println("RCS_MESSAGE_STATUS Table 수정 Query : " + resUpdatestr)
			databasepool.DB.Exec(resUpdatestr)
		}

		if len(resinsStrs) > 0 {
			stmt := fmt.Sprintf(resinsquery, s.Join(resinsStrs, ","))

			_, err := databasepool.DB.Exec(stmt, resinsValues...)

			if err != nil {
				stdlog.Println("RCS_MESSAGE_STATUS Table Insert 처리 중 오류 발생 " + err.Error())
			}

			resinsStrs = nil
			resinsValues = nil
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

		groupsql := "select distinct msg_group_id as mst_id, wms.send_mem_id as usermem_id from RCS_MESSAGE_RESULT rmr inner join cb_wt_msg_sent wms on rmr.msg_group_id = wms.mst_id where rmr.proc = 'P'"

		grows, err := db.Query(groupsql)
		if err != nil {
			stdlog.Println("RCS_MESSAGE_RESULT select 오류 ", err)
		} else {
			defer grows.Close()

			var mst_id, usermem_id sql.NullString

			ossmsStrs := []string{}
			ossmsValues := []interface{}{}

			osmmsStrs := []string{}
			osmmsValues := []interface{}{}

			for grows.Next() {

				ossmsStrs = nil //스마트미 SMS Table Insert 용
				ossmsValues = nil

				osmmsStrs = nil //스마트미 LMS/MMS Table Insert 용
				osmmsValues = nil

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
,ifnull(rmr.result_status, rmr.status)  as res_status
,ifnull(rmr.result_error,rmr.error) as res_error
,rmr.service_type
,(SELECT mi.origin1_path FROM cb_mms_images mi where wms.mst_mms_content = mi.mms_id and length(mst_mms_content ) > 5 ) as mms_file1
,(SELECT mi.origin2_path FROM cb_mms_images mi where wms.mst_mms_content = mi.mms_id and length(mst_mms_content ) > 5 ) as mms_file2
,(SELECT mi.origin3_path FROM cb_mms_images mi where wms.mst_mms_content = mi.mms_id and length(mst_mms_content ) > 5 ) as mms_file3
,wms.mst_sent_voucher  
,wms.mst_mem_id as send_mem_id
,wms.mst_type2
,wms.mst_type3
,(select ( case when rcs.msr_exptime < NOW() then 'N' WHEn rcs.msr_exptime IS NULL then 'N' else 'Y' END) from cb_wt_msg_rcs rcs where rcs.msr_mst_id = rmr.msg_group_id) as msr_exptime
,wms.mst_lms_content 
FROM RCS_MESSAGE_RESULT rmr 
     inner join cb_wt_msg_sent wms on rmr.msg_group_id = wms.mst_id 
where rmr.proc = 'P'
and rmr.msg_group_id = ?
`

				resrows, err := db.Query(ressql, mst_id.String)
				if err != nil {
					stdlog.Println("결과 처리 Select 오류 ", err)
				} else {
					defer resrows.Close()

					var msgid, phnstr, sms_sender, body, remark4, reserve_dt, userid, res_status, res_error, service_type, mms_file1, mms_file2, mms_file3, mst_sent_voucher, send_mem_id, mst_type2, mst_type3, msr_exptime, mst_lms_content sql.NullString

					amtsStrs := []string{}
					amtsValues := []interface{}{}

					var amtinsstr = ""

					amtsStrs = nil
					amtsValues = nil

					for resrows.Next() {

						var amount float64
						var memo string
						var payback float64
						var admin_amt float64

						resrows.Scan(&msgid, &phnstr, &sms_sender, &body, &remark4, &reserve_dt, &userid, &res_status, &res_error, &service_type, &mms_file1, &mms_file2, &mms_file3, &mst_sent_voucher, &send_mem_id, &mst_type2, &mst_type3, &msr_exptime, &mst_lms_content)

						cprice := baseprice.GetPrice(db, send_mem_id.String, stdlog)

						amtinsstr = "insert into cb_amt_" + userid.String + "(amt_datetime," +
							"amt_kind," +
							"amt_amount," +
							"amt_memo," +
							"amt_reason," +
							"amt_payback," +
							"amt_admin)" +
							" values %s"

						switch s.ToLower(res_status.String) {
						case "success":
							db.Exec("update cb_msg_"+userid.String+" set CODE = 'RCS', MESSAGE_TYPE='rc', MESSAGE = ?, RESULT = ? where remark4=? and msgid = ?", "RCS 성공", "Y", remark4.String, msgid.String)
							scnt++
						case "fail":
							amtsStrs = append(amtsStrs, "(now(),?,?,?,?,?,?)")
							amtsValues = append(amtsValues, "3")
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

							amtsValues = append(amtsValues, amount)

							amtsValues = append(amtsValues, memo)
							amtsValues = append(amtsValues, msgid.String+"/"+phnstr.String)
							amtsValues = append(amtsValues, payback*-1)
							amtsValues = append(amtsValues, admin_amt*-1)

							var rcsBody RcsBody
							json.Unmarshal([]byte(body.String), &rcsBody)
							if s.Contains(mst_type3.String, "wc") && s.EqualFold(msr_exptime.String, "Y") {

								stdlog.Println("RCS 실패 -> WEB(C) 발송 처리 ", mst_type3.String, msr_exptime.String, msgid.String)

								db.Exec("update cb_msg_"+userid.String+" set CODE = 'SMT', MESSAGE_TYPE='sm' where remark4=? and msgid = ?", remark4.String, msgid.String)

								if s.EqualFold(mst_type3.String, "wcs") {
									ossmsStrs = append(ossmsStrs, "(?,?,?,?,?,null,?,?,?)")
									ossmsValues = append(ossmsValues, sms_sender.String)
									ossmsValues = append(ossmsValues, phnstr.String)
									ossmsValues = append(ossmsValues, mst_lms_content.String)
									ossmsValues = append(ossmsValues, "")
									if s.EqualFold(reserve_dt.String, "00000000000000") {
										ossmsValues = append(ossmsValues, sql.NullString{})
									} else {
										ossmsValues = append(ossmsValues, reserve_dt.String)
									}
									ossmsValues = append(ossmsValues, "0")
									ossmsValues = append(ossmsValues, remark4.String)
									ossmsValues = append(ossmsValues, msgid.String)

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

									amtsStrs = append(amtsStrs, "(now(),?,?,?,?,?,?)")
									amtsValues = append(amtsValues, "P")
									amtsValues = append(amtsValues, amount)
									amtsValues = append(amtsValues, memo)
									amtsValues = append(amtsValues, msgid.String+"/"+phnstr.String)
									amtsValues = append(amtsValues, payback)
									amtsValues = append(amtsValues, admin_amt)

								} else if s.EqualFold(mst_type3.String, "wc") {
									osmmsStrs = append(osmmsStrs, "(?,?,?,?,?,?,null,?,?,?,?,?,?)")
									osmmsValues = append(osmmsValues, remark4.String)
									osmmsValues = append(osmmsValues, sms_sender.String)
									osmmsValues = append(osmmsValues, phnstr.String)
									osmmsValues = append(osmmsValues, rcsBody.Title)
									osmmsValues = append(osmmsValues, mst_lms_content.String)
									if s.EqualFold(reserve_dt.String, "00000000000000") {
										osmmsValues = append(osmmsValues, sql.NullString{})
									} else {
										osmmsValues = append(osmmsValues, reserve_dt.String)
									}
									osmmsValues = append(osmmsValues, "0")

									osmmsValues = append(osmmsValues, sql.NullString{})
									osmmsValues = append(osmmsValues, sql.NullString{})
									osmmsValues = append(osmmsValues, sql.NullString{})
									osmmsValues = append(osmmsValues, remark4.String)
									osmmsValues = append(osmmsValues, msgid.String)

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

									amtsStrs = append(amtsStrs, "(now(),?,?,?,?,?,?)")
									amtsValues = append(amtsValues, "P")
									amtsValues = append(amtsValues, amount)
									amtsValues = append(amtsValues, memo)
									amtsValues = append(amtsValues, msgid.String+"/"+phnstr.String)
									amtsValues = append(amtsValues, payback)
									amtsValues = append(amtsValues, admin_amt)
								}
							} else {
								ecnt++
								db.Exec("update cb_msg_"+userid.String+" set CODE = 'RCS', MESSAGE_TYPE='rc', MESSAGE = ?, RESULT = ? where remark4=? and msgid = ?", res_error.String, "N", remark4.String, msgid.String)
							}
						}

						db.Exec("update RCS_MESSAGE_RESULT set proc='Y' where msg_group_id = ? and proc='P' and msg_id = ?", mst_id.String, msgid.String)

					}

					if len(ossmsStrs) > 0 {
						stmt := fmt.Sprintf("insert into OShotSMS(Sender,Receiver,Msg,URL,ReserveDT,TimeoutDT,SendResult,mst_id,cb_msg_id ) values %s", s.Join(ossmsStrs, ","))
						_, err := db.Exec(stmt, ossmsValues...)

						if err != nil {
							stdlog.Println("스마트미 SMS Table Insert 처리 중 오류 발생 " + err.Error())
						}

						ossmsStrs = nil
						ossmsValues = nil
					}

					if len(osmmsStrs) > 0 {
						stmt := fmt.Sprintf("insert into OShotMMS(MsgGroupID,Sender,Receiver,Subject,Msg,ReserveDT,TimeoutDT,SendResult,File_Path1,File_Path2,File_Path3,mst_id,cb_msg_id ) values %s", s.Join(osmmsStrs, ","))
						_, err := db.Exec(stmt, osmmsValues...)

						if err != nil {
							stdlog.Println("스마트미 LMS Table Insert 처리 중 오류 발생 " + err.Error())
						}

						osmmsStrs = nil
						osmmsValues = nil
					}

					if len(amtsStrs) > 0 {
						stmt := fmt.Sprintf(amtinsstr, s.Join(amtsStrs, ","))
						_, err := db.Exec(stmt, amtsValues...)

						if err != nil {
							stdlog.Println("AMT Table Insert 처리 중 오류 발생 " + err.Error())
						}

						amtsStrs = nil
						amtsValues = nil
					}

					db.Exec("update cb_wt_msg_sent set mst_rcs = ifnull(mst_rcs,0) + ?,mst_err_rcs = ifnull(mst_err_rcs,0) + ?, mst_wait = mst_wait - ?  where mst_id=?", scnt, ecnt, (ecnt + scnt), mst_id.String)

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

		resp, err := config.Client.R().
			SetHeaders(map[string]string{"Content-Type": "application/json", "Authorization": "Bearer " + RToken2}).
			SetBody(resultReq).
			Post(config.Conf.RCSRESULTURL + "/corp/v1/querymsgstatus")

		//fmt.Println(resp, resultReq)

		if err != nil {
			config.Stdlog.Println("RCS 메시지 결과 서버 호출 오류 : ", err)
			//	return nil
		} else {
			var resultInfo RcsResultInfo
			json.Unmarshal(resp.Body(), &resultInfo)

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
