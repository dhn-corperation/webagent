package config

import (
	"os"
	"fmt"
	"log"
	"time"
	"path/filepath"

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
	KISACODE     string
}

var Conf Config
var Stdlog *log.Logger
var Client *resty.Client

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
		log.Fatalf("Failed to Initialize config File %s", err)
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