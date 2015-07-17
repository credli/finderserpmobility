package main

import (
	"errors"
	"flag"
	"fmt"
	"log"
	"net/http"

	"github.com/gorilla/mux"
)

var (
	dbHost   = flag.String("DbHost", ".", "MSSQL server host name")
	dbPort   = flag.Int("DbPort", 1433, "MSSQL server port number")
	dbName   = flag.String("DbName", "FindersERPDB", "Database name")
	dbUser   = flag.String("DbUser", "sa", "Database login")
	dbPass   = flag.String("DbPass", "P@ssw0rd", "Login password")
	hostAddr = flag.String("HostAddr", "localhost", "Host to listen on (0.0.0.0 for listening over the network).")
	hostPort = flag.Int("HostPort", 5001, "Port to listen on (Defaults to 5001).")
	logger   = log.Logger
)

func main() {
	InitRoutes()

	logger.Printf("Started running on %s\n", *hostAddr)
	if err := http.ListenAndServe(*hostAddr, nil); err != nil {
		fmt.Printf("http.ListenAndServe() failed with %q\n", err)
	}
	fmt.Printf("Exited\n")
}
