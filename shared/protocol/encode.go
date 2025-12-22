package protocol

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"net"
)

func Encode(msg *Message) ([]byte, error) {
	if len(msg.Content) > MaxPayloadSize {
		return nil, errors.New("message content exceeds maximum payload size")
	}

	if len(msg.Content) != int(msg.Header.Length) {
		return nil, errors.New("message content length does not match header length")
	}

	msg.Header.Length = uint32(len(msg.Content))

	buf := new(bytes.Buffer)

	// Write type (1 byte)
	if err := binary.Write(buf, binary.BigEndian, msg.Header.Type); err != nil {
		return nil, fmt.Errorf("failed to write message type: %w", err)
	}

	// Write length
	if err := binary.Write(buf, binary.BigEndian, msg.Header.Length); err != nil {
		return nil, fmt.Errorf("failed to write message length: %w", err)
	}

	if _, err := buf.Write(msg.Content); err != nil {
		return nil, fmt.Errorf("failed to write message content: %w", err)
	}

	return buf.Bytes(), nil
}

func Decode(conn net.Conn) (*Message, error) {
	header := &Header{}

	if err := binary.Read(conn, binary.BigEndian, &header.Type); err != nil {
		if err == io.EOF {
			return nil, err
		}
		return nil, fmt.Errorf("failed to read message type: %w", err)
	}

	if err := binary.Read(conn, binary.BigEndian, &header.Length); err != nil {
		return nil, fmt.Errorf("failed to read message length: %w", err)
	}

	if header.Length > MaxPayloadSize {
		return nil, fmt.Errorf("message length exceeds maximum payload size: %d", header.Length)
	}

	content := make([]byte, header.Length)
	if _, err := conn.Read(content); err != nil {
		return nil, fmt.Errorf("failed to read message content: %w", err)
	}

	return &Message{
		Header:  *header,
		Content: content,
	}, nil
}
