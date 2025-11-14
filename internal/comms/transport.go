package comms

import (
	"encoding/binary"
	"errors"
	"fmt"
	"math"
	"net"
)

const MaxPayloadSize int = 1024 * 1024

var PayloadTooLargeError = errors.New("message too long")

func ReadContent(conn net.Conn) (string, error) {
	header, err := readExactly(conn, 4)
	if err != nil {
		return "", fmt.Errorf("failed to read header: %w", err)
	}
	length := binary.BigEndian.Uint32(header)
	if uint64(length) > math.MaxInt || int(length) > MaxPayloadSize {
		return "", PayloadTooLargeError
	}

	data, err := readExactly(conn, int(length))
	if err != nil {
		return "", fmt.Errorf("failed to read body: %w", err)
	}

	return string(data), err
}

func readExactly(conn net.Conn, n int) ([]byte, error) {
	buf := make([]byte, n)
	read := 0
	for read < n {
		k, err := conn.Read(buf[read:])
		if err != nil {
			return nil, err
		}
		read += k
	}
	return buf, nil
}

func WriteContent(conn net.Conn, msg string) error {
	if len(msg) > MaxPayloadSize {
		return PayloadTooLargeError
	}
	data := make([]byte, len(msg)+4)
	binary.BigEndian.PutUint32(data, uint32(len(msg)))
	copy(data[4:], msg)

	_, err := conn.Write(data)
	return err
}
