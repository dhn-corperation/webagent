package webbmms

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
	var wg sync.WaitGroup
	for {
		select {
		case <- ctx.Done():
			time.Sleep(20 * time.Second)
			config.Stdlog.Println("webbmms 정상적으로 종료되었습니다.")
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
			config.Stdlog.Println("webbmms panic 발생 원인 : ", r)
			if err, ok := r.(error); ok {
				if s.Contains(err.Error(), "connection refused") {
					for {
						config.Stdlog.Println("webbmms send ping to DB")
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

	var MMSTable = "LG_MMS_LOG_" + monthStr
	var msgcnt sql.NullString
	//발송 6시간 지난 메세지는 응답과 상관 없이 성공 처리 함.

	//lms 성공 처리
	err1 := db.QueryRow("SELECT count(1) as cnt from LG_MMS_MSG WHERE STATUS='2' and date_add(REQDATE, interval 6 HOUR) < now() and ETC3 is not null and ETC2 not like 'khug%'").Scan(&msgcnt)
	if err1 != nil {
	   errlog.Println("LG_MMS_MSG Table 조회 중 중 오류 발생", err1)
	   panic(err1)
	} else {		
		if msgcnt.String != "0" {
			db.Exec("UPDATE LG_MMS_MSG SET REPORTDATE=now(), STATUS='3', RSLT='1000' WHERE STATUS='2' and date_add(REQDATE, interval 6 HOUR) < now() and ETC3 is not null and ETC2 not like 'khug%'")
		}
    }
	var groupQuery = `
		select distinct 
			a.ETC3,
			b.mst_mem_id as mem_id,
			(select mem_userid from cb_member cm where cm.mem_id = b.mst_mem_id) AS mem_userid, 
			b.mst_sent_voucher
		from
			` + MMSTable + ` a 
		inner join
			cb_wt_msg_sent b ON a.ETC3 = b.mst_id
		where
			a.status = '3'
			and a.ETC4 is null`

	groupRows, err := db.Query(groupQuery)
	if err != nil {
		errlog.Println("LGU MMS 조회 중 오류 발생")
		errlog.Println(groupQuery)
		errcode := err.Error()

		if s.Index(errcode, "1146") > 0 {
			db.Exec("Create Table IF NOT EXISTS " + MMSTable + " like LG_MMS_MSG")
			stdlog.Println(MMSTable + " 생성 !!")
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

			var mmsQuery = `
				select sql_no_cache
					a.MSGKEY,
					a.RSLT,
					a.PHONE,
					b.mst_id as REMARK4,
					(select mem_userid from cb_member cm where cm.mem_id = b.mst_mem_id) AS mem_userid,
					b.mst_mem_id AS mem_id,
					a.ETC1 as cb_msg_id,
					a.FILE_PATH1 as mms1
				from
					` + MMSTable + ` a
				inner join
					cb_wt_msg_sent b ON a.ETC3 = b.mst_id
				where
					a.STATUS = '3'
					and a.ETC4 is null
					and a.ETC3 = ?
			`

			rows, err := db.Query(mmsQuery, mst_id.String)
			if err != nil {
				errlog.Println("LGU MMS 조회 중 오류 발생")
				errlog.Println(mmsQuery)
			}
			defer rows.Close()

			tx, err := db.Begin()
			if err != nil {
				errlog.Println("LGU MMS 트랜잭션 시작 중 오류 발생")
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

			var msgkey, rslt, phone, sent_key, userid, cb_msg_id, mms1 sql.NullString
			var startNow = time.Now()
			var startTime = fmt.Sprintf("%02d:%02d:%02d", startNow.Hour(), startNow.Minute(), startNow.Second())
			for rows.Next() {
				var message = ""
				var result = ""
				mmscnt++
				
				rows.Scan(&msgkey, &rslt, &phone, &sent_key, &userid, &mem_id, &cb_msg_id, &mms1)

				if rslt.String != "1000" && conf.REFUND {
					message = rslt.String
					result = "N"
					mmserrcnt++
					amtsStrs = append(amtsStrs, "(now(),?,?,?,?,?,?)")
					amtsValues = append(amtsValues, "3")
					if len(mms1.String) > 0 {
						if s.EqualFold(mst_sent_voucher.String, "V") {
							amtsValues = append(amtsValues, cprice.V_price_nas_mms.Float64)
							amtsValues = append(amtsValues, "웹(B) 발송실패 환불,바우처")
							amtsValues = append(amtsValues, cb_msg_id.String+phone.String)
							amtsValues = append(amtsValues, ((cprice.V_price_nas_mms.Float64 - cprice.P_price_nas_mms.Float64) * -1))
							amtsValues = append(amtsValues, cprice.B_price_nas_mms.Float64*-1)
						} else {
							amtsValues = append(amtsValues, cprice.C_price_nas_mms.Float64)
							if s.EqualFold(mst_sent_voucher.String, "B") {
								amtsValues = append(amtsValues, "웹(B) 발송실패 환불,보너스")
							} else {
								amtsValues = append(amtsValues, "웹(B) 발송실패 환불")
							}										
							amtsValues = append(amtsValues, cb_msg_id.String+phone.String)
							amtsValues = append(amtsValues, ((cprice.C_price_nas_mms.Float64 - cprice.P_price_nas_mms.Float64) * -1))
							amtsValues = append(amtsValues, cprice.B_price_nas_mms.Float64*-1)
						}

					} else {
						if s.EqualFold(mst_sent_voucher.String, "V") {
							amtsValues = append(amtsValues, cprice.V_price_nas.Float64)
							amtsValues = append(amtsValues, "웹(B) 발송실패 환불,바우처")
							amtsValues = append(amtsValues, cb_msg_id.String+phone.String)
							amtsValues = append(amtsValues, ((cprice.V_price_nas.Float64 - cprice.P_price_nas.Float64) * -1))
							amtsValues = append(amtsValues, cprice.B_price_nas.Float64*-1)
						} else {
							amtsValues = append(amtsValues, cprice.C_price_nas.Float64)
							if s.EqualFold(mst_sent_voucher.String, "B") {
								amtsValues = append(amtsValues, "웹(B) 발송실패 환불,보너스")
							} else {
								amtsValues = append(amtsValues, "웹(B) 발송실패 환불")
							}										
							amtsValues = append(amtsValues, cb_msg_id.String+phone.String)
							amtsValues = append(amtsValues, ((cprice.C_price_nas.Float64 - cprice.P_price_nas.Float64) * -1))
							amtsValues = append(amtsValues, cprice.B_price_nas.Float64*-1)
						}
					}
				} else {
					message = "웹(B) 성공"
					result = "Y"
				}

				tx.Exec("update cb_msg_"+userid.String+" set MESSAGE_TYPE='lg', MESSAGE = ?, RESULT = ? where remark4=? and msgid = ?", message, result, sent_key.String, cb_msg_id.String)

				upmsgids = append(upmsgids, msgkey.String)

				if len(upmsgids) >= 1000 {

					var commastr = "update " + MMSTable + " set ETC4 = '1' where MSGKEY in ("

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

				var commastr = "update " + MMSTable + " set ETC4 = '1' where MSGKEY in ("

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

			tx.Exec("update cb_wt_msg_sent set mst_err_nas = ifnull(mst_err_nas,0) + ?, mst_nas = ifnull(mst_nas,0) + ?, mst_wait = mst_wait - ?  where mst_id=?", mmserrcnt, ( mmscnt-mmserrcnt), mmscnt, sent_key.String)
			tx.Commit()
			stdlog.Printf(" ( %s ) WEB(B) MMS 처리 - %s : %d \n", startTime, sent_key.String, mmscnt)

		}
	}
}
