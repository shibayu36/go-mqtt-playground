package main

import (
	"bufio"
	"log"
	"net"
)

func main() {
	listener, err := net.Listen("tcp", ":1883")
	if err != nil {
		log.Fatal(err)
	}
	defer listener.Close()

	log.Println("Listening on port 1883")
	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Println("Error accepting connection: ", err)
			continue
		}

		log.Println("New connection accepted")
		go handleConnection(conn)
	}
}

func handleConnection(conn net.Conn) {
	defer conn.Close()

	reader := bufio.NewReader(conn)
	writer := bufio.NewWriter(conn)

	for {
		// Read the first byte (this should be the packet type)
		packetType, err := reader.ReadByte()
		if err != nil {
			log.Println("Error reading packet type:", err)
			return
		}

		log.Println("Packet Type:", packetType)

		switch packetType >> 4 {
		case 1:
			handleConnect(reader, writer)
		}
	}
}

// handleConnect handles the CONNECT packet
func handleConnect(reader *bufio.Reader, writer *bufio.Writer) {
	// Read the remaining length
	remainingLength, err := readRemainingLength(reader)
	if err != nil {
		log.Println("Error reading remaining length:", err)
		return
	}
	log.Println("Remaining Length:", remainingLength)

	// Read the bytes specified by the Remaining Length
	payload := make([]byte, remainingLength)
	_, err = reader.Read(payload)
	if err != nil {
		log.Println("Error reading payload:", err)
		return
	}
	log.Printf("Payload: %v\n", payload)

	// TODO: Handling Connect Flags
	// User Name Flag, Password Flag, Will Retain, Will QoS, Will Flag, Clean Session

	// TODO: Handling Keep Alive.

	// Send the connack
	connack := []byte{0x20, 0x02, 0x00, 0x00}
	_, err = writer.Write(connack)
	if err != nil {
		log.Println("Error sending CONNACK:", err)
		return
	}
	err = writer.Flush()
	if err != nil {
		log.Println("Error flushing writer:", err)
		return
	}

	log.Println("Sent CONNACK packet")
}

// MQTTのFixed HeaderのRemaining Lengthを読み込む
func readRemainingLength(reader *bufio.Reader) (int, error) {
	var value int
	var multiplier int = 1

	for {
		digit, err := reader.ReadByte()
		if err != nil {
			return 0, err
		}

		// 最下位7bitをvalueとして使い、最上位1bitを継続判定として利用している
		value += int(digit&127) * multiplier
		multiplier *= 128

		if digit&128 == 0 {
			break
		}
	}

	return value, nil
}
