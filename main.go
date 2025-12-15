package main

import (
	"fmt"
	"os"
)

const port int = 8080

func main() {
	server := NewServer(port)
	err := server.Run()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
