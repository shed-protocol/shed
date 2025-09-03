package ot_test

import (
	"bytes"
	"testing"

	"github.com/shed-protocol/shed/internal/ot"
)

func TestInsert(t *testing.T) {
	cases := []struct {
		start, want []byte
		op          ot.Insertion
	}{
		{
			start: []byte(""),
			want:  []byte("hello"),
			op:    ot.Insertion{Pos: 0, Text: []byte("hello")},
		},
		{
			start: []byte("hello"),
			want:  []byte("hello world"),
			op:    ot.Insertion{Pos: 5, Text: []byte(" world")},
		},
		{
			start: []byte("hllo"),
			want:  []byte("hello"),
			op:    ot.Insertion{Pos: 1, Text: []byte("e")},
		},
	}

	for _, c := range cases {
		if got := c.op.Apply(c.start); !bytes.Equal(got, c.want) {
			t.Errorf("insertion failed: got %s, want %s", got, c.want)
		}
	}
}

func TestDelete(t *testing.T) {
	cases := []struct {
		start, want []byte
		op          ot.Deletion
	}{
		{
			start: []byte(""),
			want:  []byte(""),
			op:    ot.Deletion{Pos: 0, Len: 0},
		},
		{
			start: []byte("hello world"),
			want:  []byte("hello"),
			op:    ot.Deletion{Pos: 5, Len: 6},
		},
		{
			start: []byte("hello"),
			want:  []byte("hllo"),
			op:    ot.Deletion{Pos: 1, Len: 1},
		},
	}

	for _, c := range cases {
		if got := c.op.Apply(c.start); !bytes.Equal(got, c.want) {
			t.Errorf("deletion failed: got %s, want %s", got, c.want)
		}
	}
}
