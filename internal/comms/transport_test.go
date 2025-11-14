package comms_test

import (
	"encoding/binary"
	"errors"
	"net"
	"strings"
	"sync"
	"testing"

	"github.com/shed-protocol/shed/internal/comms"
)

func TestReadReturnsMessage(t *testing.T) {
	alice, bob := net.Pipe()
	defer alice.Close()
	defer bob.Close()

	msg := "Hello, world!"

	var wg sync.WaitGroup
	wg.Go(func() {
		if err := comms.WriteContent(alice, msg); err != nil {
			t.Errorf("error writing message: %s", err)
		}
	})
	wg.Go(func() {
		got, err := comms.ReadContent(bob)
		if err != nil {
			t.Errorf("error reading message: %s", err)
			return
		}
		if got != msg {
			t.Errorf("received message != sent message (%q != %q)", got, msg)
		}
	})
	wg.Wait()
}

func TestReadFailsOnLongMessage(t *testing.T) {
	const n = uint32(comms.MaxPayloadSize) + 1
	header := binary.BigEndian.AppendUint32(nil, n)
	alice, bob := net.Pipe()

	var wg sync.WaitGroup
	wg.Go(func() {
		alice.Write(header)
	})
	wg.Go(func() {
		if _, err := comms.ReadContent(bob); !errors.Is(err, comms.PayloadTooLargeError) {
			t.Errorf("expected PayloadTooLargeError")
		}
	})
	wg.Wait()
}

func TestReadFailsOnConnectionError(t *testing.T) {
	alice, bob := net.Pipe()
	alice.Close()
	defer bob.Close()

	_, err := comms.ReadContent(bob)
	if err == nil {
		t.Errorf("expected error")
	}
}

func TestWriteAddsLengthHeader(t *testing.T) {
	cases := []string{
		"",
		"hello",
		"abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ123456890abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ123456890abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ123456890abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ123456890abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ123456890",
		"a string whose length doesn't equal the number of runes বাংলা",
	}

	for _, s := range cases {
		alice, bob := net.Pipe()
		defer alice.Close()
		defer bob.Close()

		var wg sync.WaitGroup
		wg.Go(func() {
			go comms.WriteContent(alice, s)
		})
		wg.Go(func() {
			header := make([]byte, 4)
			bob.Read(header)

			if got := binary.BigEndian.Uint32(header); got != uint32(len(s)) {
				t.Errorf("header says %v bytes, string %q has %v bytes", got, s, len(s))
			}
		})
		wg.Wait()
	}
}

func TestWriteFailsOnLongMessage(t *testing.T) {
	msg := strings.Repeat("a", comms.MaxPayloadSize+1)
	alice, _ := net.Pipe()
	if err := comms.WriteContent(alice, msg); !errors.Is(err, comms.PayloadTooLargeError) {
		t.Errorf("expected PayloadTooLargeError")
	}
}
