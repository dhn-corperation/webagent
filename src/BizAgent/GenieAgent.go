package main

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"context"
	"time"

	_ "github.com/go-sql-driver/mysql"

	"webagent/src/config"
	"webagent/src/databasepool"
	"webagent/src/tblreqprocess"
	"webagent/src/phonemsg"
	"webagent/src/req2ndprocess"
	"webagent/src/webamms"
	"webagent/src/webasms"
	"webagent/src/webcmms"
	"webagent/src/webcsms"
	"webagent/src/rcs"
	"webagent/src/handler"
	
	"github.com/takama/daemon"
	"github.com/gin-gonic/gin"
	"github.com/jmoiron/sqlx"
)

const (
	name        = "GenieAgent"
	description = "지니 메세지 후속 처리 프로그램"
)

var dependencies = []string{"GenieAgent.service"}

var resultTable string

type Service struct {
	daemon.Daemon
}

func (service *Service) Manage() (string, error) {

	usage := "Usage: GenieAgent install | remove | start | stop | status"

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

	config.InitConfigG()

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
	contextCancel := map[string]interface{}{}

	config.Stdlog.Println("GenieAgent 시작")
	config.Stdlog.Println("---------------------------------------")
	var conf = config.Conf
	config.Stdlog.Println(conf)
	config.Stdlog.Println("---------------------------------------")

	go tblreqprocess.Process()

	go req2ndprocess.Process()

	if conf.RCS {
		config.Stdlog.Println("RCS 사용 - 시작")
		config.Stdlog.Println("RCS ID :",config.RCSID)
		config.Stdlog.Println("RCS PW :",config.RCSPW)
		
		go rcs.ResultProcess()
		
		go rcs.RetryProcess()
		
		go rcs.Process()
	}

	//결과 처리이기 때문에 항상 실행되어 있어야 함.

	//오샷 결과값 조회 및 문자 실패 환불 처리 고루틴
	go webcsms.Process()
	go webcmms.Process()
	//오샷 결과값 조회 및 문자 실패 환불 처리 고루틴

	//나노 결과값 조회 및 문자 실패 환불 처리 고루틴
	go webasms.Process()
	go webamms.Process()
	//나노 결과값 조회 및 문자 실패 환불 처리 고루틴

	//나노 저가망 결과값 조회 및 문자 실패 환불 처리 고루틴
	go webasms.Process_g()
	go webamms.Process_g()
	//나노 저가망 결과값 조회 및 문자 실패 환불 처리 고루틴

	// go nanoit.Process()
	// go webaproc.Process()

	if conf.SMTPHN {
		config.Stdlog.Println("폰문자 처리 시작")
		go phonemsg.Process()
	}

	r := gin.New()
	r.Use(gin.Recovery())

	r.GET("/", func(c *gin.Context) {
		c.String(200, "igenie api")
	})

	r.GET("/on", func(c *gin.Context) {
		target := c.Query("target")
		sd := c.Query("sd")

		if len(contextCancel) > 0 {
			c.JSON(400, gin.H{
				"code":    "error",
				"message": "이미 실행중입니다.",
			})
		}

		db, err := sqlx.Connect(config.Conf.DB, config.Conf.DBURL)
		if err != nil {
			config.Stdlog.Println("DB 접속 불가")
			c.JSON(500, gin.H{
				"code":    "error",
				"message": "DB 연결이 되지 않습니다.",
			})
		}

		parseSd, err := time.Parse("20060102150405", sd)
		if err != nil {
			c.JSON(400, gin.H{
				"code":    "error",
				"message": "잘못된 시간형식 입니다.",
				"sd":  sd,
			})
		}
		formattedSd := parseSd.Format("2006-01-02 15:04:05")

		ctx, cancel := context.WithCancel(context.Background())

		if target == "nano" || target == "oshot" {
			contextCancel[target] = cancel
			go handler.Resend(ctx, db, target, formattedSd)
		} else {
			c.JSON(400, gin.H{
				"code":    "error",
				"message": "잘못된 타겟 입니다.",
				"target": target,
			})
		}
		c.JSON(200, gin.H{
			"code":    "ok",
			"message":  "'시작' 신호가 정상적으로 되었습니다 / 타겟 : " + target,
			"sd": sd,
		})
	})

	r.GET("/off", func(c *gin.Context){
		target := c.Query("target")
		if target == "nano" || target == "oshot" {
			cancel := contextCancel[target].(context.CancelFunc)
			cancel()
			delete(contextCancel, target)
			c.JSON(200, gin.H{
				"code":    "ok",
				"message": "'종료' 신호가 정상적으로 전달되었습니다 / 타겟 : " + target,
			})
		}

	})

	r.Run(":3030")
}
