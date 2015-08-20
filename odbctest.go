package main

import (
	"database/sql"
	_ "github.com/alexbrainman/odbc"
	"log"
)

var (
	connStr = "server=j7dpgj7zuc.database.secure.windows.net;driver=FreeTDS;port=1433;uid=finderserp@j7dpgj7zuc;pwd=Pl@c10!@#;database=FindersERPDB"
)

func main() {
	db, err := sql.Open("odbc", connStr)
	if err != nil {
		reportError(err)
		return
	}
	defer db.Close()
	err = db.Ping()
	if err != nil {
		reportError(err)
		return
	}
	rows, err := db.Query("select UserId, UserName from aspnet_Users;")
	if err != nil {
		reportError(err)
		return
	}
	defer rows.Close()
	for rows.Next() {
		var (
			userId   string
			userName string
		)
		rows.Scan(&userId, &userName)
		log.Printf("UserID: %s\nUserName: %s", userId, userName)
	}
	log.Println("Connected successfully")

}

func reportError(err error) {
	log.Printf("ERROR: %s\n", err.Error())
}
