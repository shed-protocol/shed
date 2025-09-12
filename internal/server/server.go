package server

import "github.com/shed-protocol/shed/internal/ot"

type Server struct {
	doc     string
	in, out chan ot.Operation
	version int
}

func (s *Server) Init() {
	s.in = make(chan ot.Operation)
	s.out = make(chan ot.Operation)
}

func (s *Server) Run() {
	for {
		op := <-s.in
		s.doc = op.Apply(s.doc)
		s.version++
		s.out <- op
	}
}
