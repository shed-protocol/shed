package comms_test

import (
	"net"
	"sync"
	"testing"

	"github.com/shed-protocol/shed/internal/comms"
	"github.com/shed-protocol/shed/internal/ot"
)

func TestReadMessage(t *testing.T) {
	alice, bob := net.Pipe()
	defer alice.Close()
	defer bob.Close()

	msg1 := comms.InsertionMessage{Op: ot.Insertion{Pos: 2, Text: "hello"}}
	msg2 := comms.DeletionMessage{Op: ot.Deletion{Pos: 2, Len: 1}}

	var wg sync.WaitGroup
	wg.Go(func() {
		comms.WriteMessage(alice, msg1)
		comms.WriteMessage(alice, msg2)
	})
	wg.Go(func() {
		got, err := comms.ReadMessage(bob)
		if err != nil {
			t.Errorf("error reading message: %s", err)
		}
		switch got := got.(type) {
		case *comms.InsertionMessage:
			if *got != msg1 {
				t.Errorf("received message (%+v) != sent message (%+v)", *got, msg1)
			}
		default:
			t.Errorf("received message (%T) != sent message (%T)", got, msg1)
		}

		got, err = comms.ReadMessage(bob)
		if err != nil {
			t.Errorf("error reading message: %s", err)
		}
		switch got := got.(type) {
		case *comms.DeletionMessage:
			if *got != msg2 {
				t.Errorf("received message (%+v) != sent message (%+v)", *got, msg2)
			}
		default:
			t.Errorf("received message (%T) != sent message (%T)", got, msg2)
		}
	})
	wg.Wait()
}
