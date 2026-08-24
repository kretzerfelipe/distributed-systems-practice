package main

import (
	"encoding/binary"
	"fmt"
	"io"
	"net"
)

const MaxFrameSize = 64 * 1024

func WriteFrame(conn net.Conn, payload []byte) error {
	if len(payload) > MaxFrameSize {
		return fmt.Errorf("payload of %d bytes exceeds the limit", len(payload))
	}

	header := make([]byte, 4)
	binary.BigEndian.PutUint32(header, uint32(len(payload)))

	if _, err := conn.Write(header); err != nil {
		return err
	}
	if _, err := conn.Write(payload); err != nil {
		return err
	}
	return nil
}

func ReadFrame(conn net.Conn) ([]byte, error) {
	header := make([]byte, 4)
	if _, err := io.ReadFull(conn, header); err != nil {
		return nil, err
	}

	size := binary.BigEndian.Uint32(header)
	if size > MaxFrameSize {
		return nil, fmt.Errorf("invalid frame size: %d", size)
	}

	if size == 0 {
		return []byte{}, nil
	}

	payload := make([]byte, size)
	if _, err := io.ReadFull(conn, payload); err != nil {
		return nil, err
	}

	return payload, nil
}
