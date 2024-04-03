package config

import (
	"fmt"
	"log"
	"os"
	"time"

	ini "github.com/BurntSushi/toml"
	rotatelogs "github.com/lestrrat/go-file-rotatelogs"
	"github.com/go-resty/resty/v2"
)

type Config struct {
	DB           string
	DBURL        string
	GRS          bool
	IMC          bool
	NAS          bool
	SMTPHN       bool
	SMT          bool
	PMS          bool
	FUN          bool
	BKG          bool
	REFUND       bool
	SMTPHNDB     bool
	RCS          bool
	RESULTTABLE  string
	REQTABLE1    string
	REQTABLE2    string
	WP1          string
	WP2          string
	RCSID        string
	RCSPW        string
	PHNURL       string
	RCSSENDURL   string
	RCSRESULTURL string
}

var Conf Config
var Stdlog *log.Logger
var Client *resty.Client

var RCSID = ""
var RCSPW = ""

func InitConfig() {
	path := "/root/BizAgent/log/BizAgent"
	//path := "./log/BizAgent"
	loc, _ := time.LoadLocation("Asia/Seoul")
	writer, err := rotatelogs.New(
		fmt.Sprintf("%s-%s.log", path, "%Y-%m-%d"),
		//rotatelogs.WithMaxAge(time.Hour*24*7),
		//rotatelogs.WithRotationTime(time.Hour*24),
		rotatelogs.WithLocation(loc),
		rotatelogs.WithMaxAge(-1),
		rotatelogs.WithRotationCount(7),
	)

	if err != nil {
		log.Fatalf("Failed to Initialize Log File %s", err)
	}

	log.SetOutput(writer)
	stdlog := log.New(os.Stdout, "INFO -> ", log.Ldate|log.Ltime)
	stdlog.SetOutput(writer)
	Stdlog = stdlog

	Conf = readConfig()

	RCSID = Conf.RCSID
	RCSPW = Conf.RCSPW

	Client = resty.New()
}

func readConfig() Config {
	var configfile = "/root/BizAgent/config.ini"
	//var configfile = "./config.ini"
	_, err := os.Stat(configfile)
	if err != nil {
		fmt.Println("Config file is missing : ", configfile)
	}

	var result Config
	_, err1 := ini.DecodeFile(configfile, &result)

	if err1 != nil {
		fmt.Println("Config file read error : ", err1)
	}

	return result
}

func InitConfigG() {
	path := "/root/GenieAgent/log/GenieAgent"
	loc, _ := time.LoadLocation("Asia/Seoul")
	writer, err := rotatelogs.New(
		fmt.Sprintf("%s-%s.log", path, "%Y-%m-%d"),
		//rotatelogs.WithMaxAge(time.Hour*24*7),
		//rotatelogs.WithRotationTime(time.Hour*24),
		rotatelogs.WithLocation(loc),
		rotatelogs.WithMaxAge(-1),
		rotatelogs.WithRotationCount(7),
	)

	if err != nil {
		log.Fatalf("Failed to Initialize Log File %s", err)
	}

	log.SetOutput(writer)
	stdlog := log.New(os.Stdout, "INFO -> ", log.Ldate|log.Ltime)
	stdlog.SetOutput(writer)
	Stdlog = stdlog

	Conf = readConfigG()

	RCSID = Conf.RCSID
	RCSPW = Conf.RCSPW

	Client = resty.New()

}

func readConfigG() Config {
	var configfile = "/root/GenieAgent/config.ini"
	_, err := os.Stat(configfile)
	if err != nil {
		fmt.Println("Config file is missing : ", configfile)
	}

	var result Config
	_, err1 := ini.DecodeFile(configfile, &result)

	if err1 != nil {
		fmt.Println("Config file read error : ", err1)
	}

	return result
}

func InitConfigU(_path string) {
	path := _path + "/log/Agent"
	loc, _ := time.LoadLocation("Asia/Seoul")
	writer, err := rotatelogs.New(
		fmt.Sprintf("%s-%s.log", path, "%Y-%m-%d"),
		//rotatelogs.WithMaxAge(time.Hour*24*7),
		//rotatelogs.WithRotationTime(time.Hour*24),
		rotatelogs.WithLocation(loc),
		rotatelogs.WithMaxAge(-1),
		rotatelogs.WithRotationCount(7),
	)

	if err != nil {
		log.Fatalf("Failed to Initialize Log File %s", err)
	}

	log.SetOutput(writer)
	stdlog := log.New(os.Stdout, "INFO -> ", log.Ldate|log.Ltime)
	stdlog.SetOutput(writer)
	Stdlog = stdlog

	Conf = readConfigU(_path)

}

func readConfigU(_path string) Config {
	var configfile = _path + "/config.ini"
	_, err := os.Stat(configfile)
	if err != nil {
		fmt.Println("Config file is missing : ", configfile)
	}

	var result Config
	_, err1 := ini.DecodeFile(configfile, &result)

	if err1 != nil {
		fmt.Println("Config file read error : ", err1)
	}

	return result
}
