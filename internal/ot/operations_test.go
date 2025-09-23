package ot_test

import (
	"testing"

	"github.com/shed-protocol/shed/internal/ot"
)

func TestInsert(t *testing.T) {
	cases := []struct {
		start, want string
		op          ot.Insertion
	}{
		{
			start: "",
			want:  "hello",
			op:    ot.Insertion{Pos: 0, Text: "hello"},
		},
		{
			start: "hello",
			want:  "hello world",
			op:    ot.Insertion{Pos: 5, Text: " world"},
		},
		{
			start: "hllo",
			want:  "hello",
			op:    ot.Insertion{Pos: 1, Text: "e"},
		},
	}

	for _, c := range cases {
		if got := c.op.Apply(c.start); got != c.want {
			t.Errorf("insertion failed: got %s, want %s", got, c.want)
		}
	}
}

func TestDelete(t *testing.T) {
	cases := []struct {
		start, want string
		op          ot.Deletion
	}{
		{
			start: "",
			want:  "",
			op:    ot.Deletion{Pos: 0, Len: 0},
		},
		{
			start: "hello world",
			want:  "hello",
			op:    ot.Deletion{Pos: 5, Len: 6},
		},
		{
			start: "hello",
			want:  "hllo",
			op:    ot.Deletion{Pos: 1, Len: 1},
		},
	}

	for _, c := range cases {
		if got := c.op.Apply(c.start); got != c.want {
			t.Errorf("deletion failed: got %s, want %s", got, c.want)
		}
	}
}

func TestOperationCommutativity(t *testing.T) {
	var cases = []struct {
		start    string
		opA, opB ot.Operation
	}{
		{
			start: "",
			opA:   ot.Insertion{Pos: 0, Text: "hello"},
			opB:   ot.Insertion{Pos: 0, Text: "world"},
		},
		{
			start: "hello",
			opA:   ot.Insertion{Pos: 5, Text: " world"},
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

func TestRebaseDeletionPanicsOnInvalidOperation(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Error("expected panic")
		}
	}()

	var invalidOp ot.Operation
	op := ot.Insertion{Pos: 5, Text: "hello"}
	op.Rebase(invalidOp)
}

func TestRebaseInsertionPanicsOnInvalidOperation(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Error("expected panic")
		}
	}()

	var invalidOp ot.Operation
	op := ot.Deletion{Pos: 5, Len: 2}
	op.Rebase(invalidOp)
}

func valid(start string, op ot.Operation) bool {
	switch op := op.(type) {
	case ot.Insertion:
		return int(op.Pos) <= len(start)
	case ot.Deletion:
		return int(op.Pos)+int(op.Len) <= len(start)
	}
	panic("unhandled operation type")
}

func FuzzInsertInsertCommutativity(f *testing.F) {
	f.Add("0000", uint(0), uint(3), "hello", "world")
	f.Add("0000", uint(2), uint(2), "hello", "world")
	f.Add("0000", uint(0), uint(0), "hello", "hello")
	f.Fuzz(func(t *testing.T, start string, posA, posB uint, textA, textB string) {
		opA := ot.Insertion{Pos: posA, Text: textA}
		opB := ot.Insertion{Pos: posB, Text: textB}

		if !valid(start, opA) || !valid(start, opB) {
			t.Skip()
		}

		a, b := start, start

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
	f.Add("0000", uint(0), uint(3), "hello", uint(1))
	f.Add("0000", uint(1), uint(1), "hello", uint(3))
	f.Add("0000", uint(1), uint(0), "hello", uint(3))
	f.Fuzz(func(t *testing.T, start string, posA, posB uint, textA string, lenB uint) {
		opA := ot.Insertion{Pos: posA, Text: textA}
		opB := ot.Deletion{Pos: posB, Len: lenB}

		if !valid(start, opA) || !valid(start, opB) {
			t.Skip()
		}

		a, b := start, start

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

		a, b := start, start

		a = opA.Apply(a)
		a = opB.Rebase(opA).Apply(a)
		b = opB.Apply(b)
		b = opA.Rebase(opB).Apply(b)

		if string(a) != string(b) {
			t.Errorf("operations depend on order (%q != %q): %q, %+v, %+v", a, b, start, opA, opB)
		}
	})
}
