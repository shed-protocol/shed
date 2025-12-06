package comms_test

import (
	"net"
	"sync"
	"testing"

	"github.com/shed-protocol/shed/internal/comms"
	"github.com/shed-protocol/shed/internal/ot"
)

func TestReadWrite(t *testing.T) {
	connA, connB := net.Pipe()
	defer connA.Close()
	defer connB.Close()

	alice := make(chan comms.Message)
	bob := make(chan comms.Message)

	m1 := comms.OpMessage{ot.Insertion{Pos: 2, Text: "hello"}}
	m2 := comms.OpMessage{ot.Deletion{Pos: 2, Len: 3}}

	var wg sync.WaitGroup
	wg.Go(func() {
		go comms.ChanToConn(alice, connA)
		alice <- m1
		alice <- m2
		close(alice)
	})

	wg.Go(func() {
		go comms.ConnToChan(connB, bob)
		if got := <-bob; *got.(*comms.OpMessage) != m1 {
			t.Errorf("got %v, want %v", got, m1)
		}
		if got := <-bob; *got.(*comms.OpMessage) != m2 {
			t.Errorf("got %v, want %v", got, m2)
		}
	})
	wg.Wait()
}
