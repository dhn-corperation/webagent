package config

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"time"

	ini "github.com/BurntSushi/toml"

	"github.com/go-resty/resty/v2"
	rotatelogs "github.com/lestrrat/go-file-rotatelogs"
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
var GoClient *http.Client = &http.Client{
	Timeout: time.Second * 30,
	Transport: &http.Transport{
		TLSHandshakeTimeout: 10 * time.Second,
	},
}

var RCSID = ""
var RCSPW = ""

func InitConfig() {
	realpath, _ := os.Executable()
	dir := filepath.Dir(realpath)
	logDir := filepath.Join(dir, "logs")
	err := createDir(logDir)
	if err != nil {
		log.Fatalf("Failed to ensure log directory: %s", err)
	}
	path := filepath.Join(logDir, "BizAgent")
	loc, _ := time.LoadLocation("Asia/Seoul")
	writer, err := rotatelogs.New(
		fmt.Sprintf("%s-%s.log", path, "%Y-%m-%d"),
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
	realpath, _ := os.Executable()
	dir := filepath.Dir(realpath)
	var configfile = filepath.Join(dir, "config.ini")
	_, err := os.Stat(configfile)
	if err != nil {
		err := createConfig(configfile)
		if err != nil {
			Stdlog.Println("Config file create fail")
		}
		Stdlog.Println("config.ini 생성완료 작성을 해주세요.")

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
	realpath, _ := os.Executable()
	dir := filepath.Dir(realpath)
	logDir := filepath.Join(dir, "logs")
	err := createDir(logDir)
	if err != nil {
		log.Fatalf("Failed to ensure log directory: %s", err)
	}
	path := filepath.Join(logDir, "GenieAgent")
	loc, _ := time.LoadLocation("Asia/Seoul")
	writer, err := rotatelogs.New(
		fmt.Sprintf("%s-%s.log", path, "%Y-%m-%d"),
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
	realpath, _ := os.Executable()
	dir := filepath.Dir(realpath)
	var configfile = filepath.Join(dir, "config.ini")
	_, err := os.Stat(configfile)
	if err != nil {

		err := createConfig(configfile)
		if err != nil {
			Stdlog.Println("Config file create fail")
		}
		Stdlog.Println("config.ini 생성완료 작성을 해주세요.")

		fmt.Println("Config file is missing : ", configfile)
	}

	var result Config
	_, err1 := ini.DecodeFile(configfile, &result)

	if err1 != nil {
		fmt.Println("Config file read error : ", err1)
	}

	return result
}

func InitConfigU() {
	realpath, _ := os.Executable()
	dir := filepath.Dir(realpath)
	logDir := filepath.Join(dir, "logs")
	err := createDir(logDir)
	if err != nil {
		log.Fatalf("Failed to ensure log directory: %s", err)
	}
	path := filepath.Join(logDir, "Agent")
	loc, _ := time.LoadLocation("Asia/Seoul")
	writer, err := rotatelogs.New(
		fmt.Sprintf("%s-%s.log", path, "%Y-%m-%d"),
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

	Conf = readConfigU()

}

func readConfigU() Config {
	realpath, _ := os.Executable()
	dir := filepath.Dir(realpath)
	var configfile = filepath.Join(dir, "config.ini")
	_, err := os.Stat(configfile)
	if err != nil {
		err := createConfig(configfile)
		if err != nil {
			Stdlog.Println("Config file create fail")
		}
		Stdlog.Println("config.ini 생성완료 작성을 해주세요.")

		fmt.Println("Config file is missing : ", configfile)
	}

	var result Config
	_, err1 := ini.DecodeFile(configfile, &result)

	if err1 != nil {
		fmt.Println("Config file read error : ", err1)
	}

	return result
}

func createDir(dirName string) error {
	err := os.MkdirAll(dirName, os.ModePerm)
	if err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}
	return nil
}

func createConfig(dirName string) error {
	fo, err := os.Create(dirName)
	if err != nil {
		return fmt.Errorf("Config file create fail: %w", err)
	}

	configData := []string{
		`#실행 환경 설정 파일`,
		``,
		`# DB`,
		`DB = "postgres"`,
		`HOST = "210.114.225.58"`,
		`PORT = "5432"`,
		`DBID = "postgres"`,
		`DBPW = "dhn7985!"`,
		`DBNAME = "igenie"`,
		``,
		`#그린샷 사용 유무`,
		`GRS = true`,
		``,
		`#IMC 사용 유무`,
		`IMC = false`,
		``,
		`#Naself 사용 유무`,
		`NAS = false`,
		``,
		`#스마트미 PHN 사용 유무`,
		`SMTPHN = false`,
		``,
		`#스마트미 사용 유무`,
		`SMT = true`,
		``,
		`#나노 폰문자 사용 유무`,
		`PMS = false`,
		``,
		`#나노 Fun문자 사용 유무`,
		`FUN = true`,
		``,
		`#나노 BKG문자 사용 유무`,
		`BKG = false`,
		``,
		`#2발신 실패시 환불 여부`,
		`REFUND = true`,
		``,
		`#스마트미 DB 직접 연결 사용 유무`,
		`SMTPHNDB = false`,
		``,
		`#RCS 사용 유무`,
		`RCS = true`,
		``,
		`#결과 테이블`,
		`RESULTTABLE="DHN_REQUEST_RESULT"`,
		``,
		`#2차 알림톡 발송을 위한 카카오 발송용 Table1`,
		`REQTABLE1="DHN_REQUEST"`,
		``,
		`#2차 알림톡 발송을 위한 카카오 발송용 Table2`,
		`REQTABLE2="DHN_REQUEST_2ND"`,
		``,
		`#폰문자 정액제 User KEY`,
		`WP1="FE227003022D124978D41FFA0C3F71CA"`,
		``,
		`#폰문자 건당 User KEY`,
		`WP2="FE227003022D124978D41FFA0C3F71CA"`,
		``,
		`#RCS 관련 ID / PW`,
		`RCSID="dhn7137985"`,
		`RCSPW="$2a$10$wss410VSvWDh7lABAGdjvu.iJnQ4jziEnzXlDB2./PVBcTrO5L/iK"`,
	}

	for _, line := range configData {
		fmt.Fprintln(fo, line)
	}

	return nil
}
