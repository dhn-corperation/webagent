package webemsg

import (
	"fmt"
	"sync"
	"time"
	"context"
	s "strings"
	"database/sql"

	"webagent/src/config"
	"webagent/src/baseprice"
	"webagent/src/databasepool"

	_ "github.com/go-sql-driver/mysql"
)

func Process(ctx context.Context) {
	config.Stdlog.Println("JJ MSG (WEB E) - 결과 처리 프로세스 시작")
	var wg sync.WaitGroup
	for {
		select {
		case <- ctx.Done():
			config.Stdlog.Println("JJ MSG (WEB E) - process가 15초 후에 종료")
		    time.Sleep(15 * time.Second)
		    config.Stdlog.Println("JJ MSG (WEB E) - process 종료 완료")
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
			config.Stdlog.Println("JJ MSG (WEB E) - panic 발생 원인 : ", r)
			if err, ok := r.(error); ok {
				if s.Contains(err.Error(), "connection refused") {
					for {
						config.Stdlog.Println("JJ MSG (WEB E) - send ping to DB")
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
	var conf = config.Conf
	var stdlog = config.Stdlog
	var errlog = config.Stdlog

	var cprice baseprice.BasePrice

	amtsStrs := []string{}
	amtsValues := []interface{}{}

	var isProc = true
	var t = time.Now()

	if pastFlag {
		t = time.Now().Add(time.Hour * -96)
	}
	
	var monthStr = fmt.Sprintf("%d%02d", t.Year(), t.Month())

	var SMSTable = "MTMSG_LOG_" + monthStr
	var msgcnt sql.NullString

	//발송 6시간 지난 메세지는 응답과 상관 없이 성공 처리 함.
	// sms 성공 처리
	err1 := db.QueryRow("SELECT count(1) as cnt from MTMSG_DATA WHERE MSG_STATE='5' and MSG_TYPE in ('SMS', 'LMS', 'MMS') and date_add(INPUT_DATE, interval 6 HOUR) < now() and DHN_ETC3 is not null and DHN_ETC2 not like 'khug%'").Scan(&msgcnt)
	if err1 != nil {
	   errlog.Println("JJ MSG (WEB E) - 조회 중 오류 발생", err1)
	   panic(err1)
	} else {		
		if !s.EqualFold(msgcnt.String, "0") {	
			db.Exec("UPDATE MTMSG_DATA SET MSG_STATE='6', RSLT_CODE='1000', RSLT_NET='ETC', RSLT_DATE=now(), REPORT_DATE=now() WHERE MSG_STATE='5' and MSG_TYPE in ('SMS', 'LMS', 'MMS') and date_add(INPUT_DATE, interval 6 HOUR) < now() and DHN_ETC3 is not null and DHN_ETC2 not like 'khug%'")
		}
	}

	var tickCnt sql.NullInt64
	var tickSql = `
		select
			count(1) as cnt
		from
			` + SMSTable + ` a
		inner join
			cb_wt_msg_sent b ON a.DHN_ETC3 = b.mst_id
		where
			a.MSG_STATE = '6'
			and a.MSG_TYPE in ('SMS', 'LMS', 'MMS')
			and a.DHN_ETC4 is null
		limit 1`

	cnterr := databasepool.DB.QueryRow(tickSql).Scan(&tickCnt)

	if cnterr != nil && cnterr != sql.ErrNoRows {
		errlog.Println("JJ MSG (WEB E) -", SMSTable, "Table - select error : " + cnterr.Error())
		if s.Index(cnterr.Error(), "1146") > 0 {
			db.Exec("Create Table IF NOT EXISTS " + SMSTable + " like MTMSG_DATA")
			errlog.Println("JJ MSG (WEB E) -", SMSTable + " 생성 !!")
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
			a.DHN_ETC3,
			b.mst_mem_id as mem_id,
			(select mem_userid from cb_member cm where cm.mem_id = b.mst_mem_id) AS mem_userid,
			b.mst_sent_voucher
		from
			` + SMSTable + ` a
		inner join
			cb_wt_msg_sent b ON a.DHN_ETC3 = b.mst_id
		where
			a.MSG_STATE = '6'
			and a.MSG_TYPE in ('SMS', 'LMS', 'MMS')
			and a.DHN_ETC4 is null
	`

	groupRows, err := db.Query(groupQuery)
	if err != nil {
		errlog.Println("JJ MSG (WEB E) - 조회 중 오류 발생")
		errlog.Println(groupQuery)
		errcode := err.Error()

		if s.Index(errcode, "1146") > 0 {
			db.Exec("Create Table IF NOT EXISTS " + SMSTable + " like Msg_Tran")
			stdlog.Println("JJ MSG (WEB E) -", SMSTable + " 생성 !!")

		}

		isProc = false
		return
	}
	defer groupRows.Close()

	if isProc {

		for groupRows.Next() {
			var mst_id sql.NullString
			var mem_id sql.NullString
			var mem_userid sql.NullString
			var mst_sent_voucher sql.NullString

			groupRows.Scan(&mst_id, &mem_id, &mem_userid, &mst_sent_voucher)

			cprice = baseprice.GetPrice(db, mem_id.String, errlog)

			var smsQuery = `
				select sql_no_cache
					a.MSG_KEY,
					a.RSLT_CODE,
					a.CALL_TO,
					a.MSG_TYPE,
					b.mst_id AS REMARK4,
					(select mem_userid from cb_member cm where cm.mem_id = b.mst_mem_id) AS mem_userid,
					b.mst_mem_id AS mem_id,
					a.DHN_ETC1 as cb_msg_id
				from 
					` + SMSTable + ` a
				inner join
					cb_wt_msg_sent b ON a.DHN_ETC3 = b.mst_id
				where
					a.MSG_STATE = '6'
					and a.MSG_TYPE in ('SMS', 'LMS', 'MMS')
					and a.DHN_ETC3 = ?
					and a.DHN_ETC4 is null
			`

			rows, err := db.Query(smsQuery, mst_id.String)
			if err != nil {
				errlog.Println("JJ MSG (WEB E) - 조회 중 오류 발생")
				errlog.Println(smsQuery)
			}
			defer rows.Close()

			tx, err := db.Begin()
			if err != nil {
				errlog.Println("JJ MSG (WEB E) - 트랜잭션 시작 중 오류 발생")
			}

			var amtinsstr = "insert into cb_amt_" + mem_userid.String + "(amt_datetime," +
				"amt_kind," +
				"amt_amount," +
				"amt_memo," +
				"amt_reason," +
				"amt_payback," +
				"amt_admin)" +
				" values %s"

			var msgerrcnt = 0
			var msgcnt = 0

			amtsStrs = nil
			amtsValues = nil

			upmsgids := []interface{}{}

			var jjMsgKey, rsltCode, callTo, msgType, sent_key, userid, cb_msg_id sql.NullString
			var fileCount sql.NullInt16
			var startNow = time.Now()
			var startTime = fmt.Sprintf("%02d:%02d:%02d", startNow.Hour(), startNow.Minute(), startNow.Second())
			for rows.Next() {
				var message = ""
				var result = ""
				msgcnt++

				rows.Scan(&jjMsgKey, &rsltCode, &callTo, &msgType, &sent_key, &userid, &mem_id, &cb_msg_id)

				if rsltCode.String != "1000" && conf.REFUND {

					msgerrcnt++
					message = rsltCode.String
					result = "N"
					amtsStrs = append(amtsStrs, "(now(),?,?,?,?,?,?)")
					amtsValues = append(amtsValues, "3")

					if msgType.String == "4" {
						if s.EqualFold(mst_sent_voucher.String, "V") {
							amtsValues = append(amtsValues, cprice.V_price_smt_sms.Float64)
							amtsValues = append(amtsValues, "웹(E) 발송실패 환불,바우처")
							amtsValues = append(amtsValues, cb_msg_id.String+callTo.String)
							amtsValues = append(amtsValues, ((cprice.V_price_smt_sms.Float64 - cprice.P_price_smt_sms.Float64) * -1))
							amtsValues = append(amtsValues, cprice.B_price_smt_sms.Float64*-1)
						} else {
							amtsValues = append(amtsValues, cprice.C_price_smt_sms.Float64)
							if s.EqualFold(mst_sent_voucher.String, "B") {
								amtsValues = append(amtsValues, "웹(E) 발송실패 환불,보너스")
							} else {
								amtsValues = append(amtsValues, "웹(E) 발송실패 환불")
							}										
							amtsValues = append(amtsValues, cb_msg_id.String+callTo.String)
							amtsValues = append(amtsValues, ((cprice.C_price_smt_sms.Float64 - cprice.P_price_smt_sms.Float64) * -1))
							amtsValues = append(amtsValues, cprice.B_price_smt_sms.Float64*-1)
						}
					} else {
						if fileCount.Int16 > 0 {
							if s.EqualFold(mst_sent_voucher.String, "V") {
								amtsValues = append(amtsValues, cprice.V_price_smt_mms.Float64)
								amtsValues = append(amtsValues, "웹(E) 발송실패 환불,바우처")
								amtsValues = append(amtsValues, cb_msg_id.String+callTo.String)
								amtsValues = append(amtsValues, ((cprice.V_price_smt_mms.Float64 - cprice.P_price_smt_mms.Float64) * -1))
								amtsValues = append(amtsValues, cprice.B_price_smt_mms.Float64*-1)
							} else {
								amtsValues = append(amtsValues, cprice.C_price_smt_mms.Float64)
								if s.EqualFold(mst_sent_voucher.String, "B") {
									amtsValues = append(amtsValues, "웹(E) 발송실패 환불,보너스")
								} else {
									amtsValues = append(amtsValues, "웹(E) 발송실패 환불")
								}										
								amtsValues = append(amtsValues, cb_msg_id.String+callTo.String)
								amtsValues = append(amtsValues, ((cprice.C_price_smt_mms.Float64 - cprice.P_price_smt_mms.Float64) * -1))
								amtsValues = append(amtsValues, cprice.B_price_smt_mms.Float64*-1)
							}

						} else {
							if s.EqualFold(mst_sent_voucher.String, "V") {
								amtsValues = append(amtsValues, cprice.V_price_smt.Float64)
								amtsValues = append(amtsValues, "웹(E) 발송실패 환불,바우처")
								amtsValues = append(amtsValues, cb_msg_id.String+callTo.String)
								amtsValues = append(amtsValues, ((cprice.V_price_smt.Float64 - cprice.P_price_smt.Float64) * -1))
								amtsValues = append(amtsValues, cprice.B_price_smt.Float64*-1)
							} else {
								amtsValues = append(amtsValues, cprice.C_price_smt.Float64)
								if s.EqualFold(mst_sent_voucher.String, "B") {
									amtsValues = append(amtsValues, "웹(E) 발송실패 환불,보너스")
								} else {
									amtsValues = append(amtsValues, "웹(E) 발송실패 환불")
								}										
								amtsValues = append(amtsValues, cb_msg_id.String+callTo.String)
								amtsValues = append(amtsValues, ((cprice.C_price_smt.Float64 - cprice.P_price_smt.Float64) * -1))
								amtsValues = append(amtsValues, cprice.B_price_smt.Float64*-1)
							}
						}
					}
					
				} else {
					message = "웹(E) 성공"
					result = "Y"
				}

				tx.Exec("update cb_msg_"+userid.String+" set MESSAGE_TYPE='jj', MESSAGE = ?, RESULT = ? where remark4=? and msgid = ?", message, result, sent_key.String, cb_msg_id.String)

				upmsgids = append(upmsgids, jjMsgKey.String)

				if len(upmsgids) >= 1000 {

					var commastr = "update " + SMSTable + " set DHN_ETC4 = '1' where MSG_KEY in ("

					for i := 1; i < len(upmsgids); i++ {
						commastr = commastr + "?,"
					}

					commastr = commastr + "?)"

					_, err1 := tx.Exec(commastr, upmsgids...)

					if err1 != nil {
						errlog.Println("JJ MSG (WEB E) -", SMSTable + " Table Update 처리 중 오류 발생 ")
					}

					upmsgids = nil
				}

				if len(amtsStrs) >= 1000 {
					stmt := fmt.Sprintf(amtinsstr, s.Join(amtsStrs, ","))
					_, err := tx.Exec(stmt, amtsValues...)

					if err != nil {
						errlog.Println("JJ MSG (WEB E) - AMT Table Insert 처리 중 오류 발생 " + err.Error())
					}

					amtsStrs = nil
					amtsValues = nil
				}

			}

			if len(upmsgids) > 0 {

				var commastr = "update " + SMSTable + " set DHN_ETC4 = '1' where MSG_KEY in ("

				for i := 1; i < len(upmsgids); i++ {
					commastr = commastr + "?,"
				}

				commastr = commastr + "?)"

				_, err1 := tx.Exec(commastr, upmsgids...)

				if err1 != nil {
					errlog.Println("JJ MSG (WEB E) -", SMSTable + " Table Update 처리 중 오류 발생 ")
				}
			}

			if len(amtsStrs) > 0 {
				stmt := fmt.Sprintf(amtinsstr, s.Join(amtsStrs, ","))
				_, err := tx.Exec(stmt, amtsValues...)

				if err != nil {
					errlog.Println("JJ MSG (WEB E) - AMT Table Insert 처리 중 오류 발생 " + err.Error())
				}

			}

			tx.Exec("update cb_wt_msg_sent set mst_err_smt = ifnull(mst_err_smt,0) + ?, mst_smt = ifnull(mst_smt,0) + ?, mst_wait = mst_wait - ?  where mst_id=?", msgerrcnt, (msgcnt-msgerrcnt), msgcnt, sent_key.String)
			tx.Commit()
			stdlog.Printf("JJ MSG (WEB E) - ( %s ) WEB(E) MSG 처리 - %s : %d \n", startTime, sent_key.String, msgcnt)
		}
	}
}
