package main

import (
	"bufio"
	"fmt"
	"log"
	"net"
)

func main() {
	// Connect to the MQTT broker
	conn, err := net.Dial("tcp", "localhost:1883")
	if err != nil {
		log.Fatal("Error connecting to broker:", err)
	}
	defer conn.Close()

	// Send CONNECT packet
	connectPacket := []byte{
		0x10,       // CONNECT packet type
		13,         // Remaining Length
		0x00, 0x04, // Length of protocol name
		'M', 'Q', 'T', 'T', // Protocol name
		0x05,       // Protocol level (MQTT 5.0)
		0x02,       // Connect flags (Clean Start)
		0x00, 0x3C, // Keep Alive (60 seconds)
		0x00, // Properties (none)
		0x00, // Client ID (empty)
	}
	_, err = conn.Write(connectPacket)
	if err != nil {
		log.Fatal("Error sending CONNECT:", err)
	}

	// Read CONNACK packet
	reader := bufio.NewReader(conn)
	header, err := reader.ReadByte()
	if err != nil {
		log.Fatal("Error reading CONNACK header:", err)
	}

	// Check if it's a CONNACK packet (0x20)
	if header>>4 != 2 {
		log.Fatal("Received non-CONNACK packet")
	}

	_, err = reader.ReadByte() // Read Remaining Length (assuming it's small enough to be in one byte)
	if err != nil {
		log.Fatal("Error reading Remaining Length:", err)
	}

	// Read the rest of the CONNACK packet
	connAckFlags, err := reader.ReadByte()
	if err != nil {
		log.Fatal("Error reading CONNACK flags:", err)
	}
	returnCode, err := reader.ReadByte()
	if err != nil {
		log.Fatal("Error reading CONNACK return code:", err)
	}

	// Output the result
	fmt.Printf("CONNACK Flags: 0x%X, Return Code: 0x%X\n", connAckFlags, returnCode)
}
