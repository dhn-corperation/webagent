package webrcs

import (
	"fmt"
	"sync"
	"time"
	"context"
	s "strings"
	"database/sql"
	"encoding/json"

	"webagent/src/config"
	"webagent/src/databasepool"
	"webagent/src/baseprice"
)

var rToken string

var dhnTickSql string
var dhnSelMstIdSql string
var dhnSelSql string
var dhnResultSql string

var rrId, msgId, userContact, chatbotId, secondSendTitle, secondSendBody, status, resultStatus, resultCode sql.NullString
var resultMessage, msgServiceType, memId, memUserId, mstId, mstSendVoucher, mstType3, mmsFile1, mmsFile2, mmsFile3 sql.NullString

func ResultProcess(ctx context.Context) {
	config.Stdlog.Println("(신) Rcs - 결과 처리 프로세스 시작")
	rcsResultInit()
	var wg sync.WaitGroup
	
	for {
		select {
			case <- ctx.Done():
				config.Stdlog.Println("(신) Rcs - Result Process가 15초 후에 종료")
				time.Sleep(15 * time.Second)
				config.Stdlog.Println("(신) Rcs - Result Process 종료 완료")
				return
			default:
				wg.Add(1)
				go resultProcess(&wg)
				wg.Wait()
		}
	}
}

func rcsResultInit() {
	dhnTickSql = `
		select
			count(1) as cnt
		from
			` + config.Conf.RCSTABLE + ` a
		left join
			cb_wt_msg_sent_rcs b on a.mstr_id = b.mstr_id
		left join
			cb_wt_msg_sent c on b.mst_id = c.mst_id
		where
			(a.status = 2 and b.second_send_dt < NOW() and c.mst_type3 is not null)
			or a.status = 3
		limit 1`

	dhnSelMstIdSql = `
		select distinct
			b.mst_id
		from
			` + config.Conf.RCSTABLE + ` a
		left join
			cb_wt_msg_sent_rcs b on a.mstr_id = b.mstr_id
		left join
			cb_wt_msg_sent c on b.mst_id = c.mst_id
		where
			(a.status = 2 and b.second_send_dt < NOW() and c.mst_type3 is not null)
			or a.status = 3`

	dhnSelSql = `
		select
			a.rr_id,
			a.msg_id,
			a.user_contract,
			b.chatbot_id,
			b.second_send_title,
			a.second_send_body,
			a.status,
			a.result_status,
			a.result_code,
			a.result_message,
			b.msg_service_type,
			b.mem_id,
			b.mem_userid,
			c.mst_sent_voucher,
			c.mst_type3,
			(SELECT mi.origin1_path FROM cb_mms_images mi where c.mst_mms_content = mi.mms_id and length(mst_mms_content ) > 5 ) as mms_file1,
			(SELECT mi.origin2_path FROM cb_mms_images mi where c.mst_mms_content = mi.mms_id and length(mst_mms_content ) > 5 ) as mms_file2,
			(SELECT mi.origin3_path FROM cb_mms_images mi where c.mst_mms_content = mi.mms_id and length(mst_mms_content ) > 5 ) as mms_file3
		from
			` + config.Conf.RCSTABLE + ` a
		left join
			cb_wt_msg_sent_rcs b on a.mstr_id = b.mstr_id
		left join
			cb_wt_msg_sent c on b.mst_id = c.mst_id
		where
			((a.status = 2 and b.second_send_dt < NOW() and c.mst_type3 is not null)
			or a.status = 3)
			and b.mst_id = ?`

	dhnResultSql = `
		update
			` + config.Conf.RCSTABLE + `
		set
			status = 4
		where
			rr_id in (%s)`
}

func resultProcess(wg *sync.WaitGroup) {
	defer wg.Done()
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

	var rcsResultRequest RcsResultRequest

	var startNow = time.Now()
	var startTime = fmt.Sprintf("%02d:%02d:%02d.%09d", startNow.Hour(), startNow.Minute(), startNow.Second(), startNow.Nanosecond())

	rcsResultRequest.QueryType = "rcsId"
	rcsResultRequest.RcsId = config.RCSID
	rcsResultRequest.QueryId = "DHN_" + startTime

 	resAfter6 := `
 		update
 			` + config.Conf.RCSTABLE + `
 		set
 			status = 3,
 			result_status = 'success',
 			result_code = '00001',
 			result_message = 'RCS 성공',
 			manual_flag = 1,
 			result_dt = now()
		where
			date_add(send_dt, interval 6 HOUR) < NOW() and status = 2`

	_, execErr := databasepool.DB.Exec(resAfter6)
	if execErr != nil {
		config.Stdlog.Println("(신) Rcs - 메시지 결과 6시간 성공 처리 에러 : ", execErr)
		panic(execErr)
	}

// 토큰 발급 시작
	var rcsAuthResponse RcsAuthResponse
	rcsAuthResponse, err := rcsAuthRequest.getTokenInfo()

	if err != nil {
		config.Stdlog.Println("(신) Rcs - 토큰 발급 통신 실패 err : ", err)
		time.Sleep(1 * time.Second)
		return
	} else {
		if rcsAuthResponse.Status == "200" {
			rToken = rcsAuthResponse.Data.TokenInfo.AccessToken
		} else {
			config.Stdlog.Println("(신) Rcs - 토큰 발급 실패 err : ", rcsAuthResponse.Error.Message)
			time.Sleep(1 * time.Second)
			return
		}
	}
// 토큰 발급 끝

// RCS 결과 처리 시작
	resp, err := config.Client.R().
		SetHeaders(map[string]string{"Content-Type": "application/json", "Authorization": "Bearer " + rToken}).
		SetBody(rcsResultRequest).
		Post(config.Conf.RCSRESULTURL + "/corp/v1/querymsgstatus")


	if err != nil {
		config.Stdlog.Println("(신) Rcs - 메시지 결과 서버 호출 오류 : ", err)
		time.Sleep(10 * time.Second)
		return
	} else {
		var rcsResultResponse RcsResultResponse
		var rcsResponseErrorInfo RcsResponseErrorInfo

		if resp.StatusCode() != 200 {
			json.Unmarshal(resp.Body(), &rcsResponseErrorInfo)
			config.Stdlog.Println("(신) Rcs - 메시지 결과 서버 호출 결과 오류 : ", rcsResponseErrorInfo.Error.Code, " / ", rcsResponseErrorInfo.Error.Message)
		} else {
			json.Unmarshal(resp.Body(), &rcsResultResponse)

			rcsResUpdSql := `
				update
					` + config.Conf.RCSTABLE + `
				set
					status = 3,
					result_status = ?,
					result_code = ?,
					result_message = ?,
					result_mno_info = ?,
					manual_flag = 0,
					result_dt = ?
				where
					msg_id = ?`

			for _, statusInfo := range rcsResultResponse.StatusInfo {
				resultDt := statusInfo.Timestamp
				parseResultDt, err := time.Parse(time.RFC3339, resultDt)
				if err != nil {
					parseResultDt = time.Now()
				}
				convResultDt := parseResultDt.Format("2006-01-02 15:04:05")

				rCode := ""
				rMsg := ""

				if statusInfo.Status == "success" {
					rCode = "00000"
					rMsg = "RCS 성공"
				} else if statusInfo.Status == "fail" {
					rCode = statusInfo.Error.Code
					rMsg = statusInfo.Error.Message
				} else {
					rCode = "00002"
					rMsg = "RCS 성공"
				}

				_, err = db.Exec(rcsResUpdSql, statusInfo.Status, rCode, rMsg, statusInfo.MnoInfo, convResultDt, statusInfo.MsgId)
				if err != nil {
					config.Stdlog.Println("(신) Rcs - 메시지 결과 처리 중 업데이트 오류 msg_id : ", statusInfo.MsgId)
				}
			}
		}
	}
// RCS 결과 처리 끝
	
// DHN 결과 처리 시작
	var tickCnt sql.NullInt16

	db.QueryRow(dhnTickSql).Scan(&tickCnt)

	if tickCnt.Int16 > 0 {
		selUserIdRows, err := db.Query(dhnSelMstIdSql)
		if  err != nil {
			config.Stdlog.Println("(신) Rcs - DHN userid 데이터 조회 중 오류")
			time.Sleep(1 * time.Second)
			return
		} else {
			for selUserIdRows.Next() {

				selUserIdRows.Scan(&mstId)

				selRows, err := db.Query(dhnSelSql, mstId.String)

				if  err != nil {
					config.Stdlog.Println("(신) Rcs - DHN 데이터 조회 중 오류")
					time.Sleep(1 * time.Second)
					return
				} else {

					scnt := 0
					ecnt := 0

					dhnResultStr := []string{}
					dhnResultValues := []interface{}{}

					ossmsStrs := []string{}
					ossmsValues := []interface{}{}

					osmmsStrs := []string{}
					osmmsValues := []interface{}{}

					lgusmsStrs := []string{}
					lgusmsValues := []interface{}{}

					lgummsStrs := []string{}
					lgummsValues := []interface{}{}

					nnsmsStrs := []string{}
					nnsmsValues := []interface{}{}

					nnmmsStrs := []string{}
					nnmmsValues := []interface{}{}

					amtsStrs := []string{}
					amtsValues := []interface{}{}
					amtinsstr := ""

					var amount float64
					var payback float64
					var admin_amt float64
					var memo string
					var firstFlag = true

					stdlog.Println("(신) Rcs - 처리 시작 : (", mstId.String, " )")

					for selRows.Next() {

						selRows.Scan(&rrId, &msgId, &userContact, &chatbotId, &secondSendTitle, &secondSendBody, &status, &resultStatus,
							&resultCode, &resultMessage, &msgServiceType, &memId, &memUserId, &mstSendVoucher, &mstType3, &mmsFile1, &mmsFile2, &mmsFile3)


						if firstFlag {
							amtinsstr = `
								insert into cb_amt_` + memUserId.String + `(
									amt_datetime,
									amt_kind,
									amt_amount,
									amt_memo,
									amt_reason
									amt_payback,
									amt_admin)
								values %s`
							firstFlag = false
						}

						switch s.ToLower(status.String) {
							case "success":
								db.Exec("update cb_msg_" + memUserId.String + " set CODE = 'RCS', MESSAGE_TYPE='nr', MESSAGE = 'RCS " + msgServiceType.String + " 성공', RESULT = 'Y' where remark4 = ? and msgid = ?", mstId.String, msgId.String)
								scnt++
							case "fail":
								cprice := baseprice.GetPrice(db, msgId.String, stdlog)

								if s.EqualFold(mstSendVoucher.String, "V") {
									switch msgServiceType.String {
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
									switch msgServiceType.String {
										case "RCSSMS":
											amount = cprice.C_price_rcs_sms.Float64
											payback = cprice.C_price_rcs_sms.Float64 - cprice.P_price_rcs_sms.Float64
											admin_amt = cprice.B_price_rcs_sms.Float64
											if s.EqualFold(mstSendVoucher.String, "B") {
												memo = "RCS SMS 발송실패 환불,보너스"
											} else {
												memo = "RCS SMS 발송실패 환불"
											}
										case "RCSLMS":
											amount = cprice.C_price_rcs.Float64
											payback = cprice.C_price_rcs.Float64 - cprice.P_price_rcs.Float64
											admin_amt = cprice.B_price_rcs.Float64
											if s.EqualFold(mstSendVoucher.String, "B") {
												memo = "RCS LMS 발송실패 환불,보너스"
											} else {
												memo = "RCS LMS 발송실패 환불"
											}
										case "RCSMMS":
											amount = cprice.C_price_rcs_mms.Float64
											payback = cprice.C_price_rcs_mms.Float64 - cprice.P_price_rcs_mms.Float64
											admin_amt = cprice.B_price_rcs_mms.Float64
											if s.EqualFold(mstSendVoucher.String, "B") {
												memo = "RCS MMS 발송실패 환불,보너스"
											} else {
												memo = "RCS MMS 발송실패 환불"
											}
										case "RCSTMPL":
											amount = cprice.C_price_rcs_tem.Float64
											payback = cprice.C_price_rcs_tem.Float64 - cprice.P_price_rcs_tem.Float64
											admin_amt = cprice.B_price_rcs_tem.Float64
											if s.EqualFold(mstSendVoucher.String, "B") {
												memo = "RCS TMPL 발송실패 환불,보너스"
											} else {
												memo = "RCS TMPL 발송실패 환불"
											}
									}
								}

								amtsStrs = append(amtsStrs, "(now(),?,?,?,?,?,?)")
								amtsValues = append(amtsValues, "3")
								amtsValues = append(amtsValues, amount)
								amtsValues = append(amtsValues, memo)
								amtsValues = append(amtsValues, msgId.String+"/"+userContact.String)
								amtsValues = append(amtsValues, payback * -1)
								amtsValues = append(amtsValues, admin_amt * -1)

								if s.Contains(mstType3.String, "wc") {

									stdlog.Println("(신) Rcs - 발송 실패 -> WEB(C) 발송 처리 ", mstType3.String, " / ", msgId.String)

									db.Exec("update cb_msg_"+memUserId.String+" set CODE = 'SMT', MESSAGE_TYPE='sm' where remark4=? and msgid = ?", mstId.String, msgId.String)

									if s.EqualFold(mstType3.String, "wcs") {
										ossmsStrs = append(ossmsStrs, "(?,?,?,?,?,null,?,?,?)")
										ossmsValues = append(ossmsValues, chatbotId.String)
										ossmsValues = append(ossmsValues, userContact.String)
										ossmsValues = append(ossmsValues, secondSendBody.String)
										ossmsValues = append(ossmsValues, "")
										ossmsValues = append(ossmsValues, sql.NullString{})
										ossmsValues = append(ossmsValues, "0")
										ossmsValues = append(ossmsValues, mstId.String)
										ossmsValues = append(ossmsValues, msgId.String)

										if s.EqualFold(mstSendVoucher.String, "V") {
											amount = cprice.V_price_smt_sms.Float64
											payback = cprice.V_price_smt_sms.Float64 - cprice.P_price_smt_sms.Float64
											admin_amt = cprice.B_price_smt_sms.Float64
											memo = "웹(C) SMS,바우처"
										} else {
											amount = cprice.C_price_smt_sms.Float64
											payback = cprice.C_price_smt_sms.Float64 - cprice.P_price_smt_sms.Float64
											admin_amt = cprice.B_price_smt_sms.Float64
											if s.EqualFold(mstSendVoucher.String, "B") {
												memo = "웹(C) SMS,보너스"
											} else {
												memo = "웹(C) SMS"
											}
										}

										amtsStrs = append(amtsStrs, "(now(),?,?,?,?,?,?)")
										amtsValues = append(amtsValues, "P")
										amtsValues = append(amtsValues, amount)
										amtsValues = append(amtsValues, memo)
										amtsValues = append(amtsValues, msgId.String+"/"+userContact.String)
										amtsValues = append(amtsValues, payback)
										amtsValues = append(amtsValues, admin_amt)

									} else if s.EqualFold(mstType3.String, "wc") {
										osmmsStrs = append(osmmsStrs, "(?,?,?,?,?,?,null,?,?,?,?,?,?)")
										osmmsValues = append(osmmsValues, mstId.String)
										osmmsValues = append(osmmsValues, chatbotId.String)
										osmmsValues = append(osmmsValues, userContact.String)
										osmmsValues = append(osmmsValues, secondSendTitle.String)
										osmmsValues = append(osmmsValues, secondSendBody.String)
										osmmsValues = append(osmmsValues, sql.NullString{})
										osmmsValues = append(osmmsValues, "0")
										osmmsValues = append(osmmsValues, sql.NullString{})
										osmmsValues = append(osmmsValues, sql.NullString{})
										osmmsValues = append(osmmsValues, sql.NullString{})
										osmmsValues = append(osmmsValues, mstId.String)
										osmmsValues = append(osmmsValues, msgId.String)

										if len(mmsFile1.String) <= 0 {
											if s.EqualFold(mstSendVoucher.String, "V") {
												amount = cprice.V_price_smt.Float64
												payback = cprice.V_price_smt.Float64 - cprice.P_price_smt.Float64
												admin_amt = cprice.B_price_smt.Float64
												memo = "웹(C) LMS,바우처"
											} else {
												amount = cprice.C_price_smt.Float64
												payback = cprice.C_price_smt.Float64 - cprice.P_price_smt.Float64
												admin_amt = cprice.B_price_smt.Float64
												if s.EqualFold(mstSendVoucher.String, "B") {
													memo = "웹(C) LMS,보너스"
												} else {
													memo = "웹(C) LMS"
												}
											}
										} else {
											if s.EqualFold(mstSendVoucher.String, "V") {
												amount = cprice.V_price_smt_mms.Float64
												payback = cprice.V_price_smt_mms.Float64 - cprice.P_price_smt_mms.Float64
												admin_amt = cprice.B_price_smt_mms.Float64
												memo = "웹(C) MMS,바우처"
											} else {
												amount = cprice.C_price_smt_mms.Float64
												payback = cprice.C_price_smt_mms.Float64 - cprice.P_price_smt_mms.Float64
												admin_amt = cprice.B_price_smt_mms.Float64
												if s.EqualFold(mstSendVoucher.String, "B") {
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
										amtsValues = append(amtsValues, msgId.String+"/"+userContact.String)
										amtsValues = append(amtsValues, payback)
										amtsValues = append(amtsValues, admin_amt)
									}
								} else if s.Contains(mstType3.String, "wa") {

									stdlog.Println("(신) Rcs - 발송 실패 -> WEB(A) 발송 처리 ", mstType3.String, " / ", msgId.String)

									db.Exec("update cb_msg_"+memUserId.String+" set CODE = 'GRS', MESSAGE_TYPE='gr' where remark4=? and msgid = ?", mstId.String, msgId.String)

									if s.EqualFold(mstType3.String, "was") {

										nnsmsStrs = append(nnsmsStrs, "(?,?,?,?,?,?,?,?,?,'Y')")
										nnsmsValues = append(nnsmsValues, chatbotId.String)
										nnsmsValues = append(nnsmsValues, userContact.String)
										nnsmsValues = append(nnsmsValues, secondSendBody.String)
										nnsmsValues = append(nnsmsValues, time.Now().Format("2006-01-02 15:04:05"))
										nnsmsValues = append(nnsmsValues, "0")
										nnsmsValues = append(nnsmsValues, "0")
										nnsmsValues = append(nnsmsValues, msgId.String)
										nnsmsValues = append(nnsmsValues, mstId.String)
										nnsmsValues = append(nnsmsValues, config.Conf.KISACODE)

										if s.EqualFold(mstSendVoucher.String, "V") {
											amount = cprice.V_price_smt_sms.Float64
											payback = cprice.V_price_smt_sms.Float64 - cprice.P_price_smt_sms.Float64
											admin_amt = cprice.B_price_smt_sms.Float64
											memo = "웹(A) SMS,바우처"
										} else {
											amount = cprice.C_price_smt_sms.Float64
											payback = cprice.C_price_smt_sms.Float64 - cprice.P_price_smt_sms.Float64
											admin_amt = cprice.B_price_smt_sms.Float64
											if s.EqualFold(mstSendVoucher.String, "B") {
												memo = "웹(A) SMS,보너스"
											} else {
												memo = "웹(A) SMS"
											}
										}

										amtsStrs = append(amtsStrs, "(now(),?,?,?,?,?,?)")
										amtsValues = append(amtsValues, "P")
										amtsValues = append(amtsValues, amount)
										amtsValues = append(amtsValues, memo)
										amtsValues = append(amtsValues, msgId.String+"/"+userContact.String)
										amtsValues = append(amtsValues, payback)
										amtsValues = append(amtsValues, admin_amt)

									} else if s.EqualFold(mstType3.String, "wa") {

										filecnt := 0

										if len(mmsFile1.String) > 0 {
											filecnt = filecnt + 1
										}

										if len(mmsFile2.String) > 0 {
											filecnt = filecnt + 1
										}

										if len(mmsFile3.String) > 0 {
											filecnt = filecnt + 1
										}

										nnmmsStrs = append(nnmmsStrs, "( ?,?,?,?,?,?,?,?,?,?,?,?,?,'Y')")
										nnmmsValues = append(nnmmsValues, chatbotId.String)
										nnmmsValues = append(nnmmsValues, userContact.String)
										nnmmsValues = append(nnmmsValues, secondSendTitle.String)
										nnmmsValues = append(nnmmsValues, secondSendBody.String)
										nnmmsValues = append(nnmmsValues, time.Now().Format("2006-01-02 15:04:05"))
										nnmmsValues = append(nnmmsValues, "0")
										nnmmsValues = append(nnmmsValues, filecnt)
										nnmmsValues = append(nnmmsValues, mmsFile1)
										nnmmsValues = append(nnmmsValues, mmsFile2)
										nnmmsValues = append(nnmmsValues, mmsFile3)
										nnmmsValues = append(nnmmsValues, msgId.String)
										nnmmsValues = append(nnmmsValues, mstId.String)
										nnmmsValues = append(nnmmsValues, config.Conf.KISACODE)

										if len(mmsFile1.String) <= 0 {
											if s.EqualFold(mstSendVoucher.String, "V") {
												amount = cprice.V_price_smt.Float64
												payback = cprice.V_price_smt.Float64 - cprice.P_price_smt.Float64
												admin_amt = cprice.B_price_smt.Float64
												memo = "웹(A) LMS,바우처"
											} else {
												amount = cprice.C_price_smt.Float64
												payback = cprice.C_price_smt.Float64 - cprice.P_price_smt.Float64
												admin_amt = cprice.B_price_smt.Float64
												if s.EqualFold(mstSendVoucher.String, "B") {
													memo = "웹(A) LMS,보너스"
												} else {
													memo = "웹(A) LMS"
												}
											}
										} else {
											if s.EqualFold(mstSendVoucher.String, "V") {
												amount = cprice.V_price_smt_mms.Float64
												payback = cprice.V_price_smt_mms.Float64 - cprice.P_price_smt_mms.Float64
												admin_amt = cprice.B_price_smt_mms.Float64
												memo = "웹(A) MMS,바우처"
											} else {
												amount = cprice.C_price_smt_mms.Float64
												payback = cprice.C_price_smt_mms.Float64 - cprice.P_price_smt_mms.Float64
												admin_amt = cprice.B_price_smt_mms.Float64
												if s.EqualFold(mstSendVoucher.String, "B") {
													memo = "웹(A) MMS,보너스"
												} else {
													memo = "웹(A) MMS"
												}
											}

										}

										amtsStrs = append(amtsStrs, "(now(),?,?,?,?,?,?)")
										amtsValues = append(amtsValues, "P")
										amtsValues = append(amtsValues, amount)
										amtsValues = append(amtsValues, memo)
										amtsValues = append(amtsValues, msgId.String+"/"+userContact.String)
										amtsValues = append(amtsValues, payback)
										amtsValues = append(amtsValues, admin_amt)
									}
								} else if s.Contains(mstType3.String, "wb") {

									stdlog.Println("(신) Rcs - 발송 실패 -> WEB(B) 발송 처리 ", mstType3.String, " / ", msgId.String)

									db.Exec("update cb_msg_"+memUserId.String+" set CODE = 'LGU', MESSAGE_TYPE='lg' where remark4=? and msgid = ?", mstId.String, msgId.String)

									if s.EqualFold(mstType3.String, "wbs") {
										lgusmsStrs = append(lgusmsStrs, "(?,?,?,?,?,?,?,?)")
										lgusmsValues = append(lgusmsValues, time.Now().Format("2006-01-02 15:04:05"))
										lgusmsValues = append(lgusmsValues, userContact.String)
										lgusmsValues = append(lgusmsValues, chatbotId.String)
										lgusmsValues = append(lgusmsValues, secondSendBody.String)
										lgusmsValues = append(lgusmsValues, msgId.String)
										lgusmsValues = append(lgusmsValues, memUserId.String)
										lgusmsValues = append(lgusmsValues, mstId.String)
										lgusmsValues = append(lgusmsValues, config.Conf.KISACODE)

										admin_amt = cprice.B_price_nas_sms.Float64
										if s.EqualFold(mstSendVoucher.String, "V") {
											amount = cprice.V_price_nas_sms.Float64
											payback = cprice.V_price_nas_sms.Float64 - cprice.P_price_nas_sms.Float64
											memo = "웹(B) SMS,바우처"
										} else {
											amount = cprice.C_price_nas_sms.Float64
											payback = cprice.C_price_nas_sms.Float64 - cprice.P_price_nas_sms.Float64
											if s.EqualFold(mstSendVoucher.String, "B") {
												memo = "웹(B) SMS,보너스"
											} else {
												memo = "웹(B) SMS"
											}
										}

										amtsStrs = append(amtsStrs, "(now(),?,?,?,?,?,?)")
										amtsValues = append(amtsValues, "P")
										amtsValues = append(amtsValues, amount)
										amtsValues = append(amtsValues, memo)
										amtsValues = append(amtsValues, msgId.String+"/"+userContact.String)
										amtsValues = append(amtsValues, payback)
										amtsValues = append(amtsValues, admin_amt)

									} else if s.EqualFold(mstType3.String, "wb") {
										file_cnt  := 0
										if mmsFile1.String != "" {
											file_cnt++
										}
										if mmsFile2.String != "" {
											file_cnt++
										}
										if mmsFile3.String != "" {
											file_cnt++
										}
										lgummsStrs = append(lgummsStrs, "(?,?,?,?,?,?,?,?,?,?,?,?,?)")
										lgummsValues = append(lgummsValues, secondSendTitle.String)
										lgummsValues = append(lgummsValues, userContact.String)
										lgummsValues = append(lgummsValues, chatbotId.String)
										lgummsValues = append(lgummsValues, time.Now().Format("2006-01-02 15:04:05"))
										lgummsValues = append(lgummsValues, secondSendBody.String)
										lgummsValues = append(lgummsValues, file_cnt)
										lgummsValues = append(lgummsValues, mmsFile1.String)
										lgummsValues = append(lgummsValues, mmsFile2.String)
										lgummsValues = append(lgummsValues, mmsFile3.String)
										lgummsValues = append(lgummsValues, msgId.String)
										lgummsValues = append(lgummsValues, memUserId.String)
										lgummsValues = append(lgummsValues, mstId.String)
										lgummsValues = append(lgummsValues, config.Conf.KISACODE)

										if len(mmsFile1.String) <= 0 {

											admin_amt = cprice.B_price_nas.Float64
											if s.EqualFold(mstSendVoucher.String, "V") {
												amount = cprice.V_price_nas.Float64
												payback = cprice.V_price_nas.Float64 - cprice.P_price_nas.Float64
												memo = "웹(B) LMS,바우처"
											} else {
												amount = cprice.C_price_nas.Float64
												payback = cprice.C_price_nas.Float64 - cprice.P_price_nas.Float64
												if s.EqualFold(mstSendVoucher.String, "B") {
													memo = "웹(B) LMS,보너스"
												} else {
													memo = "웹(B) LMS"
												}
											}
										} else {

											admin_amt = cprice.B_price_nas_mms.Float64
											if s.EqualFold(mstSendVoucher.String, "V") {
												amount = cprice.V_price_nas_mms.Float64
												payback = cprice.V_price_nas_mms.Float64 - cprice.P_price_nas_mms.Float64
												memo = "웹(B) MMS,바우처"
											} else {
												amount = cprice.C_price_nas_mms.Float64
												payback = cprice.C_price_nas_mms.Float64 - cprice.P_price_nas_mms.Float64
												if s.EqualFold(mstSendVoucher.String, "B") {
													memo = "웹(B) MMS,보너스"
												} else {
													memo = "웹(B) MMS"
												}
											}
										}

										amtsStrs = append(amtsStrs, "(now(),?,?,?,?,?,?)")
										amtsValues = append(amtsValues, "P")
										amtsValues = append(amtsValues, amount)
										amtsValues = append(amtsValues, memo)
										amtsValues = append(amtsValues, msgId.String+"/"+userContact.String)
										amtsValues = append(amtsValues, payback)
										amtsValues = append(amtsValues, admin_amt)
									}

								} else {
									ecnt++
									db.Exec("update cb_msg_" + memUserId.String + " set CODE = 'RCS', MESSAGE_TYPE='nr', MESSAGE = ?, RESULT = 'N' where remark4 = ? and msgid = ?", resultMessage.String, mstId.String, msgId.String)
								}
						}

						dhnResultStr = append(dhnResultStr, "?")
						dhnResultValues = append(dhnResultValues, rrId)
					}

					if len(ossmsStrs) > 0 {
						stmt := fmt.Sprintf("insert into OShotSMS(Sender,Receiver,Msg,URL,ReserveDT,TimeoutDT,SendResult,mst_id,cb_msg_id) values %s", s.Join(ossmsStrs, ","))
						_, err := db.Exec(stmt, ossmsValues...)

						if err != nil {
							stdlog.Println("(신) Rcs - 스마트미 SMS Table Insert 처리 중 오류 발생 " + err.Error())
						}

						ossmsStrs = nil
						ossmsValues = nil
					}

					if len(osmmsStrs) > 0 {
						stmt := fmt.Sprintf("insert into OShotMMS(MsgGroupID,Sender,Receiver,Subject,Msg,ReserveDT,TimeoutDT,SendResult,File_Path1,File_Path2,File_Path3,mst_id,cb_msg_id ) values %s", s.Join(osmmsStrs, ","))
						_, err := db.Exec(stmt, osmmsValues...)

						if err != nil {
							stdlog.Println("(신) Rcs - 스마트미 LMS Table Insert 처리 중 오류 발생 " + err.Error())
						}

						osmmsStrs = nil
						osmmsValues = nil
					}

					if len(nnsmsStrs) > 0 {
						stmt := fmt.Sprintf("insert into SMS_MSG(TR_CALLBACK,TR_PHONE,TR_MSG,TR_SENDDATE,TR_SENDSTAT,TR_MSGTYPE,TR_ETC9,TR_ETC10,TR_IDENTIFICATION_CODE,TR_ETC8) values %s", s.Join(nnsmsStrs, ","))
						_, err := db.Exec(stmt, nnsmsValues...)

						if err != nil {
							stdlog.Println("(신) Rcs - 나노 SMS Table Insert 처리 중 오류 발생 " + err.Error())
						}

						nnsmsStrs = nil
						nnsmsValues = nil
					}

					if len(nnmmsStrs) > 0 {
						stmt := fmt.Sprintf("insert into MMS_MSG(CALLBACK,PHONE,SUBJECT,MSG,REQDATE,STATUS,FILE_CNT,FILE_PATH1,FILE_PATH2,FILE_PATH3,ETC9,ETC10,IDENTIFICATION_CODE,ETC8) values %s", s.Join(nnmmsStrs, ","))
						_, err := db.Exec(stmt, nnmmsValues...)

						if err != nil {
							stdlog.Println("(신) Rcs - 나노 LMS Table Insert 처리 중 오류 발생 " + err.Error())
						}

						nnmmsStrs = nil
						nnmmsValues = nil
					}

					if len(lgusmsStrs) > 0 {
						stmt := fmt.Sprintf("insert into LG_SC_TRAN(TR_SENDDATE,TR_PHONE,TR_CALLBACK, TR_MSG, TR_ETC1, TR_ETC2, TR_ETC3, TR_KISAORIGCODE) values %s", s.Join(lgusmsStrs, ","))
						_, err := db.Exec(stmt, lgusmsValues...)

						if err != nil {
							stdlog.Println("(신) Rcs - LGU SMS Table Insert 처리 중 오류 발생 " + err.Error())
						}

						lgusmsStrs = nil
						lgusmsValues = nil
					}

					if len(lgummsStrs) > 0 {
						stmt := fmt.Sprintf("insert into LG_MMS_MSG(SUBJECT, PHONE, CALLBACK, REQDATE, MSG, FILE_CNT, FILE_PATH1, FILE_PATH2, FILE_PATH3, ETC1, ETC2, ETC3, KISA_ORIGCODE) values %s", s.Join(lgummsStrs, ","))
						_, err := db.Exec(stmt, lgummsValues...)

						if err != nil {
							stdlog.Println("(신) Rcs - LGU LMS Table Insert 처리 중 오류 발생 " + err.Error())
						}

						lgummsStrs = nil
						lgummsValues = nil
					}

					if len(amtsStrs) > 0 {
						stmt := fmt.Sprintf(amtinsstr, s.Join(amtsStrs, ","))
						_, err := db.Exec(stmt, amtsValues...)

						if err != nil {
							stdlog.Println("(신) Rcs - AMT Table Insert 처리 중 오류 발생 " + err.Error())
						}

						amtsStrs = nil
						amtsValues = nil
					}

					if len(dhnResultStr) > 0 {
						stmt := fmt.Sprintf(dhnResultSql, s.Join(dhnResultStr, ","))
						_, err := db.Exec(stmt, dhnResultValues...)

						if err != nil {
							stdlog.Println("(신) Rcs - " + config.Conf.RCSTABLE + " status 4 err : " + err.Error())
						}

						dhnResultStr = nil
						dhnResultValues = nil
					}

					db.Exec("update cb_wt_msg_sent set mst_rcs = ifnull(mst_rcs,0) + ?,mst_err_rcs = ifnull(mst_err_rcs,0) + ?, mst_wait = mst_wait - ?  where mst_id=?", scnt, ecnt, (ecnt + scnt), mstId.String)

					stdlog.Println("(신) Rcs - 처리 끝 : (", mstId.String, " ) 성공 : ", scnt, " / 실패 : ", ecnt)

				}
				defer selRows.Close()
			}
			defer selUserIdRows.Close()
		}
	} else {
		time.Sleep(1 * time.Second)
	}
// DHN 결과 처리 끝
}
