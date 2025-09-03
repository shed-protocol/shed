package ot

import "slices"

// An Operation is an atomic edit that can be applied to a buffer.
type Operation interface {
	// Apply returns the result of applying an Operation to a buffer.
	Apply(buf []byte) []byte
}

// An Insertion adds text before a specified 0-indexed position.
type Insertion struct {
	Pos  uint
	Text []byte
}

// A Deletion removes text starting from a specified 0-indexed position.
type Deletion struct {
	Pos uint
	Len uint
}

func (op Insertion) Apply(buf []byte) []byte {
	return slices.Insert(buf, int(op.Pos), op.Text...)
}

func (op Deletion) Apply(buf []byte) []byte {
	return slices.Delete(buf, int(op.Pos), int(op.Pos+op.Len))
}
