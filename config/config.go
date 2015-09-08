package config

import (
	"flag"
	"log"
	"os"
)

type Config struct {
	DbConnectionString string
	HostAddr           string
	MemcacheHostAddr   string
}

func NewConfig() *Config {
	connectionstring := os.Getenv("FINDERSERP_MOBILITY_CONNECTIONSTRING")
	hostaddr := os.Getenv("FINDERSERP_MOBILITY_HOST_ADDRESS")
	memcachehostaddr := os.Getenv("FINDERSERP_MOBILITY_MEMCACHE_HOST_ADDRESS")

	if connectionstring == "" {
		log.Println("WARNING: environment variable FINDERSERP_MOBILITY_CONNECTIONSTRING is not set! Use the -DbConnectionString flag.")
	}
	if hostaddr == "" {
		log.Println("WARNING: environment variable FINDERSERP_MOBILITY_HOST_ADDRESS is not set! Use the -HostAddr flag.")
	}
	if memcachehostaddr == "" {
		log.Println("WARNING: environment variable FINDERSERP_MOBILITY_MEMCACHE_HOST_ADDRESS is not set! Use the -MemcacheHostAddr flag.")
	}

	return &Config{
		DbConnectionString: *flag.String("DbConnectionString", connectionstring, "Database connection string, example: \"server=ServerName;user id=username;password=p@ssw0rd;database=DatabaseName\""),
		HostAddr:           *flag.String("HostAddr", hostaddr, "Host and port to listen on."),
		MemcacheHostAddr:   *flag.String("MemcacheHostAddr", memcachehostaddr, "Host and port address of memcache server"),
	}
}
