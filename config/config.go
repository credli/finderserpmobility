package config

import (
	"flag"
	"os"
)

type Config struct {
	DbConnectionString string
	HostAddr           string
}

func NewConfig() *Config {
	connectionstring := os.Getenv("FINDERSERP_MOBILITY_CONNECTIONSTRING")
	hostaddr := os.Getenv("FINDERSERP_MOBILITY_HOST_ADDRESS")

	return &Config{
		DbConnectionString: *flag.String("DbConnectionString", connectionstring, "Database connection string, example: \"server=ServerName;user id=username;password=p@ssw0rd;database=DatabaseName\""),
		HostAddr:           *flag.String("HostAddr", hostaddr, "Host and port to listen on."),
	}
}
