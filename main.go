package main

import (
	c "./config"
	"fmt"
	"log"
	"net/http"
)

var (
	config = c.NewConfig()
)

func main() {
	InitRoutes()
	log.Printf("Started running on %s\n", config.HostAddr)
	if err := http.ListenAndServe(config.HostAddr, nil); err != nil {
		fmt.Printf("http.ListenAndServe() failed with %q\n", err)
	}
	fmt.Printf("Exited\n")
}
