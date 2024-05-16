package webcsms

import (
	"database/sql"
	"fmt"
	"webagent/src/baseprice"
	"webagent/src/common"
	"webagent/src/config"
	"webagent/src/databasepool"

	//"log"
	s "strings"
	"sync"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"github.com/lib/pq"
)

func Process() {
	var wg sync.WaitGroup
	for {
		wg.Add(1)
		go smsProcess(&wg)
		wg.Wait()
	}

}

func smsProcess(wg *sync.WaitGroup) {

	defer wg.Done()
	var db = databasepool.DB
	var conf = config.Conf
	var stdlog = config.Stdlog
	var errlog = config.Stdlog

	var cprice baseprice.BasePrice

	var isProc = true
	var t = time.Now()
	var monthStr = fmt.Sprintf("%d%02d", t.Year(), t.Month())

	var SMSTable = "OShotSMS_" + monthStr
	var msgcnt sql.NullString

	//발송 6시간 지난 메세지는 응답과 상관 없이 성공 처리 함.
	// sms 성공 처리
	err1 := db.QueryRow("SELECT count(1) as cnt FROM OShotSMS WHERE SendResult=1  AND insertdt + interval '6 hours' < now() AND mst_id IS NOT NULL").Scan(&msgcnt)
	if err1 != nil {
		errlog.Println("OShotMMS Table 조회 중 중 오류 발생", err1)
	} else {
		if !s.EqualFold(msgcnt.String, "0") {
			_, err := db.Exec("UPDATE OShotSMS SET SendDT=now(), SendResult='6', Telecom='000' WHERE SendResult=1 and insertdt + interval '6 hours' < now() and mst_id is not null")
			if err != nil {
				errlog.Println("webcsms.go / smsProcess / OshotSMS / update err : ", err)
			}
			_, err = db.Exec("insert into " + SMSTable + " SELECT * FROM OShotSMS WHERE SendResult>1 AND SendDT is not null and telecom = '000'")
			if err != nil {
				errlog.Println("webcsms.go / smsProcess / "+SMSTable+" / insert err : ", err)
			}
			_, err = db.Exec("delete FROM OShotSMS WHERE SendResult>1 AND SendDT is not null and telecom = '000'")
			if err != nil {
				errlog.Println("webcsms.go / smsProcess / OshotSMS / delete err : ", err)
			}
		}
	}

	var groupQuery = "select a.mst_id, b.mst_mem_id as mem_id,(select mem_userid from cb_member cm where cm.mem_id = b.mst_mem_id) AS mem_userid, b.mst_sent_voucher  from " + SMSTable + " a inner join cb_wt_msg_sent b ON a.mst_id = b.mst_id where a.proc_flag = 'Y'"

	groupRows, err := db.Query(groupQuery)
	if err != nil {
		errlog.Println("스마트미 SMS 조회 중 오류 발생")
		errlog.Println(groupQuery)
		errcode := err.Error()

		if s.Contains(errcode, "릴레이션(relation)이 없습니다") {
			db.Exec("CREATE TABLE " + SMSTable + " (LIKE OShotSMS INCLUDING ALL)")
			stdlog.Println(SMSTable + " 생성 !!")
		} else {
			errlog.Fatal(err)
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

			//cprice = baseprice.GetPrice(db, mem_id.String, errlog)
			cprice = baseprice.GetPrice(db, mem_id.String, errlog)

			var smsQuery = `select 
				a.msgid, 
				a.sendresult, 
				a.receiver as PHN, 
				b.mst_id as REMARK4, 
				(select mem_userid from cb_member cm where cm.mem_id = b.mst_mem_id) as mem_userid, 
				b.mst_mem_id as mem_id, 
				a.cb_msg_id 
			from 
				` + SMSTable + ` a inner join cb_wt_msg_sent b 
			on 
				a.mst_id = b.mst_id 
			where 
				a.proc_flag = 'Y' and a.mst_id = $1
			`
			rows, err := db.Query(smsQuery, mst_id.String)
			if err != nil {
				errlog.Println("스마트미 SMS 조회 중 오류 발생")
				errlog.Println(smsQuery)
				errlog.Fatal(smsQuery)
			}
			defer rows.Close()

			tx, err := db.Begin()
			if err != nil {
				errlog.Println(" 트랜잭션 시작 중 오류 발생")
				errlog.Fatal(err)
			}
			defer tx.Rollback() // 오류 발생 시 롤백

			var smserrcnt = 0
			var smscnt = 0

			AmtSmsValues := []common.AmtSmsColumn{}

			upmsgids := []interface{}{}

			var msgid, sendresult, phn, sent_key, userid, cb_msg_id sql.NullString
			var startNow = time.Now()
			var startTime = fmt.Sprintf("%02d:%02d:%02d", startNow.Hour(), startNow.Minute(), startNow.Second())

			for rows.Next() {
				var message = ""
				var result = ""
				AmtSmsValue := common.AmtSmsColumn{}

				rows.Scan(&msgid, &sendresult, &phn, &sent_key, &userid, &mem_id, &cb_msg_id)
				smscnt++

				currentTime := time.Now().Format("2006-01-02 15:04:05")

				if !s.EqualFold(sendresult.String, "6") && conf.REFUND {
					smserrcnt++
					message = sendresult.String
					result = "N"

					AmtSmsValue.Amt_datetime = currentTime
					AmtSmsValue.Amt_kind = "3"
					if s.EqualFold(mst_sent_voucher.String, "V") {
						AmtSmsValue.Amt_amount = cprice.V_price_smt_sms.Float64
						AmtSmsValue.Amt_memo = "웹(C) 발송실패 환불,바우처"
						AmtSmsValue.Amt_reason = cb_msg_id.String + phn.String
						AmtSmsValue.Amt_payback = ((cprice.V_price_smt_sms.Float64 - cprice.P_price_smt_sms.Float64) * -1)
						AmtSmsValue.Amt_admin = cprice.B_price_smt_sms.Float64 * -1
					} else {
						AmtSmsValue.Amt_amount = cprice.C_price_smt_sms.Float64
						if s.EqualFold(mst_sent_voucher.String, "B") {
							AmtSmsValue.Amt_memo = "웹(C) 발송실패 환불,보너스"
						} else {
							AmtSmsValue.Amt_memo = "웹(C) 발송실패 환불"
						}
						AmtSmsValue.Amt_reason = cb_msg_id.String + phn.String
						AmtSmsValue.Amt_payback = ((cprice.C_price_smt_sms.Float64 - cprice.P_price_smt_sms.Float64) * -1)
						AmtSmsValue.Amt_admin = cprice.B_price_smt_sms.Float64 * -1
					}
				} else {
					message = "웹(C) 성공"
					result = "Y"
				}
				AmtSmsValues = append(AmtSmsValues, AmtSmsValue)

				_, err := tx.Exec("update cb_msg_"+userid.String+" set MESSAGE_TYPE='sm', MESSAGE = $1, RESULT = $2 where remark4=$3 and msgid = $4", message, result, sent_key.String, cb_msg_id.String)
				if err != nil {
					errlog.Println("SMS Update 중 오류 발생", err)
					tx.Rollback()
					continue
				}

				upmsgids = append(upmsgids, msgid.String)

				if len(upmsgids) >= 1000 {

					err := updateProcFlag(tx, SMSTable, upmsgids)
					if err != nil {
						errlog.Println(SMSTable+" Table Update 중 오류 발생", err)
						tx.Rollback()
						break
					}
					upmsgids = nil
				}

				if len(AmtSmsValues) >= 1000 {
					insertAmtSms(tx, AmtSmsValues, mem_userid.String)
					AmtSmsValues = []common.AmtSmsColumn{}
				}

			}

			if len(upmsgids) > 0 {
				err := updateProcFlag(tx, SMSTable, upmsgids)
				if err != nil {
					errlog.Println(SMSTable+" Table Update 중 오류 발생", err)
					tx.Rollback()
				}
				upmsgids = nil
			}

			if len(AmtSmsValues) > 0 {
				insertAmtSms(tx, AmtSmsValues, mem_userid.String)
			}

			tx.Exec("update cb_wt_msg_sent set mst_err_smt = ifnull(mst_err_smt,0) + $1, mst_smt = ifnull(mst_smt,0) + $2, mst_wait = mst_wait - $3  where mst_id=$4", smserrcnt, (smscnt - smserrcnt), smscnt, sent_key.String)
			tx.Commit()
			stdlog.Printf(" ( %s ) WEB(C) SMS 처리 - %s : %d \n", startTime, sent_key.String, smscnt)

		}

	}

}

func updateProcFlag(tx *sql.Tx, tableName string, ids []interface{}) error {
	commastr := fmt.Sprintf("update %s set proc_flag='N' where msgid in (", tableName)
	for i := range ids {
		if i == 0 {
			commastr += fmt.Sprintf("$%d", i+1)
		} else {
			commastr += fmt.Sprintf(", $%d", i+1)
		}
	}
	commastr += ")"

	_, err := tx.Exec(commastr, ids...)
	return err
}

func insertAmtSms(tx *sql.Tx, RcsMsgResValue []common.AmtSmsColumn, userid string) {
	webcSStmt, err := tx.Prepare(pq.CopyIn("cb_amt_"+userid, common.GetRcsColumnPq(common.AmtSmsColumn{})...))
	if err != nil {
		config.Stdlog.Println("webcsms.go / insertAmtSms / cb_amt_"+userid+" / webcSStmt 초기화 실패 ", err)
		return
	}
	defer webcSStmt.Close()

	for _, data := range RcsMsgResValue {
		_, err := webcSStmt.Exec(data.Amt_datetime, data.Amt_kind, data.Amt_amount, data.Amt_memo, data.Amt_reason, data.Amt_payback, data.Amt_admin)
		if err != nil {
			config.Stdlog.Println("webcsms.go / insertAmtSms / cb_amt_"+userid+" / webcSStmt personal Exec ", err)
		}
	}

	_, err = webcSStmt.Exec()
	if err != nil {
		webcSStmt.Close()
		config.Stdlog.Println("webcsms.go / insertAmtSms / cb_amt_"+userid+" / webcSStmt Exec ", err)
	}
	webcSStmt.Close()
	err = tx.Commit()
	if err != nil {
		config.Stdlog.Println("webcsms.go / insertAmtSms / cb_amt_"+userid+" / webcSStmt commit ", err)
	}
}
