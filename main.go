package main

import (
	"fmt"
	"os"
)

func main() {
	server := NewServer(8080)
	err := server.Run()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
