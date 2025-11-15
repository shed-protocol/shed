package client

import (
	"net"
	"testing"

	"github.com/shed-protocol/shed/internal/comms"
	"github.com/shed-protocol/shed/internal/ot"
)

type MockEditor struct {
	client net.Conn
	local  chan comms.Message
	remote chan comms.Message
}

func (e *MockEditor) Init(c net.Conn) {
	e.client = c
	e.local = make(chan comms.Message)
	e.remote = make(chan comms.Message)
	go comms.ChanToConn(e.local, e.client)
	go comms.ConnToChan(e.client, e.remote)
}

type MockServer struct {
	client net.Conn
	cIn    chan<- comms.Message
	cOut   <-chan comms.Message
}

func (s *MockServer) Accept(c net.Conn) {
	cIn := make(chan comms.Message)
	cOut := make(chan comms.Message)
	s.client = c
	s.cIn = cIn
	s.cOut = cOut
	go comms.ChanToConn(cIn, c)
	go comms.ConnToChan(s.client, cOut)
}

func setupSingleClient() (*Client, *MockEditor, *MockServer) {
	var (
		c Client
		e MockEditor
		s MockServer
	)
	{
		a, b := net.Pipe()
		c.Attach(a)
		e.Init(b)
	}
	{
		a, b := net.Pipe()
		c.Connect(a)
		s.Accept(b)
	}
	return &c, &e, &s
}

func TestClientSendsLocalChangeToServer(t *testing.T) {
	// Given the client has no sent changes
	c, e, s := setupSingleClient()
	defer e.client.Close()
	defer c.editor.Close()
	defer c.server.Close()
	defer s.client.Close()

	// When the client receives a local change
	msg := comms.InsertionMessage{Op: ot.Insertion{Pos: 0, Text: "hello"}}
	e.local <- msg

	// Then the change should be sent to the server
	<-s.cOut
}

func TestClientSendsOneMessageAtATime(t *testing.T) {
	// When the client receives multiple local changes
	c, e, s := setupSingleClient()
	defer e.client.Close()
	defer c.editor.Close()
	defer c.server.Close()
	defer s.client.Close()

	msg1 := comms.InsertionMessage{Op: ot.Insertion{Pos: 0, Text: "hello"}}
	msg2 := comms.DeletionMessage{Op: ot.Deletion{Pos: 0, Len: 1}}
	e.local <- msg1
	e.local <- msg2

	// Then the first change should be sent to the server
	if got := <-s.cOut; *got.(*comms.InsertionMessage) != msg1 {
		t.Fatalf("server received %v, expected %v", got, msg1)
	}

	// Then the second change should not be sent to the server
	select {
	case <-s.cOut:
		t.Fatal("server received second message, expected only one")
	default:
	}

	// When the first change is acknowledged by the server
	s.cIn <- comms.AcknowledgeChange{}

	// Then the second change should be sent
	if got := <-s.cOut; *got.(*comms.DeletionMessage) != msg2 {
		t.Fatalf("server received %v, expected %v", got, msg2)
	}
}

func TestClientSendsRemoteChangeToEditor(t *testing.T) {
	// Given the client has no pending changes
	c, e, s := setupSingleClient()
	defer e.client.Close()
	defer c.editor.Close()
	defer c.server.Close()
	defer s.client.Close()

	// When the client receives a remote change
	msg := comms.InsertionMessage{Op: ot.Insertion{Pos: 0, Text: "hello"}}
	s.cIn <- msg

	// Then the change should be sent to the editor
	<-e.remote
}

func TestClientRebasesRemoteChangesForEditor(t *testing.T) {
	// Given the client has an unacknowledged change
	c, e, s := setupSingleClient()
	defer e.client.Close()
	defer c.editor.Close()
	defer c.server.Close()
	defer s.client.Close()

	localOp := ot.Insertion{Pos: 1, Text: "hello"}
	e.local <- comms.InsertionMessage{Op: localOp}

	// When the client receives a remote change
	remoteOp := ot.Deletion{Pos: 2, Len: 1}
	s.cIn <- comms.DeletionMessage{Op: remoteOp}

	// Then the remote change should be rebased and sent to the editor
	if got, ok := asOp(<-e.remote); ok {
		want := remoteOp.Rebase(localOp)
		if got != want {
			t.Errorf("editor received remote change %#v, expected %#v", got, want)
		}
	} else {
		t.Fatalf("editor received unexpected message type")
	}
}

func TestClientRebasesQueuedChangesForServer(t *testing.T) {
	// Given the client has pending changes
	c, e, s := setupSingleClient()
	defer e.client.Close()
	defer c.editor.Close()
	defer c.server.Close()
	defer s.client.Close()

	localOp1 := ot.Insertion{Pos: 1, Text: "hello"}
	msg1 := comms.InsertionMessage{Op: localOp1}
	e.local <- msg1
	for c.sent == nil {
	}

	localOp2 := ot.Insertion{Pos: 6, Text: "world"}
	msg2 := comms.InsertionMessage{Op: localOp2}
	e.local <- msg2
	for len(c.queue) != 1 {
	}

	// When the client receives a remote change
	remoteOp := ot.Deletion{Pos: 2, Len: 1}
	s.cIn <- comms.DeletionMessage{Op: remoteOp}

	// Then the remote change should be rebased and sent to the editor
	if got, ok := asOp(<-e.remote); ok {
		want := remoteOp.Rebase(localOp1).Rebase(localOp2)
		if got != want {
			t.Errorf("editor received remote change %#v, expected %#v", got, want)
		}
	} else {
		t.Fatalf("editor received unexpected message type")
	}

	// When the server acknowledges the pending change
	<-s.cOut
	s.cIn <- comms.AcknowledgeChange{}

	// Then the client should send rebased local changes
	if got, ok := asOp(<-s.cOut); ok {
		want := localOp2.Rebase(remoteOp)
		if got != want {
			t.Errorf("server received local change %#v, expected %#v", got, want)
		}
	} else {
		t.Fatalf("server received unexpected message type")
	}
}
