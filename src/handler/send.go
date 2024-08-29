package handler

import(
	"time"
	// s "strings"
	"context"

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
		time.Sleep(5 * time.Second)
	}
}

func oshotToNano(db *sqlx.DB, sd string) bool {
	infolog := config.Stdlog
	errlog := config.Stdlog

	var smsUpdateId []int
	var mmsUpdateId []int
	var oshotSmsDataList []OshotSmsTable
	var oshotMmsDataList []OshotMmsTable

	tx, err := db.Beginx()
	if err != nil {
		errlog.Println("oshotToNano / sms 트랜잭션 실행 실패 / err : ", err)
		return false
	}

	infolog.Println("oshotToNano sms 처리 시작 / sd : ", sd)
	err = tx.Select(&oshotSmsDataList, "select * from OShotSMS where resend_flag = '0' and InsertDT >= ?", sd)
	if err != nil {
		errlog.Println("oshotToNano / OShotSMS 조회 실패 / err : ", err)
		return false
	}
	
	smsInsertQuery := `
		insert into SMS_MSG(TR_SENDDATE, TR_PHONE, TR_CALLBACK, TR_MSG, TR_IDENTIFICATION_CODE, TR_ETC9, TR_ETC10)
		values (:TR_SENDDATE, :TR_PHONE, :TR_CALLBACK, :TR_MSG, :TR_IDENTIFICATION_CODE, :TR_ETC9, :TR_ETC10)
	`

	if len(oshotSmsDataList) > 0 {
		for _, smsData := range oshotSmsDataList {
			mapData := map[string]interface{}{
				"TR_SENDDATE": smsData.InsertDt,
				"TR_PHONE": smsData.Receiver,
				"TR_CALLBACK": smsData.Sender,
				"TR_MSG": smsData.Msg,
				"TR_IDENTIFICATION_CODE": "302190001",
				"TR_ETC9": smsData.MstId,
				"TR_ETC10": smsData.CbMsgId,
			}

			_, err := tx.NamedExec(smsInsertQuery, mapData)
			if err != nil {
				errlog.Println("oshotToNano / insert 실패 / OShotSMS의 MsgID 값 : ", smsData.MsgId, " / err : ", err)
			} else {
				smsUpdateId = append(smsUpdateId, smsData.MsgId)
			}
		}

		if len(smsUpdateId) > 0 {
			smsUpdateQuery, args, err := sqlx.In(`update OShotSMS set resend_flag = '1' where MsgID IN (?)`, smsUpdateId)
			if err != nil {
				errlog.Println("oshotToNano / OShotSMS 재발송 flag 변환 Sql 생성 실패 / err : ", err)
				return false
			}

			smsUpdateQuery = tx.Rebind(smsUpdateQuery)

			smsUpdateResult, err := tx.Exec(smsUpdateQuery, args...)

			if err != nil {
				errlog.Println("oshotToNano / OShotSMS 재발송 flag 변환 실패 / err : ", err)
				tx.Rollback()
				return false
			}

			smsUpdateRowCnt, _ := smsUpdateResult.RowsAffected()

			err = tx.Commit()
			if err != nil {
				errlog.Println("oshotToNano / sms commit 실패 / err : ", err)
				return false
			} else {
				infolog.Println("oshotToNano sms 처리 끝 / sd :", sd, " / 업데이트 건수 : ", smsUpdateRowCnt)
			}
		}
	}

	tx, err = db.Beginx()
	if err != nil {
		errlog.Println("oshotToNano / mms 트랜잭션 실행 실패 / err : ", err)
		return false
	}

	infolog.Println("oshotToNano mms 처리 시작 / sd : ", sd)
	err = tx.Select(&oshotMmsDataList, "select * from OShotMMS where resend_flag = '0' and InsertDT >= ?", sd)
	if err != nil {
		errlog.Println("oshotToNano / OShotMMS 조회 실패 / err : ", err)
		return false
	}
	
	mmsInsertQuery := `
		insert into MMS_MSG(SUBJECT, PHONE, CALLBACK, REQDATE, MSG, FILE_CNT, FILE_PATH1, FILE_PATH2, FILE_PATH3, IDENTIFICATION_CODE, ETC9, ETC10)
		values (:SUBJECT, :PHONE, :CALLBACK, :REQDATE, :MSG, :FILE_CNT, :FILE_PATH1, :FILE_PATH2, :FILE_PATH3, :IDENTIFICATION_CODE, :ETC9, :ETC10)
	`
	
	if len(oshotMmsDataList) > 0 {
		for _, mmsData := range oshotMmsDataList {
			fc := 0
			if len(mmsData.FilePath1.String) > 0 {
				fc++
			}
			if len(mmsData.FilePath2.String) > 0 {
				fc++
			}
			if len(mmsData.FilePath3.String) > 0 {
				fc++
			}
			mapData := map[string]interface{}{
				"SUBJECT": mmsData.Subject,
				"PHONE": mmsData.Receiver,
				"CALLBACK": mmsData.Sender,
				"REQDATE": mmsData.InsertDt,
				"MSG": mmsData.Msg,
				"FILE_CNT": fc,
				"FILE_PATH1": mmsData.FilePath1.String,
				"FILE_PATH2": mmsData.FilePath2.String,
				"FILE_PATH3": mmsData.FilePath3.String,
				"IDENTIFICATION_CODE": "302190001",
				"ETC9": mmsData.MstId,
				"ETC10": mmsData.CbMsgId,
			}

			_, err := tx.NamedExec(mmsInsertQuery, mapData)
			if err != nil {
				errlog.Println("oshotToNano / insert 실패 / OShotMMS의 MsgID 값 : ", mmsData.MsgId, " / err : ", err)
			} else {
				mmsUpdateId = append(mmsUpdateId, mmsData.MsgId)
			}
		}

		if len(mmsUpdateId) > 0 {
			mmsUpdateQuery, args, err := sqlx.In(`update OShotMMS set resend_flag = '1' where MsgID IN (?)`, mmsUpdateId)
			if err != nil {
				errlog.Println("oshotToNano / OShotMMS 재발송 flag 변환 Sql 생성 실패 / err : ", err)
				return false
			}

			mmsUpdateQuery = tx.Rebind(mmsUpdateQuery)

			mmsUpdateResult, err := tx.Exec(mmsUpdateQuery, args...)

			if err != nil {
				errlog.Println("oshotToNano / OShotMMS 재발송 flag 변환 실패 / err : ", err)
				tx.Rollback()
				return false
			}

			mmsUpdateRowCnt, _ := mmsUpdateResult.RowsAffected()

			err = tx.Commit()
			if err != nil {
				errlog.Println("oshotToNano / mms commit 실패 / err : ", err)
				return false
			} else {
				infolog.Println("oshotToNano mms 처리 끝 / sd :", sd, " / 업데이트 건수 : ", mmsUpdateRowCnt)
			}
		}
	}
	
	
	return true
}

func nanoToOshot(db *sqlx.DB, sd string) bool {
	// var nanoSmsDataList []NanoSmsTable
	// var nanoMmsDataList []NanoSmsTable
	return true
}