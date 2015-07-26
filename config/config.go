package config

import (
	"flag"
)

type Config struct {
	DbDriverName       string
	DbConnectionString string
	HostAddr           string
}

func NewConfig() *Config {
	return &Config{
		DbDriverName:       *flag.String("DbDriverName", "mssql", "SQL driver name"),
		DbConnectionString: *flag.String("DbConnectionString", "server=j7dpgj7zuc.database.secure.windows.net;user id=finderserp@j7dpgj7zuc;password=Pl@c10!@#;database=FindersERPDB", "Database connection string, example: \"server=ServerName;user id=username;password=p@ssw0rd;database=DatabaseName\""),
		HostAddr:           *flag.String("HostAddr", ":5001", "Host and port to listen on."),
	}
}
