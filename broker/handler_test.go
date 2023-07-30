package main

import (
	"bufio"
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestHandleConnect(t *testing.T) {
	handler := NewHandler()

	// Create a mock CONNECT packet
	packet := []byte{
		0x10, // Packet type for CONNECT
		0x0A, // Remaining length
		// Payload
		0x00, 0x04, // Protocol name length
		0x4D, 0x51, 0x54, 0x54, // Protocol name
		0x04,       // Protocol level
		0x02,       // Connect flags
		0x00, 0x0A, // Keep alive
	}

	reader := bufio.NewReader(bytes.NewReader(packet))
	buf := &bytes.Buffer{}
	writer := bufio.NewWriter(buf)

	handler.Handle(reader, writer)

	// Check if the CONNACK packet was written to the writer
	expectedConnack := []byte{0x20, 0x02, 0x00, 0x00}
	assert.Equal(t, expectedConnack, buf.Bytes(), "Expected CONNACK to be written to the writer")

	// Check if the client was added to the client manager
	assert.Equal(t, 1, len(handler.clientManager.List()))
	assert.NotEmpty(t, handler.clientManager.Get(&Client{ClientID("0")}))
}

func TestHandlePingreq(t *testing.T) {
	handler := NewHandler()

	// Create a mock PINGREQ packet
	packet := []byte{
		0xC0, // Packet type for PINGREQ
		0x00, // Remaining length
	}

	reader := bufio.NewReader(bytes.NewReader(packet))
	buf := &bytes.Buffer{}
	writer := bufio.NewWriter(buf)

	handler.Handle(reader, writer)

	// Check if the PINGRESP packet was written to the writer
	expectedPingresp := []byte{0xD0, 0x00}
	assert.Equal(t, expectedPingresp, buf.Bytes(), "Expected PINGRESP to be written to the writer")
}
