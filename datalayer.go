package main

import (
	"database/sql"
	"errors"
	"fmt"
	_ "github.com/alexbrainman/odbc"
	"runtime"
	"strings"
)

func GetDefaultDriver() string {
	if runtime.GOOS == "windows" {
		return "sql server"
	} else {
		return "freetds"
	}
}

func isFreeTDS() bool {
	return *dbDriver == "freetds"
}

type connParams map[string]string

func newConnParams() connParams {
	params := connParams{
		"driver":   *dbDriver,
		"server":   *dbHost,
		"database": *dbName,
	}
	if isFreeTDS() {
		params["uid"] = *dbUser
		params["pwd"] = *dbPass
		params["port"] = *dbPort
	} else {
		if len(*dbUser) == 0 {
			params["trusted_connection"] = "yes"
		} else {
			params["uid"] = *dbUser
			params["pwd"] = *dbPass
		}
	}
	a := strings.SplitN(params["server"], ",", -1)
	if len(a) == 2 {
		params["server"] = a[0]
		params["port"] = a[1]
	}
	return params
}

func (params connParams) getConnAddress() (string, error) {
	port, ok := params["port"]
	if !ok {
		return "", errors.New("no port number provided")
	}
	host, ok := params["server"]
	if !ok {
		return "", errors.New("no host name provided")
	}
	return host + ":" + port, nil
}

func (params connParams) updateConnAddress(address string) error {
	a := strings.SplitN(address, ":", -1)
	if len(a) != 2 {
		fmt.Errorf("listen address must have to fields, but %d found", len(a))
	}
	params["server"] = a[0]
	params["port"] = a[1]
	return nil
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
