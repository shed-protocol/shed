package comms

import (
	"encoding/json"

	"github.com/shed-protocol/shed/internal/ot"
)

type MessageKind uint8

const (
	BUFFER_OP MessageKind = iota + 1
	ACK_CHANGE
)

func MessageOfKind(k MessageKind) Message {
	switch k {
	case BUFFER_OP:
		return &OpMessage{}
	case ACK_CHANGE:
		return &AcknowledgeChange{}
	default:
		panic("unrecognized message kind")
	}
}

type Message interface {
	Kind() MessageKind
}

type OpMessage struct {
	Op ot.Operation `json:"op"`
}

func (OpMessage) Kind() MessageKind {
	return BUFFER_OP
}

func (m *OpMessage) UnmarshalJSON(body []byte) error {
	type wrapper struct {
		Op struct {
			Type string `json:"type"`
			Pos  uint   `json:"pos"`
			Len  uint   `json:"len"`
			Text string `json:"text"`
		} `json:"op"`
	}

	var w wrapper
	err := json.Unmarshal(body, &w)
	if err != nil {
		return err
	}

	op := w.Op
	switch op.Type {
	case "insertion":
		m.Op = ot.Insertion{Pos: op.Pos, Text: op.Text}
	case "deletion":
		m.Op = ot.Deletion{Pos: op.Pos, Len: op.Len}
	}

	return nil
}

type AcknowledgeChange struct {
}

func (AcknowledgeChange) Kind() MessageKind {
	return ACK_CHANGE
}
