package rcs

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	s "strings"
	"sync"
	"time"

	"webagent/src/config"
	"webagent/src/baseprice"
	"webagent/src/databasepool"
)


func ResultProcessMZ(ctx context.Context) {
	config.Stdlog.Println("Rcs MZ - 결과 처리 프로세스 시작")
	var wg sync.WaitGroup
	for {
		select {
		case <-ctx.Done():
			config.Stdlog.Println("Rcs MZ - Result Process 종료")
			return
		default:
			wg.Add(1)
			go resultProcMZ(&wg)
			wg.Wait()
		}
	}
}

func resultProcMZ(wg *sync.WaitGroup) {
	defer wg.Done()
	defer func() {
		if r := recover(); r != nil {
			config.Stdlog.Println("Rcs MZ - resultProcMZ panic : ", r)
			// DB Ping 등 복구 로직
		}
	}()

	db := databasepool.DB
	stdlog := config.Stdlog

	// 1. 발송은 완료(proc='N')되었으나 결과처리가 안된 MZ 데이터 조회
	sqlStr := "SELECT rcs_id, msg_id, user_contact, msg_group_id, megabird_msg_grp_no FROM RCS_MESSAGE_RESULT WHERE proc = 'N' AND platform = 'MZ' LIMIT 500"

	rows, err := db.Query(sqlStr)
	if err != nil {
		stdlog.Println("Rcs MZ - Select 오류 :", err)
		time.Sleep(1 * time.Second)
		return
	}
	defer rows.Close()

	for rows.Next() {
		var rcsID, msgID, userContact, msgGroupID sql.NullString
		var megabirdMsgGrpNo int64

		err := rows.Scan(&rcsID, &msgID, &userContact, &msgGroupID, &megabirdMsgGrpNo)
		if err != nil || megabirdMsgGrpNo == 0 {
			continue
		}

		// 2. 메가버드 API 호출 (GET 방식, Authorization 필수)
		queryUrl := fmt.Sprintf("https://api.megabird.com/v1/openapi/message/receiver/%d", megabirdMsgGrpNo)
		
		resp, err := config.Client.R().
			SetHeaders(map[string]string{
				"Content-Type":  "application/json",
				"Authorization": config.Conf.MZRCSKEY, // 💡 문서 누락분 완벽 적용
			}).
			Get(queryUrl)

		if err != nil {
			stdlog.Println("Rcs MZ - 결과 조회 API 호출 에러 :", err)
			continue
		}

		var mzResp MegabirdResultResponse
		json.Unmarshal(resp.Body(), &mzResp)

		if mzResp.Code == "0000" && len(mzResp.Result.Result) > 0 {
			// 단건 발송이므로 첫 번째 결과만 확인
			resDetail := mzResp.Result.Result[0]
			
			var finalStatus, finalError string

			// 💡 3. MTC(성공), MTF(실패) 판별하여 기존 KT 상태값(success, fail)으로 스위칭
			if resDetail.SndStsCd == "MTC" {
				finalStatus = "success"
			} else if resDetail.SndStsCd == "MTF" {
				finalStatus = "fail"
				finalError = resDetail.SndStsCzCmt
			} else {
				// MTC, MTF 외의 상태(대기 중 등)는 이번 턴에 처리하지 않고 스킵
				continue 
			}

			// 4. RCS_MESSAGE_RESULT 테이블 상태 업데이트
			updSql := `UPDATE RCS_MESSAGE_RESULT 
			           SET result_status = ?, result_error = ?, proc = 'P' 
			           WHERE rcs_id = ? AND platform = 'MZ'`
			_, dbErr := db.Exec(updSql, finalStatus, finalError, rcsID.String)
			
			if dbErr != nil {
				stdlog.Println("Rcs MZ - Update 에러 :", dbErr)
			}
		}
	}

	groupsql := "select distinct msg_group_id as mst_id, wms.send_mem_id as usermem_id from RCS_MESSAGE_RESULT rmr inner join cb_wt_msg_sent wms on rmr.msg_group_id = wms.mst_id where rmr.proc = 'P' and rmr.platform = 'MZ'"

	grows, err := db.Query(groupsql)
	if err != nil {
		stdlog.Println("Rcs MZ - RCS_MESSAGE_RESULT select 오류 ", err)
	} else {
		defer grows.Close()

		var mst_id, usermem_id sql.NullString

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

		tntsmsStrs := []string{}
		tntsmsValues := []interface{}{}

		tntmmsStrs := []string{}
		tntmmsValues := []interface{}{}

		jjsmsStrs := []string{}
		jjsmsValues := []interface{}{}

		jjmmsStrs := []string{}
		jjmmsValues := []interface{}{}

		for grows.Next() {

			ossmsStrs = nil //스마트미 SMS Table Insert 용
			ossmsValues = nil

			osmmsStrs = nil //스마트미 LMS/MMS Table Insert 용
			osmmsValues = nil

			lgusmsStrs = nil //LGU SMS Table Insert 용
			lgusmsValues = nil

			lgummsStrs = nil //LGU LMS/MMS Table Insert 용
			lgummsValues = nil

			nnsmsStrs = nil //스마트미 SMS Table Insert 용
			nnsmsValues = nil

			nnmmsStrs = nil //스마트미 LMS/MMS Table Insert 용
			nnmmsValues = nil

			tntsmsStrs = nil //SMTNT SMS Table Insert 용
			tntsmsValues = nil

			tntmmsStrs = nil //SMTNT LMS/MMS Table Insert 용
			tntmmsValues = nil

			jjsmsStrs = nil //JJ SMS Table Insert 용
			jjsmsValues = nil

			jjmmsStrs = nil //JJ LMS/MMS Table Insert 용
			jjmmsValues = nil

			var scnt int = 0
			var ecnt int = 0

			grows.Scan(&mst_id, &usermem_id)

			ressql := `
				SELECT
					rmr.msg_id as msgid
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
				FROM
					RCS_MESSAGE_RESULT rmr 
			   inner join
			   	cb_wt_msg_sent wms on rmr.msg_group_id = wms.mst_id 
				where
					rmr.proc = 'P'
					and rmr.msg_group_id = ?
					and rmr.platform = 'MZ'`

			resrows, err := db.Query(ressql, mst_id.String)
			if err != nil {
				stdlog.Println("Rcs MZ - 결과 처리 Select 오류 ", err)
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

							stdlog.Println("Rcs MZ - 발송 실패 -> WEB(C) 발송 처리 ", mst_type3.String, msr_exptime.String, msgid.String)

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

							} else if s.EqualFold(mst_type3.String, "wc") || s.EqualFold(mst_type3.String, "wcm") {
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
						} else if s.Contains(mst_type3.String, "wa") && s.EqualFold(msr_exptime.String, "Y") {

							stdlog.Println("Rcs MZ - 발송 실패 -> WEB(A) 발송 처리 ", mst_type3.String, msr_exptime.String, msgid.String)

							db.Exec("update cb_msg_"+userid.String+" set CODE = 'GRS', MESSAGE_TYPE='gr' where remark4=? and msgid = ?", remark4.String, msgid.String)

							if s.EqualFold(mst_type3.String, "was") {

								nnsmsStrs = append(nnsmsStrs, "(?,?,?,?,?,?,?,?,?,'Y')")
								nnsmsValues = append(nnsmsValues, sms_sender.String)
								nnsmsValues = append(nnsmsValues, phnstr.String)
								nnsmsValues = append(nnsmsValues, mst_lms_content.String)
								nnsmsValues = append(nnsmsValues, time.Now().Format("2006-01-02 15:04:05"))
								nnsmsValues = append(nnsmsValues, "0")
								nnsmsValues = append(nnsmsValues, "0")
								nnsmsValues = append(nnsmsValues, msgid.String)
								nnsmsValues = append(nnsmsValues, remark4.String)
								nnsmsValues = append(nnsmsValues, config.Conf.KISACODE)

								if s.EqualFold(mst_sent_voucher.String, "V") {
									amount = cprice.V_price_smt_sms.Float64
									payback = cprice.V_price_smt_sms.Float64 - cprice.P_price_smt_sms.Float64
									admin_amt = cprice.B_price_smt_sms.Float64
									memo = "웹(A) SMS,바우처"
								} else {
									amount = cprice.C_price_smt_sms.Float64
									payback = cprice.C_price_smt_sms.Float64 - cprice.P_price_smt_sms.Float64
									admin_amt = cprice.B_price_smt_sms.Float64
									if s.EqualFold(mst_sent_voucher.String, "B") {
										memo = "웹(A) SMS,보너스"
									} else {
										memo = "웹(A) SMS"
									}
								}

								amtsStrs = append(amtsStrs, "(now(),?,?,?,?,?,?)")
								amtsValues = append(amtsValues, "P")
								amtsValues = append(amtsValues, amount)
								amtsValues = append(amtsValues, memo)
								amtsValues = append(amtsValues, msgid.String+"/"+phnstr.String)
								amtsValues = append(amtsValues, payback)
								amtsValues = append(amtsValues, admin_amt)

							} else if s.EqualFold(mst_type3.String, "wa") || s.EqualFold(mst_type3.String, "wam") {

								filecnt := 0

								if len(mms_file1.String) > 0 {
									filecnt = filecnt + 1
								}

								if len(mms_file2.String) > 0 {
									filecnt = filecnt + 1
								}

								if len(mms_file3.String) > 0 {
									filecnt = filecnt + 1
								}

								nnmmsStrs = append(nnmmsStrs, "( ?,?,?,?,?,?,?,?,?,?,?,?,?,'Y')")

								nnmmsValues = append(nnmmsValues, sms_sender.String)
								nnmmsValues = append(nnmmsValues, phnstr.String)
								nnmmsValues = append(nnmmsValues, rcsBody.Title)
								nnmmsValues = append(nnmmsValues, mst_lms_content.String)
								nnmmsValues = append(nnmmsValues, time.Now().Format("2006-01-02 15:04:05"))
								nnmmsValues = append(nnmmsValues, "0")
								nnmmsValues = append(nnmmsValues, filecnt)
								nnmmsValues = append(nnmmsValues, mms_file1)
								nnmmsValues = append(nnmmsValues, mms_file2)
								nnmmsValues = append(nnmmsValues, mms_file3)
								nnmmsValues = append(nnmmsValues, msgid.String)
								nnmmsValues = append(nnmmsValues, remark4.String)
								nnmmsValues = append(nnmmsValues, config.Conf.KISACODE)

								if len(mms_file1.String) <= 0 {
									if s.EqualFold(mst_sent_voucher.String, "V") {
										amount = cprice.V_price_smt.Float64
										payback = cprice.V_price_smt.Float64 - cprice.P_price_smt.Float64
										admin_amt = cprice.B_price_smt.Float64
										memo = "웹(A) LMS,바우처"
									} else {
										amount = cprice.C_price_smt.Float64
										payback = cprice.C_price_smt.Float64 - cprice.P_price_smt.Float64
										admin_amt = cprice.B_price_smt.Float64
										if s.EqualFold(mst_sent_voucher.String, "B") {
											memo = "웹(A) LMS,보너스"
										} else {
											memo = "웹(A) LMS"
										}
									}
								} else {
									if s.EqualFold(mst_sent_voucher.String, "V") {
										amount = cprice.V_price_smt_mms.Float64
										payback = cprice.V_price_smt_mms.Float64 - cprice.P_price_smt_mms.Float64
										admin_amt = cprice.B_price_smt_mms.Float64
										memo = "웹(A) MMS,바우처"
									} else {
										amount = cprice.C_price_smt_mms.Float64
										payback = cprice.C_price_smt_mms.Float64 - cprice.P_price_smt_mms.Float64
										admin_amt = cprice.B_price_smt_mms.Float64
										if s.EqualFold(mst_sent_voucher.String, "B") {
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
								amtsValues = append(amtsValues, msgid.String+"/"+phnstr.String)
								amtsValues = append(amtsValues, payback)
								amtsValues = append(amtsValues, admin_amt)
							}
						} else if s.Contains(mst_type3.String, "wb") && s.EqualFold(msr_exptime.String, "Y") {

							stdlog.Println("Rcs MZ - 발송 실패 -> WEB(B) 발송 처리 ", mst_type3.String, msr_exptime.String, msgid.String)

							db.Exec("update cb_msg_"+userid.String+" set CODE = 'LGU', MESSAGE_TYPE='lg' where remark4=? and msgid = ?", remark4.String, msgid.String)

							if s.EqualFold(mst_type3.String, "wbs") {
								lgusmsStrs = append(lgusmsStrs, "(?,?,?,?,?,?,?,?)")
								lgusmsValues = append(lgusmsValues, time.Now().Format("2006-01-02 15:04:05"))
								lgusmsValues = append(lgusmsValues, phnstr)
								lgusmsValues = append(lgusmsValues, sms_sender)
								lgusmsValues = append(lgusmsValues, mst_lms_content.String)
								lgusmsValues = append(lgusmsValues, msgid.String)
								lgusmsValues = append(lgusmsValues, userid.String)
								lgusmsValues = append(lgusmsValues, remark4.String)
								lgusmsValues = append(lgusmsValues, config.Conf.KISACODE)

								admin_amt = cprice.B_price_nas_sms.Float64
								if s.EqualFold(mst_sent_voucher.String, "V") {
									amount = cprice.V_price_nas_sms.Float64
									payback = cprice.V_price_nas_sms.Float64 - cprice.P_price_nas_sms.Float64
									memo = "웹(B) SMS,바우처"
								} else {
									amount = cprice.C_price_nas_sms.Float64
									payback = cprice.C_price_nas_sms.Float64 - cprice.P_price_nas_sms.Float64
									if s.EqualFold(mst_sent_voucher.String, "B") {
										memo = "웹(B) SMS,보너스"
									} else {
										memo = "웹(B) SMS"
									}
								}

								amtsStrs = append(amtsStrs, "(now(),?,?,?,?,?,?)")
								amtsValues = append(amtsValues, "P")
								amtsValues = append(amtsValues, amount)
								amtsValues = append(amtsValues, memo)
								amtsValues = append(amtsValues, msgid.String+"/"+phnstr.String)
								amtsValues = append(amtsValues, payback)
								amtsValues = append(amtsValues, admin_amt)
							} else if s.EqualFold(mst_type3.String, "wb") || s.EqualFold(mst_type3.String, "wbm") {
								file_cnt := 0
								if mms_file1.String != "" {
									file_cnt++
								}
								if mms_file2.String != "" {
									file_cnt++
								}
								if mms_file3.String != "" {
									file_cnt++
								}
								lgummsStrs = append(lgummsStrs, "(?,?,?,?,?,?,?,?,?,?,?,?,?)")
								lgummsValues = append(lgummsValues, rcsBody.Title)
								lgummsValues = append(lgummsValues, phnstr)
								lgummsValues = append(lgummsValues, sms_sender)
								lgummsValues = append(lgummsValues, time.Now().Format("2006-01-02 15:04:05"))
								lgummsValues = append(lgummsValues, mst_lms_content.String)
								lgummsValues = append(lgummsValues, file_cnt)
								lgummsValues = append(lgummsValues, mms_file1)
								lgummsValues = append(lgummsValues, mms_file2)
								lgummsValues = append(lgummsValues, mms_file3)
								lgummsValues = append(lgummsValues, msgid.String)
								lgummsValues = append(lgummsValues, userid.String)
								lgummsValues = append(lgummsValues, remark4.String)
								lgummsValues = append(lgummsValues, config.Conf.KISACODE)

								if len(mms_file1.String) <= 0 {

									admin_amt = cprice.B_price_nas.Float64
									if s.EqualFold(mst_sent_voucher.String, "V") {
										amount = cprice.V_price_nas.Float64
										payback = cprice.V_price_nas.Float64 - cprice.P_price_nas.Float64
										memo = "웹(B) LMS,바우처"
									} else {
										amount = cprice.C_price_nas.Float64
										payback = cprice.C_price_nas.Float64 - cprice.P_price_nas.Float64
										if s.EqualFold(mst_sent_voucher.String, "B") {
											memo = "웹(B) LMS,보너스"
										} else {
											memo = "웹(B) LMS"
										}
									}
								} else {

									admin_amt = cprice.B_price_nas_mms.Float64
									if s.EqualFold(mst_sent_voucher.String, "V") {
										amount = cprice.V_price_nas_mms.Float64
										payback = cprice.V_price_nas_mms.Float64 - cprice.P_price_nas_mms.Float64
										memo = "웹(B) MMS,바우처"
									} else {
										amount = cprice.C_price_nas_mms.Float64
										payback = cprice.C_price_nas_mms.Float64 - cprice.P_price_nas_mms.Float64
										if s.EqualFold(mst_sent_voucher.String, "B") {
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
								amtsValues = append(amtsValues, msgid.String+"/"+phnstr.String)
								amtsValues = append(amtsValues, payback)
								amtsValues = append(amtsValues, admin_amt)
							}

						} else if s.Contains(mst_type3.String, "wd") && s.EqualFold(msr_exptime.String, "Y") {
							stdlog.Println("Rcs MZ - 발송 실패 -> WEB(D) 발송 처리 ", mst_type3.String, msr_exptime.String, msgid.String)

							db.Exec("update cb_msg_"+userid.String+" set CODE = 'TNT', MESSAGE_TYPE='tn' where remark4=? and msgid = ?", remark4.String, msgid.String)

							smtntTime := time.Now().Format("2006-01-02 15:04:05")
							if s.EqualFold(mst_type3.String, "wds") {
								tntsmsStrs = append(tntsmsStrs, "(?,?,?,?,?,?,?,?,?,?)")
								tntsmsValues = append(tntsmsValues, phnstr)                 // Phone_No 1
								tntsmsValues = append(tntsmsValues, sms_sender)             // Callback_No 2
								tntsmsValues = append(tntsmsValues, "4")                    // Msg_Type 3
								tntsmsValues = append(tntsmsValues, smtntTime)              // Send_Time 4
								tntsmsValues = append(tntsmsValues, smtntTime)              // Save_Time 5
								tntsmsValues = append(tntsmsValues, mst_lms_content.String) // Message 6
								tntsmsValues = append(tntsmsValues, config.Conf.KISACODE)   // Reseller_Code 7

								tntsmsValues = append(tntsmsValues, msgid.String)   // Etc1 8
								tntsmsValues = append(tntsmsValues, userid.String)  // Etc2 9
								tntsmsValues = append(tntsmsValues, remark4.String) // Etc3 10

								admin_amt = cprice.B_price_tnt_sms.Float64
								if s.EqualFold(mst_sent_voucher.String, "V") {
									amount = cprice.V_price_tnt_sms.Float64
									payback = cprice.V_price_tnt_sms.Float64 - cprice.P_price_tnt_sms.Float64
									memo = "웹(D) SMS,바우처"
								} else {
									amount = cprice.C_price_tnt_sms.Float64
									payback = cprice.C_price_tnt_sms.Float64 - cprice.P_price_tnt_sms.Float64
									if s.EqualFold(mst_sent_voucher.String, "B") {
										memo = "웹(D) SMS,보너스"
									} else {
										memo = "웹(D) SMS"
									}
								}

								amtsStrs = append(amtsStrs, "(now(),?,?,?,?,?,?)")
								amtsValues = append(amtsValues, "P")
								amtsValues = append(amtsValues, amount)
								amtsValues = append(amtsValues, memo)
								amtsValues = append(amtsValues, msgid.String+"/"+phnstr.String)
								amtsValues = append(amtsValues, payback)
								amtsValues = append(amtsValues, admin_amt)
							} else if s.EqualFold(mst_type3.String, "wd") || s.EqualFold(mst_type3.String, "wdm") {
								fileCnt := 0
								fileType1 := ""
								fileType2 := ""
								fileType3 := ""
								if len(mms_file1.String) > 0 {
									fileCnt++
									fileType1 = "IMG"
								}
								if len(mms_file2.String) > 0 {
									fileCnt++
									fileType2 = "IMG"
								}
								if len(mms_file3.String) > 0 {
									fileCnt++
									fileType3 = "IMG"
								}
								tntmmsStrs = append(tntmmsStrs, "(?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?)")
								tntmmsValues = append(tntmmsValues, phnstr)                 // Phone_No 1
								tntmmsValues = append(tntmmsValues, sms_sender)             // Callback_No 2
								tntmmsValues = append(tntmmsValues, "6")                    // Msg_Type 3
								tntmmsValues = append(tntmmsValues, smtntTime)              // Send_Time 4
								tntmmsValues = append(tntmmsValues, smtntTime)              // Save_Time 5
								tntmmsValues = append(tntmmsValues, rcsBody.Title)          // Subject 6
								tntmmsValues = append(tntmmsValues, mst_lms_content.String) // Message 7
								tntmmsValues = append(tntmmsValues, fileCnt)                // File_Count 8
								tntmmsValues = append(tntmmsValues, fileType1)              // File_Type1 9
								tntmmsValues = append(tntmmsValues, fileType2)              // File_Type2 10
								tntmmsValues = append(tntmmsValues, fileType3)              // File_Type3 11
								tntmmsValues = append(tntmmsValues, mms_file1)              // File_Name1 12
								tntmmsValues = append(tntmmsValues, mms_file2)              // File_Name2 13
								tntmmsValues = append(tntmmsValues, mms_file3)              // File_Name3 14
								tntmmsValues = append(tntmmsValues, config.Conf.KISACODE)   // Reseller_Code 15

								tntmmsValues = append(tntmmsValues, msgid.String)   // Etc1 16
								tntmmsValues = append(tntmmsValues, userid.String)  // Etc2 17
								tntmmsValues = append(tntmmsValues, remark4.String) // Etc3 18

								if len(mms_file1.String) <= 0 {
									admin_amt = cprice.B_price_tnt.Float64
									if s.EqualFold(mst_sent_voucher.String, "V") {
										amount = cprice.V_price_tnt.Float64
										payback = cprice.V_price_tnt.Float64 - cprice.P_price_tnt.Float64
										memo = "웹(D) LMS,바우처"
									} else {
										amount = cprice.C_price_tnt.Float64
										payback = cprice.C_price_tnt.Float64 - cprice.P_price_tnt.Float64
										if s.EqualFold(mst_sent_voucher.String, "B") {
											memo = "웹(D) LMS,보너스"
										} else {
											memo = "웹(D) LMS"
										}
									}
								} else {
									admin_amt = cprice.B_price_tnt_mms.Float64
									if s.EqualFold(mst_sent_voucher.String, "V") {
										amount = cprice.V_price_tnt_mms.Float64
										payback = cprice.V_price_tnt_mms.Float64 - cprice.P_price_tnt_mms.Float64
										memo = "웹(D) MMS,바우처"
									} else {
										amount = cprice.C_price_tnt_mms.Float64
										payback = cprice.C_price_tnt_mms.Float64 - cprice.P_price_tnt_mms.Float64
										if s.EqualFold(mst_sent_voucher.String, "B") {
											memo = "웹(D) MMS,보너스"
										} else {
											memo = "웹(D) MMS"
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
						} else if s.Contains(mst_type3.String, "we") && s.EqualFold(msr_exptime.String, "Y") {

							stdlog.Println("Rcs MZ - 발송 실패 -> WEB(E) 발송 처리 ", mst_type3.String, msr_exptime.String, msgid.String)

							db.Exec("update cb_msg_"+userid.String+" set CODE = 'JJ', MESSAGE_TYPE='jj' where remark4=? and msgid = ?", remark4.String, msgid.String)

							if s.EqualFold(mst_type3.String, "wes") {
								jjsmsStrs = append(jjsmsStrs, "(?,?,?,?,?,?,?,?,?)")
								jjsmsValues = append(jjsmsValues, phnstr.String)
								jjsmsValues = append(jjsmsValues, sms_sender.String)
								jjsmsValues = append(jjsmsValues, "SMS")
								jjsmsValues = append(jjsmsValues, rcsBody.Title)
								jjsmsValues = append(jjsmsValues, mst_lms_content.String)
								jjsmsValues = append(jjsmsValues, config.Conf.KISACODE)
								jjsmsValues = append(jjsmsValues, msgid.String)
								jjsmsValues = append(jjsmsValues, userid)
								jjsmsValues = append(jjsmsValues, remark4.String)

								if s.EqualFold(mst_sent_voucher.String, "V") {
									amount = cprice.V_price_smt_sms.Float64
									payback = cprice.V_price_smt_sms.Float64 - cprice.P_price_smt_sms.Float64
									admin_amt = cprice.B_price_smt_sms.Float64
									memo = "웹(E) SMS,바우처"
								} else {
									amount = cprice.C_price_smt_sms.Float64
									payback = cprice.C_price_smt_sms.Float64 - cprice.P_price_smt_sms.Float64
									admin_amt = cprice.B_price_smt_sms.Float64
									if s.EqualFold(mst_sent_voucher.String, "B") {
										memo = "웹(E) SMS,보너스"
									} else {
										memo = "웹(E) SMS"
									}
								}

								amtsStrs = append(amtsStrs, "(now(),?,?,?,?,?,?)")
								amtsValues = append(amtsValues, "P")
								amtsValues = append(amtsValues, amount)
								amtsValues = append(amtsValues, memo)
								amtsValues = append(amtsValues, msgid.String+"/"+phnstr.String)
								amtsValues = append(amtsValues, payback)
								amtsValues = append(amtsValues, admin_amt)

							} else if s.EqualFold(mst_type3.String, "we") || s.EqualFold(mst_type3.String, "wem") {
								jjMsgType := "LMS"
								if (mms_file1.Valid && mms_file1.String != "" && len(mms_file1.String) > 0) ||
									(mms_file2.Valid && mms_file2.String != "" && len(mms_file2.String) > 0) ||
									(mms_file3.Valid && mms_file3.String != "" && len(mms_file3.String) > 0) {
									jjMsgType = "MMS"
								}

								jjmmsStrs = append(jjmmsStrs, "(?,?,?,?,?,?,?,?,?,?,?,?)")
								jjmmsValues = append(jjmmsValues, phnstr.String)
								jjmmsValues = append(jjmmsValues, sms_sender.String)
								jjmmsValues = append(jjmmsValues, jjMsgType)
								jjmmsValues = append(jjmmsValues, rcsBody.Title)
								jjmmsValues = append(jjmmsValues, mst_lms_content.String)
								if mms_file1.Valid && mms_file1.String != "" && len(mms_file1.String) > 0 {
									jjmmsValues = append(jjmmsValues, mms_file1.String)
								} else {
									jjmmsValues = append(jjmmsValues, sql.NullString{})
								}
								if mms_file2.Valid && mms_file2.String != "" && len(mms_file2.String) > 0 {
									jjmmsValues = append(jjmmsValues, mms_file2.String)
								} else {
									jjmmsValues = append(jjmmsValues, sql.NullString{})
								}
								if mms_file3.Valid && mms_file3.String != "" && len(mms_file3.String) > 0 {
									jjmmsValues = append(jjmmsValues, mms_file3.String)
								} else {
									jjmmsValues = append(jjmmsValues, sql.NullString{})
								}
								jjmmsValues = append(jjmmsValues, config.Conf.KISACODE)
								jjmmsValues = append(jjmmsValues, msgid.String)
								jjmmsValues = append(jjmmsValues, userid)
								jjmmsValues = append(jjmmsValues, remark4.String)

								if len(mms_file1.String) <= 0 {
									if s.EqualFold(mst_sent_voucher.String, "V") {
										amount = cprice.V_price_smt.Float64
										payback = cprice.V_price_smt.Float64 - cprice.P_price_smt.Float64
										admin_amt = cprice.B_price_smt.Float64
										memo = "웹(E) LMS,바우처"
									} else {
										amount = cprice.C_price_smt.Float64
										payback = cprice.C_price_smt.Float64 - cprice.P_price_smt.Float64
										admin_amt = cprice.B_price_smt.Float64
										if s.EqualFold(mst_sent_voucher.String, "B") {
											memo = "웹(E) LMS,보너스"
										} else {
											memo = "웹(E) LMS"
										}
									}
								} else {
									if s.EqualFold(mst_sent_voucher.String, "V") {
										amount = cprice.V_price_smt_mms.Float64
										payback = cprice.V_price_smt_mms.Float64 - cprice.P_price_smt_mms.Float64
										admin_amt = cprice.B_price_smt_mms.Float64
										memo = "웹(E) MMS,바우처"
									} else {
										amount = cprice.C_price_smt_mms.Float64
										payback = cprice.C_price_smt_mms.Float64 - cprice.P_price_smt_mms.Float64
										admin_amt = cprice.B_price_smt_mms.Float64
										if s.EqualFold(mst_sent_voucher.String, "B") {
											memo = "웹(E) MMS,보너스"
										} else {
											memo = "웹(E) MMS"
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
						stdlog.Println("Rcs MZ - 스마트미 SMS Table Insert 처리 중 오류 발생 " + err.Error())
					}

					ossmsStrs = nil
					ossmsValues = nil
				}

				if len(osmmsStrs) > 0 {
					stmt := fmt.Sprintf("insert into OShotMMS(MsgGroupID,Sender,Receiver,Subject,Msg,ReserveDT,TimeoutDT,SendResult,File_Path1,File_Path2,File_Path3,mst_id,cb_msg_id ) values %s", s.Join(osmmsStrs, ","))
					_, err := db.Exec(stmt, osmmsValues...)

					if err != nil {
						stdlog.Println("Rcs MZ - 스마트미 LMS Table Insert 처리 중 오류 발생 " + err.Error())
					}

					osmmsStrs = nil
					osmmsValues = nil
				}

				if len(lgusmsStrs) > 0 {
					stmt := fmt.Sprintf("insert into LG_SC_TRAN(TR_SENDDATE,TR_PHONE,TR_CALLBACK, TR_MSG, TR_ETC1, TR_ETC2, TR_ETC3, TR_KISAORIGCODE) values %s", s.Join(lgusmsStrs, ","))
					_, err := db.Exec(stmt, lgusmsValues...)

					if err != nil {
						stdlog.Println("Rcs MZ - LGU SMS Table Insert 처리 중 오류 발생 " + err.Error())
					}

					lgusmsStrs = nil
					lgusmsValues = nil
				}

				if len(lgummsStrs) > 0 {
					stmt := fmt.Sprintf("insert into LG_MMS_MSG(SUBJECT, PHONE, CALLBACK, REQDATE, MSG, FILE_CNT, FILE_PATH1, FILE_PATH2, FILE_PATH3, ETC1, ETC2, ETC3, KISA_ORIGCODE) values %s", s.Join(lgummsStrs, ","))
					_, err := db.Exec(stmt, lgummsValues...)

					if err != nil {
						stdlog.Println("Rcs MZ - LGU LMS Table Insert 처리 중 오류 발생 " + err.Error())
					}

					lgummsStrs = nil
					lgummsValues = nil
				}

				if len(nnsmsStrs) > 0 {
					stmt := fmt.Sprintf("insert into SMS_MSG(TR_CALLBACK,TR_PHONE,TR_MSG,TR_SENDDATE,TR_SENDSTAT,TR_MSGTYPE,TR_ETC9,TR_ETC10,TR_IDENTIFICATION_CODE,TR_ETC8) values %s", s.Join(nnsmsStrs, ","))
					_, err := db.Exec(stmt, nnsmsValues...)

					if err != nil {
						stdlog.Println("Rcs MZ - 나노 SMS Table Insert 처리 중 오류 발생 " + err.Error())
					}

					nnsmsStrs = nil
					nnsmsValues = nil
				}

				if len(nnmmsStrs) > 0 {
					stmt := fmt.Sprintf("insert into MMS_MSG(CALLBACK,PHONE,SUBJECT,MSG,REQDATE,STATUS,FILE_CNT,FILE_PATH1,FILE_PATH2,FILE_PATH3,ETC9,ETC10,IDENTIFICATION_CODE,ETC8) values %s", s.Join(nnmmsStrs, ","))
					_, err := db.Exec(stmt, nnmmsValues...)

					if err != nil {
						stdlog.Println("Rcs MZ - 나노 LMS Table Insert 처리 중 오류 발생 " + err.Error())
					}

					nnmmsStrs = nil
					nnmmsValues = nil
				}

				if len(tntsmsStrs) > 0 {
					stmt := fmt.Sprintf("insert into Msg_Tran(Phone_No,Callback_No,Msg_Type,Send_Time,Save_Time,Message,Reseller_Code,Etc1,Etc2,Etc3) values %s", s.Join(tntsmsStrs, ","))
					_, err := db.Exec(stmt, tntsmsValues...)

					if err != nil {
						stdlog.Println("Rcs MZ - SMTNT SMS Table Insert 처리 중 오류 발생 " + err.Error())
					}

					tntsmsStrs = nil
					tntsmsValues = nil
				}

				if len(tntmmsStrs) > 0 {
					stmt := fmt.Sprintf("insert into Msg_Tran(Phone_No,Callback_No,Msg_Type,Send_Time,Save_Time,Subject,Message,File_Count,File_Type1,File_Type2,File_Type3,File_Name1,File_Name2,File_Name3,Reseller_Code,Etc1,Etc2,Etc3) values %s", s.Join(tntmmsStrs, ","))
					_, err := db.Exec(stmt, tntmmsValues...)

					if err != nil {
						stdlog.Println("Rcs MZ - SMTNT LMS Table Insert 처리 중 오류 발생 " + err.Error())
					}

					tntmmsStrs = nil
					tntmmsValues = nil
				}

				if len(jjsmsStrs) > 0 {
					stmt := fmt.Sprintf("insert into MTMSG_DATA(CALL_TO,CALL_FROM,MSG_TYPE,SUBJECT,MESSAGE,IDENTIFIER,DHN_ETC1,DHN_ETC2,DHN_ETC3) values %s", s.Join(jjsmsStrs, ","))
					_, err := db.Exec(stmt, jjsmsValues...)

					if err != nil {
						stdlog.Println("Rcs MZ - JJ SMS Table Insert 처리 중 오류 발생 " + err.Error())
					}

					jjsmsStrs = nil
					jjsmsValues = nil
				}

				if len(jjmmsStrs) > 0 {
					stmt := fmt.Sprintf("insert into MTMSG_DATA(CALL_TO,CALL_FROM,MSG_TYPE,SUBJECT,MESSAGE,FILE_NAME1,FILE_NAME2,FILE_NAME3,IDENTIFIER,DHN_ETC1,DHN_ETC2,DHN_ETC3) values %s", s.Join(jjmmsStrs, ","))
					_, err := db.Exec(stmt, jjmmsValues...)

					if err != nil {
						stdlog.Println("Rcs MZ - JJ LMS Table Insert 처리 중 오류 발생 " + err.Error())
					}

					jjmmsStrs = nil
					jjmmsValues = nil
				}

				if len(amtsStrs) > 0 {
					stmt := fmt.Sprintf(amtinsstr, s.Join(amtsStrs, ","))
					_, err := db.Exec(stmt, amtsValues...)

					if err != nil {
						stdlog.Println("Rcs MZ - AMT Table Insert 처리 중 오류 발생 " + err.Error())
					}

					amtsStrs = nil
					amtsValues = nil
				}

				db.Exec("update cb_wt_msg_sent set mst_rcs = ifnull(mst_rcs,0) + ?,mst_err_rcs = ifnull(mst_err_rcs,0) + ?, mst_wait = mst_wait - ?  where mst_id=?", scnt, ecnt, (ecnt + scnt), mst_id.String)

				stdlog.Println("Rcs MZ - 처리 끝 : (", mst_id.String, " ) 성공 : ", scnt, " / 실패 : ", ecnt)
			}
		}
	}

	time.Sleep(1 * time.Second)
}