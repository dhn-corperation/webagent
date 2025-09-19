package main

import (
	"os"
	"fmt"
	"time"
	"syscall"
	"context"
	"reflect"
	"os/signal"

	"webagent/src/rcs"
	"webagent/src/config"
	"webagent/src/webamms"
	"webagent/src/webasms"
	"webagent/src/webbmms"
	"webagent/src/webbsms"
	"webagent/src/webcmms"
	"webagent/src/webcsms"
	"webagent/src/webdmsg"
	// "webagent/src/webrcs"
	"webagent/src/handler"
	"webagent/src/databasepool"
	"webagent/src/tblreqprocess"
	"webagent/src/req2ndprocess"
	
	"github.com/takama/daemon"
	"github.com/gin-gonic/gin"
	"github.com/jmoiron/sqlx"
	_ "github.com/go-sql-driver/mysql"
)

const (
	name        = "BizAgent_m"
	description = "마트톡 메세지 후속 처리 프로그램"
	port  		= ":3010"

	// name        = "BizAgent_g"
	// description = "올지니 메세지 후속 처리 프로그램"
	// port  		= ":3020"

	// name        = "BizAgent_o"
	// description = "오투오 메세지 후속 처리 프로그램"
	// port  		= ":3030"

	// name        = "BizAgent_p"
	// description = "스피드톡 메세지 후속 처리 프로그램"
	// port  		= ":3040"

	// name        = "BizAgent_s"
	// description = "싸다고 메세지 후속 처리 프로그램"
	// port  		= ":3050"
)

var dependencies = []string{name+".service"}

var resultTable string

var rcc context.CancelFunc
var cc context.CancelFunc

type Service struct {
	daemon.Daemon
}

func (service *Service) Manage() (string, error) {

	usage := "Usage: "+name+" install | remove | start | stop | status"

	if len(os.Args) > 1 {
		command := os.Args[1]
		switch command {
		case "install":
			return service.Install()
		case "remove":
			return service.Remove()
		case "start":
			return service.Start()
		case "stop":
			return service.Stop()
		case "status":
			return service.Status()
		default:
			return usage, nil
		}
	}

	resultProc()
	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt, os.Kill, syscall.SIGTERM)

	for {
		select {
		case killSignal := <-interrupt:
			config.Stdlog.Println("Got signal:", killSignal)
			config.Stdlog.Println("Stoping DB Conntion : ", databasepool.DB.Stats())
			defer databasepool.DB.Close()
			if killSignal == os.Interrupt {
				return "Daemon was interrupted by system signal", nil
			}
			return "Daemon was killed", nil
		}
	}
}

func main() {
	config.InitConfig()	
	databasepool.InitDatabase()

	var rLimit syscall.Rlimit

	rLimit.Max = 50000
	rLimit.Cur = 50000

	err := syscall.Setrlimit(syscall.RLIMIT_NOFILE, &rLimit)

	if err != nil {
		config.Stdlog.Println("Error Setting Rlimit ", err)
	}

	err = syscall.Getrlimit(syscall.RLIMIT_NOFILE, &rLimit)

	if err != nil {
		config.Stdlog.Println("Error Getting Rlimit ", err)
	}

	config.Stdlog.Println("Rlimit Final", rLimit)

	srv, err := daemon.New(name, description, daemon.SystemDaemon, dependencies...)
	if err != nil {
		config.Stdlog.Println("Error: ", err)
		os.Exit(1)
	}
	service := &Service{srv}
	status, err := service.Manage()
	if err != nil {
		config.Stdlog.Println(status, "\nError: ", err)
		os.Exit(1)
	}
	fmt.Println(status)
}

func resultProc() {

	var conf = config.Conf

	val := reflect.ValueOf(conf)
	typ := reflect.TypeOf(conf)

	config.Stdlog.Println(name, " 시작")
	config.Stdlog.Println("------------------------------------------------------------------------------")

	for i := 0; i < val.NumField(); i++ {
		field := typ.Field(i)
		value := val.Field(i)
		config.Stdlog.Println(fmt.Sprintf("%s: %v", field.Name, value.Interface()))
	}

	config.Stdlog.Println("------------------------------------------------------------------------------")

	ctx, cancel := context.WithCancel(context.Background())
	cc = cancel

	go tblreqprocess.Process(ctx)
	go req2ndprocess.Process(ctx)

	if conf.RCS {
		// (구) Rcs
		go rcs.ResultProcess(ctx)
		go rcs.RetryProcess(ctx)
		go rcs.Process(ctx)

		// (신) Rcs
		// webrcs.SetAuthRequest()
		// go webrcs.RcsProc(ctx)
		// go webrcs.ResultProcess(ctx)
	}

	//결과 처리이기 때문에 항상 실행되어 있어야 함.

	//나노 결과값 조회 및 문자 실패 환불 처리 고루틴
	go webasms.Process(ctx)
	go webamms.Process(ctx)
	//나노 결과값 조회 및 문자 실패 환불 처리 고루틴

	//나노 저가망 결과값 조회 및 문자 실패 환불 처리 고루틴
	go webasms.Process_g(ctx)
	go webamms.Process_g(ctx)
	//나노 저가망 결과값 조회 및 문자 실패 환불 처리 고루틴

	//LGU 결과값 조회 및 문자 실패 환불 처리 고루틴
	go webbsms.Process(ctx)
	go webbmms.Process(ctx)
	//LGU 결과값 조회 및 문자 실패 환불 처리 고루틴

	//오샷 결과값 조회 및 문자 실패 환불 처리 고루틴
	go webcsms.Process(ctx)
	go webcmms.Process(ctx)
	//오샷 결과값 조회 및 문자 실패 환불 처리 고루틴

	//SMTNT 결과값 조회 및 문자 실패 환불 처리 고루틴
	go webdmsg.Process(ctx)
	//SMTNT 결과값 조회 및 문자 실패 환불 처리 고루틴

	r := gin.New()
	r.Use(gin.Recovery())

	r.GET("/", func(c *gin.Context) {
	c.String(200, `----------------------------------------------------------------
`+name+` API 리스트
/resendrun?target=XXX&sd=XXXXXXXXXXXXXX     description : 임시 재발송
/resendstop?uid=XXXX                        description : 임시 재발송 종료
/resendlist                                 description : 임시 재발송 리스트
/allstop?uid=XXXXXX                         description : 발송 전체 종료
----------------------------------------------------------------`+"\n")
	})

	r.GET("/resendrun", func(c *gin.Context) {
		target := c.Query("target")
		sd := c.Query("sd")

		if rcc != nil {
			c.JSON(400, gin.H{
				"code":    "error",
				"message": "이미 실행중입니다",
			})
			return
		}

		db, err := sqlx.Connect(config.Conf.DB, config.Conf.DBURL)
		if err != nil {
			config.Stdlog.Println("DB 접속 불가")
			c.JSON(500, gin.H{
				"code":    "error",
				"message": "DB 연결이 되지 않습니다",
			})
			return
		}

		parseSd, err := time.Parse("20060102150405", sd)
		if err != nil {
			c.JSON(400, gin.H{
				"code":    "error",
				"message": "잘못된 시간형식 입니다",
				"sd":  sd,
			})
			return
		}
		formattedSd := parseSd.Format("2006-01-02 15:04:05")

		if target == "nano" || target == "oshot" || target == "nano_g" {
			ctx, cancel := context.WithCancel(context.Background())
			rcc = cancel
			go handler.Resend(ctx, db, target, formattedSd)
		} else {
			c.JSON(400, gin.H{
				"code":    "error",
				"message": "잘못된 타겟 입니다",
				"target": target,
			})
			return
		}
		c.JSON(200, gin.H{
			"code":    "ok",
			"message":  "'시작' 신호가 정상적으로 되었습니다 / 타겟 : " + target,
			"sd": sd,
		})
	})

	r.GET("/resendstop", func(c *gin.Context){
		uid := c.Query("uid")
		if uid == "dhn" {
			if rcc == nil {
				c.JSON(400, gin.H{
					"code":    "error",
					"message":  "가동중인 재발송이 없습니다",
				})
				return
			}
			rcc()
			rcc = nil
			config.Stdlog.Println("'종료' 신호가 정상적으로 전달되었습니다")
			c.JSON(200, gin.H{
				"code":    "ok",
				"message": "'종료' 신호가 정상적으로 전달되었습니다",
			})
			return
		}
		c.JSON(400, gin.H{
			"code":    "error",
			"message": "종료할 수 없습니다",
		})
	})

	r.GET("/resendlist", func(c *gin.Context){
		if rcc != nil {
			c.String(200, "1")
		} else {
			c.String(200, "0")
		}
	})

	r.GET("/allstop", func(c *gin.Context){
		uid := c.Query("uid")
		if uid == "dhn" {
			config.Stdlog.Println("전체 종료 시작")
			cc()
			cc = nil
			c.String(200, "전체 종료 성공")
		} else {
			c.String(400, "전체 종료 실패")
		}
	})

	r.Run(port)
}
