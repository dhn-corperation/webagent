package handler

import(
	// "time"
	// s "strings"
	"context"
	"strconv"

	"webagent/src/config"

	"github.com/jmoiron/sqlx"
)



func Resend(ctx context.Context, db *sqlx.DB, target, sd string) {
	for {
		select {
		case <- ctx.Done():
			config.Stdlog.Println("대기 발송 전환이 종료되었습니다 / 타겟 : " + target)
			return
		default:
			config.Stdlog.Println("대기 발송 전환 루프 시작 / 타겟 : " + target)
			if target == "nano" {
				if !oshotToNano(db, sd) {
					ctx.Done()
				}
			} else if target == "oshot" {
				if !nanoToOshot(db, sd) {
					ctx.Done()
				}
			}
		}
	}
}

func oshotToNano(db *sqlx.DB, sd string) bool {
	infolog := config.Stdlog
	errlog := config.Stdlog
	
	var oshotSmsDataList []OshotSmsTable
	// var oshotMmsDataList []OshotMmsTable
	var nanoSmsDataList []NanoSmsTable
	// var nanoMmsDataList []NanoSmsTable

	infolog.Println("oshotToNano sms 처리 시작 start dt : ", sd)
	err := db.Select(&oshotSmsDataList, "select * from OShotSMS where resend_flag = '0' and InsertDT >= ?", sd)
	if err != nil {
		errlog.Println("oshotToNano / OShotSMS 조회 실패 / err : ", err)
	}
	var smsUpdateId []int
	for _, smsData := range oshotSmsDataList {
		nanoSmsDataList = append(nanoSmsDataList, NanoSmsTable{
			SendDate: smsData.InsertDt,
			Phone: smsData.Receiver,
			Callback: smsData.Sender,
			Msg: smsData.Msg,
			IdentificationCode: "302190001",
			Etc9: strconv.Itoa(smsData.MstId),
			Etc10: smsData.CbMsgId,
		})
		smsUpdateId = append(smsUpdateId, smsData.MsgId)
	}

	smsInsertQuery := `
		insert into SMS_MSG(TR_SENDATE, TR_PHONE, TR_CALLBACK, TR_MSG, TR_IDENTIFICATION_CODE, TR_ETC9, TR_ETC10)
		values (:TR_SENDATE, :TR_PHONE, :TR_CALLBACK, :TR_MSG, :TR_IDENTIFICATION_CODE, :TR_ETC9, :TR_ETC10)
	`
	if len(nanoSmsDataList) > 0 {
		_, err = db.NamedExec(smsInsertQuery, nanoSmsDataList)

		if err != nil {
			errlog.Println("oshotToNano / SMS_MSG 삽입 실패 / err : ", err)
			return false
		}

		smsUpdateQuery, args, err := sqlx.In(`update OShotSMS set resend_flag = '1' where MsgID IN (?)`, smsUpdateId)
		if err != nil {
			errlog.Println("oshotToNano / OShotSMS 재발송 flag 변환 Sql 생성 실패 / err : ", err)
			return false
		}

		smsUpdateQuery = db.Rebind(smsUpdateQuery)

		smsUpdateResult, err := db.Exec(smsUpdateQuery, args...)

		if err != nil {
			errlog.Println("oshotToNano / OShotSMS 재발송 flag 변환 실패 / err : ", err)
		}

		smsUpdateRowCnt, _ := smsUpdateResult.RowsAffected()
		infolog.Println("oshotToNano sms 처리 끝 업데이트 건수 : ", smsUpdateRowCnt)
	}
	
	return true
}

func nanoToOshot(db *sqlx.DB, sd string) bool {
	return true
}