package server

import (
	"net"
	"testing"

	"github.com/shed-protocol/shed/internal/comms"
	"github.com/shed-protocol/shed/internal/ot"
)

type MockClient struct {
	sIn  chan<- comms.Message
	sOut <-chan comms.Message
}

func (c *MockClient) Connect(conn net.Conn) {
	sIn := make(chan comms.Message)
	sOut := make(chan comms.Message)
	c.sIn = sIn
	c.sOut = sOut
	go comms.ChanToConn(sIn, conn)
	go comms.ConnToChan(conn, sOut)
}

func setupTwoClients() (alice *MockClient, bob *MockClient, s *Server, teardown func()) {
	alice = new(MockClient)
	bob = new(MockClient)
	s = new(Server)
	s.Init()

	a1, b1 := net.Pipe()
	alice.Connect(a1)
	s.Accept(b1)

	a2, b2 := net.Pipe()
	bob.Connect(a2)
	s.Accept(b2)

	go s.Start()

	return alice, bob, s, func() {
		a1.Close()
		a2.Close()
		b1.Close()
		b2.Close()
	}
}

func TestServerAcknowledgesChanges(t *testing.T) {
	// Given a client is connected to the server
	alice, _, _, teardown := setupTwoClients()
	defer teardown()

	// When the client sends a change
	go func() {
		alice.sIn <- comms.OpMessage{Op: ot.Insertion{Text: "hello", Pos: 2}}
	}()

	// Then the server should acknowledge the change
	if msg := <-alice.sOut; msg.Kind() != comms.ACK_CHANGE {
		t.Fail()
	}
}

func TestServerBroadcastsChanges(t *testing.T) {
	// Given two clients are connected to the server
	alice, bob, _, teardown := setupTwoClients()
	defer teardown()

	// When one client sends a change
	want := comms.OpMessage{Op: ot.Insertion{Text: "hello", Pos: 2}}
	go func() {
		alice.sIn <- want
	}()

	// Then the server should relay the change to the other client
	if got := <-bob.sOut; *got.(*comms.OpMessage) != want {
		t.Errorf("Bob got %v, but Alice sent %v", got, want)
	}
}
