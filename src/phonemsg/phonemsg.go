package phonemsg

import (
	"bytes"
	"webagent/src/config"
	"database/sql"
	"encoding/json"
	"io/ioutil"
	"regexp"
	//"log"
	"webagent/src/databasepool"
	"net/http"
	s "strings"
	"sync"
	"time"
	"fmt"

	_ "github.com/go-sql-driver/mysql"
)

type ListJson struct {
	Msgid    string `json:"msgid,omitempty"`
	Sender   string `json:"sender,omitempty"`
	Receiver string `json:"receiver,omitempty"`
	Title    string `json:"title,omitempty"`
	Msg      string `json:"msg,omitempty"`
}

type MsgJson struct {
	Cnt  int `json:"cnt,omitempty"`
	List []ListJson `json:"list,omitempty"`
}

var url = "http://66.232.143.52/apis/pms/send"
// var url = "http://api.martok.co.kr/apis/pms/send"

func Process() {
	var wg sync.WaitGroup
	for {
		wg.Add(1)
		go phnProcess(&wg)
		wg.Wait()
	}

}

func phnProcess(wg *sync.WaitGroup) {
	defer wg.Done()
	defer func() {
		if r := recover(); r != nil {
			config.Stdlog.Println("phonemsg panic 발생 원인 : ", r)
			if err, ok := r.(string); ok {
				if s.Contains(err, "connection refused") {
					for {
						config.Stdlog.Println("phonemsg send ping to DB")
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
	var stdlog = config.Stdlog
	var errlog = config.Stdlog

	var count int
	var url = "http://66.232.143.52/apis/pms/send"
	
	cnterr := db.QueryRow("select count(1) as cnt from SMT_SEND r where  group_no is null and SEND_STATUS = 'READY' limit 1").Scan(&count)

	if len(config.Conf.PHNURL) > 5 {
		url = config.Conf.PHNURL
	}

	if cnterr != nil {
		config.Stdlog.Println("Request Table - select 오류 : " + cnterr.Error())
		panic(cnterr)
	} else {

		if count > 0 {

			var startNow = time.Now()
			var group_no = fmt.Sprintf("%02d%02d%02d%09d", startNow.Hour(), startNow.Minute(), startNow.Second(), startNow.Nanosecond())

			updateRows, err := db.Exec("update SMT_SEND r set group_no = '" + group_no + "' where  group_no is null and SEND_STATUS = 'READY' limit 1000")
			if err != nil {
				config.Stdlog.Println("폰 문자 Request Table - Group No Update 오류" + err.Error())
			} else {
				rowcnt, _ := updateRows.RowsAffected()

				if rowcnt > 0 {

					var smtmsg_str = "select (message_id div 1000) as part, " +
						"       max(message_id) as msg_id," +
						"       group_concat( receivers) as PHN," +
						"       group_concat(message_id) as msg_ids," +
						"       max(request_id) as REMARK4, " +
						"       max(message) as msg, " +
						"       max(sender) as sender, " +
						"       max(subject) as subject, " +
						"       max(user_acct_key) as user_acct_key " +
						"  from SMT_SEND " +
						" where group_no = '" + group_no + "'" +
						" group by request_id, (message_id div 1000)"

					Rows, err := db.Query(smtmsg_str)
					if err != nil {
						errlog.Println("폰문자 조회 중 오류 발생")
						// errlog.Fatal(err)
					}

					defer Rows.Close()

					var stotalcnt = 0

					for Rows.Next() {
						var part sql.NullString
						var msg_id sql.NullString
						var phn sql.NullString
						var msg_ids sql.NullString
						var remark4 sql.NullString
						var msg sql.NullString
						var sender sql.NullString
						var subject sql.NullString
						var user_acct_key sql.NullString

						Rows.Scan(&part, &msg_id, &phn, &msg_ids, &remark4, &msg, &sender, &subject, &user_acct_key)

						var ins_cnt = len(s.Split(phn.String, ","))
						stotalcnt = stotalcnt + ins_cnt

						var msgj MsgJson

						msgj.Cnt = ins_cnt
						msgj.List = make([]ListJson, 1)
						msgj.List[0].Msgid = msg_id.String
						msgj.List[0].Sender = sender.String
						msgj.List[0].Receiver = phn.String
						msgj.List[0].Msg = msg.String
						re := regexp.MustCompile(`[\n\r]+`)
						msgj.List[0].Title = re.ReplaceAllString(subject.String, "")

						//stdlog.Println(msgj)
						param, _ := json.Marshal(msgj)
						stdlog.Println(string(param))

						buff := bytes.NewBuffer(param)

						req, err := http.NewRequest("POST", url, buff)

						req.Header.Add("Content-Type", "application/json")
						req.Header.Add("Authorization", user_acct_key.String)

						client := &http.Client{}
						resp, err := client.Do(req)

						//stdlog.Println(resp)

						if err != nil {
							//errlog.Fatal(err)
							errlog.Println("폰문자 발송 중 요류 발생")
						}
						defer resp.Body.Close()

						respBody, err := ioutil.ReadAll(resp.Body)
						if err == nil {
							str := string(respBody)
							stdlog.Println(str)
						} else {
							errlog.Println("폰문자 발송 후처리 중 요류 발생")
							//errlog.Fatal(err)
						}

						tx, err := db.Begin()
						if err != nil {
							errlog.Println("폰문자 트랜잭션 시작 중 오류 발생")
							//errlog.Fatal(err)
						} else {
							tx.Exec("update SMT_SEND set SEND_STATUS = 'SUCCESS' where message_id in (" + msg_ids.String + ") ")
							tx.Commit()
						}

					}
				}
			}
		}
	}
}
