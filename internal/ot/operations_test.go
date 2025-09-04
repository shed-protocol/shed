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

func TestOperationCommutativity(t *testing.T) {
	var cases = []struct {
		start    []byte
		opA, opB ot.Operation
	}{
		{
			start: []byte(""),
			opA:   ot.Insertion{Pos: 0, Text: []byte("hello")},
			opB:   ot.Insertion{Pos: 0, Text: []byte("world")},
		},
		{
			start: []byte("hello"),
			opA:   ot.Insertion{Pos: 5, Text: []byte(" world")},
			opB:   ot.Deletion{Pos: 0, Len: 5},
		},
	}

	for _, c := range cases {
		a, b := c.start, c.start
		opA, opB := c.opA, c.opB

		a = opA.Apply(a)
		a = opB.Rebase(opA).Apply(a)
		b = opB.Apply(b)
		b = opA.Rebase(opB).Apply(b)

		if string(a) != string(b) {
			t.Errorf("operations depend on order (%q != %q): %+v, %+v", a, b, opA, opB)
		}
	}
}

func valid(start string, op ot.Operation) bool {
	switch op := op.(type) {
	case ot.Insertion:
		return int(op.Pos) <= len([]byte(start))
	case ot.Deletion:
		return int(op.Pos)+int(op.Len) <= len([]byte(start))
	}
	panic("unhandled operation type")
}

func FuzzInsertInsertCommutativity(f *testing.F) {
	f.Add("0000", uint(0), uint(3), []byte("hello"), []byte("world"))
	f.Add("0000", uint(2), uint(2), []byte("hello"), []byte("world"))
	f.Add("0000", uint(0), uint(0), []byte("hello"), []byte("hello"))
	f.Fuzz(func(t *testing.T, start string, posA, posB uint, textA, textB []byte) {
		opA := ot.Insertion{Pos: posA, Text: textA}
		opB := ot.Insertion{Pos: posB, Text: textB}

		if !valid(start, opA) || !valid(start, opB) {
			t.Skip()
		}

		a, b := []byte(start), []byte(start)

		a = opA.Apply(a)
		a = opB.Rebase(opA).Apply(a)
		b = opB.Apply(b)
		b = opA.Rebase(opB).Apply(b)

		if string(a) != string(b) {
			t.Errorf("operations depend on order (%q != %q): %+v, %+v", a, b, opA, opB)
		}
	})
}

func FuzzInsertDeleteCommutativity(f *testing.F) {
	f.Add("0000", uint(0), uint(3), []byte("hello"), uint(1))
	f.Add("0000", uint(1), uint(1), []byte("hello"), uint(3))
	f.Add("0000", uint(1), uint(0), []byte("hello"), uint(3))
	f.Fuzz(func(t *testing.T, start string, posA, posB uint, textA []byte, lenB uint) {
		opA := ot.Insertion{Pos: posA, Text: textA}
		opB := ot.Deletion{Pos: posB, Len: lenB}

		if !valid(start, opA) || !valid(start, opB) {
			t.Skip()
		}

		a, b := []byte(start), []byte(start)

		a = opA.Apply(a)
		a = opB.Rebase(opA).Apply(a)
		b = opB.Apply(b)
		b = opA.Rebase(opB).Apply(b)

		if string(a) != string(b) {
			t.Errorf("operations depend on order (%q != %q): %+v, %+v", a, b, opA, opB)
		}
	})
}

func FuzzDeleteDeleteCommutativity(f *testing.F) {
	f.Add("0000", uint(0), uint(3), uint(2), uint(1))
	f.Add("0000", uint(3), uint(0), uint(1), uint(2))
	f.Add("0000", uint(0), uint(2), uint(3), uint(2))
	f.Add("0000", uint(1), uint(1), uint(2), uint(1))
	f.Add("0000", uint(1), uint(2), uint(3), uint(1))
	f.Add("0000", uint(2), uint(0), uint(1), uint(4))
	f.Add("0000", uint(0), uint(0), uint(1), uint(1))
	f.Add("", uint(0), uint(0), uint(0), uint(0))
	f.Fuzz(func(t *testing.T, start string, posA, posB uint, lenA, lenB uint) {
		opA := ot.Deletion{Pos: posA, Len: lenA}
		opB := ot.Deletion{Pos: posB, Len: lenB}

		if !valid(start, opA) || !valid(start, opB) {
			t.Skip()
		}

		a, b := []byte(start), []byte(start)

		a = opA.Apply(a)
		a = opB.Rebase(opA).Apply(a)
		b = opB.Apply(b)
		b = opA.Rebase(opB).Apply(b)

		if string(a) != string(b) {
			t.Errorf("operations depend on order (%q != %q): %q, %+v, %+v", a, b, start, opA, opB)
		}
	})
}
