package main

import (
	"database/sql"
	"flag"
	_ "github.com/alexbrainman/odbc"
	"runtime"
	"testing"
)

var (
	connStr = flag.String("dsn", "", "ODBC DSN to override")
)

func getConnStr() string {
	if connStr != nil && *connStr != "" {
		return *connStr
	}
	if runtime.GOOS == "windows" {
		return "DRIVER=SQL Server Native Client 11.0;Server=j7dpgj7zuc.database.secure.windows.net;uid=finderserp@j7dpgj7zuc;pwd=Pl@c10!@#;database=FindersERPDB;Encrypt=yes;TrustServerCertificate=no;"
	}
	return "server=j7dpgj7zuc.database.secure.windows.net;driver=FreeTDS;port=1433;uid=finderserp@j7dpgj7zuc;pwd=Pl@c10!@#;database=FindersERPDB"
}

func TestConnectionAndBasicQuery(t *testing.T) {
	db, err := sql.Open("odbc", getConnStr())
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
