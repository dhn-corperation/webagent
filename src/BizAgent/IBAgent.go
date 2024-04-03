package main

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"

	_ "github.com/go-sql-driver/mysql"

	"config"
	"databasepool"
	"nanoit"
	"phonemsg"
	"tblreqprocess"
	"webaproc"
	"webcmms"
	"webcsms"
	"req2ndprocess"

	"github.com/takama/daemon"
)

const (
	name        = "IBAgent"
	description = "지니 메세지 후속 처리 프로그램"
)

var dependencies = []string{"IBAgent.service"}

var resultTable string

type Service struct {
	daemon.Daemon
}

func (service *Service) Manage() (string, error) {

	usage := "Usage: IBAgent install | remove | start | stop | status"

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

	config.InitConfigU("/root/IBAgent")

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
	config.Stdlog.Println("IBAgent 시작")
	var conf = config.Conf

	go tblreqprocess.Process()

	go req2ndprocess.Process()

	if conf.SMT {
		go webcsms.Process()

		go webcmms.Process()
	}

	if conf.GRS {
		go nanoit.Process()
		go webaproc.Process()
	}

	if conf.SMTPHN {
		go phonemsg.Process()
	}

}
