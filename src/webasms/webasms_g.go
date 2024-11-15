package webasms

import (
	"webagent/src/baseprice"
	"webagent/src/config"
	// "webagent/src/mapper"
	"database/sql"
	"webagent/src/databasepool"
	"fmt"

	//"log"
	s "strings"
	"sync"
	"time"
	"context"

	_ "github.com/go-sql-driver/mysql"
)

func Process_g(ctx context.Context) {
	var wg sync.WaitGroup
	for {
		select {
		case <- ctx.Done():
			time.Sleep(20 * time.Second)
			config.Stdlog.Println("webasms_g 정상적으로 종료되었습니다.")
			return
		default:
			wg.Add(1)
			go smsProcess_g(&wg)
			wg.Wait()
		}
	}

}

func smsProcess_g(wg *sync.WaitGroup) {
	defer wg.Done()
	defer func() {
		if r := recover(); r != nil {
			config.Stdlog.Println("webasms_g panic 발생 원인 : ", r)
			if err, ok := r.(error); ok {
				if s.Contains(err.Error(), "connection refused") {
					for {
						config.Stdlog.Println("webasms_g send ping to DB")
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

	//cprice = baseprice.GetPrice(db, mem_id.String)

	amtsStrs := []string{}
	amtsValues := []interface{}{}

	var isProc = true
	var t = time.Now()
	var monthStr = fmt.Sprintf("%d%02d", t.Year(), t.Month())

	var SMSTable = "SMS_LOG_G_" + monthStr
	var msgcnt sql.NullString

	//발송 6시간 지난 메세지는 응답과 상관 없이 성공 처리 함.
	// sms 성공 처리
	err1 := db.QueryRow("SELECT count(1) as cnt from SMS_MSG_G WHERE TR_SENDSTAT='1' and date_add(TR_SENDDATE, interval 6 HOUR) < now() and TR_ETC10 is not null").Scan(&msgcnt)
	if err1 != nil {
	   errlog.Println("나노 저가망 SMS_MSG_G Table 조회 중 중 오류 발생", err1)
	   panic(err1)
	} else {		
		if !s.EqualFold(msgcnt.String, "0") {	
			db.Exec("UPDATE SMS_MSG_G SET TR_RSLTDATE=now(), TR_SENDSTAT='2', TR_NET='ETC' WHERE TR_SENDSTAT=1 and date_add(TR_SENDDATE, interval 6 HOUR) < now() and TR_ETC10 is not null")
			db.Exec("insert into " + SMSTable + " SELECT * FROM SMS_MSG_G WHERE TR_SENDSTAT='2' AND TR_RSLTDATE is not null and TR_NET = 'ETC'")
			db.Exec("delete FROM SMS_MSG_G WHERE TR_SENDSTAT='2' AND TR_RSLTDATE is not null and TR_NET = 'ETC'")
		}
	}

	var groupQuery = "select distinct a.TR_ETC10 as mst_id, b.mst_mem_id as mem_id, (select mem_userid from cb_member cm where cm.mem_id = b.mst_mem_id) AS mem_userid, b.mst_sent_voucher from " + SMSTable + " a inner join cb_wt_msg_sent b on a.TR_ETC10 = b.mst_id where a.TR_SENDSTAT = '2' and a.TR_ETC8 = 'Y' "

	groupRows, err := db.Query(groupQuery)
	if err != nil {
		errcode := err.Error()
		errlog.Println("나노 저가망 SMS 조회 중 오류 발생", groupQuery, errcode)

		if s.Index(errcode, "1146") > 0 {
			db.Exec("Create Table IF NOT EXISTS " + SMSTable + " like SMS_LOG_G")
			errlog.Println(SMSTable + " 생성 !!")
		} else {
			//errlog.Fatal(groupQuery)
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

			var smsQuery = "SELECT SQL_NO_CACHE " +
				"       a.TR_NUM as MsgID " +
				"      ,a.TR_RSLTSTAT as SendResult" +
				"      ,a.TR_PHONE AS PHN" +
				"      ,a.TR_ETC10 AS REMARK4" +
				"      ,(select mem_userid from cb_member cm where cm.mem_id = b.mst_mem_id) AS mem_userid " +
				"      ,b.mst_mem_id AS mem_id" +
				"      ,a.TR_ETC7 as resend_flag " +
				"      ,a.TR_ETC9 as cb_msg_id " +
				" from " + SMSTable + " a INNER JOIN " +
				"        cb_wt_msg_sent b ON a.TR_ETC10 = b.mst_id " +
				"WHERE a.TR_SENDSTAT = '2' and a.TR_ETC8 = 'Y' " +
				" and a.TR_ETC10 = ?"

			rows, err := db.Query(smsQuery, mst_id.String)
			if err != nil {
				errlog.Println("나노 저가망 SMS 조회 중 오류 발생")
				errlog.Println(smsQuery)
				// errlog.Fatal(smsQuery)
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

			var smserrcnt = 0
			var smscnt = 0

			amtsStrs = nil
			amtsValues = nil

			upmsgids := []interface{}{}

			var msgid, sendresult, phn, sent_key, userid, cb_msg_id, resend_flag sql.NullString
			var startNow = time.Now()
			var startTime = fmt.Sprintf("%02d:%02d:%02d", startNow.Hour(), startNow.Minute(), startNow.Second())
			for rows.Next() {
				var message = ""
				var result = ""
				rows.Scan(&msgid, &sendresult, &phn, &sent_key, &userid, &mem_id, &resend_flag, &cb_msg_id)
				if !s.EqualFold(resend_flag.String, "2"){
					smscnt++
				}
				if !s.EqualFold(sendresult.String, "0") && conf.REFUND {
					//upcbmsgids = append(upcbmsgids, cb_msg_id.String)
					// message = mapper.NanoCode(sendresult.String)
					message = sendresult.String
					result = "N"
					if !s.EqualFold(resend_flag.String, "2"){
						smserrcnt++
						amtsStrs = append(amtsStrs, "(now(),?,?,?,?,?,?)")
						amtsValues = append(amtsValues, "3")
					}
					if s.EqualFold(mst_sent_voucher.String, "V") {
						amtsValues = append(amtsValues, cprice.V_price_grs_sms.Float64)
						amtsValues = append(amtsValues, "웹(A) 발송실패 환불,바우처")
						amtsValues = append(amtsValues, cb_msg_id.String+phn.String)
						amtsValues = append(amtsValues, ((cprice.V_price_grs_sms.Float64 - cprice.P_price_grs_sms.Float64) * -1))
						amtsValues = append(amtsValues, cprice.B_price_grs_sms.Float64*-1)
					} else {
						amtsValues = append(amtsValues, cprice.C_price_grs_sms.Float64)
						if s.EqualFold(mst_sent_voucher.String, "B") {
							amtsValues = append(amtsValues, "웹(A) 발송실패 환불,보너스")
						} else {
							amtsValues = append(amtsValues, "웹(A) 발송실패 환불")
						}										
						amtsValues = append(amtsValues, cb_msg_id.String+phn.String)
						amtsValues = append(amtsValues, ((cprice.C_price_grs_sms.Float64 - cprice.P_price_grs_sms.Float64) * -1))
						amtsValues = append(amtsValues, cprice.B_price_grs_sms.Float64*-1)
					}
				} else {
					message = "웹(A) 성공"
					result = "Y"
				}

				tx.Exec("update cb_msg_"+userid.String+" set MESSAGE_TYPE='gs', MESSAGE = ?, RESULT = ? where remark4=? and msgid = ?", message, result, sent_key.String, cb_msg_id.String)

				upmsgids = append(upmsgids, msgid.String)

				if len(upmsgids) >= 1000 {

					var commastr = "update " + SMSTable + " set TR_ETC8 = 'N' where TR_NUM in ("

					for i := 1; i < len(upmsgids); i++ {
						commastr = commastr + "?,"
					}

					commastr = commastr + "?)"

					_, err1 := tx.Exec(commastr, upmsgids...)

					if err1 != nil {
						errlog.Println(SMSTable + " Table Update 처리 중 오류 발생 ")
					}

					upmsgids = nil
				}

				if len(amtsStrs) >= 1000 {
					stmt := fmt.Sprintf(amtinsstr, s.Join(amtsStrs, ","))
					_, err := tx.Exec(stmt, amtsValues...)

					if err != nil {
						errlog.Println("나노 저가망 SMS AMT Table Insert 처리 중 오류 발생 " + err.Error())
					}

					amtsStrs = nil
					amtsValues = nil
				}

			}

			if len(upmsgids) > 0 {

				var commastr = "update " + SMSTable + " set TR_ETC8 = 'N' where TR_NUM in ("

				for i := 1; i < len(upmsgids); i++ {
					commastr = commastr + "?,"
				}

				commastr = commastr + "?)"

				_, err1 := tx.Exec(commastr, upmsgids...)

				if err1 != nil {
					errlog.Println(SMSTable + " Table Update 처리 중 오류 발생 ")
				}
			}

			if len(amtsStrs) > 0 {
				stmt := fmt.Sprintf(amtinsstr, s.Join(amtsStrs, ","))
				_, err := tx.Exec(stmt, amtsValues...)

				if err != nil {
					errlog.Println("나논 SMS AMT Table Insert 처리 중 오류 발생 " + err.Error())
				}

			}

			tx.Exec("update cb_wt_msg_sent set mst_err_grs = ifnull(mst_err_grs,0) + ?, mst_grs = ifnull(mst_grs,0) + ?, mst_wait = mst_wait - ?  where mst_id=?", smserrcnt, (smscnt-smserrcnt), smscnt, sent_key.String)
			tx.Commit()
			stdlog.Printf(" ( %s ) WEB(A) 저가망 SMS 처리 - %s : %d \n", startTime, sent_key.String, smscnt)
		}
	}
}
