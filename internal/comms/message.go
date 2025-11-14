package comms

import (
	"github.com/shed-protocol/shed/internal/ot"
)

type MessageKind uint8

const (
	OP_INSERTION MessageKind = iota + 1
	OP_DELETION
)

func MessageOfKind(k MessageKind) Message {
	switch k {
	case OP_INSERTION:
		return &InsertionMessage{}
	case OP_DELETION:
		return &DeletionMessage{}
	default:
		panic("unrecognized message kind")
	}
}

type Message interface {
	Kind() MessageKind
}

type InsertionMessage struct {
	Op ot.Insertion
}

func (InsertionMessage) Kind() MessageKind {
	return OP_INSERTION
}

type DeletionMessage struct {
	Op ot.Deletion
}

func (DeletionMessage) Kind() MessageKind {
	return OP_DELETION
}
