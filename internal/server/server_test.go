package server

import (
	"testing"

	"github.com/shed-protocol/shed/internal/ot"
)

func TestReceiveMessage(t *testing.T) {
	// Given an initialized server
	var s Server
	s.Init()
	go s.Run()

	if len(s.doc) != 0 || s.version != 0 {
		t.Fatalf("invalid initial state for server")
	}

	// When we receive a message
	op := ot.Insertion{Pos: 0, Text: "hello"}
	s.in <- op

	// Then the message should appear in the out channel
	msg := <-s.out
	switch msg := msg.(type) {
	case ot.Insertion:
		if msg != op {
			t.Errorf("sent message is different from received (%+v != %+v)", msg, op)
		}
	default:
		t.Errorf("unexpected operation type: %T", msg)
	}

	// Then the version number should be incremented
	if s.version != 1 {
		t.Errorf("document version number is %v (expected 1)", s.version)
	}

	// Then the document content should be updated
	if s.doc != op.Text {
		t.Errorf("bad document content: %q (expected %q)", s.doc, op.Text)
	}
}
