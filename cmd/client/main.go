package main

import (
	"net"
	"os"

	"github.com/shed-protocol/shed/internal/client"
)

var c client.Client

func main() {
	c.Attach(stdio{})
	server, err := net.Dial("tcp", os.Args[1])
	if err != nil {
		panic(err)
	}
	c.Connect(server)
	for {
	}
}

type stdio struct{}

func (s stdio) Read(p []byte) (n int, err error) {
	return os.Stdin.Read(p)
}

func (s stdio) Write(p []byte) (n int, err error) {
	return os.Stdout.Write(p)
}
