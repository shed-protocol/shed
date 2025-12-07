package comms

import (
	"encoding/json"
	"io"
)

func ChanToWriter(ch <-chan Message, w io.Writer) {
	for m := range ch {
		if err := writeMessage(w, m); err != nil {
			break
		}
	}
}

func ReaderToChan(r io.Reader, ch chan<- Message) {
	for {
		m, err := readMessage(r)
		if err != nil {
			break
		}
		ch <- m
	}
}

type message struct {
	Kind MessageKind     `json:"kind"`
	Body json.RawMessage `json:"body"`
}

func readMessage(r io.Reader) (m Message, err error) {
	content, err := ReadContent(r)
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

func writeMessage(w io.Writer, msg Message) error {
	body, err := json.Marshal(msg)
	if err != nil {
		return err
	}
	content, err := json.Marshal(message{Kind: msg.Kind(), Body: body})
	if err != nil {
		return err
	}
	return WriteContent(w, string(content))
}
