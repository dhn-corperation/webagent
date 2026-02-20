package rcs

import(
	"fmt"
	"time"
	"sync"
	"context"
	s "strings"
	"database/sql"

	"webagent/src/config"
	"webagent/src/databasepool"
)

func ProcessSmtnt(ctx context.Context) {
	config.Stdlog.Println("Rcs SMTNT - 발송 프로세스 시작")
	var wg sync.WaitGroup
	for {
		select {
		case <- ctx.Done():
			config.Stdlog.Println("Rcs SMTNT - process가 15초 후에 종료")
			time.Sleep(15 * time.Second)
			config.Stdlog.Println("Rcs SMTNT - process 종료 완료")
			return
		default:
			wg.Add(1)
			go rcsProcessSmtnt(&wg)
			wg.Wait()
		}
	}
}

func rcsProcessSmtnt(wg *sync.WaitGroup) {
	defer wg.Done()
	defer func() {
		if r := recover(); r != nil {
			config.Stdlog.Println("Rcs SMTNT - rcssend panic 발생 원인 : ", r)
			if err, ok := r.(error); ok {
				if s.Contains(err.Error(), "connection refused") {
					for {
						config.Stdlog.Println("Rcs SMTNT - rcssend send ping to DB")
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

	reqsql := `
		SELECT
			rcs_id,
			msg_id,
			user_contact,
			schedule_type,
			msg_group_id,
			msg_service_type,
			chatbot_id,
			agency_id,
			messagebase_id,
			service_type,
			expiry_option,
			header,
			footer,
			copy_allowed,
			body,
			buttons,
			agency_key,
			brand_key
		FROM
			RCS_MESSAGE
		WHERE
			platform = 'SMTNT'
		LIMIT 0, 500`

	reqrows, err := db.Query(reqsql)
	if err != nil {
		config.Stdlog.Println("Rcs SMTNT - RCS_MESSAGE Table Select 처리 중 오류 발생 " + err.Error())
		panic(err)
	}
	defer reqrows.Close()

	var procCount int = 0
	var rcsId, msgId, userContact, scheduleType, msgGroupId, msgServiceType, chatbotId, agencyId, messagebaseId, serviceType sql.NullString
	var expiryOption, header, footer, copyAllowd, body, buttons, agencyKey, brandKey sql.NullString

	delrcsids := []interface{}{}

	resinsStrs := []string{}
	resinsValues := []interface{}{}
	resinsquery := `
		insert IGNORE into RCS_MESSAGE_RESULT(
			rcs_id,
			msg_id,
			user_contact,
			schedule_type,
			msg_group_id,
			msg_service_type,
			chatbot_id,
			agency_id,
			messagebase_id,
			service_type,
			expiry_option,
			header,
			footer,
			copy_allowed,
			body,
			buttons,
			status,
			sentTime,
			timestamp,
			error,
			proc,
			platform) values %s`

	reqinsStrs := []string{}
	reqinsValues := []interface{}{}
	reqinsquery := `
		insert IGNORE into Msg_Tran(
			Phone_No,
			Callback_No,
			Message,
			Status,
			Msg_Type,
			Send_Time,
			Save_Time,
			File_Count,
			Spam_Check,
			Rcs_MessageBaseId,
			Rcs_Button,
			Rcs_Header,
			Rcs_Footer,
			Rcs_CopyAllowed,
			Rcs_AgencyId,
			Rcs_AgencyKey,
			Rcs_BrandKey,
			Etc1,
			Etc3
		) values %s`

	for reqrows.Next() {
		reqrows.Scan(&rcsId, &msgId, &userContact, &scheduleType, &msgGroupId, &msgServiceType, &chatbotId, &agencyId, &messagebaseId, &serviceType,
			&expiryOption, &header, &footer, &copyAllowd, &body, &buttons, &agencyKey, &brandKey)

		resinsStrs = append(resinsStrs, "(?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,'200',now(),now(),'','N',?)")
		resinsValues = append(resinsValues, rcsId.String)
		resinsValues = append(resinsValues, msgId.String)
		resinsValues = append(resinsValues, userContact.String)
		resinsValues = append(resinsValues, scheduleType.String)
		resinsValues = append(resinsValues, msgGroupId.String)
		resinsValues = append(resinsValues, msgServiceType.String)
		resinsValues = append(resinsValues, chatbotId.String)
		resinsValues = append(resinsValues, agencyId.String)
		resinsValues = append(resinsValues, messagebaseId.String)
		resinsValues = append(resinsValues, serviceType.String)
		resinsValues = append(resinsValues, expiryOption.String)
		resinsValues = append(resinsValues, header.String)
		resinsValues = append(resinsValues, footer.String)
		resinsValues = append(resinsValues, copyAllowd.String)
		resinsValues = append(resinsValues, body.String)
		resinsValues = append(resinsValues, buttons.String)
		resinsValues = append(resinsValues, "SMTNT")

		var smtntMsgType string
		var smtntCopyAllowd string = "N"

		switch  serviceType.String {
			case "RCSSMS":
				smtntMsgType = "9"
				break
			case "RCSLMS":
				smtntMsgType = "10"
				break
			case "RCSMMS":
				smtntMsgType = "11"
				break
			case "RCSTMPL":
				smtntMsgType = "12"
				break
		}

		if copyAllowd.String == "1" {
			smtntCopyAllowd = "Y"
		}

		reqinsStrs = append(reqinsStrs, "(?,?,?,'0',?,now(),now(),0,'N',?,?,?,?,?,?,?,?,?,?)")
		reqinsValues = append(reqinsValues, userContact.String) // Phone_No
		reqinsValues = append(reqinsValues, chatbotId.String) // Callback_No
		reqinsValues = append(reqinsValues, body.String) // Message
		reqinsValues = append(reqinsValues, smtntMsgType) // Msg_Type
		reqinsValues = append(reqinsValues, messagebaseId.String) // Rcs_MessageBaseId
		reqinsValues = append(reqinsValues, buttons.String) // Rcs_Button
		reqinsValues = append(reqinsValues, header.String) // Rcs_Header
		reqinsValues = append(reqinsValues, footer.String) // Rcs_Footer
		reqinsValues = append(reqinsValues, smtntCopyAllowd) // Rcs_CopyAllowed
		reqinsValues = append(reqinsValues, agencyId.String) // Rcs_AgencyId
		reqinsValues = append(reqinsValues, agencyKey.String) // Rcs_AgencyKey
		reqinsValues = append(reqinsValues, brandKey.String) // Rcs_BrandKey
		reqinsValues = append(reqinsValues, msgId.String) // Etc1
		reqinsValues = append(reqinsValues, msgGroupId.String) // Etc3

		delrcsids = append(delrcsids, rcsId.String)

		procCount++
	}

	if len(resinsStrs) > 0 {
		stmt := fmt.Sprintf(resinsquery, s.Join(resinsStrs, ","))
		_, err := db.Exec(stmt, resinsValues...)

		if err != nil {
			config.Stdlog.Println("Rcs SMTNT - RCS_MESSAGE_RESULT Table Insert 처리 중 오류 발생 " + err.Error())
		} else {
			stmt := fmt.Sprintf(reqinsquery, s.Join(reqinsStrs, ","))
			_, err := db.Exec(stmt, reqinsValues...)

			if err != nil {
				config.Stdlog.Println("Rcs SMTNT - SMTNT 발송 Table Insert 처리 중 오류 발생 " + err.Error())
			} else {
				if len(delrcsids) > 0 {
					var commastr = "delete from RCS_MESSAGE where rcs_id in ("

					for i := 1; i < len(delrcsids); i++ {
						commastr = commastr + "?,"
					}

					commastr = commastr + "?)"

					_, err1 := db.Exec(commastr, delrcsids...)

					if err1 != nil {
						config.Stdlog.Println("Rcs SMTNT - RCS_MESSAGE Table Delete 처리 중 오류 발생 " + err.Error())
					}
				}
			}
		}
	}

	if procCount > 0 {
		config.Stdlog.Println("Rcs SMTNT - 발송 : ", procCount, " 건 처리 완료 ")
		SendInterval = 1
	} else {
		SendInterval = 1000
	}

	time.Sleep(time.Millisecond * time.Duration(SendInterval))

}