package webcmms

import (
	"database/sql"
	"fmt"
	"webagent/src/baseprice"
	"webagent/src/common"
	"webagent/src/config"
	"webagent/src/databasepool"

	//	"log"
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
		go mmsProcess(&wg)
		wg.Wait()
	}

}

func mmsProcess(wg *sync.WaitGroup) {

	defer wg.Done()
	var db = databasepool.DB
	var conf = config.Conf
	var stdlog = config.Stdlog
	var errlog = config.Stdlog

	var cprice baseprice.BasePrice

	var isProc = true
	var t = time.Now()
	var monthStr = fmt.Sprintf("%d%02d", t.Year(), t.Month())

	var MMSTable = "OShotMMS_" + monthStr
	var msgcnt sql.NullString
	//발송 6시간 지난 메세지는 응답과 상관 없이 성공 처리 함.

	//lms 성공 처리
	err1 := db.QueryRow("SELECT count(1) as cnt from OShotMMS WHERE SendResult=1 and insertdt + interval '6 hours' < now() and mst_id is not null").Scan(&msgcnt)
	if err1 != nil {
		errlog.Println("OShotMMS Table 조회 중 중 오류 발생", err1)
	} else {
		if !s.EqualFold(msgcnt.String, "0") {
			_, err := db.Exec("UPDATE OShotMMS SET SendDT=now(), SendResult='6', Telecom='000' WHERE SendResult=1 and insertdt + interval '6 hours' < now() and mst_id is not null")
			if err != nil {
				errlog.Println("webcmms.go / mmsProcess / OshotMMS / update err : ", err)
			}
			_, err = db.Exec("insert into " + MMSTable + " SELECT * FROM OShotMMS WHERE SendResult>1 AND SendDT is not null and telecom = '000'")
			if err != nil {
				errlog.Println("webcmms.go / mmsProcess / "+MMSTable+" / insert err : ", err)
			}
			_, err = db.Exec("delete FROM OShotMMS WHERE SendResult>1 AND SendDT is not null and telecom = '000'")
			if err != nil {
				errlog.Println("webcmms.go / mmsProcess / OshotMMS / delete err : ", err)
			}
		}
	}
	var groupQuery = "select distinct a.mst_id, b.mst_mem_id as mem_id,(select mem_userid from cb_member cm where cm.mem_id = b.mst_mem_id) AS mem_userid, b.mst_sent_voucher  from " + MMSTable + " a inner join cb_wt_msg_sent b ON a.mst_id = b.mst_id where a.proc_flag = 'Y'"

	groupRows, err := db.Query(groupQuery)
	if err != nil {
		errlog.Println("스마트미 MMS 조회 중 오류 발생")
		errlog.Println(groupQuery)
		errcode := err.Error()

		if s.Contains(errcode, "릴레이션(relation)이 없습니다") {
			db.Exec("CREATE TABLE " + MMSTable + " (LIKE OShotMMS INCLUDING ALL)")
			stdlog.Println(MMSTable + " 생성 !!")
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

			cprice = baseprice.GetPrice(db, mem_id.String, errlog)

			var mmsQuery = `select 
				a.msgid ,
				a.sendresult ,
				a.receiver as PHN,
				b.mst_id as REMARK4,
				(select mem_userid from cb_member cm where cm.mem_id = b.mst_mem_id) AS mem_userid,
				b.mst_mem_id as mem_id,
				a.cb_msg_id ,
				a.file_path1 as mms1,
				a.file_path2 as mms2,
				a.file_path3 as mms3
			from ` + MMSTable + ` a inner join cb_wt_msg_sent b 
			on a.mst_id = b.mst_id 
			where a.proc_flag = 'Y' and a.mst_id = $1`

			rows, err := db.Query(mmsQuery, mst_id.String)
			if err != nil {
				errlog.Println("스마트미 MMS 조회 중 오류 발생")
				errlog.Println(mmsQuery)
				errlog.Fatal(mmsQuery)
			}
			defer rows.Close()

			tx, err := db.Begin()
			if err != nil {
				errlog.Println(" 트랜잭션 시작 중 오류 발생")
				errlog.Fatal(err)
			}
			defer tx.Rollback() // 오류 발생 시 롤백

			var mmserrcnt = 0
			var mmscnt = 0

			AmtMmsValues := []common.AmtSmsMmsColumn{}

			upmsgids := []interface{}{}

			var msgid, sendresult, phn, sent_key, userid, cb_msg_id, mms1, mms2, mms3 sql.NullString
			var startNow = time.Now()

			var startTime = fmt.Sprintf("%02d:%02d:%02d", startNow.Hour(), startNow.Minute(), startNow.Second())
			for rows.Next() {
				var message = ""
				var result = ""
				AmtMmsValue := common.AmtSmsMmsColumn{}

				rows.Scan(&msgid, &sendresult, &phn, &sent_key, &userid, &mem_id, &cb_msg_id, &mms1, &mms2, &mms3)
				mmscnt++

				currentTime := time.Now().Format("2006-01-02 15:04:05")

				if !s.EqualFold(sendresult.String, "6") && conf.REFUND {
					mmserrcnt++
					message = sendresult.String
					result = "N"

					AmtMmsValue.Amt_datetime = currentTime
					AmtMmsValue.Amt_kind = "3"
					if len(mms1.String) > 0 {
						if s.EqualFold(mst_sent_voucher.String, "V") {
							AmtMmsValue.Amt_amount = cprice.V_price_smt_mms.Float64
							AmtMmsValue.Amt_memo = "웹(C) 발송실패 환불,바우처"
							AmtMmsValue.Amt_reason = cb_msg_id.String + phn.String
							AmtMmsValue.Amt_payback = ((cprice.V_price_smt_mms.Float64 - cprice.P_price_smt_mms.Float64) * -1)
							AmtMmsValue.Amt_admin = cprice.B_price_smt_mms.Float64 * -1

						} else {
							AmtMmsValue.Amt_amount = cprice.C_price_smt_mms.Float64
							if s.EqualFold(mst_sent_voucher.String, "B") {
								AmtMmsValue.Amt_memo = "웹(C) 발송실패 환불,보너스"
							} else {
								AmtMmsValue.Amt_memo = "웹(C) 발송실패 환불"
							}
							AmtMmsValue.Amt_reason = cb_msg_id.String + phn.String
							AmtMmsValue.Amt_payback = ((cprice.C_price_smt_mms.Float64 - cprice.P_price_smt_mms.Float64) * -1)
							AmtMmsValue.Amt_admin = cprice.B_price_smt_mms.Float64 * -1
						}

					} else {
						if s.EqualFold(mst_sent_voucher.String, "V") {
							AmtMmsValue.Amt_amount = cprice.V_price_smt.Float64
							AmtMmsValue.Amt_memo = "웹(C) 발송실패 환불,바우처"
							AmtMmsValue.Amt_reason = cb_msg_id.String + phn.String
							AmtMmsValue.Amt_payback = ((cprice.V_price_smt.Float64 - cprice.P_price_smt.Float64) * -1)
							AmtMmsValue.Amt_admin = cprice.B_price_smt.Float64 * -1
						} else {
							AmtMmsValue.Amt_amount = cprice.C_price_smt.Float64
							if s.EqualFold(mst_sent_voucher.String, "B") {
								AmtMmsValue.Amt_memo = "웹(C) 발송실패 환불,보너스"
							} else {
								AmtMmsValue.Amt_memo = "웹(C) 발송실패 환불"
							}
							AmtMmsValue.Amt_reason = cb_msg_id.String + phn.String
							AmtMmsValue.Amt_payback = ((cprice.C_price_smt.Float64 - cprice.P_price_smt.Float64) * -1)
							AmtMmsValue.Amt_admin = cprice.B_price_smt.Float64 * -1
						}
					}
				} else {
					message = "웹(C) 성공"
					result = "Y"
				}

				AmtMmsValues = append(AmtMmsValues, AmtMmsValue)

				_, err := tx.Exec("update cb_msg_"+userid.String+" set MESSAGE_TYPE='sm', MESSAGE = $1, RESULT = $2 where remark4= $3 and msgid = $4", message, result, sent_key.String, cb_msg_id.String)
				if err != nil {
					errlog.Println("MMS Update 중 오류 발생", err)
					tx.Rollback()
					continue
				}

				upmsgids = append(upmsgids, msgid.String)

				if len(upmsgids) >= 1000 {

					err := common.UpdateProcFlag(tx, MMSTable, upmsgids)
					if err != nil {
						errlog.Println(MMSTable+" Table Update 중 오류 발생", err)
						tx.Rollback()
						break
					}
					upmsgids = nil
				}

				if len(AmtMmsValues) >= 1000 {
					insertAmtMms(tx, AmtMmsValues, mem_userid.String)
					AmtMmsValues = []common.AmtSmsMmsColumn{}
				}

			}

			if len(upmsgids) > 0 {

				err := common.UpdateProcFlag(tx, MMSTable, upmsgids)
				if err != nil {
					errlog.Println(MMSTable+" Table Update 중 오류 발생", err)
					tx.Rollback()
				}
				upmsgids = nil
			}

			if len(AmtMmsValues) > 0 {
				insertAmtMms(tx, AmtMmsValues, mem_userid.String)
			}

			tx.Exec("update cb_wt_msg_sent set mst_err_smt = ifnull(mst_err_smt,0) + &1, mst_smt = ifnull(mst_smt,0) + &2, mst_wait = mst_wait - &3  where mst_id=&4", mmserrcnt, (mmscnt - mmserrcnt), mmscnt, sent_key.String)
			tx.Commit()
			stdlog.Printf(" ( %s ) WEB(C) MMS 처리 - %s : %d \n", startTime, sent_key.String, mmscnt)

		}
	}

}

func insertAmtMms(tx *sql.Tx, RcsMsgResValue []common.AmtSmsMmsColumn, userid string) {
	webcMStmt, err := tx.Prepare(pq.CopyIn("cb_amt_"+userid, common.GetRcsColumnPq(common.AmtSmsMmsColumn{})...))
	if err != nil {
		config.Stdlog.Println("webcmms.go / insertAmtMms / cb_amt_"+userid+" / webcMStmt 초기화 실패 ", err)
		return
	}
	defer webcMStmt.Close()

	for _, data := range RcsMsgResValue {
		_, err := webcMStmt.Exec(data.Amt_datetime, data.Amt_kind, data.Amt_amount, data.Amt_memo, data.Amt_reason, data.Amt_payback, data.Amt_admin)
		if err != nil {
			config.Stdlog.Println("webcmms.go / insertAmtMms / cb_amt_"+userid+" / webcMStmt personal Exec ", err)
		}
	}

	_, err = webcMStmt.Exec()
	if err != nil {
		webcMStmt.Close()
		config.Stdlog.Println("webcmms.go / insertAmtMms / cb_amt_"+userid+" / webcMStmt Exec ", err)
	}
	webcMStmt.Close()
	err = tx.Commit()
	if err != nil {
		config.Stdlog.Println("webcmms.go / insertAmtMms / cb_amt_"+userid+" / webcMStmt commit ", err)
	}
}
