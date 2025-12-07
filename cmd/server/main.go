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

	l, err := net.Listen("tcp", fmt.Sprintf(":%s", LISTEN_PORT))
	if err != nil {
		log.Fatal(err)
	}
	defer l.Close()

	var s server.Server
	s.Init()
	go s.Start()

	for {
		conn, err := l.Accept()
		if err != nil {
			continue
		}

		s.Accept(conn)
	}
}
