package client

import (
	"net"
	"slices"

	"github.com/shed-protocol/shed/internal/comms"
	"github.com/shed-protocol/shed/internal/ot"
)

type Client struct {
	eIn  chan<- comms.Message
	eOut <-chan comms.Message

	queue []comms.Message
	sent  comms.Message
	sIn   chan<- comms.Message
	sOut  <-chan comms.Message
}

func (c *Client) Attach(editor net.Conn) {
	eIn := make(chan comms.Message)
	eOut := make(chan comms.Message)
	c.eIn = eIn
	c.eOut = eOut
	go comms.ChanToConn(eIn, editor)
	go comms.ConnToChan(editor, eOut)
}

func (c *Client) Connect(server net.Conn) {
	sIn := make(chan comms.Message)
	sOut := make(chan comms.Message)
	c.sIn = sIn
	c.sOut = sOut
	c.sent = nil
	go comms.ChanToConn(sIn, server)
	go comms.ConnToChan(server, sOut)
	go c.loop()
}

func (c *Client) loop() {
	for {
		select {
		case msg := <-c.eOut:
			if msg.Kind() == comms.BUFFER_OP {
				c.queue = append(c.queue, msg)
			}
		case msg := <-c.sOut:
			switch msg.Kind() {
			case comms.ACK_CHANGE:
				c.sent = nil
			case comms.BUFFER_OP:
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
					c.eIn <- comms.OpMessage{Op: op}
				}
				{
					on, _ := asOp(msg)
					for i, m := range c.queue {
						if op, ok := asOp(m); ok {
							c.queue[i] = comms.OpMessage{Op: op.Rebase(on)}
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
	case comms.OpMessage:
		op, ok = m.Op, true
	case *comms.OpMessage:
		op, ok = m.Op, true
	}
	return
}
