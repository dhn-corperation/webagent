package databasepool

import (
	_ "github.com/go-sql-driver/mysql"
	"database/sql"
	"log"
	"webagent/src/config"
)

var DB *sql.DB

func InitDatabase() {
	db, err := sql.Open(config.Conf.DB, config.Conf.DBURL)
	if err != nil {
		log.Fatal(err)
	}

	db.SetMaxIdleConns(15)
	db.SetMaxOpenConns(30)

	DB = db

}