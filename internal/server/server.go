package server

import (
	"net"
	"sync"

	"github.com/shed-protocol/shed/internal/comms"
)

type Server struct {
	listener net.Listener
	cOuts    chan MessageWithId

	mu   sync.Mutex
	cIns map[int]chan<- comms.Message
}

type MessageWithId struct {
	msg comms.Message
	id  int
}

func (s *Server) Init() {
	s.cIns = make(map[int]chan<- comms.Message)
	s.cOuts = make(chan MessageWithId)
}

func (s *Server) Start() {
	for m := range s.cOuts {
		s.mu.Lock()
		for id, ch := range s.cIns {
			if id == m.id {
				ch <- comms.AcknowledgeChange{}
			} else {
				ch <- m.msg
			}
		}
		s.mu.Unlock()
	}
}

func (s *Server) Accept(c net.Conn) {
	in := make(chan comms.Message)
	out := make(chan comms.Message)
	go comms.ChanToConn(in, c)
	go comms.ConnToChan(c, out)

	s.mu.Lock()
	id := len(s.cIns)
	s.cIns[id] = in
	s.mu.Unlock()

	go func() {
		for {
			s.cOuts <- MessageWithId{<-out, id}
		}
	}()
}
