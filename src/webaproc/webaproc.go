package webaproc

import (
	"webagent/src/baseprice"
	"webagent/src/config"
	"database/sql"
	"fmt"

	//"log"
	"webagent/src/databasepool"
	s "strings"
	"sync"
	"time"

	_ "github.com/go-sql-driver/mysql"
)

func Process() {
	var wg sync.WaitGroup
	for {
		wg.Add(1)
		go grsProcess(&wg)
		wg.Wait()
	}

}

func grsProcess(wg *sync.WaitGroup) {
	defer wg.Done()
	defer func() {
		if r := recover(); r != nil {
			config.Stdlog.Println("webaproc panic 발생 원인 : ", r)
			if _, ok := r.(string); ok && s.Contains(r.(string), "connection refused") {
				for {
					config.Stdlog.Println("webaproc send ping to DB")
					err := databasepool.DB.Ping()
					if err == nil {
						break
					}
					time.Sleep(10 * time.Second)
				}
			}
		}
	}()
	var db = databasepool.DB
	var conf = config.Conf
	var stdlog = config.Stdlog
	var errlog = config.Stdlog

	amtsStrs := []string{}
	amtsValues := []interface{}{}
	var cprice baseprice.BasePrice

	var t = time.Now()
	var monthStr = fmt.Sprintf("%d%02d", t.Year(), t.Month())

	var updateStr = "update cb_grs_broadcast_" + monthStr + " cgb " +
		"set cgb.BC_SND_ST = '3' " +
		"WHERE cgb.bc_snd_st = '2' " +
		"and date_add(cgb.BC_SND_DTTM, interval 6 HOUR) < now()"

	_, err := db.Exec(updateStr)

	if err != nil {
		errlog.Println("GRS Proc 6시간 초과 Update 처리 중 오류 발생")
		errcode := err.Error()

		if s.Index(errcode, "1146") > 0 {
			db.Exec("Create Table IF NOT EXISTS cb_grs_broadcast_" + monthStr + " like cb_grs_broadcast")
			stdlog.Println("cb_grs_broadcast_" + monthStr + " 생성 !!")
		} else {
			// errlog.Fatal(err)
			panic(err)
		}

	}

	var grsgroup_str = "select distinct " +
		"      cgm.mst_id as gremark4" +
		"      ,cm.mem_userid as gmem_userid" +
		"      ,cm.mem_id as gmem_id" +
		"  from cb_nano_broadcast_list cgm" +
		" inner join cb_grs_broadcast_" + monthStr + " cgb" +
		"    on cgm.msg_id = cgb.msg_id" +
		"   and cgm.rcv_phone = cgb.bc_rcv_phn" +
		" inner join cb_member cm" +
		"    on cm.mem_id = cgm.mem_id" +
		" where cgb.bc_snd_st in( '3', '4') " +
		" order by cgm.mst_id "

	gRows, err := db.Query(grsgroup_str)

	if err != nil {
		errlog.Println("GRS Proc 처리 대상 조회 중 오류 발생")
		// errlog.Fatal(err)
	}

	defer gRows.Close()

	for gRows.Next() {
		var gremark4, gmem_userid, gmem_id sql.NullString

		gRows.Scan(&gremark4, &gmem_userid, &gmem_id)

		cprice = baseprice.GetPrice(db, gmem_id.String, errlog)

		var grsmsg_str = "select SQL_NO_CACHE cgm.msg_id" +
			"      ,cgm.max_sn" +
			"      ,cgb.BC_RSLT_NO" +
			"      ,cgb.bc_rslt_text as BC_RSLT_TEXT" +
			"	   ,cgb.bc_rcv_phn" +
			"      ,(case when cgb.bc_rcv_phn like '01%' then " +
			"                concat('82', right(cgb.bc_rcv_phn, length(cgb.bc_rcv_phn) - 1)) " +
			"             else " +
			"                cgb.bc_rcv_phn" +
			"        end) as PHN" +
			"      ,cgm.mst_id as REMARK4" +
			"      ,cm.mem_userid" +
			"      ,cm.mem_level" +
			"      ,cm.mem_phn_agent" +
			"      ,cm.mem_sms_agent" +
			"      ,cm.mem_2nd_send" +
			"      ,cm.mem_id" +
			"      ,cgm.FILE_PATH1 as mms1" +
			"      ,cgm.FILE_PATH2 as mms2" +
			"      ,cgm.FILE_PATH3 as mms3" +
			"      ,cgm.cb_msg_id " +
			"      ,cgb.msg_gb " +
			"      ,cgb.BC_MSG_ID " +
			"  from cb_nano_broadcast_list cgm" +
			" inner join cb_grs_broadcast_" + monthStr + " cgb" +
			"    on cgm.msg_id = cgb.msg_id" +
			"   and cgm.rcv_phone = cgb.bc_rcv_phn" +
			" inner join cb_member cm" +
			"    on cm.mem_id = cgm.mem_id" +
			" where cgb.bc_snd_st in( '3', '4') " +
			" and cgm.mst_id = '" + gremark4.String + "'"

		Rows, err := db.Query(grsmsg_str)

		if err != nil {
			errlog.Println("GRS Proc 처리 대상 조회 중 오류 발생")
			// errlog.Fatal(err)
		}

		defer Rows.Close()

		amtsStrs = nil
		amtsValues = nil

		var errcnt, proccnt, mst_biz_qty int
		errcnt = 0
		proccnt = 0
		mst_biz_qty = 0

		tx, err := db.Begin()
		if err != nil {
			errlog.Println(" 트랜잭션 시작 중 오류 발생")
			// errlog.Fatal(err)
		}

		var amtinsstr = "insert into cb_amt_" + gmem_userid.String + "(amt_datetime," +
			"amt_kind," +
			"amt_amount," +
			"amt_memo," +
			"amt_reason," +
			"amt_payback," +
			"amt_admin)" +
			" values %s"
		var startNow = time.Now()
		var startTime = fmt.Sprintf("%02d:%02d:%02d", startNow.Hour(), startNow.Minute(), startNow.Second())
		
		for Rows.Next() {
			var msg_id, max_sn, bc_rslt_no, bc_rslt_text, bc_rcv_phn, phn, remark4, mem_userid, mem_level, mem_phn_agent, mem_sms_agent, mem_2nd_send, mem_id, mms1, mms2, mms3, cb_msg_id, msg_gb, bc_msg_id sql.NullString

			Rows.Scan(&msg_id, &max_sn, &bc_rslt_no, &bc_rslt_text, &bc_rcv_phn, &phn, &remark4, &mem_userid, &mem_level, &mem_phn_agent, &mem_sms_agent, &mem_2nd_send, &mem_id, &mms1, &mms2, &mms3, &cb_msg_id, &msg_gb, &bc_msg_id)
			proccnt++
			if s.EqualFold(bc_rslt_no.String, "0") || s.EqualFold(bc_rslt_no.String, "111") || !conf.REFUND {

				tx.Exec("update cb_msg_"+gmem_userid.String+" set MESSAGE_TYPE='gs', MESSAGE = ?, RESULT = 'Y' where remark4=? and msgid = ?", "WEB(A)성공", remark4.String, cb_msg_id.String)

				if s.EqualFold(msg_gb.String, "LMS") {
					mst_biz_qty++
				}
			} else {
				if conf.REFUND {
					errcnt++

					tx.Exec("update cb_msg_"+gmem_userid.String+" set MESSAGE_TYPE='gs', MESSAGE = ?, RESULT = 'N' where remark4=? and msgid = ?", bc_rslt_no.String, remark4.String, cb_msg_id.String)

					if len(mms1.String) <= 1 {
						amtsStrs = append(amtsStrs, "(now(),?,?,?,?,?,?)")
						amtsValues = append(amtsValues, "3")
						amtsValues = append(amtsValues, cprice.C_price_grs.Float64)
						amtsValues = append(amtsValues, "웹(A) 발송실패 환불")
						amtsValues = append(amtsValues, cb_msg_id.String+phn.String)
						amtsValues = append(amtsValues, ((cprice.C_price_grs.Float64 - cprice.P_price_grs.Float64) * -1))
						amtsValues = append(amtsValues, cprice.B_price_grs.Float64*-1)
					} else {
						amtsStrs = append(amtsStrs, "(now(),?,?,?,?,?,?)")
						amtsValues = append(amtsValues, "3")
						amtsValues = append(amtsValues, cprice.C_price_grs_mms.Float64)
						amtsValues = append(amtsValues, "웹(A) 발송실패 환불")
						amtsValues = append(amtsValues, cb_msg_id.String+phn.String)
						amtsValues = append(amtsValues, ((cprice.C_price_grs_mms.Float64 - cprice.P_price_grs_mms.Float64) * -1))
						amtsValues = append(amtsValues, cprice.B_price_grs_mms.Float64*-1)
					}
				}
			}

			tx.Exec("delete from cb_nano_broadcast_list where msg_id = ? and rcv_phone = ?", msg_id.String, bc_rcv_phn.String)
			tx.Exec("update cb_grs_broadcast_"+monthStr+" set bc_snd_st = bc_snd_st + 2 where bc_msg_id = ?", bc_msg_id.String)

			if len(amtsStrs) >= 1000 {
				stmt := fmt.Sprintf(amtinsstr, s.Join(amtsStrs, ","))
				_, err := tx.Exec(stmt, amtsValues...)

				if err != nil {
					errlog.Println("GRS 환불 AMT Table Insert 처리 중 오류 발생 " + err.Error())
				}

				amtsStrs = nil
				amtsValues = nil
			}

		}

		tx.Exec("update cb_wt_msg_sent set mst_err_grs = ifnull(mst_err_grs,0) + ?, mst_grs = ifnull(mst_grs,0) + ?, mst_wait = mst_wait - ?, mst_grs_biz_qty = ifnull(mst_grs_biz_qty,0) + ?  where mst_id=?", errcnt, (proccnt-errcnt), proccnt, mst_biz_qty, gremark4.String)

		if len(amtsStrs) > 0 {
			stmt := fmt.Sprintf(amtinsstr, s.Join(amtsStrs, ","))
			_, err := tx.Exec(stmt, amtsValues...)

			if err != nil {
				errlog.Println("GRS 환불 AMT Table Insert 처리 중 오류 발생 " + err.Error())
			}

		}
		tx.Commit()
		stdlog.Printf(" ( %s ) WEB(A) LMS 처리 - %s : %d \n", startTime, gremark4.String, proccnt)
		
	}
}
