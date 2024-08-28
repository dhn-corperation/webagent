package req2ndprocess

import (
	"webagent/src/config"
	"database/sql"
	"sync"
	"webagent/src/databasepool"
	s "strings"
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
	defer wg.Done()
	var db = databasepool.DB
	var conf = config.Conf
	var stdlog = config.Stdlog
	
	var cnt sql.NullString
	
	var reqquery = "SELECT count(1) as cnt FROM " + conf.REQTABLE2 + " where remark3 = 'Y'"
	
	err := db.QueryRow(reqquery).Scan(&cnt)

	if err != nil {
		stdlog.Println("Result Table 처리 중 오류 발생")
		stdlog.Println(err)
		stdlog.Fatal(reqquery)
	}
	 
	if !s.EqualFold(cnt.String, "0") { 
		stdlog.Println("2ND 테이블에서 Req 로 복사 시작 - ", cnt)
		
		resp, _ := db.Exec("update " + conf.REQTABLE2 + " as t2 set t2.remark3 = 'P' where t2.remark3 = 'Y' and not exists (select 1 from " + conf.REQTABLE1 + " as t1 where t1.msgid = replace(t2.msgid,'AT', '') )")
		
		upp, _ := resp.RowsAffected()
		
		resins, err2 := db.Exec("insert into " + conf.REQTABLE1 + " select * from " + conf.REQTABLE2 + " as t2 where remark3 = 'P'")
		
		if err2 != nil {
			stdlog.Println("2ND 테이블에서 Req 로 복사 중 오류 발생 - Update P : ", upp)
		} else {
		
			insp, _ := resins.RowsAffected()
			
			resn, _ :=db.Exec("update " + conf.REQTABLE2 + " set remark3 = 'N' where remark3 = 'P'")
			upn, _ := resn.RowsAffected()
			stdlog.Println("2ND 테이블에서 Req 로 복사 끝 - Update P / Insert / Update N : ", upp, insp, upn)
		}
		
		
	}
}
