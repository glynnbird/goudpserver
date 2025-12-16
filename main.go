package main

import (
	"fmt"
	"log"
	"os"
	"strconv"
)

const defaultPort string = "8081"

func main() {

	// look for override of default port using environment variable
	portStr := os.Getenv("PORT")
	if portStr == "" {
		portStr = defaultPort
	}

	// convert to string to integer
	log.Println("Listening on port", portStr)
	port, err := strconv.Atoi(portStr)
	if err != nil {
		fmt.Println("Cannot parse PORT environment variable")
		os.Exit(1)
	}

	// run the server
	server := NewServer(port)
	server.Run()

}
