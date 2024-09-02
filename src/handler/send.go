package handler

import(
	"time"
	// s "strings"
	"context"

	"webagent/src/config"

	"github.com/jmoiron/sqlx"
)

const (
	oshotSmsTableName = "OShotSMS"
	oshotMmsTableName = "OShotMMS"
	nanoSmsTableName = "SMS_MSG"
	nanoMmsTableName = "MMS_MSG"
	nanoLowSmsTableName = "SMS_MSG_G"
	nanoLowMmsTableName = "MMS_MSG_G"
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
	err = tx.Select(&oshotSmsDataList, "select * from "+oshotSmsTableName+" where resend_flag = '0' and InsertDT >= ?", sd)
	if err != nil {
		errlog.Println("oshotToNano / ", oshotSmsTableName, " 조회 실패 / err : ", err)
		return false
	}
	
	smsInsertQuery := `
		insert into `+nanoSmsTableName+`(TR_SENDDATE, TR_PHONE, TR_CALLBACK, TR_MSG, TR_IDENTIFICATION_CODE, TR_ETC9, TR_ETC10)
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
				"TR_ETC9": smsData.CbMsgId,
				"TR_ETC10": smsData.MstId,
			}

			_, err := tx.NamedExec(smsInsertQuery, mapData)
			if err != nil {
				errlog.Println("oshotToNano / insert 실패 / ", oshotSmsTableName, "의 MsgID 값 : ", smsData.MsgId, " / err : ", err)
			} else {
				smsUpdateId = append(smsUpdateId, smsData.MsgId)
			}
		}

		if len(smsUpdateId) > 0 {
			smsUpdateQuery, args, err := sqlx.In(`update `+oshotSmsTableName+` set resend_flag = '1' where MsgID IN (?)`, smsUpdateId)
			if err != nil {
				errlog.Println("oshotToNano / ", oshotSmsTableName, " 재발송 flag 변환 Sql 생성 실패 / err : ", err)
				return false
			}

			smsUpdateQuery = tx.Rebind(smsUpdateQuery)

			smsUpdateResult, err := tx.Exec(smsUpdateQuery, args...)

			if err != nil {
				errlog.Println("oshotToNano / ", oshotSmsTableName, " 재발송 flag 변환 실패 / err : ", err)
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
	} else {
		infolog.Println("oshotToNano sms 처리 끝 / sd :", sd, " / 업데이트 건수 : 0")
	}

	tx, err = db.Beginx()
	if err != nil {
		errlog.Println("oshotToNano / mms 트랜잭션 실행 실패 / err : ", err)
		return false
	}

	infolog.Println("oshotToNano mms 처리 시작 / sd : ", sd)
	err = tx.Select(&oshotMmsDataList, "select * from "+oshotMmsTableName+" where resend_flag = '0' and InsertDT >= ?", sd)
	if err != nil {
		errlog.Println("oshotToNano / ", oshotMmsTableName, " 조회 실패 / err : ", err)
		return false
	}
	
	mmsInsertQuery := `
		insert into `+nanoMmsTableName+`(SUBJECT, PHONE, CALLBACK, REQDATE, MSG, FILE_CNT, FILE_PATH1, FILE_PATH2, FILE_PATH3, IDENTIFICATION_CODE, ETC9, ETC10)
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
				"ETC9": mmsData.CbMsgId,
				"ETC10": mmsData.MstId,
			}

			_, err := tx.NamedExec(mmsInsertQuery, mapData)
			if err != nil {
				errlog.Println("oshotToNano / insert 실패 / ", oshotMmsTableName, "의 MsgID 값 : ", mmsData.MsgId, " / err : ", err)
			} else {
				mmsUpdateId = append(mmsUpdateId, mmsData.MsgId)
			}
		}

		if len(mmsUpdateId) > 0 {
			mmsUpdateQuery, args, err := sqlx.In(`update `+oshotMmsTableName+` set resend_flag = '1' where MsgID IN (?)`, mmsUpdateId)
			if err != nil {
				errlog.Println("oshotToNano / ", oshotMmsTableName, " 재발송 flag 변환 Sql 생성 실패 / err : ", err)
				return false
			}

			mmsUpdateQuery = tx.Rebind(mmsUpdateQuery)

			mmsUpdateResult, err := tx.Exec(mmsUpdateQuery, args...)

			if err != nil {
				errlog.Println("oshotToNano / ", oshotMmsTableName, " 재발송 flag 변환 실패 / err : ", err)
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
	} else {
		infolog.Println("oshotToNano mms 처리 끝 / sd :", sd, " / 업데이트 건수 : 0")
	}
	
	
	return true
}

func nanoToOshot(db *sqlx.DB, sd string) bool {
	infolog := config.Stdlog
	errlog := config.Stdlog

	var smsUpdateId []int
	var mmsUpdateId []string
	var nanoSmsDataList []NanoSmsTable
	var nanoMmsDataList []NanoMmsTable

	tx, err := db.Beginx()
	if err != nil {
		errlog.Println("nanoToOshot / sms 트랜잭션 실행 실패 / err : ", err)
		return false
	}

	infolog.Println("nanoToOshot sms 처리 시작 / sd : ", sd)
	err = tx.Select(&nanoSmsDataList, "select * from "+nanoSmsTableName+" where TR_ETC7 is null and TR_SENDDATE >= ?", sd)
	if err != nil {
		errlog.Println("nanoToOshot / ", nanoSmsTableName, " 조회 실패 / err : ", err)
		return false
	}
	
	smsInsertQuery := `
		insert into `+oshotSmsTableName+`(Sender, Receiver, Msg, InsertDT, mst_id, cb_msg_id)
		values (:Sender, :Receiver, :Msg, :InsertDT, :mst_id, :cb_msg_id)
	`

	if len(nanoSmsDataList) > 0 {
		for _, smsData := range nanoSmsDataList {
			mapData := map[string]interface{}{
				"Sender": smsData.Callback,
				"Receiver": smsData.Phone,
				"Msg": smsData.Msg,
				"InsertDT": smsData.SendDate,
				"mst_id": smsData.Etc10,
				"cb_msg_id": smsData.Etc9,
			}

			_, err := tx.NamedExec(smsInsertQuery, mapData)
			if err != nil {
				errlog.Println("nanoToOshot / insert 실패 / ", nanoSmsTableName, "의 MsgID 값 : ", smsData.Num, " / err : ", err)
			} else {
				smsUpdateId = append(smsUpdateId, smsData.Num)
			}
		}

		if len(smsUpdateId) > 0 {
			smsUpdateQuery, args, err := sqlx.In(`update `+nanoSmsTableName+` set TR_ETC7 = '1' where TR_NUM IN (?)`, smsUpdateId)
			if err != nil {
				errlog.Println("nanoToOshot / ", nanoSmsTableName, " 재발송 flag 변환 Sql 생성 실패 / err : ", err)
				return false
			}

			smsUpdateQuery = tx.Rebind(smsUpdateQuery)

			smsUpdateResult, err := tx.Exec(smsUpdateQuery, args...)

			if err != nil {
				errlog.Println("nanoToOshot / ", nanoSmsTableName, " 재발송 flag 변환 실패 / err : ", err)
				tx.Rollback()
				return false
			}

			smsUpdateRowCnt, _ := smsUpdateResult.RowsAffected()

			err = tx.Commit()
			if err != nil {
				errlog.Println("nanoToOshot / sms commit 실패 / err : ", err)
				return false
			} else {
				infolog.Println("nanoToOshot sms 처리 끝 / sd :", sd, " / 업데이트 건수 : ", smsUpdateRowCnt)
			}
		}
	} else {
		infolog.Println("nanoToOshot sms 처리 끝 / sd :", sd, " / 업데이트 건수 : 0")
	}

	tx, err = db.Beginx()
	if err != nil {
		errlog.Println("nanoToOshot / mms 트랜잭션 실행 실패 / err : ", err)
		return false
	}

	infolog.Println("nanoToOshot mms 처리 시작 / sd : ", sd)
	err = tx.Select(&nanoMmsDataList, "select * from "+nanoMmsTableName+" where ETC7 is null and REQDATE >= ?", sd)
	if err != nil {
		errlog.Println("nanoToOshot / ", nanoMmsTableName, " 조회 실패 / err : ", err)
		return false
	}
	
	mmsInsertQuery := `
		insert into `+oshotMmsTableName+`(MsgGroupID, Sender, Receiver, Subject, Msg, File_Path1, File_Path2, File_Path3, mst_id, cb_msg_Id)
		values (:MsgGroupID, :Sender, :Receiver, :Subject, :Msg, :File_Path1, :File_Path2, :File_Path3, :mst_id, :cb_msg_Id)
	`
	
	if len(nanoMmsDataList) > 0 {
		groupId := "resend"+time.Now().Format("20060102150405")
		for _, mmsData := range nanoMmsDataList {
			
			mapData := map[string]interface{}{
				"MsgGroupID": groupId,
				"Sender": mmsData.Callback,
				"Receiver": mmsData.Phone,
				"Subject": mmsData.Subject,
				"Msg": mmsData.Msg,
				"File_Path1": mmsData.FilePath1.String,
				"File_Path2": mmsData.FilePath2.String,
				"File_Path3": mmsData.FilePath3.String,
				"InsertDT": mmsData.ReqDate,
				"mst_id": mmsData.Etc10.String,
				"cb_msg_Id": mmsData.Etc9.String,
			}

			_, err := tx.NamedExec(mmsInsertQuery, mapData)
			if err != nil {
				errlog.Println("nanoToOshot / insert 실패 / "+nanoMmsTableName+"의 MSGKEY 값 : ", mmsData.MsgKey, " / err : ", err)
			} else {
				mmsUpdateId = append(mmsUpdateId, mmsData.MsgKey)
			}
		}

		if len(mmsUpdateId) > 0 {
			mmsUpdateQuery, args, err := sqlx.In(`update `+nanoMmsTableName+` set ETC7 = '1' where MSGKEY IN (?)`, mmsUpdateId)
			if err != nil {
				errlog.Println("nanoToOshot / OShotMMS 재발송 flag 변환 Sql 생성 실패 / err : ", err)
				return false
			}

			mmsUpdateQuery = tx.Rebind(mmsUpdateQuery)

			mmsUpdateResult, err := tx.Exec(mmsUpdateQuery, args...)

			if err != nil {
				errlog.Println("nanoToOshot / ", nanoMmsTableName, " 재발송 flag 변환 실패 / err : ", err)
				tx.Rollback()
				return false
			}

			mmsUpdateRowCnt, _ := mmsUpdateResult.RowsAffected()

			err = tx.Commit()
			if err != nil {
				errlog.Println("nanoToOshot / mms commit 실패 / err : ", err)
				return false
			} else {
				infolog.Println("nanoToOshot mms 처리 끝 / sd :", sd, " / 업데이트 건수 : ", mmsUpdateRowCnt)
			}
		}
	} else {
		infolog.Println("nanoToOshot mms 처리 끝 / sd :", sd, " / 업데이트 건수 : 0")
	}
	
	
	return true
}

func nanoLowToOshot(db *sqlx.DB, sd string) bool {
	infolog := config.Stdlog
	errlog := config.Stdlog

	var smsUpdateId []int
	var mmsUpdateId []string
	var nanoSmsDataList []NanoSmsTable
	var nanoMmsDataList []NanoMmsTable

	tx, err := db.Beginx()
	if err != nil {
		errlog.Println("nanoLowToOshot / sms 트랜잭션 실행 실패 / err : ", err)
		return false
	}

	infolog.Println("nanoLowToOshot sms 처리 시작 / sd : ", sd)
	err = tx.Select(&nanoSmsDataList, "select * from "+nanoLowSmsTableName+" where TR_ETC7 is null and TR_SENDDATE >= ?", sd)
	if err != nil {
		errlog.Println("nanoLowToOshot / ", nanoLowSmsTableName, " 조회 실패 / err : ", err)
		return false
	}
	
	smsInsertQuery := `
		insert into `+oshotSmsTableName+`(Sender, Receiver, Msg, InsertDT, mst_id, cb_msg_id)
		values (:Sender, :Receiver, :Msg, :InsertDT, :mst_id, :cb_msg_id)
	`

	if len(nanoSmsDataList) > 0 {
		for _, smsData := range nanoSmsDataList {
			mapData := map[string]interface{}{
				"Sender": smsData.Callback,
				"Receiver": smsData.Phone,
				"Msg": smsData.Msg,
				"InsertDT": smsData.SendDate,
				"mst_id": smsData.Etc10,
				"cb_msg_id": smsData.Etc9,
			}

			_, err := tx.NamedExec(smsInsertQuery, mapData)
			if err != nil {
				errlog.Println("nanoLowToOshot / insert 실패 / ", nanoLowSmsTableName, "의 MsgID 값 : ", smsData.Num, " / err : ", err)
			} else {
				smsUpdateId = append(smsUpdateId, smsData.Num)
			}
		}

		if len(smsUpdateId) > 0 {
			smsUpdateQuery, args, err := sqlx.In(`update `+nanoLowSmsTableName+` set TR_ETC7 = '1' where TR_NUM IN (?)`, smsUpdateId)
			if err != nil {
				errlog.Println("nanoLowToOshot / ", nanoLowSmsTableName, " 재발송 flag 변환 Sql 생성 실패 / err : ", err)
				return false
			}

			smsUpdateQuery = tx.Rebind(smsUpdateQuery)

			smsUpdateResult, err := tx.Exec(smsUpdateQuery, args...)

			if err != nil {
				errlog.Println("nanoLowToOshot / ", nanoLowSmsTableName, " 재발송 flag 변환 실패 / err : ", err)
				tx.Rollback()
				return false
			}

			smsUpdateRowCnt, _ := smsUpdateResult.RowsAffected()

			err = tx.Commit()
			if err != nil {
				errlog.Println("nanoLowToOshot / sms commit 실패 / err : ", err)
				return false
			} else {
				infolog.Println("nanoLowToOshot sms 처리 끝 / sd :", sd, " / 업데이트 건수 : ", smsUpdateRowCnt)
			}
		}
	} else {
		infolog.Println("nanoLowToOshot sms 처리 끝 / sd :", sd, " / 업데이트 건수 : 0")
	}

	tx, err = db.Beginx()
	if err != nil {
		errlog.Println("nanoLowToOshot / mms 트랜잭션 실행 실패 / err : ", err)
		return false
	}

	infolog.Println("nanoLowToOshot mms 처리 시작 / sd : ", sd)
	err = tx.Select(&nanoMmsDataList, "select * from "+nanoLowMmsTableName+" where ETC7 is null and REQDATE >= ?", sd)
	if err != nil {
		errlog.Println("nanoLowToOshot / ", nanoLowMmsTableName, " 조회 실패 / err : ", err)
		return false
	}
	
	mmsInsertQuery := `
		insert into `+oshotMmsTableName+`(MsgGroupID, Sender, Receiver, Subject, Msg, File_Path1, File_Path2, File_Path3, mst_id, cb_msg_Id)
		values (:MsgGroupID, :Sender, :Receiver, :Subject, :Msg, :File_Path1, :File_Path2, :File_Path3, :mst_id, :cb_msg_Id)
	`
	
	if len(nanoMmsDataList) > 0 {
		groupId := "resend"+time.Now().Format("20060102150405")
		for _, mmsData := range nanoMmsDataList {
			
			mapData := map[string]interface{}{
				"MsgGroupID": groupId,
				"Sender": mmsData.Callback,
				"Receiver": mmsData.Phone,
				"Subject": mmsData.Subject,
				"Msg": mmsData.Msg,
				"File_Path1": mmsData.FilePath1.String,
				"File_Path2": mmsData.FilePath2.String,
				"File_Path3": mmsData.FilePath3.String,
				"InsertDT": mmsData.ReqDate,
				"mst_id": mmsData.Etc10.String,
				"cb_msg_Id": mmsData.Etc9.String,
			}

			_, err := tx.NamedExec(mmsInsertQuery, mapData)
			if err != nil {
				errlog.Println("nanoLowToOshot / insert 실패 / "+nanoLowMmsTableName+"의 MSGKEY 값 : ", mmsData.MsgKey, " / err : ", err)
			} else {
				mmsUpdateId = append(mmsUpdateId, mmsData.MsgKey)
			}
		}

		if len(mmsUpdateId) > 0 {
			mmsUpdateQuery, args, err := sqlx.In(`update `+nanoLowMmsTableName+` set ETC7 = '1' where MSGKEY IN (?)`, mmsUpdateId)
			if err != nil {
				errlog.Println("nanoLowToOshot / OShotMMS 재발송 flag 변환 Sql 생성 실패 / err : ", err)
				return false
			}

			mmsUpdateQuery = tx.Rebind(mmsUpdateQuery)

			mmsUpdateResult, err := tx.Exec(mmsUpdateQuery, args...)

			if err != nil {
				errlog.Println("nanoLowToOshot / ", nanoLowMmsTableName, " 재발송 flag 변환 실패 / err : ", err)
				tx.Rollback()
				return false
			}

			mmsUpdateRowCnt, _ := mmsUpdateResult.RowsAffected()

			err = tx.Commit()
			if err != nil {
				errlog.Println("nanoLowToOshot / mms commit 실패 / err : ", err)
				return false
			} else {
				infolog.Println("nanoLowToOshot mms 처리 끝 / sd :", sd, " / 업데이트 건수 : ", mmsUpdateRowCnt)
			}
		}
	} else {
		infolog.Println("nanoLowToOshot mms 처리 끝 / sd :", sd, " / 업데이트 건수 : 0")
	}
	
	
	return true
}