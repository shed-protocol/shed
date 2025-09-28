package comms

import (
	"encoding/json"
	"net"
)

type message struct {
	Kind MessageKind     `json:"kind"`
	Body json.RawMessage `json:"body"`
}

func ReadMessage(conn net.Conn) (m Message, err error) {
	content, err := ReadContent(conn)
	if err != nil {
		return
	}

	var wrapper message
	if err = json.Unmarshal([]byte(content), &wrapper); err != nil {
		return
	}
	m = MessageOfKind(wrapper.Kind)
	err = json.Unmarshal(wrapper.Body, &m)
	return
}

func WriteMessage(conn net.Conn, msg Message) error {
	body, err := json.Marshal(msg)
	if err != nil {
		return err
	}
	content, err := json.Marshal(message{Kind: msg.Kind(), Body: body})
	if err != nil {
		return err
	}
	return WriteContent(conn, string(content))
}
