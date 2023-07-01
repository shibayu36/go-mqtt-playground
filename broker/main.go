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
		case 8:
			handleSubscribe(reader, writer)
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

// handleSubscribe handles the SUBSCRIBE packet
func handleSubscribe(reader *bufio.Reader, writer *bufio.Writer) {
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

	// Extract the packet ID from the payload
	packetID := payload[0:2]

	// TODO: Parse other bytes in the payload, and sends appropriate messages to the client.

	// TODO: Extract QoS levels from the payload

	// Define the return codes for the subscription (for this example, assuming success for one subscription)
	returnCodes := []byte{0x00}

	// Send the SUBACK
	sendSubAck(writer, packetID, returnCodes)
}

func sendSubAck(writer *bufio.Writer, packetID []byte, returnCodes []byte) {
	// Packet Type for SUBACK is 1001 0000 (0x90)
	packetType := byte(0x90)

	// Remaining Length = length of variable header + length of payload
	remainingLength := len(packetID) + len(returnCodes)
	remainingLengthBytes := encodeRemainingLength(remainingLength)

	// Write Fixed Header
	writer.WriteByte(packetType)
	writer.Write(remainingLengthBytes)

	// Write Variable Header
	writer.Write(packetID)

	// Write Payload
	writer.Write(returnCodes)

	// Flush the writer to ensure all data is sent
	writer.Flush()
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

func encodeRemainingLength(length int) []byte {
	encoded := make([]byte, 0)

	for {
		digit := length % 128
		length = length / 128

		if length > 0 {
			digit = digit | 0x80
		}

		encoded = append(encoded, byte(digit))

		if length == 0 {
			break
		}
	}

	return encoded
}
