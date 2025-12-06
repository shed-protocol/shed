package comms

import (
	"testing"
)

func TestMessageKinds(t *testing.T) {
	kinds := []MessageKind{
		BUFFER_OP,
		ACK_CHANGE,
	}
	for _, k := range kinds {
		msg := MessageOfKind(k)
		if got := msg.Kind(); got != k {
			t.Errorf("%T{}.Kind() returned %v", msg, got)
		}
	}
}
