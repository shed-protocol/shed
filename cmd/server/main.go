package main

import (
	"fmt"
	"log"
	"net"
	"os"

	"github.com/shed-protocol/shed/internal/server"
)

func main() {
	LISTEN_PORT := os.Args[1]

	fmt.Println("Starting server...")

	l, err := net.Listen("tcp", fmt.Sprintf(":%s", LISTEN_PORT))
	if err != nil {
		log.Fatal(err)
	}
	defer l.Close()

	s := new(server.Server)
	s.Init()
	fmt.Println("Listening on port", LISTEN_PORT)
	s.Start()
}
