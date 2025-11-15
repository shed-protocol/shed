package client

import (
	"net"
	"slices"

	"github.com/shed-protocol/shed/internal/comms"
	"github.com/shed-protocol/shed/internal/ot"
)

type Client struct {
	editor net.Conn
	eIn    chan<- comms.Message
	eOut   <-chan comms.Message

	server net.Conn
	queue  []comms.Message
	sent   comms.Message
	sIn    chan<- comms.Message
	sOut   <-chan comms.Message
}

func (c *Client) Attach(editor net.Conn) {
	eIn := make(chan comms.Message)
	eOut := make(chan comms.Message)
	c.editor = editor
	c.eIn = eIn
	c.eOut = eOut
	go comms.ChanToConn(eIn, c.editor)
	go comms.ConnToChan(c.editor, eOut)
}

func (c *Client) Connect(server net.Conn) {
	sIn := make(chan comms.Message)
	sOut := make(chan comms.Message)
	c.server = server
	c.sIn = sIn
	c.sOut = sOut
	c.sent = nil
	go comms.ChanToConn(sIn, c.server)
	go comms.ConnToChan(c.server, sOut)
	go c.loop()
}

func (c *Client) loop() {
	for {
		select {
		case msg := <-c.eOut:
			if k := msg.Kind(); k == comms.OP_DELETION || k == comms.OP_INSERTION {
				c.queue = append(c.queue, msg)
			}
		case msg := <-c.sOut:
			switch msg.Kind() {
			case comms.ACK_CHANGE:
				c.sent = nil
			case comms.OP_INSERTION, comms.OP_DELETION:
				{
					op, _ := asOp(msg)
					if c.sent != nil {
						if on, ok := asOp(c.sent); ok {
							op = op.Rebase(on)
						}
					}
					for _, m := range slices.Backward(c.queue) {
						if on, ok := asOp(m); ok {
							op = op.Rebase(on)
						}
					}
					switch op := op.(type) {
					case ot.Insertion:
						c.eIn <- comms.InsertionMessage{Op: op}
					case ot.Deletion:
						c.eIn <- comms.DeletionMessage{Op: op}
					}
				}
				{
					on, _ := asOp(msg)
					for i, m := range c.queue {
						if op, ok := asOp(m); ok {
							switch op := op.Rebase(on).(type) {
							case ot.Insertion:
								c.queue[i] = comms.InsertionMessage{Op: op}
							case ot.Deletion:
								c.queue[i] = comms.DeletionMessage{Op: op}
							}
						}
					}
				}
			}
		default:
			if c.sent == nil && len(c.queue) > 0 {
				msg := c.queue[0]
				c.queue = c.queue[1:]
				c.sIn <- msg
				c.sent = msg
			}
		}
	}
}

func asOp(m comms.Message) (op ot.Operation, ok bool) {
	switch m := m.(type) {
	case comms.InsertionMessage:
		op = m.Op
		ok = true
	case *comms.InsertionMessage:
		op = m.Op
		ok = true
	case comms.DeletionMessage:
		op = m.Op
		ok = true
	case *comms.DeletionMessage:
		op = m.Op
		ok = true
	}
	return
}
