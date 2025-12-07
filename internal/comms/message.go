package comms

import (
	"encoding/json"
	"fmt"

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
	type opWrapper struct {
		Op json.RawMessage `json:"op"`
	}
	var w1 opWrapper
	if err := json.Unmarshal(body, &w1); err != nil {
		return err
	}

	type typeWrapper struct {
		Type string `json:"type"`
	}
	var w2 typeWrapper
	if err := json.Unmarshal(w1.Op, &w2); err != nil {
		return err
	}

	switch w2.Type {
	case "insertion":
		var op ot.Insertion
		if err := json.Unmarshal(w1.Op, &op); err != nil {
			return err
		}
		m.Op = op
	case "deletion":
		var op ot.Deletion
		if err := json.Unmarshal(w1.Op, &op); err != nil {
			return err
		}
		m.Op = op
	default:
		return fmt.Errorf("unrecognized operation type: %q", w2.Type)
	}
	return nil
}

type AcknowledgeChange struct {
}

func (AcknowledgeChange) Kind() MessageKind {
	return ACK_CHANGE
}
