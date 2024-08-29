package nanoit

import (
	//"baseprice"
	"webagent/src/config"
	"database/sql"
	"fmt"

	//"log"
	"webagent/src/databasepool"
	"regexp"
	s "strings"
	"sync"

	//	"time"

	_ "github.com/go-sql-driver/mysql"
)

func Process() {
	var wg sync.WaitGroup
	for {
		wg.Add(1)
		go nanoProcess(&wg)
		wg.Wait()
	}

}

func nanoProcess(wg *sync.WaitGroup) {
	defer wg.Done()
	defer func() {
		if r := recover(); r != nil {
			config.Stdlog.Println("nanoit panic 발생 원인 : ", r)
			if err, ok := r.(error); ok {
				if s.Contains(err.Error(), "connection refused") {
					for {
						config.Stdlog.Println("nanoit send ping to DB")
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
	//var stdlog = config.Stdlog
	var errlog = config.Stdlog

	nanobcStrs := []string{}
	nanobcValues := []interface{}{}

	//		var t = time.Now()
	//		var monthStr = fmt.Sprintf("%d%02d", t.Year(), t.Month())

	var nanomsg_str = "select SQL_NO_CACHE cnm.remark4" +
		"   ,(sn div 50) as part" +
		"     ,group_concat(cnm.phn) as PHN" +
		"     ,group_concat(concat('''','82', right(phn, length(phn)-1)), '''') as msg_phn" +
		"     ,group_concat(cnm.sn) as sn" +
		"     ,wms.mst_lms_content as MSG_SMS" +
		"     ,wms.mst_sms_callback as SMS_SENDER" +
		"     ,cnm.msg_type" +
		"     ,cm.mem_userid as user_id" +
		"     ,cm.mem_id" +
		"     ,cm.mem_level" +
		"     ,wms.mst_reserved_dt as RESERVE_DT" +
		"     ,max(sn) as max_sn" +
		"     ,wms.mst_mms_content" +
		"     ,group_concat(cnm.cb_msg_id) as cb_msg_id" +
		"     ,(select origin1_path from cb_mms_images cmi where cmi.mem_id = cm.mem_id and mms_id = wms.mst_mms_content)  as mms1" +
		"     ,(select origin2_path from cb_mms_images cmi where cmi.mem_id = cm.mem_id and mms_id = wms.mst_mms_content)  as mms2" +
		"     ,(select origin3_path from cb_mms_images cmi where cmi.mem_id = cm.mem_id and mms_id = wms.mst_mms_content)  as mms3" +
		" from cb_nanoit_msg cnm " +
		"inner join cb_wt_msg_sent wms" +
		"   on cnm.remark4 = wms.mst_id " +
		"inner join cb_member cm" +
		"   on wms.mst_mem_id = cm.mem_id " +
		"group by cnm.remark4" +
		"        ,(sn div 50)" +
		"        ,wms.mst_lms_content" +
		"     ,wms.mst_sms_callback" +
		"     ,cnm.msg_type " +
		"order by remark4" +
		"        ,sn"

	Rows, err := db.Query(nanomsg_str)

	if err != nil {
		errlog.Println("cb_nanoit_msg 조회 중 오류 발생")
		errlog.Println(err)
		panic(err)
		// errlog.Fatal(nanomsg_str)
	}

	defer Rows.Close()

	nanobcStrs = nil
	nanobcValues = nil

	for Rows.Next() {
		var remark4, part, PHN, msg_phn, sn, msg_sms, sms_sender, msg_type, user_id, mem_id, mem_level, reserved_dt, max_sn, mms_content, cb_msg_id, mms1, mms2, mms3 sql.NullString

		Rows.Scan(&remark4, &part, &PHN, &msg_phn, &sn, &msg_sms, &sms_sender, &msg_type, &user_id, &mem_id, &mem_level, &reserved_dt, &max_sn, &mms_content, &cb_msg_id, &mms1, &mms2, &mms3)

		var msg_id int64

		if s.EqualFold(msg_type.String, "GRS") {
			var msgtype = "LMS"

			if mms1.Valid {
				msgtype = "MMS"
			}

			var insstr = "insert into cb_grs_msg(msg_gb" +
				",msg_st" +
				",msg_snd_phn" +
				",msg_rcv_phn" +
				",subject" +
				",text" +
				",cb_msg_id" +
				",file_path1" +
				",file_path2" +
				",file_path3" +
				",remark4" +
				",max_sn" +
				",msg_req_dttm" +
				",msg_ins_dttm)" +
				"values(?" +
				",?" +
				",?" +
				",?" +
				",?" +
				",?" +
				",?" +
				",?" +
				",?" +
				",?" +
				",?" +
				",?" +
				",now()" +
				",now())"

			var re = regexp.MustCompile(`\r?\n`)
			var str1 = re.ReplaceAllString(msg_sms.String, "")
			var subject = ""

			for idx, char := range str1 {
				subject = subject + string(char)
				if idx >= 35 {
					break
				}
			}

			//stdlog.Println(subject)
			res, err := db.Exec(insstr, msgtype, "0", sms_sender.String, PHN.String, subject, msg_sms.String, mem_id.String, mms1.String, mms2.String, mms3.String, remark4.String, max_sn.String)

			if err != nil {
				errlog.Println("cb_grs_msg Insert 처리 중 오류 발생")
				// errlog.Fatal(err)
			}
			msg_id, _ = res.LastInsertId()
		}

		var rcv_phn = s.Split(PHN.String, ",")
		var cb_msg_ids = s.Split(cb_msg_id.String, ",")
		var idx = 0

		for _, p := range rcv_phn {
			nanobcStrs = append(nanobcStrs, "(?,?,?,?,?,?,?,?,?,?)")
			nanobcValues = append(nanobcValues, msg_id)
			nanobcValues = append(nanobcValues, msg_type.String)
			nanobcValues = append(nanobcValues, p)
			nanobcValues = append(nanobcValues, remark4.String)
			nanobcValues = append(nanobcValues, mem_id.String)
			nanobcValues = append(nanobcValues, max_sn.String)
			nanobcValues = append(nanobcValues, mms1.String)
			nanobcValues = append(nanobcValues, mms2.String)
			nanobcValues = append(nanobcValues, mms3.String)
			nanobcValues = append(nanobcValues, cb_msg_ids[idx])

			//stdlog.Println(p, cb_msg_ids[idx])
			idx++
		}

		if len(nanobcStrs) > 0 {
			stmt := fmt.Sprintf("insert into cb_nano_broadcast_list(msg_id,type,rcv_phone,mst_id,mem_id,max_sn,FILE_PATH1,FILE_PATH2,FILE_PATH3,cb_msg_id) values %s", s.Join(nanobcStrs, ","))
			_, err := db.Exec(stmt, nanobcValues...)

			if err != nil {
				errlog.Println("Nano it Table Insert 처리 중 오류 발생 " + err.Error())
			}

			nanobcStrs = nil
			nanobcValues = nil

		}

		db.Exec("delete from cb_nanoit_msg where sn in (" + sn.String + ")")

	}
}
