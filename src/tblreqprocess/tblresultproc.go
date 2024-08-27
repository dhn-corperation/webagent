package tblreqprocess

import (
	"webagent/src/config"
	"database/sql"
	"fmt"
	"sync"
	//"encoding/json"
	//	"os"
	//"strconv"
	"webagent/src/baseprice"
	"webagent/src/databasepool"
	//"rcs"
	s "strings"
	"time"
	
)

func Process() {
	var wg sync.WaitGroup
	for {
		wg.Add(1)
		go resProcess(&wg)
		wg.Wait()
	}

}

func resProcess(wg *sync.WaitGroup) {
	//var name string
	//stdlog.SetPrefix(log.Ldate|log.Ltime, "Result 처리 : ")
	//errlog.SetPrefix(log.Ldate|log.Ltime, "Result 오류 : ")
	defer wg.Done()
	var db = databasepool.DB
	var conf = config.Conf
	var stdlog = config.Stdlog
	var errlog = config.Stdlog

	var msgid, ad_flag, button1, button2, button3, button4, button5, code, image_link, image_url, kind, message, message_type sql.NullString
	var msg, msg_sms, only_sms, p_com, p_invoice, phn, profile, reg_dt, remark1, remark2, remark3, remark4, remark5, res_dt, reserve_dt sql.NullString
	var result, s_code, sms_kind, sms_lms_tit, sms_sender, sync, tmpl_id, wide, supplement, price, currency_type, mem_userid, mem_id, mem_level, mem_phn_agent, mem_sms_agent, mem_2nd_send, mms_id, mst_type2, mst_type3, mst_2nd_alim, msgcnt, vancnt, mst_sent_voucher, mem_lp_flag sql.NullString
	var mms_file1, mms_file2, mms_file3 sql.NullString
	var msgtype, phnstr /*, mem_2nd_type*/ string
	var cprice baseprice.BasePrice
	var isPass bool
	var cnt int
	
	msginsStrs := []string{}
	msginsValues := []interface{}{}

	atinsids := []interface{}{} // 알림톡 2차 발신시 2nd 테이블을 이용한 insert 용 id
	//atdelids := []interface{}{} // 알림톡 2차 발신시 2nd 테이블 삭제를 위한 id

	upmsgids := []interface{}{}

	amtsStrs := []string{}
	amtsValues := []interface{}{}

	ftlistsStrs := []string{}
	ftlistsValues := []interface{}{}

	nanoitStrs := []string{}
	nanoitValues := []interface{}{}

	mmsmsgStrs := []string{}
	mmsmsgValues := []interface{}{}

	ossmsStrs := []string{}
	ossmsValues := []interface{}{}

	osmmsStrs := []string{}
	osmmsValues := []interface{}{}

	nnsmsStrs := []string{}
	nnsmsValues := []interface{}{}

	nnmmsStrs := []string{}
	nnmmsValues := []interface{}{}

	nnlpsmsStrs := []string{}
	nnlpsmsValues := []interface{}{}

	nnlpmmsStrs := []string{}
	nnlpmmsValues := []interface{}{}
	
	rcsStrs := []string{}
	rcsValues := []interface{}{}

	smtpStrs := []string{}
	smtpValues := []interface{}{}
	var resquery = "SELECT trr.REMARK4 AS ressendkey, cm.mem_userid as username, cm.mem_id as user_mem_id, trr.sms_sender gsms_sender FROM " + conf.RESULTTABLE + " trr INNER JOIN cb_member cm ON trr.remark2 = cm.mem_id WHERE trr.remark3 IS NOT null and ( trr.reserve_dt < DATE_FORMAT(NOW(), '%Y%m%d%H%i%S') or trr.reserve_dt = '00000000000000') GROUP BY trr.REMARK4, cm.mem_userid"

	//for {

	resrows, err := db.Query(resquery)

	if err != nil {
		errlog.Println("Result Table 처리 중 오류 발생")
		errlog.Println(err)
		//errlog.Fatal(resquery)
	}
	defer resrows.Close()

	for resrows.Next() {
		var ressendkey sql.NullString
		var username sql.NullString
		var usermem_id sql.NullString
		var gsms_sender sql.NullString
		var rcsTemplate, rcsBrand, rcsDatetime, rcsKind, rcsContent, rcsBtn1, rcsBtn2, rcsBtn3, rcsBtn4, rcsBtn5, rcsChatbotID, rcsBtns, rcsBody, rcsBrandkey sql.NullString
		
		resrows.Scan(&ressendkey, &username, &usermem_id, &gsms_sender)
		var t = time.Now()
		var nowstr = fmt.Sprintf("%d%02d%02d%02d%02d%02d", t.Year(), t.Month(), t.Day(), t.Hour(), t.Minute(), t.Second())
		var resultsql string
		resultsql = " select SQL_NO_CACHE msgid, ad_flag, button1, button2, button3, button4, button5, code, image_link, image_url, kind, message, message_type," +
			" msg, msg_sms, only_sms, p_com, p_invoice, phn, profile, reg_dt, remark1, remark2, remark3, remark4, remark5, res_dt, reserve_dt, " +
			" result, s_code, sms_kind, sms_lms_tit, sms_sender, sync, tmpl_id, wide, supplement, price, currency_type " +
			"       ,b.mem_userid " +
			"       ,b.mem_id " +
			"       ,b.mem_level" +
			"       ,b.mem_phn_agent" +
			"       ,b.mem_sms_agent" +
			"       ,b.mem_2nd_send" +
			"       ,b.mem_lp_flag" +
			"       ,(select mst_mms_content from cb_wt_msg_sent wms where wms.mst_id = a.remark4) as mms_id" +
			"       ,(select mst_type2 from cb_wt_msg_sent wms where wms.mst_id = a.remark4) as mst_type2" +
			"       ,(select mst_type3 from cb_wt_msg_sent wms where wms.mst_id = a.remark4) as mst_type3" +
			"       ,(select mst_2nd_alim from cb_wt_msg_sent wms where wms.mst_id = a.remark4) as mst_2nd_alim" +
			"       ,(select count(1) as msgcnt from cb_msg_" + username.String + " cbmsg where cbmsg.msgid = a.msgid) as msgcnt" +
			" ,(SELECT mi.origin1_path FROM cb_wt_msg_sent wms INNER join cb_mms_images mi ON wms.mst_mms_content = mi.mms_id  WHERE  length(mst_mms_content ) > 5 AND wms.mst_id = a.remark4 ) as mms_file1" +
			" ,(SELECT mi.origin2_path FROM cb_wt_msg_sent wms INNER join cb_mms_images mi ON wms.mst_mms_content = mi.mms_id  WHERE  length(mst_mms_content ) > 5 AND wms.mst_id = a.remark4 ) as mms_file2" +
			" ,(SELECT mi.origin3_path FROM cb_wt_msg_sent wms INNER join cb_mms_images mi ON wms.mst_mms_content = mi.mms_id  WHERE  length(mst_mms_content ) > 5 AND wms.mst_id = a.remark4 ) as mms_file3" +
			" ,(select count(1) as vancnt from cb_block_lists cbl where cbl.sender = '" + gsms_sender.String + "' AND cbl.phn = CONCAT('0', SUBSTR(a.phn, 3,20))) AS vancnt " +
			"       ,(select mst_sent_voucher from cb_wt_msg_sent wms where wms.mst_id = a.remark4) as mst_sent_voucher" +
			"   from " + conf.RESULTTABLE + " a inner join cb_member b on b.mem_id = a.REMARK2" +
			"  where a.remark4 = ? and a.remark3 is not null"
		
		//stdlog.Println("RCS Flag", conf.RCS)
		
		rows, err := db.Query(resultsql, ressendkey.String)
		if err != nil {
			errlog.Println(" Result Table 처리 중 오류 발생")
			errlog.Println(err)
			errlog.Fatal(resultsql)
		}
		defer rows.Close()

		cnt = 0

		//tx, err := db.Begin()
		if err != nil {
			errlog.Println(" 트랜잭션 시작 중 오류 발생")
			errlog.Fatal(err)
		}

		msginsStrs = nil // cb_msg Table Insert 용
		msginsValues = nil

		upmsgids = nil // 처리된 message 처리를 위한 msgid 저장

		amtsStrs = nil // 요금 차감용
		amtsValues = nil

		ftlistsStrs = nil // 친구톡 성공 List 저장용
		ftlistsValues = nil

		nanoitStrs = nil // 나노IT Table Insert 용
		nanoitValues = nil

		mmsmsgStrs = nil //웹(A) Table Insert 용
		mmsmsgValues = nil

		ossmsStrs = nil //스마트미 SMS Table Insert 용
		ossmsValues = nil

		osmmsStrs = nil //스마트미 LMS/MMS Table Insert 용
		osmmsValues = nil

		nnsmsStrs = nil //나노 SMS Table Insert 용
		nnsmsValues = nil

		nnmmsStrs = nil //나노 LMS/MMS Table Insert 용
		nnmmsValues = nil

		nnlpsmsStrs = nil //나노 저가망 SMS Table Insert 용
		nnlpsmsValues = nil

		nnlpmmsStrs = nil //나노 저가망 LMS/MMS Table Insert 용
		nnlpmmsValues = nil

		smtpStrs = nil //스마트미 폰문자 Table Insert 용
		smtpValues = nil

		var insstr = ""
		var amtinsstr = ""

		// mst msg sent 에 발송 수량 Count 용 변수들
		var ftcnt = 0
		var fticnt = 0
		var atcnt = 0
		var ftilcnt = 0
		var ftcscnt = 0
		var lms_015cnt = 0
		var lms_phncnt = 0
		var lms_smtcnt = 0
		var lms_grscnt = 0
		var lms_imccnt = 0
		var lms_nascnt = 0
		//var rcs_cnt = 0

		var err_015cnt = 0
		var err_phncnt = 0
		var err_smtcnt = 0
		var err_grscnt = 0
		var err_imccnt = 0
		var err_nascnt = 0
		var err_ftcnt = 0
		var err_fticnt = 0
		var err_atcnt = 0
		var err_rcscnt = 0
		var err_ftilcnt = 0
		var err_ftcscnt = 0

		var mst_waitcnt = 0

		var cb_msg_message_type = ""
		var cb_msg_code = ""
		var cb_msg_message = ""
		var result_flag = ""

		// var resend_message_type = ""
		// var resend_code = ""

		//var aicnt = 0

		var sendkey = ""
		var mem_resend = ""

		cprice = baseprice.GetPrice(db, usermem_id.String, errlog)

		insstr = "insert IGNORE  into cb_msg_" + username.String + "(MSGID," +
			"AD_FLAG," +
			"BUTTON1," +
			"BUTTON2," +
			"BUTTON3," +
			"BUTTON4," +
			"BUTTON5," +
			"CODE," +
			"IMAGE_LINK," +
			"IMAGE_URL," +
			"KIND," +
			"MESSAGE," +
			"MESSAGE_TYPE," +
			"MSG," +
			"MSG_SMS," +
			"ONLY_SMS," +
			"P_COM," +
			"P_INVOICE," +
			"PHN," +
			"PROFILE," +
			"REG_DT," +
			"REMARK1," +
			"REMARK2," +
			"REMARK3," +
			"REMARK4," +
			"REMARK5," +
			"RES_DT," +
			"RESERVE_DT," +
			"RESULT," +
			"S_CODE," +
			"SMS_KIND," +
			"SMS_LMS_TIT," +
			"SMS_SENDER," +
			"SYNC," +
			"TMPL_ID," +
			"mem_userid," +
			"wide)" +
			"	  values %s"

		amtinsstr = "insert into cb_amt_" + username.String + "(amt_datetime," +
			"amt_kind," +
			"amt_amount," +
			"amt_memo," +
			"amt_reason," +
			"amt_payback," +
			"amt_admin)" +
			" values %s"

		var isPayment bool
		var startNow = time.Now()
		var startTime = fmt.Sprintf("%02d:%02d:%02d", startNow.Hour(), startNow.Minute(), startNow.Second())
		for rows.Next() {

			isPass = false
			isPayment = true

			var kko_kind string
			var amount float64
			var memo string
			var payback float64
			var admin_amt float64
			var ph_msg_type string
			if cnt == 0 {
				//stdlog.Println(ressendkey.String + " - Result 처리 시작 - ")
				sendkey = ressendkey.String
			}
			cnt++

			rows.Scan(&msgid,
				&ad_flag,
				&button1,
				&button2,
				&button3,
				&button4,
				&button5,
				&code,
				&image_link,
				&image_url,
				&kind,
				&message,
				&message_type,
				&msg,
				&msg_sms,
				&only_sms,
				&p_com,
				&p_invoice,
				&phn,
				&profile,
				&reg_dt,
				&remark1,
				&remark2,
				&remark3,
				&remark4,
				&remark5,
				&res_dt,
				&reserve_dt,
				&result,
				&s_code,
				&sms_kind,
				&sms_lms_tit,
				&sms_sender,
				&sync,
				&tmpl_id,
				&wide,
				&supplement,
				&price,
				&currency_type,
				&mem_userid,
				&mem_id,
				&mem_level,
				&mem_phn_agent,
				&mem_sms_agent,
				&mem_2nd_send,
				&mem_lp_flag,
				&mms_id,
				&mst_type2,
				&mst_type3,
				&mst_2nd_alim,
				&msgcnt,
				&mms_file1,
				&mms_file2,
				&mms_file3,
				&vancnt,
				&mst_sent_voucher)

			cb_msg_code = code.String
			cb_msg_message_type = message_type.String
			cb_msg_message = message.String

			if s.EqualFold(msgcnt.String, "0") {

				mem_resend = ""
				ph_msg_type = ""
				
				if s.EqualFold(mst_type2.String, "AI") || s.EqualFold(mst_type2.String, "AT") {
					if s.Contains(mst_type3.String, "s") {
						msgtype = "SMS"
					} else {
						msgtype = "LMS"
					}
					
					if s.Contains(mst_type3.String, "wa") && s.EqualFold(mem_lp_flag.String, "0") {
						mem_resend = "GREEN_SHOT"
					} else if s.Contains(mst_type3.String, "wa") && s.EqualFold(mem_lp_flag.String, "1") {
						mem_resend = "GREEN_SHOT_G"
					}
					
					if s.Contains(mst_type3.String, "wc") {
						mem_resend = "SMART"
					}
					
					if s.Contains(mst_type3.String, "rc") {
						mem_resend = "RCS"

						if s.EqualFold(msgtype, "LMS") {
							mem_resend = "SMART"
						}						
					}
					
					if s.Contains(mst_type3.String, "wp") {
						mem_resend = "SMT_PHN"
						ph_msg_type = mst_type3.String
					}

				} else { 
					if s.Contains(mst_type2.String, "s") {
						msgtype = "SMS"
					} else {
						msgtype = "LMS"
					}
	
					if s.Contains(mst_type2.String, "wa") && s.EqualFold(mem_lp_flag.String, "0") {
						mem_resend = "GREEN_SHOT"
					} else if s.Contains(mst_type2.String, "wa") && s.EqualFold(mem_lp_flag.String, "1") {
						mem_resend = "GREEN_SHOT_G"
					}

					if s.Contains(mst_type2.String, "wc") {
						mem_resend = "SMART"
					}
					
					if s.Contains(mst_type2.String, "rc") {
						mem_resend = "RCS"				
					}

					if s.Contains(mst_type2.String, "wp") {
						mem_resend = "SMT_PHN"
						ph_msg_type = mst_type2.String
					}
										
				}

				errlog.Println(mst_type3.String)
				errlog.Println(mem_lp_flag.String)
				errlog.Println(mem_resend)

				phnstr = phn.String

				//mem_resend = mem_2nd_send.String

				if len(p_invoice.String) > 0 && s.EqualFold(message_type.String, "ph") {
					
					if len(mem_resend) <= 0 {
						mem_resend = p_invoice.String
					}

					switch mem_resend {
					case "GREEN_SHOT":
						cb_msg_message_type = "gs"
						break
					case "GREEN_SHOT_G":
						cb_msg_message_type = "nl"
						break
					case "NASELF":
						cb_msg_message_type = "ns"
						break
					case "SMART":
						cb_msg_message_type = "sm"
						break
					case "RCS":
						cb_msg_message_type = "rc"
						break
					case "SMT_PHN":
						cb_msg_message_type = "wp"
						break
					}
				} else {
					cb_msg_message_type = message_type.String
				}
	
				// if !s.EqualFold(sendkey, remark4.String) {
				// 	var cntupdate = "update cb_wt_msg_sent set mst_ft = ifnull(mst_ft,0) + ?, mst_ft_img = ifnull(mst_ft_img,0) + ?, mst_at = ifnull(mst_at,0) + ? where mst_id = ?"
				// 	_, err := tx.Exec(cntupdate, ftcnt, fticnt, atcnt, sendkey)

				// 	if err != nil {
				// 		errlog.Println("WT_MSG_SENT 카카오 메세지 수량 처리 중 오류 발생 " + err.Error())
				// 	}
				// 	ftcnt = 0
				// 	fticnt = 0
				// 	atcnt = 0
				// }
				if s.HasPrefix(s.ToUpper(message_type.String), "F") && len(mst_2nd_alim.String) > 0 && s.EqualFold(mst_2nd_alim.String, "0") == false {

					//errlog.Println(result.String, msgid.String+"AT")

					if s.EqualFold(result.String, "N") {
						isPass = true

						atinsids = append(atinsids, msgid.String+"AT")

					}

					if len(atinsids) >= 100 {
 
						var copystr = "update " + conf.REQTABLE2 + " set remark3 = 'Y' where MSGID in ("

						for i := 1; i < len(atinsids); i++ {
							copystr = copystr + "?,"
						}

						copystr = copystr + "?)"

						_, err1 := db.Exec(copystr, atinsids...)

						if err1 != nil {
							errlog.Println("2ND 테이블에서 복사 처리 중 오류 발생 ")
							errlog.Println(err1)
							errlog.Println(copystr)
						} else {
							errlog.Println("2ND 테이블에서 복사 처리 완료 : ",len(atinsids))
						}
						atinsids = nil
 
					}

				}
				result_flag = result.String
				//stdlog.Println("MST Type ", mem_resend, msgtype, isPass, "sms", msg_sms.String)
				if isPass == false {
					// 발신 성공 시 금액 차감 처리
					if s.EqualFold(result.String, "Y") { // 카카오 메세지 성공시 차감 처리
						//errlog.Println("성공 처리 : ",result.String, message_type.String)
						if s.HasPrefix(s.ToUpper(message_type.String), "F") { // 친구톡 이면
							if s.EqualFold(message_type.String, "FC") {
								ftcscnt++
								kko_kind = "F"							
								if s.EqualFold(mst_sent_voucher.String, "V") {
									amount = cprice.V_price_ft_cs.Float64
									payback = cprice.V_price_ft_cs.Float64 - cprice.P_price_ft_cs.Float64
									admin_amt = cprice.B_price_ft_cs.Float64
									memo = "친구톡(CAROUSEL),바우처"
								} else {
									amount = cprice.C_price_ft_cs.Float64
									payback = cprice.C_price_ft_cs.Float64 - cprice.P_price_ft_cs.Float64
									admin_amt = cprice.B_price_ft_cs.Float64
									if s.EqualFold(mst_sent_voucher.String, "B") {
										memo = "친구톡(CAROUSEL),보너스"
									} else {
										memo = "친구톡(CAROUSEL)"
									}
								}							
							} else if s.EqualFold(message_type.String, "FL") {
								ftilcnt++
								kko_kind = "I"							
								if s.EqualFold(mst_sent_voucher.String, "V") {
									amount = cprice.V_price_ft_il.Float64
									payback = cprice.V_price_ft_il.Float64 - cprice.P_price_ft_il.Float64
									admin_amt = cprice.B_price_ft_il.Float64
									memo = "친구톡(이미지리스트),바우처"
								} else {
									amount = cprice.C_price_ft_il.Float64
									payback = cprice.C_price_ft_il.Float64 - cprice.P_price_ft_il.Float64
									admin_amt = cprice.B_price_ft_il.Float64
									if s.EqualFold(mst_sent_voucher.String, "B") {
										memo = "친구톡(이미지리스트),보너스"
									} else {
										memo = "친구톡(이미지리스트)"
									}									
								}						
							} else if len(image_url.String) <= 1 { // Image Url 이 null 이면 텍스트 친구톡
								ftcnt++
								kko_kind = "F"
								if s.EqualFold(mst_sent_voucher.String, "V") {
									amount = cprice.V_price_ft.Float64
									payback = cprice.V_price_ft.Float64 - cprice.P_price_ft.Float64
									admin_amt = cprice.B_price_ft.Float64
									memo = "친구톡(텍스트),바우처"
								} else {
									amount = cprice.C_price_ft.Float64
									payback = cprice.C_price_ft.Float64 - cprice.P_price_ft.Float64
									admin_amt = cprice.B_price_ft.Float64
									if s.EqualFold(mst_sent_voucher.String, "B") {
										memo = "친구톡(텍스트),보너스"
									} else {
										memo = "친구톡(텍스트)"
									}									
								}
							} else {
								fticnt++
								kko_kind = "I"
								if s.EqualFold(wide.String, "Y") { // 친구톡 WIDE 이면..
									if s.EqualFold(mst_sent_voucher.String, "V") {
										amount = cprice.V_price_ft_w_img.Float64
										payback = cprice.V_price_ft_w_img.Float64 - cprice.P_price_ft_w_img.Float64
										admin_amt = cprice.B_price_ft_w_img.Float64
										memo = "친구톡(와이드이미지),바우처"
									} else {
										amount = cprice.C_price_ft_w_img.Float64
										payback = cprice.C_price_ft_w_img.Float64 - cprice.P_price_ft_w_img.Float64
										admin_amt = cprice.B_price_ft_w_img.Float64
										if s.EqualFold(mst_sent_voucher.String, "B") {
											memo = "친구톡(와이드이미지),보너스"
										} else {
											memo = "친구톡(와이드이미지)"
										}										
									}
								} else {
									if s.EqualFold(mst_sent_voucher.String, "V") {
										amount = cprice.V_price_ft_img.Float64
										payback = cprice.V_price_ft_img.Float64 - cprice.P_price_ft_img.Float64
										admin_amt = cprice.B_price_ft_img.Float64
										memo = "친구톡(이미지),바우처"
									} else {
										amount = cprice.C_price_ft_img.Float64
										payback = cprice.C_price_ft_img.Float64 - cprice.P_price_ft_img.Float64
										admin_amt = cprice.B_price_ft_img.Float64
										if s.EqualFold(mst_sent_voucher.String, "B") {
											memo = "친구톡(이미지),보너스"
										} else {
											memo = "친구톡(이미지)"
										}										
										
									}
								}
							}

							// 친구 List 추가

							ftlistsStrs = append(ftlistsStrs, "(?, ?, now())")

							ftlistsValues = append(ftlistsValues, mem_id.String)
							ftlistsValues = append(ftlistsValues, phnstr)

						} else if s.EqualFold(message_type.String, "at") || s.EqualFold(message_type.String, "al") { // 알림톡 이면
							atcnt++
							kko_kind = "A"
							if s.EqualFold(mst_sent_voucher.String, "V") {
								amount = cprice.V_price_at.Float64
								payback = cprice.V_price_at.Float64 - cprice.P_price_at.Float64
								admin_amt = cprice.B_price_at.Float64
								memo = "알림톡(텍스트),바우처"
							} else {
								amount = cprice.C_price_at.Float64
								payback = cprice.C_price_at.Float64 - cprice.P_price_at.Float64
								admin_amt = cprice.B_price_at.Float64
								if s.EqualFold(mst_sent_voucher.String, "B") {
									memo = "알림톡(텍스트),보너스"
								} else {
									memo = "알림톡(텍스트)"
								}										
								
							}
						} else if s.EqualFold(message_type.String, "ai") { // 알림톡 이미지 이면
							atcnt++
							kko_kind = "E"
							if s.EqualFold(mst_sent_voucher.String, "V") {
								amount = cprice.V_price_at.Float64
								payback = cprice.V_price_at.Float64 - cprice.P_price_at.Float64
								admin_amt = cprice.B_price_at.Float64
								memo = "알림톡(이미지),바우처"
							} else {
								amount = cprice.C_price_at.Float64
								payback = cprice.C_price_at.Float64 - cprice.P_price_at.Float64
								admin_amt = cprice.B_price_at.Float64
								if s.EqualFold(mst_sent_voucher.String, "B") {
									memo = "알림톡(이미지),보너스"
								} else {
									memo = "알림톡(이미지)"
								}										
								
							}
						}

					} else { //  카카오 메세지 실패 시 혹은 메세지 전용 일 경우 처리
						if !s.EqualFold(message.String, "InvalidPhoneNumber") && len(mem_resend) > 0 && !s.EqualFold(mem_resend, "NONE") && len(sms_sender.String) > 0 {

							//var vansql = "select count(1) as vancnt from cb_block_lists cbl where cbl.sender = '" + sms_sender.String + "' AND cbl.phn = '" + phnstr + "'"
							//stdlog.Println(vansql)
							//err = db.QueryRow(vansql).Scan(&vancnt)

							if s.HasPrefix(phnstr, "82") {
								phnstr = "0" + phnstr[2:len(phnstr)]
							}

							//stdlog.Println("MST Type ", mem_resend, msgtype)

							if !s.EqualFold(vancnt.String, "0") { // 수신거부 List 에 있으면 수신거부 메세지 처리

								cb_msg_message = "수신거부"
								isPayment = false

								switch mem_resend {
								case "015":
									if s.EqualFold(msgtype, "SMS") {
										err_015cnt++
										cb_msg_message_type = "15"
										cb_msg_code = "015"
									} else if s.EqualFold(msgtype, "LMS") {
										if len(mms_file1.String) <= 0 {
											err_015cnt++
											cb_msg_message_type = "15"
											cb_msg_code = "015"
										} else {
											err_015cnt++
											cb_msg_message_type = "15"
											cb_msg_code = "015"
										}
									}
								case "PHONE":
									if s.EqualFold(msgtype, "SMS") {
										err_phncnt++
										cb_msg_message_type = "ph"
										cb_msg_code = "PHN"
									} else if s.EqualFold(msgtype, "LMS") {
										if len(mms_file1.String) <= 0 {
											err_phncnt++
											cb_msg_message_type = "ph"
											cb_msg_code = "PHN"
										} else {
											err_phncnt++
											cb_msg_message_type = "ph"
											cb_msg_code = "PHN"
										}
									}
								case "BKG":
									if s.EqualFold(msgtype, "SMS") {
										err_grscnt++
										cb_msg_message_type = "gs"
										cb_msg_code = "GRS"
									} else if s.EqualFold(msgtype, "LMS") {
										if len(mms_file1.String) <= 0 {
											err_grscnt++
											cb_msg_message_type = "gs"
											cb_msg_code = "GRS"
										} else {
											err_grscnt++
											cb_msg_message_type = "gs"
											cb_msg_code = "GRS"
										}
									}
								case "SMART":
									if s.EqualFold(msgtype, "SMS") {
										err_smtcnt++
										cb_msg_message_type = "SM"
										cb_msg_code = "SMT"
									} else if s.EqualFold(msgtype, "LMS") {
										if len(mms_file1.String) <= 0 {
											err_smtcnt++
											cb_msg_message_type = "SM"
											cb_msg_code = "SMT"
										} else {
											err_smtcnt++
											cb_msg_message_type = "SM"
											cb_msg_code = "SMT"
										}
									}
								case "GREEN_SHOT":
									if s.EqualFold(msgtype, "SMS") {
										err_grscnt++
										cb_msg_message_type = "gs"
										cb_msg_code = "GRS"
									} else if s.EqualFold(msgtype, "LMS") {
										if len(mms_file1.String) <= 0 {
											err_grscnt++
											cb_msg_message_type = "gs"
											cb_msg_code = "GRS"
										} else {
											err_grscnt++
											cb_msg_message_type = "gs"
											cb_msg_code = "GRS"
										}
									}
								case "GREEN_SHOT_G":
									if s.EqualFold(msgtype, "SMS") {
										err_grscnt++
										cb_msg_message_type = "nl"
										cb_msg_code = "GRS"
									} else if s.EqualFold(msgtype, "LMS") {
										if len(mms_file1.String) <= 0 {
											err_grscnt++
											cb_msg_message_type = "nl"
											cb_msg_code = "GRS"
										} else {
											err_grscnt++
											cb_msg_message_type = "nl"
											cb_msg_code = "GRS"
										}
									}
								case "IMC":
									if s.EqualFold(msgtype, "SMS") {
										err_smtcnt++
										cb_msg_message_type = "SM"
										cb_msg_code = "SMT"
									} else if s.EqualFold(msgtype, "LMS") {
										if len(mms_file1.String) <= 0 {
											err_imccnt++
											cb_msg_message_type = "IM"
											cb_msg_code = "IMC"
										} else {
											err_smtcnt++
											cb_msg_message_type = "SM"
											cb_msg_code = "SMT"
										}
									}
								case "SMT_PHN", "SMT_PHN_DB":
									if s.EqualFold(msgtype, "SMS") {
										err_smtcnt++
										cb_msg_message_type = "SM"
										cb_msg_code = "SMT"
									} else if s.EqualFold(msgtype, "LMS") {
										if len(mms_file1.String) <= 0 {
											err_imccnt++
											cb_msg_message_type = "WP"
											cb_msg_code = "SPH"
										} else {
											err_smtcnt++
											cb_msg_message_type = "WP"
											cb_msg_code = "SPH"
										}
									}
								case "NASELF":
									if s.EqualFold(msgtype, "SMS") {
										err_smtcnt++
										cb_msg_message_type = "SM"
										cb_msg_code = "SMT"
									} else if s.EqualFold(msgtype, "LMS") {
										if len(mms_file1.String) <= 0 {
											err_nascnt++
											cb_msg_message_type = "ns"
											cb_msg_code = "NAS"
										} else {
											err_smtcnt++
											cb_msg_message_type = "SM"
											cb_msg_code = "SMT"
										}
									}
								case "RCS":
									if s.EqualFold(msgtype, "SMS") {
										err_rcscnt++
										cb_msg_message_type = "RC"
										cb_msg_code = "RCS"
									} else if s.EqualFold(msgtype, "LMS") {
										if len(mms_file1.String) <= 0 {
											err_rcscnt++
											cb_msg_message_type = "RC"
											cb_msg_code = "RCS"
										} else {
											err_rcscnt++
											cb_msg_message_type = "RC"
											cb_msg_code = "RCS"
										}
									}

								}
							} else { // 수신거부 처리 끝
								// 2차 발신 처리 시작

								mst_waitcnt++
								cb_msg_message = "결과 수신대기"

								switch mem_resend {
								case "015":
									cb_msg_message_type = "15"
									cb_msg_code = "015"
									lms_015cnt++

									nanoitStrs = append(nanoitStrs, "(?,?,?,?)")
									nanoitValues = append(nanoitValues, "015")
									nanoitValues = append(nanoitValues, remark4.String)
									nanoitValues = append(nanoitValues, phnstr)
									nanoitValues = append(nanoitValues, msgid.String)

									kko_kind = "P"
									amount = cprice.C_price_015.Float64
									payback = cprice.C_price_015.Float64 - cprice.P_price_015.Float64
									admin_amt = cprice.B_price_015.Float64
									memo = "015저가문자"

								case "PHONE":
									cb_msg_message_type = "ph"
									cb_msg_code = "phn"
									lms_phncnt++
									nanoitStrs = append(nanoitStrs, "(?,?,?,?)")
									nanoitValues = append(nanoitValues, "PHONE")
									nanoitValues = append(nanoitValues, remark4.String)
									nanoitValues = append(nanoitValues, phnstr)
									nanoitValues = append(nanoitValues, msgid.String)

									kko_kind = "P"
									amount = cprice.C_price_phn.Float64
									payback = cprice.C_price_phn.Float64 - cprice.P_price_phn.Float64
									admin_amt = cprice.B_price_phn.Float64
									memo = "폰문자"

								case "BKG":
									cb_msg_message_type = "gs"
									cb_msg_code = "GRS"
									lms_grscnt++
									mmsmsgStrs = append(mmsmsgStrs, "(?,?,?,?,?,?,?,?,?,?,?)")
									mmsmsgValues = append(mmsmsgValues, sms_lms_tit)
									mmsmsgValues = append(mmsmsgValues, phnstr)
									mmsmsgValues = append(mmsmsgValues, sms_sender)
									mmsmsgValues = append(mmsmsgValues, "0")
									mmsmsgValues = append(mmsmsgValues, msg_sms.String)
									mmsmsgValues = append(mmsmsgValues, sms_sender)
									mmsmsgValues = append(mmsmsgValues, "0")
									mmsmsgValues = append(mmsmsgValues, msgid)
									mmsmsgValues = append(mmsmsgValues, remark4)
									mmsmsgValues = append(mmsmsgValues, mem_id)
									if s.EqualFold(reserve_dt.String, "00000000000000") {
										mmsmsgValues = append(mmsmsgValues, nowstr)
									} else {
										mmsmsgValues = append(mmsmsgValues, reserve_dt)
									}

									kko_kind = "P"
									amount = cprice.C_price_grs.Float64
									payback = cprice.C_price_grs.Float64 - cprice.P_price_grs.Float64
									admin_amt = cprice.B_price_grs.Float64
									memo = "웹(A)"

								case "RCS":
									cb_msg_message_type = "rc"
									cb_msg_code = "RCS"
									msgbaseid := ""
									srv_type := "RCSSMS"
									
									if conf.RCS {
										err = db.QueryRow("SELECT msr_template,msr_brand,msr_datetime,msr_kind,msr_content,msr_button1,msr_button2,msr_button3, msr_button4, msr_button5,msr_chatbotid,msr_button, msr_body, msr_brandkey FROM cb_wt_msg_rcs WHERE msr_mst_id = '" + ressendkey.String + "'").Scan(&rcsTemplate, &rcsBrand, &rcsDatetime, &rcsKind, &rcsContent, &rcsBtn1, &rcsBtn2, &rcsBtn3, &rcsBtn4, &rcsBtn5, &rcsChatbotID, &rcsBtns, &rcsBody, &rcsBrandkey)
									    //stdlog.Println("RCS BTN", rcsBtns.String)
									    if err != nil {
									        errlog.Println("RCS Table 조회 중 중 오류 발생", err)
									        errlog.Println("SELECT msr_template,msr_brand,msr_datetime,msr_kind,msr_content,msr_button1,msr_button2,msr_button3, msr_button4, msr_button5,msr_chatbotid,msr_button FROM cb_wt_msg_rcs WHERE msr_mst_id = '" + ressendkey.String + "'")
									    }			
									}									
									
									//var body rcs.RcsBody 
									
									if s.EqualFold(msgtype, "SMS") {
										//msgbaseid = "SS000000"				
										//srv_type = "RCSSMS"
										msgbaseid = rcsTemplate.String //"UBR.0Uv12uh7R4-GG000F"				
										srv_type = rcsKind.String //"RCSTMPL"
										//body.Description = rcsContent.String
									} else if s.EqualFold(msgtype, "LMS") {
										srv_type = rcsKind.String //"RCSLMS"
										msgbaseid = rcsTemplate.String //"SL000000"
										//body.Title = sms_lms_tit.String
										//body.Description = rcsContent.String
									}
									
									kko_kind = "R"
									if s.EqualFold(mst_sent_voucher.String, "V") {
										switch rcsKind.String {
											case "RCSSMS":
												amount = cprice.V_price_rcs_sms.Float64
												payback = cprice.V_price_rcs_sms.Float64 - cprice.P_price_rcs_sms.Float64
												admin_amt = cprice.B_price_rcs_sms.Float64
												memo = "RCS SMS,바우처"						
											case "RCSLMS":
												amount = cprice.V_price_rcs.Float64
												payback = cprice.V_price_rcs.Float64 - cprice.P_price_rcs.Float64
												admin_amt = cprice.B_price_rcs.Float64
												memo = "RCS LMS,바우처"						
											case "RCSMMS":
												amount = cprice.V_price_rcs_mms.Float64
												payback = cprice.V_price_rcs_mms.Float64 - cprice.P_price_rcs_mms.Float64
												admin_amt = cprice.B_price_rcs_mms.Float64
												memo = "RCS MMS,바우처"						
											case "RCSTMPL":
												amount = cprice.V_price_rcs_tem.Float64
												payback = cprice.V_price_rcs_tem.Float64 - cprice.P_price_rcs_tem.Float64
												admin_amt = cprice.B_price_rcs_tem.Float64
												memo = "RCS TMPL,바우처"						
										}									
									} else {
										switch rcsKind.String {
											case "RCSSMS":
												amount = cprice.C_price_rcs_sms.Float64
												payback = cprice.C_price_rcs_sms.Float64 - cprice.P_price_rcs_sms.Float64
												admin_amt = cprice.B_price_rcs_sms.Float64
												if s.EqualFold(mst_sent_voucher.String, "B") {
													memo = "RCS SMS,보너스"
												} else {
													memo = "RCS SMS"
												}										
											case "RCSLMS":
												amount = cprice.C_price_rcs.Float64
												payback = cprice.C_price_rcs.Float64 - cprice.P_price_rcs.Float64
												admin_amt = cprice.B_price_rcs.Float64
												if s.EqualFold(mst_sent_voucher.String, "B") {
													memo = "RCS LMS,보너스"
												} else {
													memo = "RCS LMS"
												}										
											case "RCSMMS":
												amount = cprice.C_price_rcs_mms.Float64
												payback = cprice.C_price_rcs_mms.Float64 - cprice.P_price_rcs_mms.Float64
												admin_amt = cprice.B_price_rcs_mms.Float64
												if s.EqualFold(mst_sent_voucher.String, "B") {
													memo = "RCS MMS,보너스"
												} else {
													memo = "RCS MMS"
												}										
											case "RCSTMPL":
												amount = cprice.C_price_rcs_tem.Float64
												payback = cprice.C_price_rcs_tem.Float64 - cprice.P_price_rcs_tem.Float64
												admin_amt = cprice.B_price_rcs_tem.Float64
												if s.EqualFold(mst_sent_voucher.String, "B") {
													memo = "RCS TMPL,보너스"
												} else {
													memo = "RCS TMPL"
												}										
										}
									}
									
									//bodyBytes, _ := json.Marshal(body)
									//bodyJson := string(bodyBytes)
									
									rcsStrs = append(rcsStrs, "(?,?,'0',?,'rcs',?,?,?,?,1,'0',null,1,?,?,?)")
									rcsValues = append(rcsValues, msgid.String)
									rcsValues = append(rcsValues, phnstr)
									rcsValues = append(rcsValues, remark4.String)
									rcsValues = append(rcsValues, s.Replace(sms_sender.String, "-", "", -1)) // 발신자 전화 번호
									rcsValues = append(rcsValues, "dhn2021g") // Agency ID -> 대행사라는데...나중에 다른걸로 변경 해야 할지..
									rcsValues = append(rcsValues, msgbaseid)
									rcsValues = append(rcsValues, srv_type)
									rcsValues = append(rcsValues, rcsBody.String)
									rcsValues = append(rcsValues, rcsBtns.String)
									rcsValues = append(rcsValues, rcsBrandkey.String)
									
									//stdlog.Println("RCS msg Infor ", rcsBtns.String)
									
 
								case "SMART":
									cb_msg_message_type = "sm"
									cb_msg_code = "SMT"
									//lms_smtcnt++

									if s.EqualFold(msgtype, "SMS") {
										ossmsStrs = append(ossmsStrs, "(?,?,?,?,?,null,?,?,?)")
										ossmsValues = append(ossmsValues, sms_sender)
										ossmsValues = append(ossmsValues, phnstr)
										ossmsValues = append(ossmsValues, msg_sms)
										ossmsValues = append(ossmsValues, "")
										if s.EqualFold(reserve_dt.String, "00000000000000") {
											ossmsValues = append(ossmsValues, sql.NullString{})
										} else {
											ossmsValues = append(ossmsValues, reserve_dt)
										}
										ossmsValues = append(ossmsValues, "0")
										ossmsValues = append(ossmsValues, remark4)
										ossmsValues = append(ossmsValues, msgid)

										kko_kind = "P"
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
									} else if s.EqualFold(msgtype, "LMS") {
										osmmsStrs = append(osmmsStrs, "(?,?,?,?,?,?,null,?,?,?,?,?,?)")
										osmmsValues = append(osmmsValues, remark4)
										osmmsValues = append(osmmsValues, sms_sender)
										osmmsValues = append(osmmsValues, phnstr)
										osmmsValues = append(osmmsValues, sms_lms_tit)
										osmmsValues = append(osmmsValues, msg_sms)
										if s.EqualFold(reserve_dt.String, "00000000000000") {
											osmmsValues = append(osmmsValues, sql.NullString{})
										} else {
											osmmsValues = append(osmmsValues, reserve_dt)
										}
										osmmsValues = append(osmmsValues, "0")
										osmmsValues = append(osmmsValues, mms_file1)
										osmmsValues = append(osmmsValues, mms_file2)
										osmmsValues = append(osmmsValues, mms_file3)
										osmmsValues = append(osmmsValues, remark4)
										osmmsValues = append(osmmsValues, msgid)

										if len(mms_file1.String) <= 0 {
											kko_kind = "P"
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
											kko_kind = "P"
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
									}
								case "GREEN_SHOT":
									cb_msg_message_type = "gs"
									cb_msg_code = "GRS"
									//lms_smtcnt++

									if s.EqualFold(msgtype, "SMS") {
										nnsmsStrs = append(nnsmsStrs, "(?,?,?,?,?,?,?,?,?,'Y')")
										nnsmsValues = append(nnsmsValues, sms_sender)
										nnsmsValues = append(nnsmsValues, phnstr)
										nnsmsValues = append(nnsmsValues, msg_sms)
										nnsmsValues = append(nnsmsValues, time.Now().Format("2006-01-02 15:04:05"))
										nnsmsValues = append(nnsmsValues, "0")
										nnsmsValues = append(nnsmsValues, "0")
										nnsmsValues = append(nnsmsValues, msgid)
										nnsmsValues = append(nnsmsValues, remark4)
										nnsmsValues = append(nnsmsValues, "302190001")

										kko_kind = "P"
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
									} else if s.EqualFold(msgtype, "LMS") {
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

										nnmmsValues = append(nnmmsValues, sms_sender)
										nnmmsValues = append(nnmmsValues, phnstr)
										nnmmsValues = append(nnmmsValues, sms_lms_tit)
										nnmmsValues = append(nnmmsValues, msg_sms)
										nnmmsValues = append(nnmmsValues, time.Now().Format("2006-01-02 15:04:05"))
										nnmmsValues = append(nnmmsValues, "0")
										nnmmsValues = append(nnmmsValues, filecnt)
										nnmmsValues = append(nnmmsValues, mms_file1)
										nnmmsValues = append(nnmmsValues, mms_file2)
										nnmmsValues = append(nnmmsValues, mms_file3)
										nnmmsValues = append(nnmmsValues, msgid)
										nnmmsValues = append(nnmmsValues, remark4)
										nnmmsValues = append(nnmmsValues, "302190001")

										if len(mms_file1.String) <= 0 {
											kko_kind = "P"
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
											kko_kind = "P"
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
									}
								case "GREEN_SHOT_G":
									cb_msg_message_type = "nl"
									cb_msg_code = "GRS"

									if s.HasPrefix(sms_sender.String, "010") {
										if s.EqualFold(msgtype, "SMS") {
											nnsmsStrs = append(nnsmsStrs, "(?,?,?,?,?,?,?,?,?,'Y')")
											nnsmsValues = append(nnsmsValues, sms_sender)
											nnsmsValues = append(nnsmsValues, phnstr)
											nnsmsValues = append(nnsmsValues, msg_sms)
											nnsmsValues = append(nnsmsValues, time.Now().Format("2006-01-02 15:04:05"))
											nnsmsValues = append(nnsmsValues, "0")
											nnsmsValues = append(nnsmsValues, "0")
											nnsmsValues = append(nnsmsValues, msgid)
											nnsmsValues = append(nnsmsValues, remark4)
											nnsmsValues = append(nnsmsValues, "302190001")

											kko_kind = "P"
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
										} else if s.EqualFold(msgtype, "LMS") {
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

											nnmmsValues = append(nnmmsValues, sms_sender)
											nnmmsValues = append(nnmmsValues, phnstr)
											nnmmsValues = append(nnmmsValues, sms_lms_tit)
											nnmmsValues = append(nnmmsValues, msg_sms)
											nnmmsValues = append(nnmmsValues, time.Now().Format("2006-01-02 15:04:05"))
											nnmmsValues = append(nnmmsValues, "0")
											nnmmsValues = append(nnmmsValues, filecnt)
											nnmmsValues = append(nnmmsValues, mms_file1)
											nnmmsValues = append(nnmmsValues, mms_file2)
											nnmmsValues = append(nnmmsValues, mms_file3)
											nnmmsValues = append(nnmmsValues, msgid)
											nnmmsValues = append(nnmmsValues, remark4)
											nnmmsValues = append(nnmmsValues, "302190001")

											if len(mms_file1.String) <= 0 {
												kko_kind = "P"
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
												kko_kind = "P"
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
										}
									} else {
										if s.EqualFold(msgtype, "SMS") {
											nnlpsmsStrs = append(nnlpsmsStrs, "(?,?,?,?,?,?,?,?,?,'Y')")
											nnlpsmsValues = append(nnlpsmsValues, sms_sender)
											nnlpsmsValues = append(nnlpsmsValues, phnstr)
											nnlpsmsValues = append(nnlpsmsValues, msg_sms)
											nnlpsmsValues = append(nnlpsmsValues, time.Now().Format("2006-01-02 15:04:05"))
											nnlpsmsValues = append(nnlpsmsValues, "0")
											nnlpsmsValues = append(nnlpsmsValues, "0")
											nnlpsmsValues = append(nnlpsmsValues, msgid)
											nnlpsmsValues = append(nnlpsmsValues, remark4)
											nnlpsmsValues = append(nnlpsmsValues, "302190001")

											kko_kind = "P"
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
										} else if s.EqualFold(msgtype, "LMS") {
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

											nnlpmmsStrs = append(nnlpmmsStrs, "( ?,?,?,?,?,?,?,?,?,?,?,?,?,'Y')")
											nnlpmmsValues = append(nnlpmmsValues, sms_sender)
											nnlpmmsValues = append(nnlpmmsValues, phnstr)
											nnlpmmsValues = append(nnlpmmsValues, sms_lms_tit)
											nnlpmmsValues = append(nnlpmmsValues, msg_sms)
											nnlpmmsValues = append(nnlpmmsValues, time.Now().Format("2006-01-02 15:04:05"))
											nnlpmmsValues = append(nnlpmmsValues, "0")
											nnlpmmsValues = append(nnlpmmsValues, filecnt)
											nnlpmmsValues = append(nnlpmmsValues, mms_file1)
											nnlpmmsValues = append(nnlpmmsValues, mms_file2)
											nnlpmmsValues = append(nnlpmmsValues, mms_file3)
											nnlpmmsValues = append(nnlpmmsValues, msgid)
											nnlpmmsValues = append(nnlpmmsValues, remark4)
											nnlpmmsValues = append(nnlpmmsValues, "302190001")

											if len(mms_file1.String) <= 0 {
												kko_kind = "P"
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
												kko_kind = "P"
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
										}
									}
								case "SMT_PHN":
									cb_msg_message_type = "sm"
									cb_msg_code = "SMT"

									// 폰문자는 자동 성공 처리 하기 위해 대기 차감함.
									lms_imccnt++
									mst_waitcnt--

									cb_msg_message = "폰문자 성공"
									result_flag = "Y"

									smtpStrs = append(smtpStrs, "(?,?,?,?,?,?,?,?,?,?,?,?,?,?,?)")
									smtpValues = append(smtpValues, "dhn7985")
									smtpValues = append(smtpValues, "")
									smtpValues = append(smtpValues, msgtype)
									smtpValues = append(smtpValues, sms_sender)
									smtpValues = append(smtpValues, sms_lms_tit)
									smtpValues = append(smtpValues, msg_sms)
									smtpValues = append(smtpValues, "")
									smtpValues = append(smtpValues, phnstr)
									smtpValues = append(smtpValues, "N")
									smtpValues = append(smtpValues, "")
									smtpValues = append(smtpValues, remark4)
									smtpValues = append(smtpValues, nowstr)
									smtpValues = append(smtpValues, "READY")
									if s.EqualFold(ph_msg_type, "wp1") {
										smtpValues = append(smtpValues, conf.WP1)
										smtpValues = append(smtpValues, ph_msg_type)
									} else {
										smtpValues = append(smtpValues, conf.WP2)
										smtpValues = append(smtpValues, ph_msg_type)
									}

									kko_kind = "P"
									if s.EqualFold(mst_sent_voucher.String, "V") {
										amount = cprice.V_price_imc.Float64
										payback = cprice.V_price_imc.Float64 - cprice.P_price_imc.Float64
										admin_amt = cprice.B_price_imc.Float64
										memo = "SMT PHN,바우처"
									} else {
										amount = cprice.C_price_imc.Float64
										payback = cprice.C_price_imc.Float64 - cprice.P_price_imc.Float64
										admin_amt = cprice.B_price_imc.Float64
										if s.EqualFold(mst_sent_voucher.String, "B") {
											memo = "SMT PHN,보너스"
										} else {
											memo = "SMT PHN"
										}										
										
									}


								}
							}

						} else {
						    // 2차 발신 없는 kakaotalk 처리.
						    isPayment = false  // 2차 발신 없으면 과금에서 제외
							if s.HasPrefix(s.ToUpper(message_type.String), "F") { // 친구톡 이면
								if s.EqualFold(message_type.String, "FC") {
									err_ftcscnt++
								} else if s.EqualFold(message_type.String, "FL") {
									err_ftilcnt++
								} else {
									if  len(image_url.String) <= 0 { // Image Url 이 null 이면 텍스트 친구톡
										err_ftcnt++
									} else {
										err_fticnt++
									}
								}
							} else if s.EqualFold(message_type.String, "at") || s.EqualFold(message_type.String, "al") { // 알림톡 이면
								err_atcnt++
							} else if s.EqualFold(message_type.String, "ai") { // 알림톡 이미지 이면
								err_atcnt++
							}

						}
					}
				}

				if isPass == false {
					// 알림톡 2 차 발신이 아니면 cb_msg 에 insert 처리
					msginsStrs = append(msginsStrs, "	(?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?)")
					msginsValues = append(msginsValues, msgid)
					msginsValues = append(msginsValues, ad_flag)
					msginsValues = append(msginsValues, button1)
					msginsValues = append(msginsValues, button2)
					msginsValues = append(msginsValues, button3)
					msginsValues = append(msginsValues, button4)
					msginsValues = append(msginsValues, button5)
					msginsValues = append(msginsValues, cb_msg_code)
					msginsValues = append(msginsValues, image_link)
					msginsValues = append(msginsValues, image_url)
					msginsValues = append(msginsValues, kind)
					msginsValues = append(msginsValues, cb_msg_message)
					msginsValues = append(msginsValues, cb_msg_message_type)
					msginsValues = append(msginsValues, "")
					msginsValues = append(msginsValues, "")
					msginsValues = append(msginsValues, only_sms)
					msginsValues = append(msginsValues, p_com)
					msginsValues = append(msginsValues, p_invoice)
					msginsValues = append(msginsValues, phn)
					msginsValues = append(msginsValues, profile)
					msginsValues = append(msginsValues, reg_dt)
					msginsValues = append(msginsValues, remark1)
					msginsValues = append(msginsValues, remark2)
					msginsValues = append(msginsValues, "")
					msginsValues = append(msginsValues, remark4)
					msginsValues = append(msginsValues, remark5)
					msginsValues = append(msginsValues, res_dt)
					msginsValues = append(msginsValues, reserve_dt)
					msginsValues = append(msginsValues, result_flag)
					msginsValues = append(msginsValues, s_code)
					msginsValues = append(msginsValues, sms_kind)
					msginsValues = append(msginsValues, sms_lms_tit)
					msginsValues = append(msginsValues, sms_sender)
					msginsValues = append(msginsValues, sync)
					msginsValues = append(msginsValues, tmpl_id)
					msginsValues = append(msginsValues, mem_userid)
					msginsValues = append(msginsValues, wide)

					if s.EqualFold(vancnt.String, "0") && isPayment { // 수신 거부 이면 금액 차감 에서 제외
						if amount <= 0 {
							amount = admin_amt
						}
						amtsStrs = append(amtsStrs, "(now(),?,?,?,?,?,?)")
						amtsValues = append(amtsValues, kko_kind)
						amtsValues = append(amtsValues, amount)
						amtsValues = append(amtsValues, memo)
						amtsValues = append(amtsValues, msgid.String)
						amtsValues = append(amtsValues, payback)
						amtsValues = append(amtsValues, admin_amt)
					}
				}

			}

			upmsgids = append(upmsgids, msgid.String)

			sendkey = remark4.String

			// Mst_id 별 Loop 중 1000 개 이상건을 DB Insert / Update 처리
			if len(msginsStrs) >= 1000 {
				stmt := fmt.Sprintf(insstr, s.Join(msginsStrs, ","))
				_, err := db.Exec(stmt, msginsValues...)

				if err != nil {
					errlog.Println("MSG Table Insert 처리 중 오류 발생 " + err.Error())
				}

				msginsStrs = nil
				msginsValues = nil
			}

			if len(ftlistsStrs) >= 1000 {
				stmt := fmt.Sprintf("insert IGNORE into cb_friend_list(mem_id, phn, last_send_date) values %s", s.Join(ftlistsStrs, ","))
				_, err := db.Exec(stmt, ftlistsValues...)

				if err != nil {
					errlog.Println("FT List Table Insert 처리 중 오류 발생 " + err.Error())
				}

				ftlistsStrs = nil
				ftlistsValues = nil
			}

			if len(upmsgids) >= 1000 {

				var commastr = "delete from " + conf.RESULTTABLE + " where MSGID in ("

				for i := 1; i < len(upmsgids); i++ {
					commastr = commastr + "?,"
				}

				commastr = commastr + "?)"

				_, err1 := db.Exec(commastr, upmsgids...)

				if err1 != nil {
					errlog.Println("Result Table Update 처리 중 오류 발생 ")
				}

				upmsgids = nil
			}

			if len(amtsStrs) >= 1000 {
				stmt := fmt.Sprintf(amtinsstr, s.Join(amtsStrs, ","))
				_, err := db.Exec(stmt, amtsValues...)

				if err != nil {
					errlog.Println("AMT Table Insert 처리 중 오류 발생 " + err.Error())
				}

				amtsStrs = nil
				amtsValues = nil
			}

			if len(nanoitStrs) >= 1000 {
				stmt := fmt.Sprintf("insert into cb_nanoit_msg(msg_type, remark4, phn, cb_msg_id) values %s", s.Join(nanoitStrs, ","))
				_, err := db.Exec(stmt, nanoitValues...)

				if err != nil {
					errlog.Println("Nano it Table Insert 처리 중 오류 발생 " + err.Error())
				}

				nanoitStrs = nil
				nanoitValues = nil
			}

			if len(ossmsStrs) >= 1000 {
				stmt := fmt.Sprintf("insert into OShotSMS(Sender,Receiver,Msg,URL,ReserveDT,TimeoutDT,SendResult,mst_id,cb_msg_id ) values %s", s.Join(ossmsStrs, ","))
				_, err := db.Exec(stmt, ossmsValues...)

				if err != nil {
					errlog.Println("스마트미 SMS Table Insert 처리 중 오류 발생 " + err.Error())
				}

				ossmsStrs = nil
				ossmsValues = nil
			}

			if len(osmmsStrs) >= 1000 {
				stmt := fmt.Sprintf("insert into OShotMMS(MsgGroupID,Sender,Receiver,Subject,Msg,ReserveDT,TimeoutDT,SendResult,File_Path1,File_Path2,File_Path3,mst_id,cb_msg_id ) values %s", s.Join(osmmsStrs, ","))
				_, err := db.Exec(stmt, osmmsValues...)

				if err != nil {
					errlog.Println("스마트미 LMS Table Insert 처리 중 오류 발생 " + err.Error())
				}

				osmmsStrs = nil
				osmmsValues = nil
			}

			if len(nnsmsStrs) >= 1000 {
				stmt := fmt.Sprintf("insert into SMS_MSG(TR_CALLBACK,TR_PHONE,TR_MSG,TR_SENDDATE,TR_SENDSTAT,TR_MSGTYPE,TR_ETC9,TR_ETC10,TR_IDENTIFICATION_CODE,TR_ETC8) values %s", s.Join(nnsmsStrs, ","))
				_, err := db.Exec(stmt, nnsmsValues...)

				if err != nil {
					errlog.Println("나노 SMS Table Insert 처리 중 오류 발생 " + err.Error())
				}

				nnsmsStrs = nil
				nnsmsValues = nil
			}

			if len(nnmmsStrs) >= 1000 {
				stmt := fmt.Sprintf("insert into MMS_MSG(CALLBACK,PHONE,SUBJECT,MSG,REQDATE,STATUS,FILE_CNT,FILE_PATH1,FILE_PATH2,FILE_PATH3,ETC9,ETC10,IDENTIFICATION_CODE,ETC8) values %s", s.Join(nnmmsStrs, ","))
				_, err := db.Exec(stmt, nnmmsValues...)

				if err != nil {
					errlog.Println("나노 LMS Table Insert 처리 중 오류 발생 " + err.Error())
				}

				nnmmsStrs = nil
				nnmmsValues = nil
			}

			if len(nnlpsmsStrs) >= 1000 {
				stmt := fmt.Sprintf("insert into SMS_MSG_G(TR_CALLBACK,TR_PHONE,TR_MSG,TR_SENDDATE,TR_SENDSTAT,TR_MSGTYPE,TR_ETC9,TR_ETC10,TR_IDENTIFICATION_CODE,TR_ETC8) values %s", s.Join(nnlpsmsStrs, ","))
				_, err := db.Exec(stmt, nnlpsmsValues...)

				if err != nil {
					errlog.Println("나노 SMS Table Insert 처리 중 오류 발생 " + err.Error())
				}

				nnlpsmsStrs = nil
				nnlpsmsValues = nil
			}

			if len(nnlpmmsStrs) >= 1000 {
				stmt := fmt.Sprintf("insert into MMS_MSG_G(CALLBACK,PHONE,SUBJECT,MSG,REQDATE,STATUS,FILE_CNT,FILE_PATH1,FILE_PATH2,FILE_PATH3,ETC9,ETC10,IDENTIFICATION_CODE,ETC8) values %s", s.Join(nnlpmmsStrs, ","))
				_, err := db.Exec(stmt, nnlpmmsValues...)

				if err != nil {
					errlog.Println("나노 LMS Table Insert 처리 중 오류 발생 " + err.Error())
				}

				nnlpmmsStrs = nil
				nnlpmmsValues = nil
			}

			if len(rcsStrs) >= 1000 {
				stmt := fmt.Sprintf("insert into RCS_MESSAGE(msg_id,user_contact ,schedule_type,msg_group_id,msg_service_type ,chatbot_id,agency_id ,messagebase_id ,service_type ,expiry_option ,header  ,footer  ,copy_allowed ,body ,buttons,brand_key) values %s", s.Join(rcsStrs, ","))
				_, err := db.Exec(stmt, rcsValues...)

				if err != nil {
					errlog.Println("스마트미 LMS Table Insert 처리 중 오류 발생 " + err.Error())
				}

				rcsStrs = nil
				rcsValues = nil
			}

			if len(smtpStrs) >= 1000 {
				stmt := fmt.Sprintf("insert into SMT_SEND(user_id,sub_id,send_type,sender,subject,message,file_url,receivers,reserve_yn,reserve_dt,request_id,request_dt,send_status, user_acct_key,user_acct_type) values %s", s.Join(smtpStrs, ","))
				_, err := db.Exec(stmt, smtpValues...)

				if err != nil {
					errlog.Println("스마트미 폰문자 Table Insert 처리 중 오류 발생 " + err.Error())
				}

				smtpStrs = nil
				smtpValues = nil
			}

		}
		// mst_id 별 Loop 끝

		// DB Insert / Update 처리
		if len(msginsValues) > 0 {

			stmt := fmt.Sprintf(insstr, s.Join(msginsStrs, ","))
			_, err := db.Exec(stmt, msginsValues...)

			if err != nil {
				errlog.Println("MSG Table Insert 처리 중 오류 발생 " + err.Error())
			}
		}

		if len(upmsgids) > 0 {
			var commastr = "delete from " + conf.RESULTTABLE + " where MSGID in ("
			for i := 1; i < len(upmsgids); i++ {
				commastr = commastr + "?,"
			}

			commastr = commastr + "?)"

			_, err1 := db.Exec(commastr, upmsgids...)

			if err1 != nil {
				errlog.Println("Result Table Update 처리 중 오류 발생 ")
			}
		}

		if len(ftlistsStrs) > 0 {
			stmt := fmt.Sprintf("insert IGNORE into cb_friend_list(mem_id, phn, last_send_date) values %s", s.Join(ftlistsStrs, ","))
			_, err := db.Exec(stmt, ftlistsValues...)

			if err != nil {
				errlog.Println("FT List Table Insert 처리 중 오류 발생 " + err.Error())
			}

			ftlistsStrs = nil
			ftlistsValues = nil
		}

		if len(atinsids) > 0 {
			var copystr = "update " + conf.REQTABLE2 + " set remark3 = 'Y' where MSGID in ("

			for i := 1; i < len(atinsids); i++ {
				copystr = copystr + "?,"
			}

			copystr = copystr + "?)"

			_, err1 := db.Exec(copystr, atinsids...)

			if err1 != nil {
				errlog.Println("2ND 테이블에서 복사 처리 중 오류 발생 ")
				errlog.Println(err1)
				errlog.Println(copystr)
			} else {
				errlog.Println("2ND 테이블에서 복사 처리 완료 : ", len(atinsids))
			}
			atinsids = nil
		}

		if len(amtsStrs) > 0 {
			stmt := fmt.Sprintf(amtinsstr, s.Join(amtsStrs, ","))
			_, err := db.Exec(stmt, amtsValues...)

			if err != nil {
				errlog.Println("AMT Table Insert 처리 중 오류 발생 " + err.Error())
			}
		}

		if len(nanoitStrs) > 0 {
			stmt := fmt.Sprintf("insert into cb_nanoit_msg(msg_type, remark4, phn, cb_msg_id) values %s", s.Join(nanoitStrs, ","))
			_, err := db.Exec(stmt, nanoitValues...)

			if err != nil {
				errlog.Println("Nano it Table Insert 처리 중 오류 발생 " + err.Error())
			}

		}

		if len(ossmsStrs) > 0 {
			stmt := fmt.Sprintf("insert into OShotSMS(Sender,Receiver,Msg,URL,ReserveDT,TimeoutDT,SendResult,mst_id,cb_msg_id ) values %s", s.Join(ossmsStrs, ","))
			_, err := db.Exec(stmt, ossmsValues...)

			if err != nil {
				errlog.Println("스마트미 SMS Table Insert 처리 중 오류 발생 " + err.Error())
			}

		}

		if len(osmmsStrs) > 0 {
			stmt := fmt.Sprintf("insert into OShotMMS(MsgGroupID,Sender,Receiver,Subject,Msg,ReserveDT,TimeoutDT,SendResult,File_Path1,File_Path2,File_Path3,mst_id,cb_msg_id ) values %s", s.Join(osmmsStrs, ","))
			_, err := db.Exec(stmt, osmmsValues...)

			if err != nil {
				errlog.Println("스마트미 LMS Table Insert 처리 중 오류 발생 " + err.Error())
			}

		}

		if len(nnsmsStrs) > 0 {
			stmt := fmt.Sprintf("insert into SMS_MSG(TR_CALLBACK,TR_PHONE,TR_MSG,TR_SENDDATE,TR_SENDSTAT,TR_MSGTYPE,TR_ETC9,TR_ETC10,TR_IDENTIFICATION_CODE,TR_ETC8) values %s", s.Join(nnsmsStrs, ","))
			_, err := db.Exec(stmt, nnsmsValues...)

			if err != nil {
				errlog.Println("나노 SMS Table Insert 처리 중 오류 발생 " + err.Error())
			}
		}

		if len(nnmmsStrs) > 0 {
			stmt := fmt.Sprintf("insert into MMS_MSG(CALLBACK,PHONE,SUBJECT,MSG,REQDATE,STATUS,FILE_CNT,FILE_PATH1,FILE_PATH2,FILE_PATH3,ETC9,ETC10,IDENTIFICATION_CODE,ETC8) values %s", s.Join(nnmmsStrs, ","))
			_, err := db.Exec(stmt, nnmmsValues...)

			if err != nil {
				errlog.Println("나노 LMS Table Insert 처리 중 오류 발생 " + err.Error())
			}
		}

		if len(nnlpsmsStrs) > 0 {
			stmt := fmt.Sprintf("insert into SMS_MSG_G(TR_CALLBACK,TR_PHONE,TR_MSG,TR_SENDDATE,TR_SENDSTAT,TR_MSGTYPE,TR_ETC9,TR_ETC10,TR_IDENTIFICATION_CODE,TR_ETC8) values %s", s.Join(nnlpsmsStrs, ","))
			_, err := db.Exec(stmt, nnlpsmsValues...)

			if err != nil {
				errlog.Println("나노 저가망 SMS Table Insert 처리 중 오류 발생 " + err.Error())
			}
		}

		if len(nnlpmmsStrs) > 0 {
			stmt := fmt.Sprintf("insert into MMS_MSG_G(CALLBACK,PHONE,SUBJECT,MSG,REQDATE,STATUS,FILE_CNT,FILE_PATH1,FILE_PATH2,FILE_PATH3,ETC9,ETC10,IDENTIFICATION_CODE,ETC8) values %s", s.Join(nnlpmmsStrs, ","))
			_, err := db.Exec(stmt, nnlpmmsValues...)

			if err != nil {
				errlog.Println("나노 저가망 LMS Table Insert 처리 중 오류 발생 " + err.Error())
			}
		}
	
		if len(rcsStrs) > 0 {
			stmt := fmt.Sprintf("insert into RCS_MESSAGE(msg_id,user_contact ,schedule_type,msg_group_id,msg_service_type ,chatbot_id,agency_id ,messagebase_id ,service_type ,expiry_option ,header  ,footer  ,copy_allowed ,body ,buttons, brand_key) values %s", s.Join(rcsStrs, ","))
			_, err := db.Exec(stmt, rcsValues...)

			if err != nil {
				errlog.Println("RCS Table Insert 처리 중 오류 발생 " + err.Error())
			}
		}
			
		if len(smtpStrs) > 0 {
			stmt := fmt.Sprintf("insert into SMT_SEND(user_id,sub_id,send_type,sender,subject,message,file_url,receivers,reserve_yn,reserve_dt,request_id,request_dt,send_status, user_acct_key, user_acct_type) values %s", s.Join(smtpStrs, ","))
			_, err := db.Exec(stmt, smtpValues...)

			if err != nil {
				errlog.Println("스마트미 폰문자 Table Insert 처리 중 오류 발생 " + err.Error())
			}
		}

		if len(remark4.String) > 0 {
			var cntupdate = `update cb_wt_msg_sent 
				                   set mst_ft = ifnull(mst_ft,0) + ?
								    , mst_ft_img = ifnull(mst_ft_img,0) + ?
									, mst_at = ifnull(mst_at,0) + ? 
									, mst_phn = ifnull(mst_phn,0) + ? 
									, mst_015 = ifnull(mst_015,0) + ? 
									, mst_grs = ifnull(mst_grs,0) + ? 
									, mst_nas = ifnull(mst_nas,0) + ? 
									, mst_smt = ifnull(mst_smt,0) + ? 
									, mst_imc = ifnull(mst_imc,0) + ? 
									, mst_cs = ifnull(mst_cs,0) + ? 
									, mst_il = ifnull(mst_il,0) + ? 
									, mst_err_ft = ifnull(mst_err_ft,0) + ?
								    , mst_err_ft_img = ifnull(mst_err_ft_img,0) + ?
									, mst_err_at = ifnull(mst_err_at,0) + ? 
									, mst_err_phn = ifnull(mst_err_phn,0) + ? 
									, mst_err_015 = ifnull(mst_err_015,0) + ? 
									, mst_err_grs = ifnull(mst_err_grs,0) + ? 
									, mst_err_nas = ifnull(mst_err_nas,0) + ? 
									, mst_err_smt = ifnull(mst_err_smt,0) + ? 
									, mst_err_imc = ifnull(mst_err_imc,0) + ?  
									, mst_err_rcs = ifnull(mst_err_rcs,0) + ?  
									, mst_err_cs = ifnull(mst_err_cs,0) + ?  
									, mst_err_il = ifnull(mst_err_il,0) + ?  
									, mst_wait = ifnull(mst_wait,0) + ?  
								where mst_id = ?`
			_, err := db.Exec(cntupdate, ftcnt, fticnt, atcnt, lms_phncnt, lms_015cnt, lms_grscnt, lms_nascnt, lms_smtcnt, lms_imccnt, ftcscnt, ftilcnt, err_ftcnt, err_fticnt, err_atcnt, err_phncnt, err_015cnt, err_grscnt, err_nascnt, err_smtcnt, err_imccnt, err_rcscnt, err_ftcscnt, err_ftilcnt, mst_waitcnt, sendkey)

			if err != nil {
				errlog.Println("WT_MSG_SENT 카카오 메세지 수량 처리 중 오류 발생 " + err.Error())
			}
		}

		if cnt > 0 {
			stdlog.Printf("( %s ) Result 처리 - %s : %d 건 처리 완료", startTime, ressendkey.String, cnt)
		}

		//err = tx.Commit()
		
		//if err != nil {
		//	errlog.Println("Result Proc Commit 처리 중 오류 발생 ")
		//}
		
		// 2차 알림톡 2일 지난건 삭제 함.
		db.Exec("delete a from " + conf.REQTABLE2 + " a where  ( ( a.reserve_dt < DATE_FORMAT(ADDDATE(now(), INTERVAL -2 DAY), '%Y%m%d%H%i%S') and a.reserve_dt <> '00000000000000') or ( a.REG_DT < ADDDATE(now(), INTERVAL -2 DAY) and a.reserve_dt = '00000000000000'))");
	}
	//}
}

func getSubstring(str string, cnt int) string {
	var temp, value string
	str1 := []rune(str)
	for _, char := range str1 {
		temp += fmt.Sprintf("%c", char)

		if len(temp) > cnt {
			return value
		} else {
			value += fmt.Sprintf("%c", char)
		}
	}
	return value
}

