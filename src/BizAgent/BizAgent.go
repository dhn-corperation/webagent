package main

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	_ "github.com/go-sql-driver/mysql"

	"webagent/src/config"
	"webagent/src/databasepool"
	// "webagent/src/nanoit"
	"webagent/src/phonemsg"
	"webagent/src/tblreqprocess"
	// "webagent/src/webaproc"
	"webagent/src/webamms"
	"webagent/src/webasms"
	"webagent/src/webcmms"
	"webagent/src/webcsms"
	"webagent/src/req2ndprocess"
	"webagent/src/rcs"

	"github.com/takama/daemon"
)

const (
	name        = "BizAgent"
	description = "DHN 메세지 후속 처리 프로그램"
)

var dependencies = []string{"BizAgent.service"}

var resultTable string

type Service struct {
	daemon.Daemon
}

func (service *Service) Manage() (string, error) {

	usage := "Usage: BizAgent install | remove | start | stop | status"

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
	defer func() {
		if r := recover(); r != nil {
			for {
				err := databasepool.DB.Ping()
				if err == nil {
					break
				}
				time.Sleep(10 * time.Second)
			}
		}
	}()
	config.Stdlog.Println("BizAgent 시작")
	var conf = config.Conf

	if conf.RCS {
		config.Stdlog.Println("RCS 사용 - 시작")
		config.Stdlog.Println("RCS ID :",config.RCSID)
		config.Stdlog.Println("RCS PW :",config.RCSPW)
		
		go rcs.ResultProcess()
		
		go rcs.RetryProcess()
		
		go rcs.Process()
	}

	go tblreqprocess.Process()
	
	go req2ndprocess.Process()

	//결과 처리이기 때문에 항상 실행되어 있어야 함.

	//오샷 결과값 조회 및 문자 실패 환불 처리 고루틴
	go webcsms.Process()
	go webcmms.Process()
	//오샷 결과값 조회 및 문자 실패 환불 처리 고루틴

	//나노 결과값 조회 및 문자 실패 환불 처리 고루틴
	go webasms.Process()
	go webamms.Process()
	//나노 결과값 조회 및 문자 실패 환불 처리 고루틴

	// if conf.GRS {
	// 	config.Stdlog.Println("GRS 사용 - 시작")
	// 	go nanoit.Process()
	// 	go webaproc.Process()
	// }

	if conf.SMTPHN {
		config.Stdlog.Println("Phon 사용 - 시작")
		go phonemsg.Process()
	}
	for {
		time.Sleep(1 * time.Second)
	}
}
