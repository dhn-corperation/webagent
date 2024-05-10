package databasepool

import (
	"database/sql"
	"fmt"
	"log"
	"webagent/src/config"

	//_ "github.com/go-sql-driver/mysql"
	_ "github.com/lib/pq"
)

var DB *sql.DB

func InitDatabase() {
	psqlInfo := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable", config.Conf.HOST, config.Conf.PORT, config.Conf.DBID, config.Conf.DBPW, config.Conf.DBNAME)
	db, err := sql.Open(config.Conf.DB, psqlInfo)
	if err != nil {
		log.Fatal(err)
	}

	db.SetMaxIdleConns(15)
	db.SetMaxOpenConns(30)

	DB = db

}
