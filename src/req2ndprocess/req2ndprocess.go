package req2ndprocess

import (
	"webagent/src/config"
	"database/sql"
	"sync"
	"webagent/src/databasepool"
	s "strings"
	"time"
	"context"
)

var insColumn string = `
	MSGID,
	AD_FLAG,
	BUTTON1,
	BUTTON2,
	BUTTON3,
	BUTTON4,
	BUTTON5,
	IMAGE_LINK,
	IMAGE_URL,
	MESSAGE_TYPE,
	MSG,
	MSG_SMS,
	ONLY_SMS,
	P_COM,
	P_INVOICE,
	PHN,
	PROFILE,
	REG_DT,
	REMARK1,
	REMARK2,
	REMARK3,
	REMARK4,
	REMARK5,
	RESERVE_DT,
	S_CODE,
	SMS_KIND,
	SMS_LMS_TIT,
	SMS_SENDER,
	TMPL_ID,
	WIDE,
	SUPPLEMENT,
	PRICE,
	CURRENCY_TYPE,
	group_no,
	TITLE,
	HEADER,
	ATT_ITEMS,
	CAROUSEL,
	ATT_COUPON,
	KIND,
	ATTACHMENTS`

func Process(ctx context.Context) {
	config.Stdlog.Println("req2ndprocess - 프로세스 시작")
	var wg sync.WaitGroup
	for {
		select {
			case <- ctx.Done():
				config.Stdlog.Println("req2ndprocess - process가 15초 후에 종료")
			    time.Sleep(15 * time.Second)
			    config.Stdlog.Println("req2ndprocess - process 종료 완료")
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
			config.Stdlog.Println("req2ndprocess - panic 발생 원인 : ", r)
			if err, ok := r.(error); ok {
				if s.Contains(err.Error(), "connection refused") {
					for {
						config.Stdlog.Println("req2ndprocess - send ping to DB")
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
	
	var cnt sql.NullString
	
	var reqquery = "SELECT count(1) as cnt FROM " + conf.REQTABLE2 + " where remark3 = 'Y'"
	
	err := db.QueryRow(reqquery).Scan(&cnt)

	if err != nil {
		stdlog.Println("req2ndprocess - Result Table2 처리 중 오류 발생")
		stdlog.Println(err)
		// stdlog.Fatal(reqquery)
		panic(err)
	}
	 
	if cnt.String != "0" { 
		stdlog.Println("req2ndprocess - 2ND 테이블에서 Req 로 복사 시작 - ", cnt)
		
		resp, _ := db.Exec("update " + conf.REQTABLE2 + " as t2 set t2.remark3 = 'P' where t2.remark3 = 'Y' and not exists (select 1 from " + conf.REQTABLE1 + " as t1 where t1.msgid = replace(t2.msgid,'AT', '') )")
		
		upp, _ := resp.RowsAffected()

		insSql := `
			insert into ` + conf.REQTABLE1 + `(` + insColumn + `)
			select
				` + insColumn + `
			from
				` + conf.REQTABLE2 + ` where remark3 = 'P'`
		
		resins, err2 := db.Exec(insSql)
		
		if err2 != nil {
			stdlog.Println("req2ndprocess - 2ND 테이블에서 Req 로 복사 중 오류 발생 - Update P : ", upp)
		} else {
		
			insp, _ := resins.RowsAffected()
			
			resn, _ :=db.Exec("update " + conf.REQTABLE2 + " set remark3 = 'N' where remark3 = 'P'")
			upn, _ := resn.RowsAffected()
			stdlog.Println("req2ndprocess - 2ND 테이블에서 Req 로 복사 끝 - Update P / Insert / Update N : ", upp, insp, upn)
		}
	} else {
		time.Sleep(100 * time.Millisecond)
	}
}
