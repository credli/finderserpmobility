package main

import (
	"database/sql"
	"flag"
	"fmt"
	_ "github.com/alexbrainman/odbc"
	"runtime"
)

var (
	dbDriver = flag.String("dbDriver", defaultDriver(), "MSSQL ODBC driver name")
	dbHost   = flag.String("dbHost", "j7dpgj7zuc.database.secure.windows.net", "Database Host DNS or IP address")
	dbPort   = flag.Int("dbPort", 1433, "Database Port, defaults to 1433")
	dbName   = flag.String("dbName", "FindersERPDB", "Database name")
	dbUser   = flag.String("dbUser", "finderserp@j7dpgj7zuc", "Database user")
	dbPass   = flag.String("dbPass", "Pl@c10!@#", "Database password")
)

func defautDriver() {
	if runtime.GOOS == "windows" {
		return "sql server"
	} else {
		return "freetds"
	}
}

func isFreeTDS() {
	return *dbDriver == "freetds"
}

type connParams map[string]string

func newConnParams() connParams {
	params := connParams{
		"driver":   dbDriver,
		"server":   dbHost,
		"database": dbName,
	}
	if isFreeTDS() {
		params["uid"] = *dbUser
	}

}

func ConnectDatabase() {
	db, err := sql.Open("odbc", "driver=freetds;server=j7dpgj7zuc.database.secure.windows.net,1433;uid=finderserp@j7dpgj7zuc;pwd=Pl@c10!@#;")
	if err != nil {
		fmt.Printf("Error connecting to mssql, reason: %s\n", err)
	}
	err = db.Ping()
	if err != nil {
		fmt.Printf("Database server ping failed, reason: %s\n", err)
	}
	defer db.Close()

	//rows, err := db.Query("SELECT * FROM aspnet_Users", ...)
}
