package main

import (
	"database/sql"
	"flag"
	_ "github.com/alexbrainman/odbc"
	"os"
	"testing"
)

var (
	connStr = flag.String("connStr", "", "ODBC DSN to override")
)

func getConnStr(t *testing.T) string {
	if connStr != nil && *connStr != "" {
		return *connStr
	}
	connectionstring := os.Getenv("FINDERSERP_MOBILITY_CONNECTIONSTRING")
	if connectionstring == "" {
		t.Fatal("No connection string was available. Try running the test again specifying the connStr flag.")
	}
	return connectionstring
}

func TestConnectionAndBasicQuery(t *testing.T) {
	db, err := sql.Open("odbc", getConnStr(t))
	if err != nil {
		reportError(t, err)
		return
	}
	defer db.Close()
	err = db.Ping()
	if err != nil {
		reportError(t, err)
		return
	}
	rows, err := db.Query("select UserId, UserName from aspnet_Users;")
	if err != nil {
		reportError(t, err)
		return
	}
	defer rows.Close()
	for rows.Next() {
		var (
			userId   string
			userName string
		)
		rows.Scan(&userId, &userName)
		t.Logf("UserID: %s\nUserName: %s", userId, userName)
	}
	t.Logf("Connected successfully")

}

func reportError(t *testing.T, err error) {
	t.Errorf("ERROR: %s\n", err.Error())
}
