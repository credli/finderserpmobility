package config

import (
	"flag"
	"os"
)

type Config struct {
	DbDriverName       string
	DbConnectionString string
	HostAddr           string
}

func NewConfig() *Config {
	drivername := os.Getenv("FINDERSERP_MOBILITY_DRIVER_NAME")
	connectionstring := os.Getenv("FINDERSERP_MOBILITY_CONNECTIONSTRING")
	hostaddr := os.Getenv("FINDERSERP_MOBILITY_HOST_ADDRESS")

	return &Config{
		DbDriverName:       *flag.String("DbDriverName", drivername, "SQL driver name"),
		DbConnectionString: *flag.String("DbConnectionString", connectionstring, "Database connection string, example: \"server=ServerName;user id=username;password=p@ssw0rd;database=DatabaseName\""),
		HostAddr:           *flag.String("HostAddr", hostaddr, "Host and port to listen on."),
	}
}
