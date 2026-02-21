package tblreqprocess

import (
	"context"
	"database/sql"
	"fmt"
	"strconv"
	s "strings"
	"sync"
	"time"

	"webagent/src/baseprice"
	"webagent/src/config"
	"webagent/src/databasepool"
)

// 나노 -> grs
// LGU -> nas
// 오샷 -> smt
// RCS -> rcs, nrc

var BMMESSAGETYPE = map[string]string{
	"B1": "TEXT",
	"B2": "IMAGE",
	"B3": "WIDE",
	"B4": "WIDE_ITEM_LIST",
	"B5": "CAROUSEL_FEED",
	"B6": "PREMIUM_VIDEO",
	"B7": "COMMERCE",
	"B8": "CAROUSEL_COMMERCE",
	"C1": "TEXT",
	"C2": "IMAGE",
	"C3": "WIDE",
	"C4": "WIDE_ITEM_LIST",
	"C5": "CAROUSEL_FEED",
	"C6": "PREMIUM_VIDEO",
	"C7": "COMMERCE",
	"C8": "CAROUSEL_COMMERCE",
	"D1": "TEXT",
	"D2": "IMAGE",
	"D3": "WIDE",
	"D4": "WIDE_ITEM_LIST",
	"D5": "CAROUSEL_FEED",
	"D6": "PREMIUM_VIDEO",
	"D7": "COMMERCE",
	"D8": "CAROUSEL_COMMERCE",
}

func Process(ctx context.Context) {
	config.Stdlog.Println("tblresultproc - 프로세스 시작")
	var wg sync.WaitGroup
	for {
		select {
		case <-ctx.Done():
			config.Stdlog.Println("tblresultproc - process가 15초 후에 종료")
			time.Sleep(15 * time.Second)
			config.Stdlog.Println("tblresultproc - process 종료 완료")
			return
		default:
			wg.Add(1)
			go resProcess(&wg)
			wg.Wait()
		}
	}
}

func resProcess(wg *sync.WaitGroup) {
	defer wg.Done()
	defer func() {
		if r := recover(); r != nil {
			config.Stdlog.Println("tblresultproc - panic 발생 원인 : ", r)
			if err, ok := r.(error); ok {
				if s.Contains(err.Error(), "connection refused") {
					for {
						config.Stdlog.Println("tblresultproc - send ping to DB")
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

	var msgid, ad_flag, button1, button2, button3, button4, button5, code, image_link, image_url, kind, message, message_type sql.NullString
	var msg, msg_sms, only_sms, p_com, p_invoice, phn, profile, reg_dt, remark1, remark2, remark3, remark4, remark5, res_dt, reserve_dt sql.NullString
	var result, s_code, sms_kind, sms_lms_tit, sms_sender, sync, tmpl_id, wide, supplement, price, currency_type, mem_userid, mem_id sql.NullString
	var mem_level, mem_phn_agent, mem_sms_agent, mem_2nd_send, mms_id, mst_type2, mst_type3, mst_2nd_alim, msgcnt, vancnt, mst_sent_voucher sql.NullString
	var mem_rcs_send, mem_rcs_send2, mem_rcs_send3, mem_rcs_send4 sql.NullString
	var mem_lp_flag, mms_file1, mms_file2, mms_file3 sql.NullString
	var cprice baseprice.BasePrice
	var msgtype, phnstr /*, mem_2nd_type*/ string
	var isPass bool
	var cnt int

	now := time.Now()
	reserveFormat := now.Format("20060102150405")

	msginsStrs := []string{}
	msginsValues := []interface{}{}

	atinsids := []interface{}{} // 알림톡 2차 발신시 2nd 테이블을 이용한 insert 용 id

	upmsgids := []interface{}{}

	amtsStrs := []string{}
	amtsValues := []interface{}{}

	ftlistsStrs := []string{}
	ftlistsValues := []interface{}{}

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

	lgusmsStrs := []string{}
	lgusmsValues := []interface{}{}

	lgummsStrs := []string{}
	lgummsValues := []interface{}{}

	tntsmsStrs := []string{}
	tntsmsValues := []interface{}{}

	tntmmsStrs := []string{}
	tntmmsValues := []interface{}{}

	jjsmsStrs := []string{}
	jjsmsValues := []interface{}{}

	jjmmsStrs := []string{}
	jjmmsValues := []interface{}{}

	var tickCnt sql.NullInt64
	var tickSql = `
		SELECT
			count(1) as cnt
		FROM 
			` + conf.RESULTTABLE + `
		WHERE
			(remark3 IS NOT null 
			and reserve_dt < ?)
			or
			(remark3 IS NOT null
			and kind = 'F')
		limit 1`

	cnterr := databasepool.DB.QueryRow(tickSql, reserveFormat).Scan(&tickCnt)

	if cnterr != nil && cnterr != sql.ErrNoRows {
		config.Stdlog.Println("tblresultproc -", conf.RESULTTABLE, "Table - select error : "+cnterr.Error())
		time.Sleep(10 * time.Second)
	} else {
		if tickCnt.Int64 <= 0 {
			time.Sleep(500 * time.Millisecond)
			return
		}
	}

	var resquery = `
		SELECT DISTINCT
			trr.REMARK4 AS ressendkey,
			cm.mem_userid as username,
			cm.mem_id as user_mem_id,
			trr.sms_sender gsms_sender
		FROM 
			` + conf.RESULTTABLE + ` trr
		INNER JOIN 
			cb_member cm ON trr.remark2 = cm.mem_id
		WHERE
			(trr.remark3 IS NOT null
			and trr.reserve_dt < ?)
			or
			(trr.remark3 IS NOT null
			and trr.kind = 'F')`

	resrows, err := db.Query(resquery, reserveFormat)

	if err != nil {
		errlog.Println("tblresultproc - Result Table 처리 중 오류 발생")
		errlog.Println(err)
		time.Sleep(500 * time.Millisecond)
		panic(err)
	}
	defer resrows.Close()

	var ressendkey sql.NullString
	var username sql.NullString
	var usermem_id sql.NullString
	var gsms_sender sql.NullString

	for resrows.Next() {
		resrows.Scan(&ressendkey, &username, &usermem_id, &gsms_sender)

		var resultsql string = `
			select SQL_NO_CACHE
				msgid
			  , ad_flag
			  , button1
			  , button2
			  , button3
			  , button4
			  , button5
			  , code
			  , image_link
			  , image_url
			  , kind
			  , message
			  , message_type
			  , msg
			  , msg_sms
			  , only_sms
			  , p_com
			  , p_invoice
			  , phn
			  , profile
			  , reg_dt
			  , remark1
			  , remark2
			  , remark3
			  , remark4
			  , remark5
			  , res_dt
			  , reserve_dt
			  , result
			  , s_code
			  , sms_kind
			  , sms_lms_tit
			  , sms_sender
			  , sync
			  , tmpl_id
			  , wide
			  , supplement
			  , price
			  , currency_type
			  , b.mem_userid
			  , b.mem_id
			  , b.mem_level
			  , b.mem_phn_agent
			  , b.mem_sms_agent
			  , b.mem_2nd_send
			  , b.mem_lp_flag
			  , (select mst_mms_content from cb_wt_msg_sent wms where wms.mst_id = a.remark4) as mms_id
			  , (select mst_type2 from cb_wt_msg_sent wms where wms.mst_id = a.remark4) as mst_type2
			  , (select mst_type3 from cb_wt_msg_sent wms where wms.mst_id = a.remark4) as mst_type3
			  , (select mst_2nd_alim from cb_wt_msg_sent wms where wms.mst_id = a.remark4) as mst_2nd_alim
			  , (select count(1) as msgcnt from cb_msg_` + username.String + ` cbmsg where cbmsg.msgid = a.msgid) as msgcnt
			  , (SELECT mi.origin1_path FROM cb_wt_msg_sent wms INNER join cb_mms_images mi ON wms.mst_mms_content = mi.mms_id  WHERE  length(mst_mms_content ) > 5 AND wms.mst_id = a.remark4 ) as mms_file1
			  , (SELECT mi.origin2_path FROM cb_wt_msg_sent wms INNER join cb_mms_images mi ON wms.mst_mms_content = mi.mms_id  WHERE  length(mst_mms_content ) > 5 AND wms.mst_id = a.remark4 ) as mms_file2
			  , (SELECT mi.origin3_path FROM cb_wt_msg_sent wms INNER join cb_mms_images mi ON wms.mst_mms_content = mi.mms_id  WHERE  length(mst_mms_content ) > 5 AND wms.mst_id = a.remark4 ) as mms_file3
			  , (select count(1) as vancnt from cb_block_lists cbl where cbl.sender = '` + gsms_sender.String + `' AND cbl.phn = CONCAT('0', SUBSTR(a.phn, 3,20))) AS vancnt
			  , (select mst_sent_voucher from cb_wt_msg_sent wms where wms.mst_id = a.remark4) as mst_sent_voucher
			  , b.mem_rcs_send
			  , b.mem_rcs_send2
			  , b.mem_rcs_send3
			  , b.mem_rcs_send4
			from
				` + conf.RESULTTABLE + ` a 
			inner join
				cb_member b on b.mem_id = a.REMARK2
			where
				a.remark3 is not null
				and a.remark4 = ?`

		rows, err := db.Query(resultsql, ressendkey.String)
		if err != nil {
			errlog.Println("tblresultproc - Result Table 처리 중 오류 발생")
			errlog.Println(err)
			time.Sleep(500 * time.Millisecond)
			panic(err)
		}
		defer rows.Close()

		cnt = 0

		msginsStrs = nil // cb_msg Table Insert 용
		msginsValues = nil

		upmsgids = nil // 처리된 message 처리를 위한 msgid 저장

		amtsStrs = nil // 요금 차감용
		amtsValues = nil

		ftlistsStrs = nil // 친구톡 성공 List 저장용
		ftlistsValues = nil

		ossmsStrs = nil //스마트미 SMS Table Insert 용
		ossmsValues = nil

		osmmsStrs = nil //스마트미 LMS/MMS Table Insert 용
		osmmsValues = nil

		lgusmsStrs = nil //LGU SMS Table Insert 용
		lgusmsValues = nil

		lgummsStrs = nil //LGU LMS/MMS Table Insert 용
		lgummsValues = nil

		nnsmsStrs = nil //나노 SMS Table Insert 용
		nnsmsValues = nil

		nnmmsStrs = nil //나노 LMS/MMS Table Insert 용
		nnmmsValues = nil

		nnlpsmsStrs = nil //나노 저가망 SMS Table Insert 용
		nnlpsmsValues = nil

		nnlpmmsStrs = nil //나노 저가망 LMS/MMS Table Insert 용
		nnlpmmsValues = nil

		rcsStrs = nil
		rcsValues = nil

		tntsmsStrs = nil //SMTNT SMS Table Insert 용
		tntsmsValues = nil

		tntmmsStrs = nil //SMTNT LMS/MMS Table Insert 용
		tntmmsValues = nil

		jjsmsStrs = nil //JJ SMS Table Insert 용
		jjsmsValues = nil

		jjmmsStrs = nil //JJ LMS/MMS Table Insert 용
		jjmmsValues = nil

		var insstr = ""
		var amtinsstr = ""

		// mst msg sent 에 발송 수량 Count 용 변수들
		var atcnt = 0
		var ftcnt = 0
		var ftncnt = 0
		var fticnt = 0
		var bcCnt = 0
		var ftilcnt = 0
		var ftcscnt = 0

		var err_atcnt = 0
		var err_ftcnt = 0
		var err_fticnt = 0
		var err_ftilcnt = 0
		var err_ftcscnt = 0
		var err_smtcnt = 0
		var err_rcscnt = 0

		var mst_waitcnt = 0

		var cb_msg_message_type = ""
		var cb_msg_code = ""
		var cb_msg_message = ""
		var result_flag = ""

		var sendkey = ""
		var mem_resend = ""

		var bmTargetingPriceV float64
		var bmTargetingPriceP float64
		var bmTargetingPriceB float64
		var bmTargetingPriceC float64

		var bmIsFriendPriceV float64
		var bmIsFriendPriceP float64
		var bmIsFriendPriceB float64
		var bmIsFriendPriceC float64

		cprice = baseprice.GetPrice(db, usermem_id.String, errlog)

		insstr = `
			insert IGNORE into
				cb_msg_` + username.String + `(
					MSGID
				  , AD_FLAG
				  , BUTTON1
				  , BUTTON2
				  , BUTTON3
				  , BUTTON4
				  , BUTTON5
				  , CODE
				  , IMAGE_LINK
				  , IMAGE_URL
				  , KIND
				  , MESSAGE
				  , MESSAGE_TYPE
				  , MSG
				  , MSG_SMS
				  , ONLY_SMS
				  , P_COM
				  , P_INVOICE
				  , PHN
				  , PROFILE
				  , REG_DT
				  , REMARK1
				  , REMARK2
				  , REMARK3
				  , REMARK4
				  , REMARK5
				  , RES_DT
				  , RESERVE_DT
				  , RESULT
				  , S_CODE
				  , SMS_KIND
				  , SMS_LMS_TIT
				  , SMS_SENDER
				  , SYNC
				  , TMPL_ID
				  , mem_userid
				  , wide)
			values %s`

		amtinsstr = `
			insert into
				cb_amt_` + username.String + `(
					amt_datetime
				  , amt_kind
				  , amt_amount
				  , amt_memo
				  , amt_reason
				  , amt_payback
				  , amt_admin)
			values %s`

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

			if cnt == 0 {
				sendkey = ressendkey.String
			}
			cnt++

			rows.Scan(
				&msgid,
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
				&mst_sent_voucher,
				&mem_rcs_send,
				&mem_rcs_send2,
				&mem_rcs_send3,
				&mem_rcs_send4)

			cb_msg_code = code.String
			cb_msg_message_type = message_type.String
			cb_msg_message = message.String

			if msgcnt.String == "0" {

				mem_resend = ""

				if s.EqualFold(mst_type2.String, "AI") || s.EqualFold(mst_type2.String, "AT") {
					if s.Contains(mst_type3.String, "s") {
						msgtype = "SMS"
					} else {
						msgtype = "LMS"
					}

					if s.Contains(mst_type3.String, "wa") && mem_lp_flag.String == "0" {
						mem_resend = "GREEN_SHOT"
					} else if s.Contains(mst_type3.String, "wa") && mem_lp_flag.String == "1" {
						mem_resend = "GREEN_SHOT_G"
					}

					if s.Contains(mst_type3.String, "wb") {
						mem_resend = "LGU"
					}

					if s.Contains(mst_type3.String, "wc") {
						mem_resend = "SMART"
					}

					if s.Contains(mst_type3.String, "wd") {
						mem_resend = "SMTNT"
					}

					if s.Contains(mst_type3.String, "we") {
						mem_resend = "JJ"
					}

					if s.Contains(mst_type3.String, "rc") {
						mem_resend = "RCS"

						if s.EqualFold(msgtype, "LMS") {
							mem_resend = "SMART"
						}
					}
				} else {
					if s.Contains(mst_type2.String, "s") {
						msgtype = "SMS"
					} else {
						msgtype = "LMS"
					}

					if s.Contains(mst_type2.String, "wa") && mem_lp_flag.String == "0" {
						mem_resend = "GREEN_SHOT"
					} else if s.Contains(mst_type2.String, "wa") && mem_lp_flag.String == "1" {
						mem_resend = "GREEN_SHOT_G"
					}

					if s.Contains(mst_type2.String, "wb") {
						mem_resend = "LGU"
					}

					if s.Contains(mst_type2.String, "wc") {
						mem_resend = "SMART"
					}

					if s.Contains(mst_type2.String, "wd") {
						mem_resend = "SMTNT"
					}

					if s.Contains(mst_type2.String, "we") {
						mem_resend = "JJ"
					}

					if s.Contains(mst_type2.String, "rc") {
						mem_resend = "RCS"
						if s.EqualFold(mst_type2.String, "rcm") {
							msgtype = "MMS"
						} else if s.EqualFold(mst_type2.String, "rct") {
							msgtype = "TEM"
						}
					}
				}

				phnstr = phn.String

				// 존재 의미가 명확하지 않은 처리
				if p_invoice.Valid && len(p_invoice.String) > 0 && s.EqualFold(message_type.String, "ph") {

					if len(mem_resend) <= 0 {
						mem_resend = p_invoice.String
					}

					switch mem_resend {
					case "SMART":
						cb_msg_message_type = "sm"
						break
					case "LGU":
						cb_msg_message_type = "lg"
						break
					case "GREEN_SHOT":
						cb_msg_message_type = "gs"
						break
					case "GREEN_SHOT_G":
						cb_msg_message_type = "nl"
						break
					case "SMTNT":
						cb_msg_message_type = "tn"
						break
					case "JJ":
						cb_msg_message_type = "jj"
						break
					case "RCS":
						cb_msg_message_type = "rc"
						break
					}
				} else {
					cb_msg_message_type = message_type.String
				}

				// BrandMsg 1차 브랜드메시지 2차 알림톡 조건 추가
				if (s.HasPrefix(s.ToUpper(message_type.String), "F") || s.HasPrefix(s.ToUpper(message_type.String), "B") || s.HasPrefix(s.ToUpper(message_type.String), "C")) && len(mst_2nd_alim.String) > 0 && mst_2nd_alim.String != "0" {

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
							errlog.Println("tblresultproc - 2ND 테이블에서 복사 처리 중 오류 발생 ")
							errlog.Println(err1)
							errlog.Println(copystr)
							time.Sleep(500 * time.Millisecond)
							panic(err1)
						} else {
							errlog.Println("tblresultproc - 2ND 테이블에서 복사 처리 완료 : ", len(atinsids))
						}
						atinsids = nil
					}

				}

				result_flag = result.String

				if isPass == false {
					// 발신 성공 시 금액 차감 처리
					if s.EqualFold(result.String, "Y") { // 카카오 메세지 성공시 차감 처리
						if s.HasPrefix(s.ToUpper(message_type.String), "F") { // 친구톡 이면
							if s.EqualFold(message_type.String, "FC") {
								ftcscnt++
								kko_kind = "F"

								admin_amt = cprice.B_price_ft_cs.Float64
								if s.EqualFold(mst_sent_voucher.String, "V") {
									amount = cprice.V_price_ft_cs.Float64
									payback = cprice.V_price_ft_cs.Float64 - cprice.P_price_ft_cs.Float64
									memo = "친구톡(CAROUSEL),바우처"
								} else {
									amount = cprice.C_price_ft_cs.Float64
									payback = cprice.C_price_ft_cs.Float64 - cprice.P_price_ft_cs.Float64
									if s.EqualFold(mst_sent_voucher.String, "B") {
										memo = "친구톡(CAROUSEL),보너스"
									} else {
										memo = "친구톡(CAROUSEL)"
									}
								}
							} else if s.EqualFold(message_type.String, "FL") {
								ftilcnt++
								kko_kind = "I"

								admin_amt = cprice.B_price_ft_il.Float64
								if s.EqualFold(mst_sent_voucher.String, "V") {
									amount = cprice.V_price_ft_il.Float64
									payback = cprice.V_price_ft_il.Float64 - cprice.P_price_ft_il.Float64
									memo = "친구톡(이미지리스트),바우처"
								} else {
									amount = cprice.C_price_ft_il.Float64
									payback = cprice.C_price_ft_il.Float64 - cprice.P_price_ft_il.Float64
									if s.EqualFold(mst_sent_voucher.String, "B") {
										memo = "친구톡(이미지리스트),보너스"
									} else {
										memo = "친구톡(이미지리스트)"
									}
								}
							} else if len(image_url.String) <= 1 { // Image Url 이 null 이면 텍스트 친구톡
								ftcnt++
								kko_kind = "F"

								admin_amt = cprice.B_price_ft.Float64
								if s.EqualFold(mst_sent_voucher.String, "V") {
									amount = cprice.V_price_ft.Float64
									payback = cprice.V_price_ft.Float64 - cprice.P_price_ft.Float64
									memo = "친구톡(텍스트),바우처"
								} else {
									amount = cprice.C_price_ft.Float64
									payback = cprice.C_price_ft.Float64 - cprice.P_price_ft.Float64
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
									admin_amt = cprice.B_price_ft_w_img.Float64
									if s.EqualFold(mst_sent_voucher.String, "V") {
										amount = cprice.V_price_ft_w_img.Float64
										payback = cprice.V_price_ft_w_img.Float64 - cprice.P_price_ft_w_img.Float64
										memo = "친구톡(와이드이미지),바우처"
									} else {
										amount = cprice.C_price_ft_w_img.Float64
										payback = cprice.C_price_ft_w_img.Float64 - cprice.P_price_ft_w_img.Float64
										if s.EqualFold(mst_sent_voucher.String, "B") {
											memo = "친구톡(와이드이미지),보너스"
										} else {
											memo = "친구톡(와이드이미지)"
										}
									}
								} else {
									admin_amt = cprice.B_price_ft_img.Float64
									if s.EqualFold(mst_sent_voucher.String, "V") {
										amount = cprice.V_price_ft_img.Float64
										payback = cprice.V_price_ft_img.Float64 - cprice.P_price_ft_img.Float64
										memo = "친구톡(이미지),바우처"
									} else {
										amount = cprice.C_price_ft_img.Float64
										payback = cprice.C_price_ft_img.Float64 - cprice.P_price_ft_img.Float64
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

							admin_amt = cprice.B_price_at.Float64
							if s.EqualFold(mst_sent_voucher.String, "V") {
								amount = cprice.V_price_at.Float64
								payback = cprice.V_price_at.Float64 - cprice.P_price_at.Float64
								memo = "알림톡(텍스트),바우처"
							} else {
								amount = cprice.C_price_at.Float64
								payback = cprice.C_price_at.Float64 - cprice.P_price_at.Float64
								if s.EqualFold(mst_sent_voucher.String, "B") {
									memo = "알림톡(텍스트),보너스"
								} else {
									memo = "알림톡(텍스트)"
								}

							}
						} else if s.EqualFold(message_type.String, "ai") { // 알림톡 이미지 이면
							atcnt++
							kko_kind = "E"

							admin_amt = cprice.B_price_at.Float64
							if s.EqualFold(mst_sent_voucher.String, "V") {
								amount = cprice.V_price_at.Float64
								payback = cprice.V_price_at.Float64 - cprice.P_price_at.Float64
								memo = "알림톡(이미지),바우처"
							} else {
								amount = cprice.C_price_at.Float64
								payback = cprice.C_price_at.Float64 - cprice.P_price_at.Float64
								if s.EqualFold(mst_sent_voucher.String, "B") {
									memo = "알림톡(이미지),보너스"
								} else {
									memo = "알림톡(이미지)"
								}

							}
						} else if s.HasPrefix(s.ToUpper(message_type.String), "B") { // BrandMsg 자유형 성공 차감
							ftcnt++
							kko_kind = "B"

							bmTargeting := s.ToUpper(kind.String)

							if bmTargeting == "M" {
								bmTargetingPriceV = cprice.V_price_bm_t_m.Float64
								bmTargetingPriceP = cprice.P_price_bm_t_m.Float64
								bmTargetingPriceB = cprice.B_price_bm_t_m.Float64
								bmTargetingPriceC = cprice.C_price_bm_t_m.Float64

								//TODO : 리스폰스로 친구인지 아닌지 판단할 수 있어야함
								//TODO : 일단은 친구가 아닌 값으로 넣어놓음
								bmIsFriendPriceV = cprice.V_price_bm_nf.Float64
								bmIsFriendPriceP = cprice.P_price_bm_nf.Float64
								bmIsFriendPriceB = cprice.B_price_bm_nf.Float64
								bmIsFriendPriceC = cprice.C_price_bm_nf.Float64
							} else if bmTargeting == "N" {
								bmTargetingPriceV = cprice.V_price_bm_t_n.Float64
								bmTargetingPriceP = cprice.P_price_bm_t_n.Float64
								bmTargetingPriceB = cprice.B_price_bm_t_n.Float64
								bmTargetingPriceC = cprice.C_price_bm_t_n.Float64

								bmIsFriendPriceV = cprice.V_price_bm_nf.Float64
								bmIsFriendPriceP = cprice.P_price_bm_nf.Float64
								bmIsFriendPriceB = cprice.B_price_bm_nf.Float64
								bmIsFriendPriceC = cprice.C_price_bm_nf.Float64

								ftncnt++
							} else if bmTargeting == "I" {
								bmTargetingPriceV = cprice.V_price_bm_t_i.Float64
								bmTargetingPriceP = cprice.P_price_bm_t_i.Float64
								bmTargetingPriceB = cprice.B_price_bm_t_i.Float64
								bmTargetingPriceC = cprice.C_price_bm_t_i.Float64

								bmIsFriendPriceV = cprice.V_price_bm_f.Float64
								bmIsFriendPriceP = cprice.P_price_bm_f.Float64
								bmIsFriendPriceB = cprice.B_price_bm_f.Float64
								bmIsFriendPriceC = cprice.C_price_bm_f.Float64
							} else if bmTargeting == "F" {
								bmTargetingPriceV = cprice.V_price_bm_t_f.Float64
								bmTargetingPriceP = cprice.P_price_bm_t_f.Float64
								bmTargetingPriceB = cprice.B_price_bm_t_f.Float64
								bmTargetingPriceC = cprice.C_price_bm_t_f.Float64

								bmIsFriendPriceV = cprice.V_price_bm_f.Float64
								bmIsFriendPriceP = cprice.P_price_bm_f.Float64
								bmIsFriendPriceB = cprice.B_price_bm_f.Float64
								bmIsFriendPriceC = cprice.C_price_bm_f.Float64
							} else {
								bmTargetingPriceV = cprice.V_price_bm_t_m.Float64
								bmTargetingPriceP = cprice.P_price_bm_t_m.Float64
								bmTargetingPriceB = cprice.B_price_bm_t_m.Float64
								bmTargetingPriceC = cprice.C_price_bm_t_m.Float64

								bmIsFriendPriceV = cprice.V_price_bm_nf.Float64
								bmIsFriendPriceP = cprice.P_price_bm_nf.Float64
								bmIsFriendPriceB = cprice.B_price_bm_nf.Float64
								bmIsFriendPriceC = cprice.C_price_bm_nf.Float64
							}

							memo = "브랜드자유형(" + BMMESSAGETYPE[s.ToUpper(message_type.String)] + ")"

							// BrandMsg 메세지 타입별 정산
							if s.EqualFold(message_type.String, "B1") {
								admin_amt = cprice.B_price_bm_b1.Float64 + bmTargetingPriceB + bmIsFriendPriceB
								if s.EqualFold(mst_sent_voucher.String, "V") {
									amount = cprice.V_price_bm_b1.Float64 + bmTargetingPriceV + bmIsFriendPriceV
									payback = (cprice.V_price_bm_b1.Float64 - cprice.P_price_bm_b1.Float64) + (bmTargetingPriceV - bmTargetingPriceP) + (bmIsFriendPriceV - bmIsFriendPriceP)
									memo = memo + ",바우처"
								} else {
									amount = cprice.C_price_bm_b1.Float64 + bmTargetingPriceC + bmIsFriendPriceC
									payback = (cprice.C_price_bm_b1.Float64 - cprice.P_price_bm_b1.Float64) + (bmTargetingPriceC - bmTargetingPriceP) + (bmIsFriendPriceC - bmIsFriendPriceP)
									if s.EqualFold(mst_sent_voucher.String, "B") {
										memo = memo + ",보너스"
									}
								}
							} else if s.EqualFold(message_type.String, "B2") {
								admin_amt = cprice.B_price_bm_b2.Float64 + bmTargetingPriceB + bmIsFriendPriceB
								if s.EqualFold(mst_sent_voucher.String, "V") {
									amount = cprice.V_price_bm_b2.Float64 + bmTargetingPriceV + bmIsFriendPriceV
									payback = (cprice.V_price_bm_b2.Float64 - cprice.P_price_bm_b2.Float64) + (bmTargetingPriceV - bmTargetingPriceP) + (bmIsFriendPriceV - bmIsFriendPriceP)
									memo = memo + ",바우처"
								} else {
									amount = cprice.C_price_bm_b2.Float64 + bmTargetingPriceC + bmIsFriendPriceC
									payback = (cprice.C_price_bm_b2.Float64 - cprice.P_price_bm_b2.Float64) + (bmTargetingPriceC - bmTargetingPriceP) + (bmIsFriendPriceC - bmIsFriendPriceP)
									if s.EqualFold(mst_sent_voucher.String, "B") {
										memo = memo + ",보너스"
									}
								}
							} else if s.EqualFold(message_type.String, "B3") {
								admin_amt = cprice.B_price_bm_b3.Float64 + bmTargetingPriceB + bmIsFriendPriceB
								if s.EqualFold(mst_sent_voucher.String, "V") {
									amount = cprice.V_price_bm_b3.Float64 + bmTargetingPriceV + bmIsFriendPriceV
									payback = (cprice.V_price_bm_b3.Float64 - cprice.P_price_bm_b3.Float64) + (bmTargetingPriceV - bmTargetingPriceP) + (bmIsFriendPriceV - bmIsFriendPriceP)
									memo = memo + ",바우처"
								} else {
									amount = cprice.C_price_bm_b3.Float64 + bmTargetingPriceC + bmIsFriendPriceC
									payback = (cprice.C_price_bm_b3.Float64 - cprice.P_price_bm_b3.Float64) + (bmTargetingPriceC - bmTargetingPriceP) + (bmIsFriendPriceC - bmIsFriendPriceP)
									if s.EqualFold(mst_sent_voucher.String, "B") {
										memo = memo + ",보너스"
									}
								}
							} else if s.EqualFold(message_type.String, "B4") {
								admin_amt = cprice.B_price_bm_b4.Float64 + bmTargetingPriceB + bmIsFriendPriceB
								if s.EqualFold(mst_sent_voucher.String, "V") {
									amount = cprice.V_price_bm_b4.Float64 + bmTargetingPriceV + bmIsFriendPriceV
									payback = (cprice.V_price_bm_b4.Float64 - cprice.P_price_bm_b4.Float64) + (bmTargetingPriceV - bmTargetingPriceP) + (bmIsFriendPriceV - bmIsFriendPriceP)
									memo = memo + ",바우처"
								} else {
									amount = cprice.C_price_bm_b4.Float64 + bmTargetingPriceC + bmIsFriendPriceC
									payback = (cprice.C_price_bm_b4.Float64 - cprice.P_price_bm_b4.Float64) + (bmTargetingPriceC - bmTargetingPriceP) + (bmIsFriendPriceC - bmIsFriendPriceP)
									if s.EqualFold(mst_sent_voucher.String, "B") {
										memo = memo + ",보너스"
									}
								}
							} else if s.EqualFold(message_type.String, "B5") {
								admin_amt = cprice.B_price_bm_b5.Float64 + bmTargetingPriceB + bmIsFriendPriceB
								if s.EqualFold(mst_sent_voucher.String, "V") {
									amount = cprice.V_price_bm_b5.Float64 + bmTargetingPriceV + bmIsFriendPriceV
									payback = (cprice.V_price_bm_b5.Float64 - cprice.P_price_bm_b5.Float64) + (bmTargetingPriceV - bmTargetingPriceP) + (bmIsFriendPriceV - bmIsFriendPriceP)
									memo = memo + ",바우처"
								} else {
									amount = cprice.C_price_bm_b5.Float64 + bmTargetingPriceC + bmIsFriendPriceC
									payback = (cprice.C_price_bm_b5.Float64 - cprice.P_price_bm_b5.Float64) + (bmTargetingPriceC - bmTargetingPriceP) + (bmIsFriendPriceC - bmIsFriendPriceP)
									if s.EqualFold(mst_sent_voucher.String, "B") {
										memo = memo + ",보너스"
									}
								}
							} else if s.EqualFold(message_type.String, "B6") {
								admin_amt = cprice.B_price_bm_b6.Float64 + bmTargetingPriceB + bmIsFriendPriceB
								if s.EqualFold(mst_sent_voucher.String, "V") {
									amount = cprice.V_price_bm_b6.Float64 + bmTargetingPriceV + bmIsFriendPriceV
									payback = (cprice.V_price_bm_b6.Float64 - cprice.P_price_bm_b6.Float64) + (bmTargetingPriceV - bmTargetingPriceP) + (bmIsFriendPriceV - bmIsFriendPriceP)
									memo = memo + ",바우처"
								} else {
									amount = cprice.C_price_bm_b6.Float64 + bmTargetingPriceC + bmIsFriendPriceC
									payback = (cprice.C_price_bm_b6.Float64 - cprice.P_price_bm_b6.Float64) + (bmTargetingPriceC - bmTargetingPriceP) + (bmIsFriendPriceC - bmIsFriendPriceP)
									if s.EqualFold(mst_sent_voucher.String, "B") {
										memo = memo + ",보너스"
									}
								}
							} else if s.EqualFold(message_type.String, "B7") {
								admin_amt = cprice.B_price_bm_b7.Float64 + bmTargetingPriceB + bmIsFriendPriceB
								if s.EqualFold(mst_sent_voucher.String, "V") {
									amount = cprice.V_price_bm_b7.Float64 + bmTargetingPriceV + bmIsFriendPriceV
									payback = (cprice.V_price_bm_b7.Float64 - cprice.P_price_bm_b7.Float64) + (bmTargetingPriceV - bmTargetingPriceP) + (bmIsFriendPriceV - bmIsFriendPriceP)
									memo = memo + ",바우처"
								} else {
									amount = cprice.C_price_bm_b7.Float64 + bmTargetingPriceC + bmIsFriendPriceC
									payback = (cprice.C_price_bm_b7.Float64 - cprice.P_price_bm_b7.Float64) + (bmTargetingPriceC - bmTargetingPriceP) + (bmIsFriendPriceC - bmIsFriendPriceP)
									if s.EqualFold(mst_sent_voucher.String, "B") {
										memo = memo + ",보너스"
									}
								}
							} else if s.EqualFold(message_type.String, "B8") {
								admin_amt = cprice.B_price_bm_b8.Float64 + bmTargetingPriceB + bmIsFriendPriceB
								if s.EqualFold(mst_sent_voucher.String, "V") {
									amount = cprice.V_price_bm_b8.Float64 + bmTargetingPriceV + bmIsFriendPriceV
									payback = (cprice.V_price_bm_b8.Float64 - cprice.P_price_bm_b8.Float64) + (bmTargetingPriceV - bmTargetingPriceP) + (bmIsFriendPriceV - bmIsFriendPriceP)
									memo = memo + ",바우처"
								} else {
									amount = cprice.C_price_bm_b8.Float64 + bmTargetingPriceC + bmIsFriendPriceC
									payback = (cprice.C_price_bm_b8.Float64 - cprice.P_price_bm_b8.Float64) + (bmTargetingPriceC - bmTargetingPriceP) + (bmIsFriendPriceC - bmIsFriendPriceP)
									if s.EqualFold(mst_sent_voucher.String, "B") {
										memo = memo + ",보너스"
									}
								}
							}
						} else if s.HasPrefix(s.ToUpper(message_type.String), "C") { // BrandMsg 기본형 성공 차감
							ftcnt++
							kko_kind = "C"

							bmTargeting := s.ToUpper(kind.String)

							if bmTargeting == "M" {
								bmTargetingPriceV = cprice.V_price_bm_t_m.Float64
								bmTargetingPriceP = cprice.P_price_bm_t_m.Float64
								bmTargetingPriceB = cprice.B_price_bm_t_m.Float64
								bmTargetingPriceC = cprice.C_price_bm_t_m.Float64

								//TODO : 리스폰스로 친구인지 아닌지 판단할 수 있어야함
								//TODO : 일단은 친구가 아닌 값으로 넣어놓음
								bmIsFriendPriceV = cprice.V_price_bm_nf.Float64
								bmIsFriendPriceP = cprice.P_price_bm_nf.Float64
								bmIsFriendPriceB = cprice.B_price_bm_nf.Float64
								bmIsFriendPriceC = cprice.C_price_bm_nf.Float64
							} else if bmTargeting == "N" {
								bmTargetingPriceV = cprice.V_price_bm_t_n.Float64
								bmTargetingPriceP = cprice.P_price_bm_t_n.Float64
								bmTargetingPriceB = cprice.B_price_bm_t_n.Float64
								bmTargetingPriceC = cprice.C_price_bm_t_n.Float64

								bmIsFriendPriceV = cprice.V_price_bm_nf.Float64
								bmIsFriendPriceP = cprice.P_price_bm_nf.Float64
								bmIsFriendPriceB = cprice.B_price_bm_nf.Float64
								bmIsFriendPriceC = cprice.C_price_bm_nf.Float64

								ftncnt++
							} else if bmTargeting == "I" {
								bmTargetingPriceV = cprice.V_price_bm_t_i.Float64
								bmTargetingPriceP = cprice.P_price_bm_t_i.Float64
								bmTargetingPriceB = cprice.B_price_bm_t_i.Float64
								bmTargetingPriceC = cprice.C_price_bm_t_i.Float64

								bmIsFriendPriceV = cprice.V_price_bm_f.Float64
								bmIsFriendPriceP = cprice.P_price_bm_f.Float64
								bmIsFriendPriceB = cprice.B_price_bm_f.Float64
								bmIsFriendPriceC = cprice.C_price_bm_f.Float64
							} else if bmTargeting == "F" {
								bmTargetingPriceV = cprice.V_price_bm_t_f.Float64
								bmTargetingPriceP = cprice.P_price_bm_t_f.Float64
								bmTargetingPriceB = cprice.B_price_bm_t_f.Float64
								bmTargetingPriceC = cprice.C_price_bm_t_f.Float64

								bmIsFriendPriceV = cprice.V_price_bm_f.Float64
								bmIsFriendPriceP = cprice.P_price_bm_f.Float64
								bmIsFriendPriceB = cprice.B_price_bm_f.Float64
								bmIsFriendPriceC = cprice.C_price_bm_f.Float64
							} else {
								bmTargetingPriceV = cprice.V_price_bm_t_m.Float64
								bmTargetingPriceP = cprice.P_price_bm_t_m.Float64
								bmTargetingPriceB = cprice.B_price_bm_t_m.Float64
								bmTargetingPriceC = cprice.C_price_bm_t_m.Float64

								bmIsFriendPriceV = cprice.V_price_bm_nf.Float64
								bmIsFriendPriceP = cprice.P_price_bm_nf.Float64
								bmIsFriendPriceB = cprice.B_price_bm_nf.Float64
								bmIsFriendPriceC = cprice.C_price_bm_nf.Float64
							}

							memo = "브랜드기본형(" + BMMESSAGETYPE[s.ToUpper(message_type.String)] + ")"

							// BrandMsg 메세지 타입별 정산
							if s.EqualFold(message_type.String, "C1") {
								admin_amt = cprice.B_price_bm_c1.Float64 + bmTargetingPriceB + bmIsFriendPriceB
								if s.EqualFold(mst_sent_voucher.String, "V") {
									amount = cprice.V_price_bm_c1.Float64 + bmTargetingPriceV + bmIsFriendPriceV
									payback = (cprice.V_price_bm_c1.Float64 - cprice.P_price_bm_c1.Float64) + (bmTargetingPriceV - bmTargetingPriceP) + (bmIsFriendPriceV - bmIsFriendPriceP)
									memo = memo + ",바우처"
								} else {
									amount = cprice.C_price_bm_c1.Float64 + bmTargetingPriceC
									payback = (cprice.C_price_bm_c1.Float64 - cprice.P_price_bm_c1.Float64) + (bmTargetingPriceC - bmTargetingPriceP) + (bmIsFriendPriceC - bmIsFriendPriceP)
									if s.EqualFold(mst_sent_voucher.String, "B") {
										memo = memo + ",보너스"
									}
								}
							} else if s.EqualFold(message_type.String, "C2") {
								admin_amt = cprice.B_price_bm_c2.Float64 + bmTargetingPriceB + bmIsFriendPriceB
								if s.EqualFold(mst_sent_voucher.String, "V") {
									amount = cprice.V_price_bm_c2.Float64 + bmTargetingPriceV + bmIsFriendPriceV
									payback = (cprice.V_price_bm_c2.Float64 - cprice.P_price_bm_c2.Float64) + (bmTargetingPriceV - bmTargetingPriceP) + (bmIsFriendPriceV - bmIsFriendPriceP)
									memo = memo + ",바우처"
								} else {
									amount = cprice.C_price_bm_c2.Float64 + bmTargetingPriceC + bmIsFriendPriceC
									payback = (cprice.C_price_bm_c2.Float64 - cprice.P_price_bm_c2.Float64) + (bmTargetingPriceC - bmTargetingPriceP) + (bmIsFriendPriceC - bmIsFriendPriceP)
									if s.EqualFold(mst_sent_voucher.String, "B") {
										memo = memo + ",보너스"
									}
								}
							} else if s.EqualFold(message_type.String, "C3") {
								admin_amt = cprice.B_price_bm_c3.Float64 + bmTargetingPriceB + bmIsFriendPriceB
								if s.EqualFold(mst_sent_voucher.String, "V") {
									amount = cprice.V_price_bm_c3.Float64 + bmTargetingPriceV + bmIsFriendPriceV
									payback = (cprice.V_price_bm_c3.Float64 - cprice.P_price_bm_c3.Float64) + (bmTargetingPriceV - bmTargetingPriceP) + (bmIsFriendPriceV - bmIsFriendPriceP)
									memo = memo + ",바우처"
								} else {
									amount = cprice.C_price_bm_c3.Float64 + bmTargetingPriceC + bmIsFriendPriceC
									payback = (cprice.C_price_bm_c3.Float64 - cprice.P_price_bm_c3.Float64) + (bmTargetingPriceC - bmTargetingPriceP) + (bmIsFriendPriceC - bmIsFriendPriceP)
									if s.EqualFold(mst_sent_voucher.String, "B") {
										memo = memo + ",보너스"
									}
								}
							} else if s.EqualFold(message_type.String, "C4") {
								admin_amt = cprice.B_price_bm_c4.Float64 + bmTargetingPriceB + bmIsFriendPriceB
								if s.EqualFold(mst_sent_voucher.String, "V") {
									amount = cprice.V_price_bm_c4.Float64 + bmTargetingPriceV + bmIsFriendPriceV
									payback = (cprice.V_price_bm_c4.Float64 - cprice.P_price_bm_c4.Float64) + (bmTargetingPriceV - bmTargetingPriceP) + (bmIsFriendPriceV - bmIsFriendPriceP)
									memo = memo + ",바우처"
								} else {
									amount = cprice.C_price_bm_c4.Float64 + bmTargetingPriceC + bmIsFriendPriceC
									payback = (cprice.C_price_bm_c4.Float64 - cprice.P_price_bm_c4.Float64) + (bmTargetingPriceC - bmTargetingPriceP) + (bmIsFriendPriceC - bmIsFriendPriceP)
									if s.EqualFold(mst_sent_voucher.String, "B") {
										memo = memo + ",보너스"
									}
								}
							} else if s.EqualFold(message_type.String, "C5") {
								admin_amt = cprice.B_price_bm_c5.Float64 + bmTargetingPriceB + bmIsFriendPriceB
								if s.EqualFold(mst_sent_voucher.String, "V") {
									amount = cprice.V_price_bm_c5.Float64 + bmTargetingPriceV + bmIsFriendPriceV
									payback = (cprice.V_price_bm_c5.Float64 - cprice.P_price_bm_c5.Float64) + (bmTargetingPriceV - bmTargetingPriceP) + (bmIsFriendPriceV - bmIsFriendPriceP)
									memo = memo + ",바우처"
								} else {
									amount = cprice.C_price_bm_c5.Float64 + bmTargetingPriceC + bmIsFriendPriceC
									payback = (cprice.C_price_bm_c5.Float64 - cprice.P_price_bm_c5.Float64) + (bmTargetingPriceC - bmTargetingPriceP) + (bmIsFriendPriceC - bmIsFriendPriceP)
									if s.EqualFold(mst_sent_voucher.String, "B") {
										memo = memo + ",보너스"
									}
								}
							} else if s.EqualFold(message_type.String, "C6") {
								admin_amt = cprice.B_price_bm_c6.Float64 + bmTargetingPriceB + bmIsFriendPriceB
								if s.EqualFold(mst_sent_voucher.String, "V") {
									amount = cprice.V_price_bm_c6.Float64 + bmTargetingPriceV + bmIsFriendPriceV
									payback = (cprice.V_price_bm_c6.Float64 - cprice.P_price_bm_c6.Float64) + (bmTargetingPriceV - bmTargetingPriceP) + (bmIsFriendPriceV - bmIsFriendPriceP)
									memo = memo + ",바우처"
								} else {
									amount = cprice.C_price_bm_c6.Float64 + bmTargetingPriceC + bmIsFriendPriceC
									payback = (cprice.C_price_bm_c6.Float64 - cprice.P_price_bm_c6.Float64) + (bmTargetingPriceC - bmTargetingPriceP) + (bmIsFriendPriceC - bmIsFriendPriceP)
									if s.EqualFold(mst_sent_voucher.String, "B") {
										memo = memo + ",보너스"
									}
								}
							} else if s.EqualFold(message_type.String, "C7") {
								admin_amt = cprice.B_price_bm_c7.Float64 + bmTargetingPriceB + bmIsFriendPriceB
								if s.EqualFold(mst_sent_voucher.String, "V") {
									amount = cprice.V_price_bm_c7.Float64 + bmTargetingPriceV + bmIsFriendPriceV
									payback = (cprice.V_price_bm_c7.Float64 - cprice.P_price_bm_c7.Float64) + (bmTargetingPriceV - bmTargetingPriceP) + (bmIsFriendPriceV - bmIsFriendPriceP)
									memo = memo + ",바우처"
								} else {
									amount = cprice.C_price_bm_c7.Float64 + bmTargetingPriceC + bmIsFriendPriceC
									payback = (cprice.C_price_bm_c7.Float64 - cprice.P_price_bm_c7.Float64) + (bmTargetingPriceC - bmTargetingPriceP) + (bmIsFriendPriceC - bmIsFriendPriceP)
									if s.EqualFold(mst_sent_voucher.String, "B") {
										memo = memo + ",보너스"
									}
								}
							} else if s.EqualFold(message_type.String, "C8") {
								admin_amt = cprice.B_price_bm_c8.Float64 + bmTargetingPriceB + bmIsFriendPriceB
								if s.EqualFold(mst_sent_voucher.String, "V") {
									amount = cprice.V_price_bm_c8.Float64 + bmTargetingPriceV + bmIsFriendPriceV
									payback = (cprice.V_price_bm_c8.Float64 - cprice.P_price_bm_c8.Float64) + (bmTargetingPriceV - bmTargetingPriceP) + (bmIsFriendPriceV - bmIsFriendPriceP)
									memo = memo + ",바우처"
								} else {
									amount = cprice.C_price_bm_c8.Float64 + bmTargetingPriceC + bmIsFriendPriceC
									payback = (cprice.C_price_bm_c8.Float64 - cprice.P_price_bm_c8.Float64) + (bmTargetingPriceC - bmTargetingPriceP) + (bmIsFriendPriceC - bmIsFriendPriceP)
									if s.EqualFold(mst_sent_voucher.String, "B") {
										memo = memo + ",보너스"
									}
								}
							}
						} else if s.HasPrefix(s.ToUpper(message_type.String), "D") { // BrandMsg 자유형 성공 차감
							bmTargeting := s.ToUpper(kind.String)

							if bmTargeting == "M" {
								bmTargetingPriceV = cprice.V_price_bm_t_m.Float64
								bmTargetingPriceP = cprice.P_price_bm_t_m.Float64
								bmTargetingPriceB = cprice.B_price_bm_t_m.Float64
								bmTargetingPriceC = cprice.C_price_bm_t_m.Float64

								//TODO : 리스폰스로 친구인지 아닌지 판단할 수 있어야함
								//TODO : 일단은 친구가 아닌 값으로 넣어놓음
								bmIsFriendPriceV = cprice.V_price_bm_nf.Float64
								bmIsFriendPriceP = cprice.P_price_bm_nf.Float64
								bmIsFriendPriceB = cprice.B_price_bm_nf.Float64
								bmIsFriendPriceC = cprice.C_price_bm_nf.Float64
							} else if bmTargeting == "N" {
								bmTargetingPriceV = cprice.V_price_bm_t_n.Float64
								bmTargetingPriceP = cprice.P_price_bm_t_n.Float64
								bmTargetingPriceB = cprice.B_price_bm_t_n.Float64
								bmTargetingPriceC = cprice.C_price_bm_t_n.Float64

								bmIsFriendPriceV = cprice.V_price_bm_nf.Float64
								bmIsFriendPriceP = cprice.P_price_bm_nf.Float64
								bmIsFriendPriceB = cprice.B_price_bm_nf.Float64
								bmIsFriendPriceC = cprice.C_price_bm_nf.Float64
							} else if bmTargeting == "I" {
								bmTargetingPriceV = cprice.V_price_bm_t_i.Float64
								bmTargetingPriceP = cprice.P_price_bm_t_i.Float64
								bmTargetingPriceB = cprice.B_price_bm_t_i.Float64
								bmTargetingPriceC = cprice.C_price_bm_t_i.Float64

								bmIsFriendPriceV = cprice.V_price_bm_f.Float64
								bmIsFriendPriceP = cprice.P_price_bm_f.Float64
								bmIsFriendPriceB = cprice.B_price_bm_f.Float64
								bmIsFriendPriceC = cprice.C_price_bm_f.Float64
							} else if bmTargeting == "F" {
								bmTargetingPriceV = cprice.V_price_bm_t_f.Float64
								bmTargetingPriceP = cprice.P_price_bm_t_f.Float64
								bmTargetingPriceB = cprice.B_price_bm_t_f.Float64
								bmTargetingPriceC = cprice.C_price_bm_t_f.Float64

								bmIsFriendPriceV = cprice.V_price_bm_f.Float64
								bmIsFriendPriceP = cprice.P_price_bm_f.Float64
								bmIsFriendPriceB = cprice.B_price_bm_f.Float64
								bmIsFriendPriceC = cprice.C_price_bm_f.Float64
							} else {
								bmTargetingPriceV = cprice.V_price_bm_t_m.Float64
								bmTargetingPriceP = cprice.P_price_bm_t_m.Float64
								bmTargetingPriceB = cprice.B_price_bm_t_m.Float64
								bmTargetingPriceC = cprice.C_price_bm_t_m.Float64

								bmIsFriendPriceV = cprice.V_price_bm_nf.Float64
								bmIsFriendPriceP = cprice.P_price_bm_nf.Float64
								bmIsFriendPriceB = cprice.B_price_bm_nf.Float64
								bmIsFriendPriceC = cprice.C_price_bm_nf.Float64
							}

							expectedCnt, err := strconv.ParseFloat(remark3.String, 64)
							if err != nil {
								expectedCnt = 0
							}
							sendCnt, _ := strconv.ParseFloat(price.String, 64)

							memo = "브랜드동보(" + BMMESSAGETYPE[s.ToUpper(message_type.String)] + " | 예상/실발송건수 : " + remark3.String + "/" + price.String + ")"

							var subCnt float64 = 0
							if code.String == "0001" {
								err_ftcnt++
								kko_kind = "3"
								subCnt = expectedCnt
							} else if code.String == "0000" {
								bcCnt, _ = strconv.Atoi(price.String)
								ftcnt++
								if sendCnt > expectedCnt {
									kko_kind = "D"
									subCnt = sendCnt - expectedCnt
								} else if sendCnt < expectedCnt {
									kko_kind = "3"
									subCnt = expectedCnt - sendCnt
								} else {
									isPayment = false
								}
							} else {
								err_ftcnt++
								isPayment = false
							}

							subCnt = float64(subCnt)

							// BrandMsg 메세지 타입별 정산
							if subCnt > 0 {
								if s.EqualFold(message_type.String, "D1") {
									admin_amt = (cprice.B_price_bm_d1.Float64 + bmTargetingPriceB + bmIsFriendPriceB) * subCnt
									if s.EqualFold(mst_sent_voucher.String, "V") {
										amount = (cprice.V_price_bm_d1.Float64 + bmTargetingPriceV + bmIsFriendPriceV) * subCnt
										payback = ((cprice.V_price_bm_d1.Float64 - cprice.P_price_bm_d1.Float64) + (bmTargetingPriceV - bmTargetingPriceP) + (bmIsFriendPriceV - bmIsFriendPriceP)) * subCnt
										memo = memo + ",바우처"
									} else {
										amount = (cprice.C_price_bm_d1.Float64 + bmTargetingPriceC) * subCnt
										payback = ((cprice.C_price_bm_d1.Float64 - cprice.P_price_bm_d1.Float64) + (bmTargetingPriceC - bmTargetingPriceP) + (bmIsFriendPriceC - bmIsFriendPriceP)) * subCnt
										if s.EqualFold(mst_sent_voucher.String, "B") {
											memo = memo + ",보너스"
										}
									}
								} else if s.EqualFold(message_type.String, "D2") {
									admin_amt = (cprice.B_price_bm_d2.Float64 + bmTargetingPriceB + bmIsFriendPriceB) * subCnt
									if s.EqualFold(mst_sent_voucher.String, "V") {
										amount = (cprice.V_price_bm_d2.Float64 + bmTargetingPriceV + bmIsFriendPriceV) * subCnt
										payback = ((cprice.V_price_bm_d2.Float64 - cprice.P_price_bm_d2.Float64) + (bmTargetingPriceV - bmTargetingPriceP) + (bmIsFriendPriceV - bmIsFriendPriceP)) * subCnt
										memo = memo + ",바우처"
									} else {
										amount = (cprice.C_price_bm_d2.Float64 + bmTargetingPriceC + bmIsFriendPriceC) * subCnt
										payback = ((cprice.C_price_bm_d2.Float64 - cprice.P_price_bm_d2.Float64) + (bmTargetingPriceC - bmTargetingPriceP) + (bmIsFriendPriceC - bmIsFriendPriceP)) * subCnt
										if s.EqualFold(mst_sent_voucher.String, "B") {
											memo = memo + ",보너스"
										}
									}
								} else if s.EqualFold(message_type.String, "D3") {
									admin_amt = (cprice.B_price_bm_d3.Float64 + bmTargetingPriceB + bmIsFriendPriceB) * subCnt
									if s.EqualFold(mst_sent_voucher.String, "V") {
										amount = (cprice.V_price_bm_d3.Float64 + bmTargetingPriceV + bmIsFriendPriceV) * subCnt
										payback = ((cprice.V_price_bm_d3.Float64 - cprice.P_price_bm_d3.Float64) + (bmTargetingPriceV - bmTargetingPriceP) + (bmIsFriendPriceV - bmIsFriendPriceP)) * subCnt
										memo = memo + ",바우처"
									} else {
										amount = (cprice.C_price_bm_d3.Float64 + bmTargetingPriceC + bmIsFriendPriceC) * subCnt
										payback = ((cprice.C_price_bm_d3.Float64 - cprice.P_price_bm_d3.Float64) + (bmTargetingPriceC - bmTargetingPriceP) + (bmIsFriendPriceC - bmIsFriendPriceP)) * subCnt
										if s.EqualFold(mst_sent_voucher.String, "B") {
											memo = memo + ",보너스"
										}
									}
								} else if s.EqualFold(message_type.String, "D4") {
									admin_amt = (cprice.B_price_bm_d4.Float64 + bmTargetingPriceB + bmIsFriendPriceB) * subCnt
									if s.EqualFold(mst_sent_voucher.String, "V") {
										amount = (cprice.V_price_bm_d4.Float64 + bmTargetingPriceV + bmIsFriendPriceV) * subCnt
										payback = ((cprice.V_price_bm_d4.Float64 - cprice.P_price_bm_d4.Float64) + (bmTargetingPriceV - bmTargetingPriceP) + (bmIsFriendPriceV - bmIsFriendPriceP)) * subCnt
										memo = memo + ",바우처"
									} else {
										amount = (cprice.C_price_bm_d4.Float64 + bmTargetingPriceC + bmIsFriendPriceC) * subCnt
										payback = ((cprice.C_price_bm_d4.Float64 - cprice.P_price_bm_d4.Float64) + (bmTargetingPriceC - bmTargetingPriceP) + (bmIsFriendPriceC - bmIsFriendPriceP)) * subCnt
										if s.EqualFold(mst_sent_voucher.String, "B") {
											memo = memo + ",보너스"
										}
									}
								} else if s.EqualFold(message_type.String, "D5") {
									admin_amt = (cprice.B_price_bm_d5.Float64 + bmTargetingPriceB + bmIsFriendPriceB) * subCnt
									if s.EqualFold(mst_sent_voucher.String, "V") {
										amount = (cprice.V_price_bm_d5.Float64 + bmTargetingPriceV + bmIsFriendPriceV) * subCnt
										payback = ((cprice.V_price_bm_d5.Float64 - cprice.P_price_bm_d5.Float64) + (bmTargetingPriceV - bmTargetingPriceP) + (bmIsFriendPriceV - bmIsFriendPriceP)) * subCnt
										memo = memo + ",바우처"
									} else {
										amount = (cprice.C_price_bm_d5.Float64 + bmTargetingPriceC + bmIsFriendPriceC) * subCnt
										payback = ((cprice.C_price_bm_d5.Float64 - cprice.P_price_bm_d5.Float64) + (bmTargetingPriceC - bmTargetingPriceP) + (bmIsFriendPriceC - bmIsFriendPriceP)) * subCnt
										if s.EqualFold(mst_sent_voucher.String, "B") {
											memo = memo + ",보너스"
										}
									}
								} else if s.EqualFold(message_type.String, "D6") {
									admin_amt = (cprice.B_price_bm_d6.Float64 + bmTargetingPriceB + bmIsFriendPriceB) * subCnt
									if s.EqualFold(mst_sent_voucher.String, "V") {
										amount = (cprice.V_price_bm_d6.Float64 + bmTargetingPriceV + bmIsFriendPriceV) * subCnt
										payback = ((cprice.V_price_bm_d6.Float64 - cprice.P_price_bm_d6.Float64) + (bmTargetingPriceV - bmTargetingPriceP) + (bmIsFriendPriceV - bmIsFriendPriceP)) * subCnt
										memo = memo + ",바우처"
									} else {
										amount = (cprice.C_price_bm_d6.Float64 + bmTargetingPriceC + bmIsFriendPriceC) * subCnt
										payback = ((cprice.C_price_bm_d6.Float64 - cprice.P_price_bm_d6.Float64) + (bmTargetingPriceC - bmTargetingPriceP) + (bmIsFriendPriceC - bmIsFriendPriceP)) * subCnt
										if s.EqualFold(mst_sent_voucher.String, "B") {
											memo = memo + ",보너스"
										}
									}
								} else if s.EqualFold(message_type.String, "D7") {
									admin_amt = (cprice.B_price_bm_d7.Float64 + bmTargetingPriceB + bmIsFriendPriceB) * subCnt
									if s.EqualFold(mst_sent_voucher.String, "V") {
										amount = (cprice.V_price_bm_d7.Float64 + bmTargetingPriceV + bmIsFriendPriceV) * subCnt
										payback = ((cprice.V_price_bm_d7.Float64 - cprice.P_price_bm_d7.Float64) + (bmTargetingPriceV - bmTargetingPriceP) + (bmIsFriendPriceV - bmIsFriendPriceP)) * subCnt
										memo = memo + ",바우처"
									} else {
										amount = (cprice.C_price_bm_d7.Float64 + bmTargetingPriceC + bmIsFriendPriceC) * subCnt
										payback = ((cprice.C_price_bm_d7.Float64 - cprice.P_price_bm_d7.Float64) + (bmTargetingPriceC - bmTargetingPriceP) + (bmIsFriendPriceC - bmIsFriendPriceP)) * subCnt
										if s.EqualFold(mst_sent_voucher.String, "B") {
											memo = memo + ",보너스"
										}
									}
								} else if s.EqualFold(message_type.String, "D8") {
									admin_amt = (cprice.B_price_bm_d8.Float64 + bmTargetingPriceB + bmIsFriendPriceB) * subCnt
									if s.EqualFold(mst_sent_voucher.String, "V") {
										amount = (cprice.V_price_bm_d8.Float64 + bmTargetingPriceV + bmIsFriendPriceV) * subCnt
										payback = ((cprice.V_price_bm_d8.Float64 - cprice.P_price_bm_d8.Float64) + (bmTargetingPriceV - bmTargetingPriceP) + (bmIsFriendPriceV - bmIsFriendPriceP)) * subCnt
										memo = memo + ",바우처"
									} else {
										amount = (cprice.C_price_bm_d8.Float64 + bmTargetingPriceC + bmIsFriendPriceC) * subCnt
										payback = ((cprice.C_price_bm_d8.Float64 - cprice.P_price_bm_d8.Float64) + (bmTargetingPriceC - bmTargetingPriceP) + (bmIsFriendPriceC - bmIsFriendPriceP)) * subCnt
										if s.EqualFold(mst_sent_voucher.String, "B") {
											memo = memo + ",보너스"
										}
									}
								}
							}
						}

					} else { // 카카오 메세지 실패 시 혹은 메세지 전용 일 경우 처리
						if !s.EqualFold(message.String, "InvalidPhoneNumber") && len(mem_resend) > 0 && !s.EqualFold(mem_resend, "NONE") && len(sms_sender.String) > 0 {

							if s.HasPrefix(phnstr, "82") {
								phnstr = "0" + phnstr[2:len(phnstr)]
							}

							if vancnt.String != "0" { // 수신거부 List 에 있으면 수신거부 메세지 처리

								cb_msg_message = "수신거부"
								isPayment = false

								switch mem_resend {
								case "SMART":
									if s.EqualFold(msgtype, "SMS") || s.EqualFold(msgtype, "LMS") {
										err_smtcnt++
										cb_msg_message_type = "SM"
										cb_msg_code = "SMT"
									}
								case "LGU":
									if s.EqualFold(msgtype, "SMS") || s.EqualFold(msgtype, "LMS") {
										err_smtcnt++
										cb_msg_message_type = "LG"
										cb_msg_code = "LGU"
									}
								case "GREEN_SHOT":
									if s.EqualFold(msgtype, "SMS") || s.EqualFold(msgtype, "LMS") {
										err_smtcnt++
										cb_msg_message_type = "GS"
										cb_msg_code = "GRS"
									}
								case "GREEN_SHOT_G":
									if s.EqualFold(msgtype, "SMS") || s.EqualFold(msgtype, "LMS") {
										err_smtcnt++
										cb_msg_message_type = "NL"
										cb_msg_code = "GRS"
									}
								case "SMTNT":
									if s.EqualFold(msgtype, "SMS") || s.EqualFold(msgtype, "LMS") {
										err_smtcnt++
										cb_msg_message_type = "TN"
										cb_msg_code = "TNT"
									}
								case "JJ":
									if s.EqualFold(msgtype, "SMS") || s.EqualFold(msgtype, "LMS") {
										err_smtcnt++
										cb_msg_message_type = "JJ"
										cb_msg_code = "JJ"
									}
								case "RCS":
									if s.EqualFold(msgtype, "SMS") || s.EqualFold(msgtype, "LMS") || s.EqualFold(msgtype, "MMS") || s.EqualFold(msgtype, "TEM") {
										err_rcscnt++
										cb_msg_message_type = "RC"
										cb_msg_code = "RCS"
									}
								}
							} else { // 수신거부 처리 끝
								// 2차 발신 처리 시작

								mst_waitcnt++
								cb_msg_message = "결과 수신대기"
								switch mem_resend {
								case "RCS":
									var rcsTemplate, rcsBrand, rcsDatetime, rcsKind, rcsContent, rcsBtn1, rcsBtn2, rcsBtn3, rcsBtn4, rcsBtn5, rcsChatbotID, rcsBtns, rcsBody, rcsBrandkey, rcsPlatform sql.NullString

									cb_msg_message_type = "rc"
									cb_msg_code = "RCS"
									msgbaseid := ""
									srv_type := "RCSSMS"

									rcsRowSql := `
											SELECT
												msr_template
											  , msr_brand
											  , msr_datetime
											  , msr_kind
											  , msr_content
											  , msr_button1
											  , msr_button2
											  , msr_button3
											  , msr_button4
											  , msr_button5
											  , msr_chatbotid
											  , msr_button
											  , msr_body
											  , msr_brandkey
											  , msr_platform
											FROM
												cb_wt_msg_rcs
											WHERE
												msr_mst_id = ?`

									if conf.RCS {
										err = db.QueryRow(rcsRowSql, ressendkey.String).Scan(&rcsTemplate, &rcsBrand, &rcsDatetime, &rcsKind, &rcsContent, &rcsBtn1, &rcsBtn2, &rcsBtn3, &rcsBtn4, &rcsBtn5, &rcsChatbotID, &rcsBtns, &rcsBody, &rcsBrandkey, &rcsPlatform)
										if err != nil {
											errlog.Println("tblresultproc - RCS Table 조회 중 중 오류 발생 err : ", err)
											errlog.Println("tblresultproc - msr_mst_id : ", ressendkey.String, " / sql : ", rcsRowSql)
										}
									}

									srv_type = rcsKind.String
									msgbaseid = rcsTemplate.String

									// if s.EqualFold(msgtype, "SMS") {
									// 	srv_type = rcsKind.String      //"RCSTMPL"
									// 	msgbaseid = rcsTemplate.String //"UBR.0Uv12uh7R4-GG000F"
									// } else if s.EqualFold(msgtype, "LMS") {
									// 	srv_type = rcsKind.String      //"RCSLMS"
									// 	msgbaseid = rcsTemplate.String //"SL000000"
									// } else if s.EqualFold(msgtype, "MMS") {
									// 	srv_type = rcsKind.String      //"RCMMS"
									// 	msgbaseid = rcsTemplate.String
									// } else if s.EqualFold(msgtype, "TEM") {
									// 	srv_type = rcsKind.String      //"RCSTMPL"
									// 	msgbaseid = rcsTemplate.String
									// }

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

									rcsStrs = append(rcsStrs, "(?,?,'0',?,'rcs',?,?,?,?,1,'0',null,1,?,?,?,?)")
									rcsValues = append(rcsValues, msgid.String)
									rcsValues = append(rcsValues, phnstr)
									rcsValues = append(rcsValues, remark4.String)
									rcsValues = append(rcsValues, s.Replace(sms_sender.String, "-", "", -1)) // 발신자 전화 번호
									rcsValues = append(rcsValues, "dhn2021g")
									rcsValues = append(rcsValues, msgbaseid)
									rcsValues = append(rcsValues, srv_type)
									rcsValues = append(rcsValues, rcsBody.String)
									rcsValues = append(rcsValues, rcsBtns.String)
									rcsValues = append(rcsValues, rcsBrandkey.String)
									rcsValues = append(rcsValues, rcsPlatform.String)
								case "SMART":
									cb_msg_message_type = "sm"
									cb_msg_code = "SMT"

									if s.EqualFold(msgtype, "SMS") {
										ossmsStrs = append(ossmsStrs, "(?,?,?,?,?,null,?,?,?,?)")
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
										ossmsValues = append(ossmsValues, config.Conf.KISACODE)

										kko_kind = "P"

										admin_amt = cprice.B_price_smt_sms.Float64
										if s.EqualFold(mst_sent_voucher.String, "V") {
											amount = cprice.V_price_smt_sms.Float64
											payback = cprice.V_price_smt_sms.Float64 - cprice.P_price_smt_sms.Float64
											memo = "웹(C) SMS,바우처"
										} else {
											amount = cprice.C_price_smt_sms.Float64
											payback = cprice.C_price_smt_sms.Float64 - cprice.P_price_smt_sms.Float64
											if s.EqualFold(mst_sent_voucher.String, "B") {
												memo = "웹(C) SMS,보너스"
											} else {
												memo = "웹(C) SMS"
											}

										}
									} else if s.EqualFold(msgtype, "LMS") {
										osmmsStrs = append(osmmsStrs, "(?,?,?,?,?,?,null,?,?,?,?,?,?,?)")
										osmmsValues = append(osmmsValues, remark4)
										osmmsValues = append(osmmsValues, sms_sender)
										osmmsValues = append(osmmsValues, phnstr)
										osmmsValues = append(osmmsValues, sms_lms_tit)
										osmmsValues = append(osmmsValues, msg_sms)
										if reserve_dt.String == "00000000000000" {
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
										osmmsValues = append(osmmsValues, config.Conf.KISACODE)

										if len(mms_file1.String) <= 0 {
											kko_kind = "P"

											admin_amt = cprice.B_price_smt.Float64
											if s.EqualFold(mst_sent_voucher.String, "V") {
												amount = cprice.V_price_smt.Float64
												payback = cprice.V_price_smt.Float64 - cprice.P_price_smt.Float64
												memo = "웹(C) LMS,바우처"
											} else {
												amount = cprice.C_price_smt.Float64
												payback = cprice.C_price_smt.Float64 - cprice.P_price_smt.Float64
												if s.EqualFold(mst_sent_voucher.String, "B") {
													memo = "웹(C) LMS,보너스"
												} else {
													memo = "웹(C) LMS"
												}
											}
										} else {
											kko_kind = "P"

											admin_amt = cprice.B_price_smt_mms.Float64
											if s.EqualFold(mst_sent_voucher.String, "V") {
												amount = cprice.V_price_smt_mms.Float64
												payback = cprice.V_price_smt_mms.Float64 - cprice.P_price_smt_mms.Float64
												memo = "웹(C) MMS,바우처"
											} else {
												amount = cprice.C_price_smt_mms.Float64
												payback = cprice.C_price_smt_mms.Float64 - cprice.P_price_smt_mms.Float64
												if s.EqualFold(mst_sent_voucher.String, "B") {
													memo = "웹(C) MMS,보너스"
												} else {
													memo = "웹(C) MMS"
												}
											}
										}
									}
								case "LGU":
									cb_msg_message_type = "lg"
									cb_msg_code = "LGU"

									if s.EqualFold(msgtype, "SMS") {
										lgusmsStrs = append(lgusmsStrs, "(?,?,?,?,?,?,?,?)")
										lgusmsValues = append(lgusmsValues, time.Now().Format("2006-01-02 15:04:05"))
										lgusmsValues = append(lgusmsValues, phnstr)
										lgusmsValues = append(lgusmsValues, sms_sender)
										lgusmsValues = append(lgusmsValues, msg_sms)
										lgusmsValues = append(lgusmsValues, msgid)
										lgusmsValues = append(lgusmsValues, mem_userid)
										lgusmsValues = append(lgusmsValues, remark4)
										lgusmsValues = append(lgusmsValues, config.Conf.KISACODE)

										kko_kind = "P"

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
									} else if s.EqualFold(msgtype, "LMS") {
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
										lgummsValues = append(lgummsValues, sms_lms_tit)
										lgummsValues = append(lgummsValues, phnstr)
										lgummsValues = append(lgummsValues, sms_sender)
										lgummsValues = append(lgummsValues, time.Now().Format("2006-01-02 15:04:05"))
										lgummsValues = append(lgummsValues, msg_sms)
										lgummsValues = append(lgummsValues, file_cnt)
										lgummsValues = append(lgummsValues, mms_file1)
										lgummsValues = append(lgummsValues, mms_file2)
										lgummsValues = append(lgummsValues, mms_file3)
										lgummsValues = append(lgummsValues, msgid)
										lgummsValues = append(lgummsValues, mem_userid)
										lgummsValues = append(lgummsValues, remark4)
										lgummsValues = append(lgummsValues, config.Conf.KISACODE)

										if len(mms_file1.String) <= 0 {
											kko_kind = "P"

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
											kko_kind = "P"

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
									}
								case "GREEN_SHOT":
									cb_msg_message_type = "gs"
									cb_msg_code = "GRS"

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
										nnsmsValues = append(nnsmsValues, config.Conf.KISACODE)

										kko_kind = "P"

										admin_amt = cprice.B_price_smt_sms.Float64
										if s.EqualFold(mst_sent_voucher.String, "V") {
											amount = cprice.V_price_smt_sms.Float64
											payback = cprice.V_price_smt_sms.Float64 - cprice.P_price_smt_sms.Float64
											memo = "웹(A) SMS,바우처"
										} else {
											amount = cprice.C_price_smt_sms.Float64
											payback = cprice.C_price_smt_sms.Float64 - cprice.P_price_smt_sms.Float64
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
										nnmmsValues = append(nnmmsValues, config.Conf.KISACODE)

										if len(mms_file1.String) <= 0 {
											kko_kind = "P"

											admin_amt = cprice.B_price_smt.Float64
											if s.EqualFold(mst_sent_voucher.String, "V") {
												amount = cprice.V_price_smt.Float64
												payback = cprice.V_price_smt.Float64 - cprice.P_price_smt.Float64
												memo = "웹(A) LMS,바우처"
											} else {
												amount = cprice.C_price_smt.Float64
												payback = cprice.C_price_smt.Float64 - cprice.P_price_smt.Float64
												if s.EqualFold(mst_sent_voucher.String, "B") {
													memo = "웹(A) LMS,보너스"
												} else {
													memo = "웹(A) LMS"
												}

											}
										} else {
											kko_kind = "P"

											admin_amt = cprice.B_price_smt_mms.Float64
											if s.EqualFold(mst_sent_voucher.String, "V") {
												amount = cprice.V_price_smt_mms.Float64
												payback = cprice.V_price_smt_mms.Float64 - cprice.P_price_smt_mms.Float64
												memo = "웹(A) MMS,바우처"
											} else {
												amount = cprice.C_price_smt_mms.Float64
												payback = cprice.C_price_smt_mms.Float64 - cprice.P_price_smt_mms.Float64
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
											nnsmsValues = append(nnsmsValues, config.Conf.KISACODE)

											kko_kind = "P"

											admin_amt = cprice.B_price_smt_sms.Float64
											if s.EqualFold(mst_sent_voucher.String, "V") {
												amount = cprice.V_price_smt_sms.Float64
												payback = cprice.V_price_smt_sms.Float64 - cprice.P_price_smt_sms.Float64
												memo = "웹(A) SMS,바우처"
											} else {
												amount = cprice.C_price_smt_sms.Float64
												payback = cprice.C_price_smt_sms.Float64 - cprice.P_price_smt_sms.Float64
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
											nnmmsValues = append(nnmmsValues, config.Conf.KISACODE)

											if len(mms_file1.String) <= 0 {
												kko_kind = "P"

												admin_amt = cprice.B_price_smt.Float64
												if s.EqualFold(mst_sent_voucher.String, "V") {
													amount = cprice.V_price_smt.Float64
													payback = cprice.V_price_smt.Float64 - cprice.P_price_smt.Float64
													memo = "웹(A) LMS,바우처"
												} else {
													amount = cprice.C_price_smt.Float64
													payback = cprice.C_price_smt.Float64 - cprice.P_price_smt.Float64
													if s.EqualFold(mst_sent_voucher.String, "B") {
														memo = "웹(A) LMS,보너스"
													} else {
														memo = "웹(A) LMS"
													}

												}
											} else {
												kko_kind = "P"

												admin_amt = cprice.B_price_smt_mms.Float64
												if s.EqualFold(mst_sent_voucher.String, "V") {
													amount = cprice.V_price_smt_mms.Float64
													payback = cprice.V_price_smt_mms.Float64 - cprice.P_price_smt_mms.Float64
													memo = "웹(A) MMS,바우처"
												} else {
													amount = cprice.C_price_smt_mms.Float64
													payback = cprice.C_price_smt_mms.Float64 - cprice.P_price_smt_mms.Float64
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
											nnlpsmsValues = append(nnlpsmsValues, config.Conf.KISACODE)

											kko_kind = "P"

											admin_amt = cprice.B_price_smt_sms.Float64
											if s.EqualFold(mst_sent_voucher.String, "V") {
												amount = cprice.V_price_smt_sms.Float64
												payback = cprice.V_price_smt_sms.Float64 - cprice.P_price_smt_sms.Float64
												memo = "웹(A) SMS,바우처"
											} else {
												amount = cprice.C_price_smt_sms.Float64
												payback = cprice.C_price_smt_sms.Float64 - cprice.P_price_smt_sms.Float64
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
											nnlpmmsValues = append(nnlpmmsValues, config.Conf.KISACODE)

											if len(mms_file1.String) <= 0 {
												kko_kind = "P"

												admin_amt = cprice.B_price_smt.Float64
												if s.EqualFold(mst_sent_voucher.String, "V") {
													amount = cprice.V_price_smt.Float64
													payback = cprice.V_price_smt.Float64 - cprice.P_price_smt.Float64
													memo = "웹(A) LMS,바우처"
												} else {
													amount = cprice.C_price_smt.Float64
													payback = cprice.C_price_smt.Float64 - cprice.P_price_smt.Float64
													if s.EqualFold(mst_sent_voucher.String, "B") {
														memo = "웹(A) LMS,보너스"
													} else {
														memo = "웹(A) LMS"
													}

												}
											} else {
												kko_kind = "P"

												admin_amt = cprice.B_price_smt_mms.Float64
												if s.EqualFold(mst_sent_voucher.String, "V") {
													amount = cprice.V_price_smt_mms.Float64
													payback = cprice.V_price_smt_mms.Float64 - cprice.P_price_smt_mms.Float64
													memo = "웹(A) MMS,바우처"
												} else {
													amount = cprice.C_price_smt_mms.Float64
													payback = cprice.C_price_smt_mms.Float64 - cprice.P_price_smt_mms.Float64
													if s.EqualFold(mst_sent_voucher.String, "B") {
														memo = "웹(A) MMS,보너스"
													} else {
														memo = "웹(A) MMS"
													}
												}

											}
										}
									}
								case "SMTNT":
									cb_msg_message_type = "tn"
									cb_msg_code = "TNT"

									smtntTime := time.Now().Format("2006-01-02 15:04:05")

									if s.EqualFold(msgtype, "SMS") {
										tntsmsStrs = append(tntsmsStrs, "(?,?,?,?,?,?,?,?,?,?)")
										tntsmsValues = append(tntsmsValues, phnstr)               // Phone_No 1
										tntsmsValues = append(tntsmsValues, sms_sender)           // Callback_No 2
										tntsmsValues = append(tntsmsValues, "4")                  // Msg_Type 3
										tntsmsValues = append(tntsmsValues, smtntTime)            // Send_Time 4
										tntsmsValues = append(tntsmsValues, smtntTime)            // Save_Time 5
										tntsmsValues = append(tntsmsValues, msg_sms)              // Message 6
										tntsmsValues = append(tntsmsValues, config.Conf.KISACODE) // Reseller_Code 7

										tntsmsValues = append(tntsmsValues, msgid)      // Etc1 8
										tntsmsValues = append(tntsmsValues, mem_userid) // Etc2 9
										tntsmsValues = append(tntsmsValues, remark4)    // Etc3 10

										kko_kind = "P"

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
									} else if s.EqualFold(msgtype, "LMS") {
										fileCnt := 0
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
										tntmmsValues = append(tntmmsValues, phnstr)               // Phone_No 1
										tntmmsValues = append(tntmmsValues, sms_sender)           // Callback_No 2
										tntmmsValues = append(tntmmsValues, "6")                  // Msg_Type 3
										tntmmsValues = append(tntmmsValues, smtntTime)            // Send_Time 4
										tntmmsValues = append(tntmmsValues, smtntTime)            // Save_Time 5
										tntmmsValues = append(tntmmsValues, sms_lms_tit)          // Subject 6
										tntmmsValues = append(tntmmsValues, msg_sms)              // Message 7
										tntmmsValues = append(tntmmsValues, fileCnt)              // File_Count 8
										tntmmsValues = append(tntmmsValues, fileType1)            // File_Type1 9
										tntmmsValues = append(tntmmsValues, fileType2)            // File_Type2 10
										tntmmsValues = append(tntmmsValues, fileType3)            // File_Type3 11
										tntmmsValues = append(tntmmsValues, mms_file1)            // File_Name1 12
										tntmmsValues = append(tntmmsValues, mms_file2)            // File_Name2 13
										tntmmsValues = append(tntmmsValues, mms_file3)            // File_Name3 14
										tntmmsValues = append(tntmmsValues, config.Conf.KISACODE) // Reseller_Code 15

										tntmmsValues = append(tntmmsValues, msgid)      // Etc1 16
										tntmmsValues = append(tntmmsValues, mem_userid) // Etc2 17
										tntmmsValues = append(tntmmsValues, remark4)    // Etc3 18

										if len(mms_file1.String) <= 0 {
											kko_kind = "P"

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
											kko_kind = "P"

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
									}
								case "JJ":
									cb_msg_message_type = "jj"
									cb_msg_code = "JJ"

									if s.EqualFold(msgtype, "SMS") {
										jjsmsStrs = append(jjsmsStrs, "(?,?,?,?,?,?,?,?,?)")
										jjsmsValues = append(jjsmsValues, sms_sender)
										jjsmsValues = append(jjsmsValues, phnstr)
										jjsmsValues = append(jjsmsValues, "SMS")
										jjsmsValues = append(jjsmsValues, sms_lms_tit)
										jjsmsValues = append(jjsmsValues, msg_sms)
										jjsmsValues = append(jjsmsValues, config.Conf.KISACODE)
										jjsmsValues = append(jjsmsValues, msgid)
										jjsmsValues = append(jjsmsValues, mem_userid)
										jjsmsValues = append(jjsmsValues, remark4)

										kko_kind = "P"

										admin_amt = cprice.B_price_smt_sms.Float64
										if s.EqualFold(mst_sent_voucher.String, "V") {
											amount = cprice.V_price_smt_sms.Float64
											payback = cprice.V_price_smt_sms.Float64 - cprice.P_price_smt_sms.Float64
											memo = "웹(E) SMS,바우처"
										} else {
											amount = cprice.C_price_smt_sms.Float64
											payback = cprice.C_price_smt_sms.Float64 - cprice.P_price_smt_sms.Float64
											if s.EqualFold(mst_sent_voucher.String, "B") {
												memo = "웹(E) SMS,보너스"
											} else {
												memo = "웹(E) SMS"
											}

										}
									} else if s.EqualFold(msgtype, "LMS") {
										jjMsgType := "LMS"
										if (mms_file1.Valid && mms_file1.String != "" && len(mms_file1.String) > 0) ||
										   (mms_file2.Valid && mms_file2.String != "" && len(mms_file2.String) > 0) ||
										   (mms_file3.Valid && mms_file3.String != "" && len(mms_file3.String) > 0) {
											jjMsgType = "MMS"
										}

										jjmmsStrs = append(jjmmsStrs, "(?,?,?,?,?,?,?,?,?,?,?,?)")
										jjmmsValues = append(jjmmsValues, sms_sender)
										jjmmsValues = append(jjmmsValues, phnstr)
										jjmmsValues = append(jjmmsValues, jjMsgType)
										jjmmsValues = append(jjmmsValues, sms_lms_tit)
										jjmmsValues = append(jjmmsValues, msg_sms)
										if mms_file1.Valid && mms_file1.String != "" && len(mms_file1.String) > 0 {
											jjmmsValues = append(jjmmsValues, mms_file1.String)
										} else {
											jjmmsValues = append(jjmmsValues, sql.NullString{})
										}
										if mms_file2.Valid && mms_file2.String != "" && len(mms_file2.String) > 0 {
											jjmmsValues = append(jjmmsValues, mms_file2.String)
										} else {
											jjmmsValues = append(jjmmsValues, sql.NullString{})
										}
										if mms_file3.Valid && mms_file3.String != "" && len(mms_file3.String) > 0 {
											jjmmsValues = append(jjmmsValues, mms_file3.String)
										} else {
											jjmmsValues = append(jjmmsValues, sql.NullString{})
										}
										jjmmsValues = append(jjmmsValues, config.Conf.KISACODE)
										jjmmsValues = append(jjmmsValues, msgid)
										jjmmsValues = append(jjmmsValues, mem_userid)
										jjmmsValues = append(jjmmsValues, remark4)

										if len(mms_file1.String) <= 0 {
											kko_kind = "P"

											admin_amt = cprice.B_price_smt.Float64
											if s.EqualFold(mst_sent_voucher.String, "V") {
												amount = cprice.V_price_smt.Float64
												payback = cprice.V_price_smt.Float64 - cprice.P_price_smt.Float64
												memo = "웹(E) LMS,바우처"
											} else {
												amount = cprice.C_price_smt.Float64
												payback = cprice.C_price_smt.Float64 - cprice.P_price_smt.Float64
												if s.EqualFold(mst_sent_voucher.String, "B") {
													memo = "웹(E) LMS,보너스"
												} else {
													memo = "웹(E) LMS"
												}
											}
										} else {
											kko_kind = "P"

											admin_amt = cprice.B_price_smt_mms.Float64
											if s.EqualFold(mst_sent_voucher.String, "V") {
												amount = cprice.V_price_smt_mms.Float64
												payback = cprice.V_price_smt_mms.Float64 - cprice.P_price_smt_mms.Float64
												memo = "웹(E) MMS,바우처"
											} else {
												amount = cprice.C_price_smt_mms.Float64
												payback = cprice.C_price_smt_mms.Float64 - cprice.P_price_smt_mms.Float64
												if s.EqualFold(mst_sent_voucher.String, "B") {
													memo = "웹(E) MMS,보너스"
												} else {
													memo = "웹(E) MMS"
												}
											}
										}
									}
								}
							}

						} else {
							// 2차 발신 없는 kakaotalk 처리.
							isPayment = false                                     // 2차 발신 없으면 과금에서 제외
							if s.HasPrefix(s.ToUpper(message_type.String), "F") { // 친구톡 이면
								if s.EqualFold(message_type.String, "FC") {
									err_ftcscnt++
								} else if s.EqualFold(message_type.String, "FL") {
									err_ftilcnt++
								} else {
									if len(image_url.String) <= 0 { // Image Url 이 null 이면 텍스트 친구톡
										err_ftcnt++
									} else {
										err_fticnt++
									}
								}
							} else if s.EqualFold(message_type.String, "at") || s.EqualFold(message_type.String, "al") || s.EqualFold(message_type.String, "ai") {
								err_atcnt++
							} else if s.HasPrefix(s.ToUpper(message_type.String), "B") || s.HasPrefix(s.ToUpper(message_type.String), "C") { // BrandMsg 실패 카운트 증가
								err_ftcnt++
							}
						}
					}

					// 알림톡 2 차 발신이 아니면 cb_msg 에 insert 처리
					msginsStrs = append(msginsStrs, "(?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?)")
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
					if s.HasPrefix(s.ToUpper(cb_msg_message_type), "D") && kind.String == "F" {
						msginsValues = append(msginsValues, remark3.String+"/"+price.String)
					} else {
						msginsValues = append(msginsValues, remark3)
					}
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
					tmplId := ""
					if len(tmpl_id.String) > 30 {
						tmplId = tmpl_id.String[:30]
					}
					msginsValues = append(msginsValues, tmplId)
					msginsValues = append(msginsValues, mem_userid)
					msginsValues = append(msginsValues, wide)

					if vancnt.String == "0" && isPayment { // 수신 거부 이면 금액 차감 에서 제외
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
					errlog.Println("tblresultproc - MSG Table Insert 처리 중 오류 발생 " + err.Error())
				}

				msginsStrs = nil
				msginsValues = nil
			}

			if len(ftlistsStrs) >= 1000 {
				stmt := fmt.Sprintf("insert IGNORE into cb_friend_list(mem_id, phn, last_send_date) values %s", s.Join(ftlistsStrs, ","))
				_, err := db.Exec(stmt, ftlistsValues...)

				if err != nil {
					errlog.Println("tblresultproc - FT List Table Insert 처리 중 오류 발생 " + err.Error())
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
					errlog.Println("tblresultproc - Result Table Update 처리 중 오류 발생 ")
				}

				upmsgids = nil
			}

			if len(amtsStrs) >= 1000 {
				stmt := fmt.Sprintf(amtinsstr, s.Join(amtsStrs, ","))
				_, err := db.Exec(stmt, amtsValues...)

				if err != nil {
					errlog.Println("tblresultproc - AMT Table Insert 처리 중 오류 발생 " + err.Error())
				}

				amtsStrs = nil
				amtsValues = nil
			}

			if len(ossmsStrs) >= 1000 {
				stmt := fmt.Sprintf("insert into OShotSMS(Sender,Receiver,Msg,URL,ReserveDT,TimeoutDT,SendResult,mst_id,cb_msg_id,IdentityCode) values %s", s.Join(ossmsStrs, ","))
				_, err := db.Exec(stmt, ossmsValues...)

				if err != nil {
					errlog.Println("tblresultproc - 스마트미 SMS Table Insert 처리 중 오류 발생 " + err.Error())
				}

				ossmsStrs = nil
				ossmsValues = nil
			}

			if len(osmmsStrs) >= 1000 {
				stmt := fmt.Sprintf("insert into OShotMMS(MsgGroupID,Sender,Receiver,Subject,Msg,ReserveDT,TimeoutDT,SendResult,File_Path1,File_Path2,File_Path3,mst_id,cb_msg_id,IdentityCode) values %s", s.Join(osmmsStrs, ","))
				_, err := db.Exec(stmt, osmmsValues...)

				if err != nil {
					errlog.Println("tblresultproc - 스마트미 LMS Table Insert 처리 중 오류 발생 " + err.Error())
				}

				osmmsStrs = nil
				osmmsValues = nil
			}

			if len(lgusmsStrs) >= 1000 {
				stmt := fmt.Sprintf("insert into LG_SC_TRAN(TR_SENDDATE,TR_PHONE,TR_CALLBACK, TR_MSG, TR_ETC1, TR_ETC2, TR_ETC3, TR_KISAORIGCODE) values %s", s.Join(lgusmsStrs, ","))
				_, err := db.Exec(stmt, lgusmsValues...)

				if err != nil {
					errlog.Println("tblresultproc - LGU SMS Table Insert 처리 중 오류 발생 " + err.Error())
				}

				lgusmsStrs = nil
				lgusmsValues = nil
			}

			if len(lgummsStrs) >= 1000 {
				stmt := fmt.Sprintf("insert into LG_MMS_MSG(SUBJECT, PHONE, CALLBACK, REQDATE, MSG, FILE_CNT, FILE_PATH1, FILE_PATH2, FILE_PATH3, ETC1, ETC2, ETC3, KISA_ORIGCODE) values %s", s.Join(lgummsStrs, ","))
				_, err := db.Exec(stmt, lgummsValues...)

				if err != nil {
					errlog.Println("tblresultproc - LGU LMS Table Insert 처리 중 오류 발생 " + err.Error())
				}

				lgummsStrs = nil
				lgummsValues = nil
			}

			if len(nnsmsStrs) >= 1000 {
				stmt := fmt.Sprintf("insert into SMS_MSG(TR_CALLBACK,TR_PHONE,TR_MSG,TR_SENDDATE,TR_SENDSTAT,TR_MSGTYPE,TR_ETC9,TR_ETC10,TR_IDENTIFICATION_CODE,TR_ETC8) values %s", s.Join(nnsmsStrs, ","))
				_, err := db.Exec(stmt, nnsmsValues...)

				if err != nil {
					errlog.Println("tblresultproc - 나노 SMS Table Insert 처리 중 오류 발생 " + err.Error())
				}

				nnsmsStrs = nil
				nnsmsValues = nil
			}

			if len(nnmmsStrs) >= 1000 {
				stmt := fmt.Sprintf("insert into MMS_MSG(CALLBACK,PHONE,SUBJECT,MSG,REQDATE,STATUS,FILE_CNT,FILE_PATH1,FILE_PATH2,FILE_PATH3,ETC9,ETC10,IDENTIFICATION_CODE,ETC8) values %s", s.Join(nnmmsStrs, ","))
				_, err := db.Exec(stmt, nnmmsValues...)

				if err != nil {
					errlog.Println("tblresultproc - 나노 LMS Table Insert 처리 중 오류 발생 " + err.Error())
				}

				nnmmsStrs = nil
				nnmmsValues = nil
			}

			if len(nnlpsmsStrs) >= 1000 {
				stmt := fmt.Sprintf("insert into SMS_MSG_G(TR_CALLBACK,TR_PHONE,TR_MSG,TR_SENDDATE,TR_SENDSTAT,TR_MSGTYPE,TR_ETC9,TR_ETC10,TR_IDENTIFICATION_CODE,TR_ETC8) values %s", s.Join(nnlpsmsStrs, ","))
				_, err := db.Exec(stmt, nnlpsmsValues...)

				if err != nil {
					errlog.Println("tblresultproc - 나노 SMS Table Insert 처리 중 오류 발생 " + err.Error())
				}

				nnlpsmsStrs = nil
				nnlpsmsValues = nil
			}

			if len(nnlpmmsStrs) >= 1000 {
				stmt := fmt.Sprintf("insert into MMS_MSG_G(CALLBACK,PHONE,SUBJECT,MSG,REQDATE,STATUS,FILE_CNT,FILE_PATH1,FILE_PATH2,FILE_PATH3,ETC9,ETC10,IDENTIFICATION_CODE,ETC8) values %s", s.Join(nnlpmmsStrs, ","))
				_, err := db.Exec(stmt, nnlpmmsValues...)

				if err != nil {
					errlog.Println("tblresultproc - 나노 LMS Table Insert 처리 중 오류 발생 " + err.Error())
				}

				nnlpmmsStrs = nil
				nnlpmmsValues = nil
			}

			if len(tntsmsStrs) >= 1000 {
				stmt := fmt.Sprintf("insert into Msg_Tran(Phone_No,Callback_No,Msg_Type,Send_Time,Save_Time,Message,Reseller_Code,Etc1,Etc2,Etc3) values %s", s.Join(tntsmsStrs, ","))
				_, err := db.Exec(stmt, tntsmsValues...)

				if err != nil {
					errlog.Println("tblresultproc - SMTNT SMS Table Insert 처리 중 오류 발생 " + err.Error())
				}

				tntsmsStrs = nil
				tntsmsValues = nil
			}

			if len(tntmmsStrs) >= 1000 {
				stmt := fmt.Sprintf("insert into Msg_Tran(Phone_No,Callback_No,Msg_Type,Send_Time,Save_Time,Subject,Message,File_Count,File_Type1,File_Type2,File_Type3,File_Name1,File_Name2,File_Name3,Reseller_Code,Etc1,Etc2,Etc3) values %s", s.Join(tntmmsStrs, ","))
				_, err := db.Exec(stmt, tntmmsValues...)

				if err != nil {
					errlog.Println("tblresultproc - SMTNT LMS Table Insert 처리 중 오류 발생 " + err.Error())
				}

				tntmmsStrs = nil
				tntmmsValues = nil
			}

			if len(jjsmsStrs) >= 1000 {
				stmt := fmt.Sprintf("insert into MTMSG_DATA(CALL_TO,CALL_FROM,MSG_TYPE,SUBJECT,MESSAGE,IDENTIFIER,DHN_ETC1,DHN_ETC2,DHN_ETC3) values %s", s.Join(jjsmsStrs, ","))
				_, err := db.Exec(stmt, jjsmsValues...)

				if err != nil {
					errlog.Println("tblresultproc - JJ SMS Table Insert 처리 중 오류 발생 " + err.Error())
				}

				jjsmsStrs = nil
				jjsmsValues = nil
			}

			if len(jjmmsStrs) >= 1000 {
				stmt := fmt.Sprintf("insert into MTMSG_DATA(CALL_TO,CALL_FROM,MSG_TYPE,SUBJECT,MESSAGE,FILE_NAME1,FILE_NAME2,FILE_NAME3,IDENTIFIER,DHN_ETC1,DHN_ETC2,DHN_ETC3) values %s", s.Join(jjmmsStrs, ","))
				_, err := db.Exec(stmt, jjmmsValues...)

				if err != nil {
					errlog.Println("tblresultproc - JJ LMS Table Insert 처리 중 오류 발생 " + err.Error())
				}

				jjmmsStrs = nil
				jjmmsValues = nil
			}

			if len(rcsStrs) >= 1000 {
				stmt := fmt.Sprintf("insert into RCS_MESSAGE(msg_id, user_contact, schedule_type, msg_group_id, msg_service_type, chatbot_id,agency_id, messagebase_id, service_type, expiry_option ,header  ,footer  ,copy_allowed ,body, buttons, brand_key, platform) values %s", s.Join(rcsStrs, ","))
				_, err := db.Exec(stmt, rcsValues...)

				if err != nil {
					errlog.Println("tblresultproc - (구) Rcs Table Insert 처리 중 오류 발생 " + err.Error())
				}

				rcsStrs = nil
				rcsValues = nil
			}

		}
		// mst_id 별 Loop 끝

		// DB Insert / Update 처리
		if len(msginsValues) > 0 {
			stmt := fmt.Sprintf(insstr, s.Join(msginsStrs, ","))
			_, err := db.Exec(stmt, msginsValues...)

			if err != nil {
				errlog.Println("tblresultproc - MSG Table Insert 처리 중 오류 발생 " + err.Error())
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
				errlog.Println("tblresultproc - Result Table Update 처리 중 오류 발생 ")
			}
		}

		if len(ftlistsStrs) > 0 {
			stmt := fmt.Sprintf("insert IGNORE into cb_friend_list(mem_id, phn, last_send_date) values %s", s.Join(ftlistsStrs, ","))
			_, err := db.Exec(stmt, ftlistsValues...)

			if err != nil {
				errlog.Println("tblresultproc - FT List Table Insert 처리 중 오류 발생 " + err.Error())
			}
		}

		if len(atinsids) > 0 {
			var copystr = "update " + conf.REQTABLE2 + " set remark3 = 'Y' where MSGID in ("

			for i := 1; i < len(atinsids); i++ {
				copystr = copystr + "?,"
			}

			copystr = copystr + "?)"

			_, err1 := db.Exec(copystr, atinsids...)

			if err1 != nil {
				errlog.Println("tblresultproc - 2ND 테이블에서 복사 처리 중 오류 발생 ")
				errlog.Println(err1)
				errlog.Println(copystr)
			} else {
				errlog.Println("tblresultproc - 2ND 테이블에서 복사 처리 완료 : ", len(atinsids))
			}
		}

		if len(amtsStrs) > 0 {
			stmt := fmt.Sprintf(amtinsstr, s.Join(amtsStrs, ","))
			_, err := db.Exec(stmt, amtsValues...)

			if err != nil {
				errlog.Println("tblresultproc - AMT Table Insert 처리 중 오류 발생 " + err.Error())
			}
		}

		if len(ossmsStrs) > 0 {
			stmt := fmt.Sprintf("insert into OShotSMS(Sender,Receiver,Msg,URL,ReserveDT,TimeoutDT,SendResult,mst_id,cb_msg_id,IdentityCode) values %s", s.Join(ossmsStrs, ","))
			_, err := db.Exec(stmt, ossmsValues...)

			if err != nil {
				errlog.Println("tblresultproc - 스마트미 SMS Table Insert 처리 중 오류 발생 " + err.Error())
			}
		}

		if len(osmmsStrs) > 0 {
			stmt := fmt.Sprintf("insert into OShotMMS(MsgGroupID,Sender,Receiver,Subject,Msg,ReserveDT,TimeoutDT,SendResult,File_Path1,File_Path2,File_Path3,mst_id,cb_msg_id,IdentityCode) values %s", s.Join(osmmsStrs, ","))
			_, err := db.Exec(stmt, osmmsValues...)

			if err != nil {
				errlog.Println("tblresultproc - 스마트미 LMS Table Insert 처리 중 오류 발생 " + err.Error())
			}
		}

		if len(lgusmsStrs) > 0 {
			stmt := fmt.Sprintf("insert into LG_SC_TRAN(TR_SENDDATE,TR_PHONE,TR_CALLBACK, TR_MSG, TR_ETC1, TR_ETC2, TR_ETC3, TR_KISAORIGCODE) values %s", s.Join(lgusmsStrs, ","))
			_, err := db.Exec(stmt, lgusmsValues...)

			if err != nil {
				errlog.Println("tblresultproc - LGU SMS Table Insert 처리 중 오류 발생 " + err.Error())
			}
		}

		if len(lgummsStrs) > 0 {
			stmt := fmt.Sprintf("insert into LG_MMS_MSG(SUBJECT, PHONE, CALLBACK, REQDATE, MSG, FILE_CNT, FILE_PATH1, FILE_PATH2, FILE_PATH3, ETC1, ETC2, ETC3, KISA_ORIGCODE) values %s", s.Join(lgummsStrs, ","))
			_, err := db.Exec(stmt, lgummsValues...)

			if err != nil {
				errlog.Println("tblresultproc - LGU LMS Table Insert 처리 중 오류 발생 " + err.Error())
			}
		}

		if len(nnsmsStrs) > 0 {
			stmt := fmt.Sprintf("insert into SMS_MSG(TR_CALLBACK,TR_PHONE,TR_MSG,TR_SENDDATE,TR_SENDSTAT,TR_MSGTYPE,TR_ETC9,TR_ETC10,TR_IDENTIFICATION_CODE,TR_ETC8) values %s", s.Join(nnsmsStrs, ","))
			_, err := db.Exec(stmt, nnsmsValues...)

			if err != nil {
				errlog.Println("tblresultproc - 나노 SMS Table Insert 처리 중 오류 발생 " + err.Error())
			}
		}

		if len(nnmmsStrs) > 0 {
			stmt := fmt.Sprintf("insert into MMS_MSG(CALLBACK,PHONE,SUBJECT,MSG,REQDATE,STATUS,FILE_CNT,FILE_PATH1,FILE_PATH2,FILE_PATH3,ETC9,ETC10,IDENTIFICATION_CODE,ETC8) values %s", s.Join(nnmmsStrs, ","))
			_, err := db.Exec(stmt, nnmmsValues...)

			if err != nil {
				errlog.Println("tblresultproc - 나노 LMS Table Insert 처리 중 오류 발생 " + err.Error())
			}
		}

		if len(nnlpsmsStrs) > 0 {
			stmt := fmt.Sprintf("insert into SMS_MSG_G(TR_CALLBACK,TR_PHONE,TR_MSG,TR_SENDDATE,TR_SENDSTAT,TR_MSGTYPE,TR_ETC9,TR_ETC10,TR_IDENTIFICATION_CODE,TR_ETC8) values %s", s.Join(nnlpsmsStrs, ","))
			_, err := db.Exec(stmt, nnlpsmsValues...)

			if err != nil {
				errlog.Println("tblresultproc - 나노 저가망 SMS Table Insert 처리 중 오류 발생 " + err.Error())
			}
		}

		if len(nnlpmmsStrs) > 0 {
			stmt := fmt.Sprintf("insert into MMS_MSG_G(CALLBACK,PHONE,SUBJECT,MSG,REQDATE,STATUS,FILE_CNT,FILE_PATH1,FILE_PATH2,FILE_PATH3,ETC9,ETC10,IDENTIFICATION_CODE,ETC8) values %s", s.Join(nnlpmmsStrs, ","))
			_, err := db.Exec(stmt, nnlpmmsValues...)

			if err != nil {
				errlog.Println("tblresultproc - 나노 저가망 LMS Table Insert 처리 중 오류 발생 " + err.Error())
			}
		}

		if len(tntsmsStrs) > 0 {
			stmt := fmt.Sprintf("insert into Msg_Tran(Phone_No,Callback_No,Msg_Type,Send_Time,Save_Time,Message,Reseller_Code,Etc1,Etc2,Etc3) values %s", s.Join(tntsmsStrs, ","))
			_, err := db.Exec(stmt, tntsmsValues...)

			if err != nil {
				errlog.Println("tblresultproc - SMTNT SMS Table Insert 처리 중 오류 발생 " + err.Error())
			}
		}

		if len(tntmmsStrs) > 0 {
			stmt := fmt.Sprintf("insert into Msg_Tran(Phone_No,Callback_No,Msg_Type,Send_Time,Save_Time,Subject,Message,File_Count,File_Type1,File_Type2,File_Type3,File_Name1,File_Name2,File_Name3,Reseller_Code,Etc1,Etc2,Etc3) values %s", s.Join(tntmmsStrs, ","))
			_, err := db.Exec(stmt, tntmmsValues...)

			if err != nil {
				errlog.Println("tblresultproc - SMTNT LMS Table Insert 처리 중 오류 발생 " + err.Error())
			}
		}

		if len(jjsmsStrs) > 0 {
			stmt := fmt.Sprintf("insert into MTMSG_DATA(CALL_TO,CALL_FROM,MSG_TYPE,SUBJECT,MESSAGE,IDENTIFIER,DHN_ETC1,DHN_ETC2,DHN_ETC3) values %s", s.Join(jjsmsStrs, ","))
			_, err := db.Exec(stmt, jjsmsValues...)

			if err != nil {
				errlog.Println("tblresultproc - JJ SMS Table Insert 처리 중 오류 발생 " + err.Error())
			}
		}

		if len(jjmmsStrs) > 0 {
			stmt := fmt.Sprintf("insert into MTMSG_DATA(CALL_TO,CALL_FROM,MSG_TYPE,SUBJECT,MESSAGE,FILE_NAME1,FILE_NAME2,FILE_NAME3,IDENTIFIER,DHN_ETC1,DHN_ETC2,DHN_ETC3) values %s", s.Join(jjmmsStrs, ","))
			_, err := db.Exec(stmt, jjmmsValues...)

			if err != nil {
				errlog.Println("tblresultproc - JJ LMS Table Insert 처리 중 오류 발생 " + err.Error())
			}
		}

		if len(rcsStrs) > 0 {
			stmt := fmt.Sprintf("insert into RCS_MESSAGE(msg_id, user_contact, schedule_type, msg_group_id, msg_service_type, chatbot_id,agency_id, messagebase_id, service_type, expiry_option ,header  ,footer  ,copy_allowed ,body, buttons, brand_key, platform) values %s", s.Join(rcsStrs, ","))
			_, err := db.Exec(stmt, rcsValues...)

			if err != nil {
				errlog.Println("tblresultproc - Rcs KT Table Insert 처리 중 오류 발생 " + err.Error())
			}
		}

		if len(remark4.String) > 0 {
			var cntupdate = `
				update
					cb_wt_msg_sent 
				set
					mst_ft = ifnull(mst_ft,0) + ?
		          , mst_ft_img = ifnull(mst_ft_img,0) + ?
				  , mst_at = ifnull(mst_at,0) + ? 
				  , mst_smt = ifnull(mst_smt,0) + 0
				  , mst_cs = ifnull(mst_cs,0) + ? 
				  , mst_il = ifnull(mst_il,0) + ?
				  , mst_err_ft = ifnull(mst_err_ft,0) + ?
			      , mst_err_ft_img = ifnull(mst_err_ft_img,0) + ?
				  , mst_err_at = ifnull(mst_err_at,0) + ?
				  , mst_err_smt = ifnull(mst_err_smt,0) + ? 
				  , mst_err_rcs = ifnull(mst_err_rcs,0) + ?  
				  , mst_err_cs = ifnull(mst_err_cs,0) + ?  
				  , mst_err_il = ifnull(mst_err_il,0) + ?  
				  , mst_wait = ifnull(mst_wait,0) + ? 
				where
					mst_id = ?`
			_, err := db.Exec(cntupdate, ftcnt, fticnt, atcnt, ftcscnt, ftilcnt, err_ftcnt, err_fticnt, err_atcnt, err_smtcnt, err_rcscnt, err_ftcscnt, err_ftilcnt, mst_waitcnt, bcCnt, ftncnt, sendkey)

			if err != nil {
				errlog.Println("tblresultproc - cb_wt_msg_sent 카카오 메세지 수량 처리 중 오류 발생 " + err.Error())
			}
		}

		if cnt > 0 {
			stdlog.Printf("tblresultproc - ( %s ) Result 처리 - %s : %d 건 처리 완료", startTime, ressendkey.String, cnt)
		}

		// 2차 알림톡 2일 지난건 삭제 함.
		db.Exec("delete a from " + conf.REQTABLE2 + " a where  ( ( a.reserve_dt < DATE_FORMAT(ADDDATE(now(), INTERVAL -2 DAY), '%Y%m%d%H%i%S') and a.reserve_dt <> '00000000000000') or ( a.REG_DT < ADDDATE(now(), INTERVAL -2 DAY) and a.reserve_dt = '00000000000000'))")
	}
}
