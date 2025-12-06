package ot

import "encoding/json"

// An Operation is an atomic edit that can be applied to a buffer.
type Operation interface {
	// Apply returns the result of applying an Operation to a buffer.
	Apply(buf string) string

	// Rebase returns an Operation that will have the same effect as the receiver
	// if applied immediately after the provided operation.
	Rebase(on Operation) Operation
}

// An Insertion adds text before a specified 0-indexed position.
type Insertion struct {
	Pos  uint   `json:"pos"`
	Text string `json:"text"`
}

// A Deletion removes text starting from a specified 0-indexed position.
type Deletion struct {
	Pos uint `json:"pos"`
	Len uint `json:"len"`
}

func (op Insertion) Apply(buf string) string {
	return buf[:op.Pos] + op.Text + buf[op.Pos:]
}

func (op Insertion) Rebase(on Operation) Operation {
	switch on := on.(type) {
	case Insertion:
		switch {
		case op.Pos < on.Pos:
			return op
		case op.Pos == on.Pos && op.Text < on.Text:
			return op
		default:
			return Insertion{Pos: op.Pos + uint(len(on.Text)), Text: op.Text}
		}
	case Deletion:
		switch {
		case op.Pos < on.Pos:
			return op
		case op.Pos >= on.Pos+on.Len:
			return Insertion{Pos: op.Pos - on.Len, Text: op.Text}
		default:
			return Insertion{}
		}
	default:
		panic("unhandled operation type")
	}
}

func (op Deletion) Apply(buf string) string {
	return buf[:op.Pos] + buf[op.Pos+op.Len:]
}

func (op Deletion) Rebase(on Operation) Operation {
	switch on := on.(type) {
	case Insertion:
		switch {
		case op.Pos+op.Len <= on.Pos:
			return op
		case op.Pos > on.Pos:
			return Deletion{Pos: op.Pos + uint(len(on.Text)), Len: op.Len}
		case op.Pos == on.Pos:
			return Deletion{Pos: op.Pos, Len: op.Len + uint(len(on.Text))}
		default:
			return Deletion{Pos: op.Pos, Len: op.Len + uint(len(on.Text))}
		}
	case Deletion:
		switch {
		case op.Pos+op.Len <= on.Pos:
			return op
		case op.Pos >= on.Pos+on.Len:
			return Deletion{Pos: op.Pos - on.Len, Len: op.Len}
		case op.Pos < on.Pos && on.Pos < op.Pos+op.Len && op.Pos+op.Len <= on.Pos+on.Len:
			return Deletion{Pos: op.Pos, Len: on.Pos - op.Pos}
		case on.Pos <= op.Pos && op.Pos < on.Pos+on.Len && on.Pos+on.Len < op.Pos+op.Len:
			return Deletion{Pos: on.Pos, Len: (op.Pos + op.Len) - (on.Pos + on.Len)}
		case op.Pos < on.Pos && on.Pos+on.Len < op.Pos+op.Len:
			return Deletion{Pos: op.Pos, Len: op.Len - on.Len}
		default:
			return Deletion{}
		}
	default:
		panic("unhandled operation type")
	}
}

func (op Insertion) MarshalJSON() ([]byte, error) {
	type insertion Insertion

	return json.Marshal(struct {
		insertion
		Type string `json:"type"`
	}{
		insertion(op), "insertion",
	})
}

func (op Deletion) MarshalJSON() ([]byte, error) {
	type deletion Deletion

	return json.Marshal(struct {
		deletion
		Type string `json:"type"`
	}{
		deletion(op), "deletion",
	})
}
