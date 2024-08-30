package webcmms

import (
	"webagent/src/baseprice"
	"webagent/src/config"
	"database/sql"
	"webagent/src/databasepool"
	"fmt"

	//	"log"
	s "strings"
	"sync"
	"time"
	"context"

	_ "github.com/go-sql-driver/mysql"
)

func Process(ctx context.Context) {
	var wg sync.WaitGroup
	for {
		select {
		case <- ctx.Done():
			time.Sleep(20 * time.Second)
			config.Stdlog.Println("webcmms 정상적으로 종료되었습니다.")
			return
		default:
			wg.Add(1)
			go mmsProcess(&wg)
			wg.Wait()
		}
	}
}

func mmsProcess(wg *sync.WaitGroup) {
	defer wg.Done()
	defer func() {
		if r := recover(); r != nil {
			config.Stdlog.Println("webcmms panic 발생 원인 : ", r)
			if err, ok := r.(error); ok {
				if s.Contains(err.Error(), "connection refused") {
					for {
						config.Stdlog.Println("webcmms send ping to DB")
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
	var monthStr = fmt.Sprintf("%d%02d", t.Year(), t.Month())

	var MMSTable = "OShotMMS_" + monthStr
	var msgcnt sql.NullString
	//발송 6시간 지난 메세지는 응답과 상관 없이 성공 처리 함.

	//lms 성공 처리
	err1 := db.QueryRow("SELECT count(1) as cnt from OShotMMS WHERE SendResult=1 and date_add(insertdt, interval 6 HOUR) < now() and mst_id is not null").Scan(&msgcnt)
	if err1 != nil {
	   errlog.Println("OShotMMS Table 조회 중 중 오류 발생", err1)
	   panic(err1)
	} else {		
		if !s.EqualFold(msgcnt.String, "0") {
			db.Exec("UPDATE OShotMMS SET SendDT=now(), SendResult='6', Telecom='000' WHERE SendResult=1 and date_add(insertdt, interval 6 HOUR) < now() and mst_id is not null")
			db.Exec("insert into " + MMSTable + " SELECT * FROM OShotMMS WHERE SendResult>1 AND SendDT is not null and telecom = '000'")
			db.Exec("delete FROM OShotMMS WHERE SendResult>1 AND SendDT is not null and telecom = '000'")
		}
    }
	var groupQuery = "select distinct a.mst_id, b.mst_mem_id as mem_id,(select mem_userid from cb_member cm where cm.mem_id = b.mst_mem_id) AS mem_userid, b.mst_sent_voucher  from " + MMSTable + " a inner join cb_wt_msg_sent b ON a.mst_id = b.mst_id where a.proc_flag = 'Y'"

	groupRows, err := db.Query(groupQuery)
	if err != nil {
		errlog.Println("스마트미 MMS 조회 중 오류 발생")
		errlog.Println(groupQuery)
		errcode := err.Error()

		if s.Index(errcode, "1146") > 0 {
			db.Exec("Create Table IF NOT EXISTS " + MMSTable + " like OShotMMS")
			stdlog.Println(MMSTable + " 생성 !!")
		} else {
			// errlog.Fatal(groupQuery)
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

			var mmsQuery = "SELECT SQL_NO_CACHE " +
				"       a.MsgID " +
				"      ,a.SendResult" +
				"      ,a.Receiver AS PHN" +
				"      ,b.mst_id AS REMARK4" +
				"      ,(select mem_userid from cb_member cm where cm.mem_id = b.mst_mem_id) AS mem_userid " +
				"      ,b.mst_mem_id AS mem_id" +
				"      ,a.cb_msg_id " +
				"      ,a.File_Path1 as mms1 " +
				"      ,a.File_Path2 as mms2 " +
				"      ,a.File_Path3 as mm3 " +
				" from " + MMSTable + " a INNER JOIN " +
				"        cb_wt_msg_sent b ON a.mst_id = b.mst_id " +
				"WHERE a.proc_flag = 'Y' " +
				" and a.mst_id = ?"

			rows, err := db.Query(mmsQuery, mst_id.String)
			if err != nil {
				errlog.Println("스마트미 MMS 조회 중 오류 발생")
				errlog.Println(mmsQuery)
				// errlog.Fatal(mmsQuery)
			}
			defer rows.Close()

			tx, err := db.Begin()
			if err != nil {
				errlog.Println(" 트랜잭션 시작 중 오류 발생")
				// errlog.Fatal(err)
			}

			var amtinsstr = "insert into cb_amt_" + mem_userid.String + "(amt_datetime," +
				"amt_kind," +
				"amt_amount," +
				"amt_memo," +
				"amt_reason," +
				"amt_payback," +
				"amt_admin)" +
				" values %s"

			var mmserrcnt = 0
			var mmscnt = 0

			amtsStrs = nil
			amtsValues = nil

			upmsgids := []interface{}{}

			var msgid, sendresult, phn, sent_key, userid, cb_msg_id, mms1, mms2, mms3 sql.NullString
			var startNow = time.Now()
			var startTime = fmt.Sprintf("%02d:%02d:%02d", startNow.Hour(), startNow.Minute(), startNow.Second())
			for rows.Next() {
				var message = ""
				var result = ""
				rows.Scan(&msgid, &sendresult, &phn, &sent_key, &userid, &mem_id, &cb_msg_id, &mms1, &mms2, &mms3)
				mmscnt++
				if !s.EqualFold(sendresult.String, "6") && conf.REFUND {
					mmserrcnt++
					//upcbmsgids = append(upcbmsgids, cb_msg_id.String)
					message = sendresult.String
					result = "N"

					amtsStrs = append(amtsStrs, "(now(),?,?,?,?,?,?)")
					amtsValues = append(amtsValues, "3")
					if len(mms1.String) > 0 {
						if s.EqualFold(mst_sent_voucher.String, "V") {
							amtsValues = append(amtsValues, cprice.V_price_smt_mms.Float64)
							amtsValues = append(amtsValues, "웹(C) 발송실패 환불,바우처")
							amtsValues = append(amtsValues, cb_msg_id.String+phn.String)
							amtsValues = append(amtsValues, ((cprice.V_price_smt_mms.Float64 - cprice.P_price_smt_mms.Float64) * -1))
							amtsValues = append(amtsValues, cprice.B_price_smt_mms.Float64*-1)
						} else {
							amtsValues = append(amtsValues, cprice.C_price_smt_mms.Float64)
							if s.EqualFold(mst_sent_voucher.String, "B") {
								amtsValues = append(amtsValues, "웹(C) 발송실패 환불,보너스")
							} else {
								amtsValues = append(amtsValues, "웹(C) 발송실패 환불")
							}										
							amtsValues = append(amtsValues, cb_msg_id.String+phn.String)
							amtsValues = append(amtsValues, ((cprice.C_price_smt_mms.Float64 - cprice.P_price_smt_mms.Float64) * -1))
							amtsValues = append(amtsValues, cprice.B_price_smt_mms.Float64*-1)
						}

					} else {
						if s.EqualFold(mst_sent_voucher.String, "V") {
							amtsValues = append(amtsValues, cprice.V_price_smt.Float64)
							amtsValues = append(amtsValues, "웹(C) 발송실패 환불,바우처")
							amtsValues = append(amtsValues, cb_msg_id.String+phn.String)
							amtsValues = append(amtsValues, ((cprice.V_price_smt.Float64 - cprice.P_price_smt.Float64) * -1))
							amtsValues = append(amtsValues, cprice.B_price_smt.Float64*-1)
						} else {
							amtsValues = append(amtsValues, cprice.C_price_smt.Float64)
							if s.EqualFold(mst_sent_voucher.String, "B") {
								amtsValues = append(amtsValues, "웹(C) 발송실패 환불,보너스")
							} else {
								amtsValues = append(amtsValues, "웹(C) 발송실패 환불")
							}										
							amtsValues = append(amtsValues, cb_msg_id.String+phn.String)
							amtsValues = append(amtsValues, ((cprice.C_price_smt.Float64 - cprice.P_price_smt.Float64) * -1))
							amtsValues = append(amtsValues, cprice.B_price_smt.Float64*-1)
						}
					}
				} else {
					message = "웹(C) 성공"
					result = "Y"
				}

				tx.Exec("update cb_msg_"+userid.String+" set MESSAGE_TYPE='sm', MESSAGE = ?, RESULT = ? where remark4=? and msgid = ?", message, result, sent_key.String, cb_msg_id.String)

				upmsgids = append(upmsgids, msgid.String)

				if len(upmsgids) >= 1000 {

					var commastr = "update " + MMSTable + " set proc_flag = 'N' where msgid in ("

					for i := 1; i < len(upmsgids); i++ {
						commastr = commastr + "?,"
					}

					commastr = commastr + "?)"

					_, err1 := tx.Exec(commastr, upmsgids...)

					if err1 != nil {
						errlog.Println(MMSTable + " Table Update 처리 중 오류 발생 ")
					}

					upmsgids = nil
				}

				if len(amtsStrs) >= 1000 {
					stmt := fmt.Sprintf(amtinsstr, s.Join(amtsStrs, ","))
					_, err := tx.Exec(stmt, amtsValues...)

					if err != nil {
						errlog.Println("AMT Table Insert 처리 중 오류 발생 " + err.Error())
					}

					amtsStrs = nil
					amtsValues = nil
				}

			}

			if len(upmsgids) > 0 {

				var commastr = "update " + MMSTable + " set proc_flag = 'N' where msgid in ("

				for i := 1; i < len(upmsgids); i++ {
					commastr = commastr + "?,"
				}

				commastr = commastr + "?)"

				_, err1 := tx.Exec(commastr, upmsgids...)

				if err1 != nil {
					errlog.Println(MMSTable + " Table Update 처리 중 오류 발생 ")
				}
			}

			if len(amtsStrs) > 0 {
				stmt := fmt.Sprintf(amtinsstr, s.Join(amtsStrs, ","))
				_, err := tx.Exec(stmt, amtsValues...)

				if err != nil {
					errlog.Println("AMT Table Insert 처리 중 오류 발생 " + err.Error())
				}

			}

			tx.Exec("update cb_wt_msg_sent set mst_err_smt = ifnull(mst_err_smt,0) + ?, mst_smt = ifnull(mst_smt,0) + ?, mst_wait = mst_wait - ?  where mst_id=?", mmserrcnt, ( mmscnt-mmserrcnt), mmscnt, sent_key.String)
			tx.Commit()
			stdlog.Printf(" ( %s ) WEB(C) MMS 처리 - %s : %d \n", startTime, sent_key.String, mmscnt)

		}
	}
}
