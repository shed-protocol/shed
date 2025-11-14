package comms

import (
	"testing"
)

func TestMessageKinds(t *testing.T) {
	kinds := []MessageKind{
		OP_INSERTION,
		OP_DELETION,
	}
	for _, k := range kinds {
		msg := MessageOfKind(k)
		if got := msg.Kind(); got != k {
			t.Errorf("%T{}.Kind() returned %v", msg, got)
		}
	}
}
