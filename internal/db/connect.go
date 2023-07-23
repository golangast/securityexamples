package connect

import (
	"database/sql"
	"embed"
	"fmt"
	"log"
	"time"

	_ "modernc.org/sqlite"
)

//go:embed database.db
var Dashboarddb embed.FS

const file string = "./internal/db/database.db"

func DbConnection() (*sql.DB, error) {
	db, err := sql.Open("sqlite", file)
	if err != nil {
		fmt.Println(err)
	}
	db.SetMaxOpenConns(20)
	db.SetMaxIdleConns(20)
	db.SetConnMaxLifetime(time.Minute * 5)
	//check if it pings
	err = db.Ping()
	if err != nil {
		fmt.Println(err)
	}

	log.Printf("Connected to DB %s successfully\n", file)
	return db, nil
} //end of connect
func ErrorCheck(err error) {
	if err != nil {
		fmt.Println(err.Error())
	}
}
