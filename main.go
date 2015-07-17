package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
)

var (
	dbDriver = flag.String("dbDriver", GetDefaultDriver(), "MSSQL ODBC driver name")
	dbHost   = flag.String("DbHost", ".", "MSSQL server host name")
	dbPort   = flag.String("DbPort", "1433", "MSSQL server port number")
	dbName   = flag.String("DbName", "FindersERPDB", "Database name")
	dbUser   = flag.String("DbUser", "sa", "Database login")
	dbPass   = flag.String("DbPass", "P@ssw0rd", "Login password")
	hostAddr = flag.String("HostAddr", ":5001", "Host and port to listen on.")
)

func main() {
	InitRoutes()
	ConnectDatabase()

	log.Printf("Started running on %s\n", *hostAddr)
	if err := http.ListenAndServe(*hostAddr, nil); err != nil {
		fmt.Printf("http.ListenAndServe() failed with %q\n", err)
	}
	fmt.Printf("Exited\n")
}
