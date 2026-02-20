package rcs

import (
	"fmt"
	"sync"
	"time"
	"context"
	s "strings"
	"database/sql"
	"encoding/json"

	"webagent/src/config"
	"webagent/src/baseprice"
	"webagent/src/databasepool"

	_ "github.com/go-sql-driver/mysql"
)

func ResultProcessSmtnt(ctx context.Context) {
	config.Stdlog.Println("Rcs SMTNT - 결과 처리 프로세스 시작")
	var wg sync.WaitGroup
	for {
		select {
		case <- ctx.Done():
			config.Stdlog.Println("Rcs SMTNT - process가 15초 후에 종료")
		    time.Sleep(15 * time.Second)
		    config.Stdlog.Println("Rcs SMTNT - process 종료 완료")
			return
		default:
			var t = time.Now()

			if t.Day() <= 3 {
				wg.Add(1)
				go msgProcess(&wg, true)
			}

			wg.Add(1)
			go msgProcess(&wg, false)
			wg.Wait()
		}
	}
}

func msgProcess(wg *sync.WaitGroup, pastFlag bool) {
	defer wg.Done()
	defer func() {
		if r := recover(); r != nil {
			config.Stdlog.Println("Rcs SMTNT - panic 발생 원인 : ", r)
			if err, ok := r.(error); ok {
				if s.Contains(err.Error(), "connection refused") {
					for {
						config.Stdlog.Println("Rcs SMTNT - send ping to DB")
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
	var errlog = config.Stdlog

	var isProc = true
	var t = time.Now()

	if pastFlag {
		t = time.Now().Add(time.Hour * -96)
	}
	
	var monthStr = fmt.Sprintf("%d%02d", t.Year(), t.Month())

	var SMSTable = "Msg_Log_" + monthStr
	var msgcnt sql.NullString

	//발송 6시간 지난 메세지는 응답과 상관 없이 성공 처리 함.
	// sms 성공 처리
	err1 := db.QueryRow("SELECT count(1) as cnt from Msg_Tran WHERE Status='2' and Msg_Type in (9, 10, 11, 12) and date_add(Send_Time, interval 6 HOUR) < now() and Etc3 is not null and Etc2 not like 'khug%'").Scan(&msgcnt)
	if err1 != nil {
	   errlog.Println("Rcs SMTNT - 조회 중 오류 발생", err1)
	   panic(err1)
	} else {		
		if !s.EqualFold(msgcnt.String, "0") {	
			db.Exec("UPDATE Msg_Tran SET Status=3, Result=0, Telecom='000', Delivery_Time=now(), Result_Time=now() WHERE Status='2' and Msg_Type in (9, 10, 11, 12) and date_add(Send_Time, interval 6 HOUR) < now() and Etc3 is not null and Etc2 not like 'khug%'")
		}
	}

	var tickCnt sql.NullInt64
	var tickSql = `
		select
			count(1) as cnt
		from
			` + SMSTable + ` a
		inner join
			cb_wt_msg_sent b ON a.Etc3 = b.mst_id
		where
			a.Status = '3'
			and a.Msg_Type in (9, 10, 11, 12)
			and a.Etc4 is null
		limit 1`

	cnterr := databasepool.DB.QueryRow(tickSql).Scan(&tickCnt)

	if cnterr != nil && cnterr != sql.ErrNoRows {
		errlog.Println("Rcs SMTNT -", SMSTable, "Table - select error : " + cnterr.Error())
		if s.Index(cnterr.Error(), "1146") > 0 {
			db.Exec("Create Table IF NOT EXISTS " + SMSTable + " like Msg_Tran")
			errlog.Println("Rcs SMTNT -", SMSTable + " 생성 !!")
		}
		time.Sleep(10 * time.Second)
		return
	} else {
		if tickCnt.Int64 <= 0 {
			time.Sleep(100 * time.Millisecond)
			return
		}
	}

	var groupQuery = `
		select distinct
			a.Etc3,
			b.mst_mem_id as mem_id,
			(select mem_userid from cb_member cm where cm.mem_id = b.mst_mem_id) AS mem_userid,
			b.mst_sent_voucher
		from
			` + SMSTable + ` a
		inner join
			cb_wt_msg_sent b ON a.Etc3 = b.mst_id
		where
			a.Status = '3'
			and a.Msg_Type in (9, 10, 11, 12)
			and a.Etc4 is null
	`

	groupRows, err := db.Query(groupQuery)
	if err != nil {
		errlog.Println("Rcs SMTNT - 조회 중 오류 발생")
		errlog.Println(groupQuery)
		errcode := err.Error()

		if s.Index(errcode, "1146") > 0 {
			db.Exec("Create Table IF NOT EXISTS " + SMSTable + " like Msg_Tran")
			stdlog.Println("Rcs SMTNT -", SMSTable + " 생성 !!")

		}

		isProc = false
		return
	}
	defer groupRows.Close()

	if isProc {

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

		for groupRows.Next() {
			var mst_id sql.NullString
			var mem_id sql.NullString
			var mem_userid sql.NullString
			var mst_sent_voucher sql.NullString

			groupRows.Scan(&mst_id, &mem_id, &mem_userid, &mst_sent_voucher)

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

			var scnt int = 0
			var ecnt int = 0

			var ressql = `
				select sql_no_cache
					a.Msg_Id,
					a.Etc1 as msgid,
					a.Phone_No as phnstr,
					a.Callback_No as sms_sender,
					a.Message,
					a.Etc3 as remark4,
					'00000000000000' as reserve_dt,
					(select mem_userid from cb_member cm where cm.mem_id = b.mst_mem_id) as userid,
					a.Result,
					a.Msg_Type,
					(SELECT mi.origin1_path FROM cb_mms_images mi where b.mst_mms_content = mi.mms_id and length(b.mst_mms_content ) > 5 ) as mms_file1,
					(SELECT mi.origin2_path FROM cb_mms_images mi where b.mst_mms_content = mi.mms_id and length(b.mst_mms_content ) > 5 ) as mms_file2,
					(SELECT mi.origin3_path FROM cb_mms_images mi where b.mst_mms_content = mi.mms_id and length(b.mst_mms_content ) > 5 ) as mms_file3,
					b.mst_sent_voucher,
					b.mst_mem_id as send_mem_id,
					b.mst_type2,
					b.mst_type3,
					b.mst_lms_content
				from 
					` + SMSTable + ` a
				inner join
					cb_wt_msg_sent b ON a.Etc3 = b.mst_id
				where
					a.Status = '3'
					and a.Msg_Type in (9, 10, 11, 12)
					and a.Etc4 is null
					and a.Etc3 = ?
			`

			resrows, err := db.Query(ressql, mst_id.String)
			if err != nil {
				stdlog.Println("Rcs SMTNT - 결과 처리 Select 오류 ", err)
			} else {
				defer resrows.Close()

				var smtntMsgId, msgid, phnstr, sms_sender, body, remark4, reserve_dt, userid, resCode, msgType, mms_file1, mms_file2, mms_file3, mst_sent_voucher, send_mem_id, mst_type2, mst_type3, mst_lms_content sql.NullString

				amtsStrs := []string{}
				amtsValues := []interface{}{}

				var amtinsstr = ""

				amtsStrs = nil
				amtsValues = nil

				upmsgids := []interface{}{}

				for resrows.Next() {

					var amount float64
					var memo string
					var payback float64
					var admin_amt float64

					resrows.Scan(&smtntMsgId, &msgid, &phnstr, &sms_sender, &body, &remark4, &reserve_dt, &userid, &resCode, &msgType, &mms_file1, &mms_file2, &mms_file3, &mst_sent_voucher, &send_mem_id, &mst_type2, &mst_type3, &mst_lms_content)

					cprice := baseprice.GetPrice(db, send_mem_id.String, stdlog)

					amtinsstr = "insert into cb_amt_" + userid.String + "(amt_datetime," +
						"amt_kind," +
						"amt_amount," +
						"amt_memo," +
						"amt_reason," +
						"amt_payback," +
						"amt_admin)" +
						" values %s"

					switch resCode.String {
					case "0":
						db.Exec("update cb_msg_"+userid.String+" set CODE = 'RCS', MESSAGE_TYPE='rc', MESSAGE = ?, RESULT = ? where remark4=? and msgid = ?", "RCS 성공", "Y", remark4.String, msgid.String)
						scnt++
						break
					default:
						amtsStrs = append(amtsStrs, "(now(),?,?,?,?,?,?)")
						amtsValues = append(amtsValues, "3")
						if s.EqualFold(mst_sent_voucher.String, "V") {
							switch msgType.String {
							case "9":
								amount = cprice.V_price_rcs_sms.Float64
								payback = cprice.V_price_rcs_sms.Float64 - cprice.P_price_rcs_sms.Float64
								admin_amt = cprice.B_price_rcs_sms.Float64
								memo = "RCS SMS 발송실패 환불,바우처"
							case "10":
								amount = cprice.V_price_rcs.Float64
								payback = cprice.V_price_rcs.Float64 - cprice.P_price_rcs.Float64
								admin_amt = cprice.B_price_rcs.Float64
								memo = "RCS LMS 발송실패 환불,바우처"
							case "11":
								amount = cprice.V_price_rcs_mms.Float64
								payback = cprice.V_price_rcs_mms.Float64 - cprice.P_price_rcs_mms.Float64
								admin_amt = cprice.B_price_rcs_mms.Float64
								memo = "RCS MMS 발송실패 환불,바우처"
							case "12":
								amount = cprice.V_price_rcs_tem.Float64
								payback = cprice.V_price_rcs_tem.Float64 - cprice.P_price_rcs_tem.Float64
								admin_amt = cprice.B_price_rcs_tem.Float64
								memo = "RCS TMPL 발송실패 환불,바우처"
							}

						} else {
							switch msgType.String {
							case "9":
								amount = cprice.C_price_rcs_sms.Float64
								payback = cprice.C_price_rcs_sms.Float64 - cprice.P_price_rcs_sms.Float64
								admin_amt = cprice.B_price_rcs_sms.Float64
								if s.EqualFold(mst_sent_voucher.String, "B") {
									memo = "RCS SMS 발송실패 환불,보너스"
								} else {
									memo = "RCS SMS 발송실패 환불"
								}
							case "10":
								amount = cprice.C_price_rcs.Float64
								payback = cprice.C_price_rcs.Float64 - cprice.P_price_rcs.Float64
								admin_amt = cprice.B_price_rcs.Float64
								if s.EqualFold(mst_sent_voucher.String, "B") {
									memo = "RCS LMS 발송실패 환불,보너스"
								} else {
									memo = "RCS LMS 발송실패 환불"
								}
							case "11":
								amount = cprice.C_price_rcs_mms.Float64
								payback = cprice.C_price_rcs_mms.Float64 - cprice.P_price_rcs_mms.Float64
								admin_amt = cprice.B_price_rcs_mms.Float64
								if s.EqualFold(mst_sent_voucher.String, "B") {
									memo = "RCS MMS 발송실패 환불,보너스"
								} else {
									memo = "RCS MMS 발송실패 환불"
								}
							case "12":
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
						if s.Contains(mst_type3.String, "wc") {

							stdlog.Println("Rcs SMTNT - 발송 실패 -> WEB(C) 발송 처리 ", mst_type3.String, msgid.String)

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
						} else if s.Contains(mst_type3.String, "wa") {

							stdlog.Println("Rcs SMTNT - 발송 실패 -> WEB(A) 발송 처리 ", mst_type3.String, msgid.String)

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

							} else if s.EqualFold(mst_type3.String, "wa") {

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
						} else if s.Contains(mst_type3.String, "wb") {

							stdlog.Println("Rcs SMTNT - 발송 실패 -> WEB(B) 발송 처리 ", mst_type3.String, msgid.String)

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

								admin_amt = cprice.B_price_smt_sms.Float64
								if s.EqualFold(mst_sent_voucher.String, "V") {
									amount = cprice.V_price_smt_sms.Float64
									payback = cprice.V_price_smt_sms.Float64 - cprice.P_price_smt_sms.Float64
									memo = "웹(B) SMS,바우처"
								} else {
									amount = cprice.C_price_smt_sms.Float64
									payback = cprice.C_price_smt_sms.Float64 - cprice.P_price_smt_sms.Float64
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
							} else if s.EqualFold(mst_type3.String, "wb") {
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

									admin_amt = cprice.B_price_smt.Float64
									if s.EqualFold(mst_sent_voucher.String, "V") {
										amount = cprice.V_price_smt.Float64
										payback = cprice.V_price_smt.Float64 - cprice.P_price_smt.Float64
										memo = "웹(B) LMS,바우처"
									} else {
										amount = cprice.C_price_smt.Float64
										payback = cprice.C_price_smt.Float64 - cprice.P_price_smt.Float64
										if s.EqualFold(mst_sent_voucher.String, "B") {
											memo = "웹(B) LMS,보너스"
										} else {
											memo = "웹(B) LMS"
										}
									}
								} else {

									admin_amt = cprice.B_price_smt_mms.Float64
									if s.EqualFold(mst_sent_voucher.String, "V") {
										amount = cprice.V_price_smt_mms.Float64
										payback = cprice.V_price_smt_mms.Float64 - cprice.P_price_smt_mms.Float64
										memo = "웹(B) MMS,바우처"
									} else {
										amount = cprice.C_price_smt_mms.Float64
										payback = cprice.C_price_smt_mms.Float64 - cprice.P_price_smt_mms.Float64
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

						} else if s.Contains(mst_type3.String, "wd") {
							stdlog.Println("Rcs SMTNT - 발송 실패 -> WEB(D) 발송 처리 ", mst_type3.String, msgid.String)

							db.Exec("update cb_msg_"+userid.String+" set CODE = 'TNT', MESSAGE_TYPE='tn' where remark4=? and msgid = ?", remark4.String, msgid.String)

							smtntTime := time.Now().Format("2006-01-02 15:04:05")
							if s.EqualFold(mst_type3.String, "wds") {
								tntsmsStrs = append(tntsmsStrs, "(?,?,?,?,?,?,?,?,?,?)")
								tntsmsValues = append(tntsmsValues, phnstr) // Phone_No 1
								tntsmsValues = append(tntsmsValues, sms_sender) // Callback_No 2
								tntsmsValues = append(tntsmsValues, "4") // Msg_Type 3
								tntsmsValues = append(tntsmsValues, smtntTime) // Send_Time 4
								tntsmsValues = append(tntsmsValues, smtntTime) // Save_Time 5
								tntsmsValues = append(tntsmsValues, mst_lms_content.String) // Message 6
								tntsmsValues = append(tntsmsValues, config.Conf.KISACODE) // Reseller_Code 7

								tntsmsValues = append(tntsmsValues, msgid.String) // Etc1 8
								tntsmsValues = append(tntsmsValues, userid.String) // Etc2 9
								tntsmsValues = append(tntsmsValues, remark4.String) // Etc3 10

								admin_amt = cprice.B_price_smt_sms.Float64
								if s.EqualFold(mst_sent_voucher.String, "V") {
									amount = cprice.V_price_smt_sms.Float64
									payback = cprice.V_price_smt_sms.Float64 - cprice.P_price_smt_sms.Float64
									memo = "웹(D) SMS,바우처"
								} else {
									amount = cprice.C_price_smt_sms.Float64
									payback = cprice.C_price_smt_sms.Float64 - cprice.P_price_smt_sms.Float64
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
							} else if s.EqualFold(mst_type3.String, "wd") {
								fileCnt  := 0
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
								tntmmsValues = append(tntmmsValues, phnstr) // Phone_No 1
								tntmmsValues = append(tntmmsValues, sms_sender) // Callback_No 2
								tntmmsValues = append(tntmmsValues, "6") // Msg_Type 3
								tntmmsValues = append(tntmmsValues, smtntTime) // Send_Time 4
								tntmmsValues = append(tntmmsValues, smtntTime) // Save_Time 5
								tntmmsValues = append(tntmmsValues, rcsBody.Title) // Subject 6
								tntmmsValues = append(tntmmsValues, mst_lms_content.String) // Message 7
								tntmmsValues = append(tntmmsValues, fileCnt) // File_Count 8
								tntmmsValues = append(tntmmsValues, fileType1) // File_Type1 9 
								tntmmsValues = append(tntmmsValues, fileType2) // File_Type2 10
								tntmmsValues = append(tntmmsValues, fileType3) // File_Type3 11
								tntmmsValues = append(tntmmsValues, mms_file1) // File_Name1 12
								tntmmsValues = append(tntmmsValues, mms_file2) // File_Name2 13
								tntmmsValues = append(tntmmsValues, mms_file3) // File_Name3 14
								tntmmsValues = append(tntmmsValues, config.Conf.KISACODE) // Reseller_Code 15

								tntmmsValues = append(tntmmsValues, msgid.String) // Etc1 16
								tntmmsValues = append(tntmmsValues, userid.String) // Etc2 17
								tntmmsValues = append(tntmmsValues, remark4.String) // Etc3 18

								if len(mms_file1.String) <= 0 {
									admin_amt = cprice.B_price_smt.Float64
									if s.EqualFold(mst_sent_voucher.String, "V") {
										amount = cprice.V_price_smt.Float64
										payback = cprice.V_price_smt.Float64 - cprice.P_price_smt.Float64
										memo = "웹(D) LMS,바우처"
									} else {
										amount = cprice.C_price_smt.Float64
										payback = cprice.C_price_smt.Float64 - cprice.P_price_smt.Float64
										if s.EqualFold(mst_sent_voucher.String, "B") {
											memo = "웹(D) LMS,보너스"
										} else {
											memo = "웹(D) LMS"
										}
									}
								} else {
									admin_amt = cprice.B_price_smt_mms.Float64
									if s.EqualFold(mst_sent_voucher.String, "V") {
										amount = cprice.V_price_smt_mms.Float64
										payback = cprice.V_price_smt_mms.Float64 - cprice.P_price_smt_mms.Float64
										memo = "웹(D) MMS,바우처"
									} else {
										amount = cprice.C_price_smt_mms.Float64
										payback = cprice.C_price_smt_mms.Float64 - cprice.P_price_smt_mms.Float64
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
						} else {
							ecnt++
							db.Exec("update cb_msg_"+userid.String+" set CODE = 'RCS', MESSAGE_TYPE='rc', MESSAGE = ?, RESULT = ? where remark4=? and msgid = ?", resCode.String, "N", remark4.String, msgid.String)
						}
					}

					upmsgids = append(upmsgids, smtntMsgId.String)
				}

				if len(ossmsStrs) > 0 {
					stmt := fmt.Sprintf("insert into OShotSMS(Sender,Receiver,Msg,URL,ReserveDT,TimeoutDT,SendResult,mst_id,cb_msg_id ) values %s", s.Join(ossmsStrs, ","))
					_, err := db.Exec(stmt, ossmsValues...)

					if err != nil {
						stdlog.Println("Rcs SMTNT - 스마트미 SMS Table Insert 처리 중 오류 발생 " + err.Error())
					}

					ossmsStrs = nil
					ossmsValues = nil
				}

				if len(osmmsStrs) > 0 {
					stmt := fmt.Sprintf("insert into OShotMMS(MsgGroupID,Sender,Receiver,Subject,Msg,ReserveDT,TimeoutDT,SendResult,File_Path1,File_Path2,File_Path3,mst_id,cb_msg_id ) values %s", s.Join(osmmsStrs, ","))
					_, err := db.Exec(stmt, osmmsValues...)

					if err != nil {
						stdlog.Println("Rcs SMTNT - 스마트미 LMS Table Insert 처리 중 오류 발생 " + err.Error())
					}

					osmmsStrs = nil
					osmmsValues = nil
				}

				if len(lgusmsStrs) >= 1000 {
					stmt := fmt.Sprintf("insert into LG_SC_TRAN(TR_SENDDATE,TR_PHONE,TR_CALLBACK, TR_MSG, TR_ETC1, TR_ETC2, TR_ETC3, TR_KISAORIGCODE) values %s", s.Join(lgusmsStrs, ","))
					_, err := db.Exec(stmt, lgusmsValues...)

					if err != nil {
						stdlog.Println("Rcs SMTNT - LGU SMS Table Insert 처리 중 오류 발생 " + err.Error())
					}

					lgusmsStrs = nil
					lgusmsValues = nil
				}

				if len(lgummsStrs) >= 1000 {
					stmt := fmt.Sprintf("insert into LG_MMS_MSG(SUBJECT, PHONE, CALLBACK, REQDATE, MSG, FILE_CNT, FILE_PATH1, FILE_PATH2, FILE_PATH3, ETC1, ETC2, ETC3, KISA_ORIGCODE) values %s", s.Join(lgummsStrs, ","))
					_, err := db.Exec(stmt, lgummsValues...)

					if err != nil {
						stdlog.Println("Rcs SMTNT - LGU LMS Table Insert 처리 중 오류 발생 " + err.Error())
					}

					lgummsStrs = nil
					lgummsValues = nil
				}

				if len(nnsmsStrs) > 0 {
					stmt := fmt.Sprintf("insert into SMS_MSG(TR_CALLBACK,TR_PHONE,TR_MSG,TR_SENDDATE,TR_SENDSTAT,TR_MSGTYPE,TR_ETC9,TR_ETC10,TR_IDENTIFICATION_CODE,TR_ETC8) values %s", s.Join(nnsmsStrs, ","))
					_, err := db.Exec(stmt, nnsmsValues...)

					if err != nil {
						stdlog.Println("Rcs SMTNT - 나노 SMS Table Insert 처리 중 오류 발생 " + err.Error())
					}

					nnsmsStrs = nil
					nnsmsValues = nil
				}

				if len(nnmmsStrs) > 0 {
					stmt := fmt.Sprintf("insert into MMS_MSG(CALLBACK,PHONE,SUBJECT,MSG,REQDATE,STATUS,FILE_CNT,FILE_PATH1,FILE_PATH2,FILE_PATH3,ETC9,ETC10,IDENTIFICATION_CODE,ETC8) values %s", s.Join(nnmmsStrs, ","))
					_, err := db.Exec(stmt, nnmmsValues...)

					if err != nil {
						stdlog.Println("Rcs SMTNT - 나노 LMS Table Insert 처리 중 오류 발생 " + err.Error())
					}

					nnmmsStrs = nil
					nnmmsValues = nil
				}

				if len(tntsmsStrs) > 0 {
					stmt := fmt.Sprintf("insert into Msg_Tran(Phone_No,Callback_No,Msg_Type,Send_Time,Save_Time,Message,Reseller_Code,Etc1,Etc2,Etc3) values %s", s.Join(tntsmsStrs, ","))
					_, err := db.Exec(stmt, tntsmsValues...)

					if err != nil {
						stdlog.Println("Rcs SMTNT - SMTNT SMS Table Insert 처리 중 오류 발생 " + err.Error())
					}

					tntsmsStrs = nil
					tntsmsValues = nil
				}

				if len(tntmmsStrs) > 0 {
					stmt := fmt.Sprintf("insert into Msg_Tran(Phone_No,Callback_No,Msg_Type,Send_Time,Save_Time,Subject,Message,File_Count,File_Type1,File_Type2,File_Type3,File_Name1,File_Name2,File_Name3,Reseller_Code,Etc1,Etc2,Etc3) values %s", s.Join(tntmmsStrs, ","))
					_, err := db.Exec(stmt, tntmmsValues...)

					if err != nil {
						stdlog.Println("Rcs SMTNT - SMTNT LMS Table Insert 처리 중 오류 발생 " + err.Error())
					}

					tntmmsStrs = nil
					tntmmsValues = nil
				}

				if len(amtsStrs) > 0 {
					stmt := fmt.Sprintf(amtinsstr, s.Join(amtsStrs, ","))
					_, err := db.Exec(stmt, amtsValues...)

					if err != nil {
						stdlog.Println("Rcs SMTNT - AMT Table Insert 처리 중 오류 발생 " + err.Error())
					}

					amtsStrs = nil
					amtsValues = nil
				}

				if len(upmsgids) > 0 {
					var commastr = "update " + SMSTable + " set Etc4 = '1' where Msg_Id in ("

					for i := 1; i < len(upmsgids); i++ {
						commastr = commastr + "?,"
					}

					commastr = commastr + "?)"

					_, err1 := db.Exec(commastr, upmsgids...)

					if err1 != nil {
						errlog.Println("Rcs SMTNT -", SMSTable + "Table Update 처리 중 오류 발생 ")
					}
				}

				db.Exec("update cb_wt_msg_sent set mst_rcs = ifnull(mst_rcs,0) + ?,mst_err_rcs = ifnull(mst_err_rcs,0) + ?, mst_wait = mst_wait - ?  where mst_id=?", scnt, ecnt, (ecnt + scnt), mst_id.String)

				stdlog.Println("Rcs SMTNT - 처리 끝 : (", mst_id.String, " ) 성공 : ", scnt, " / 실패 : ", ecnt)
			}
		}
	}
}