package main

import (
	"bufio"
	"fmt"
	"log"
	"net"

	"github.com/davecgh/go-spew/spew"
)

func main() {
	listener, err := net.Listen("tcp", ":1883")
	if err != nil {
		log.Fatal(err)
	}
	defer listener.Close()

	log.Println("Listening on port 1883")

	topicTree := NewTopicTree()
	clientManager := NewClientManager()

	// TODO: ClientID should be the ID sent by CONNECT
	clientId := 0
	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Println("Error accepting connection: ", err)
			continue
		}

		log.Println("New connection accepted")
		go handleConnection(conn, topicTree, clientManager, clientId)
		clientId++
	}
}

func handleConnection(conn net.Conn, topicTree *TopicTree, clientManager *ClientManager, clientId int) {
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
			handleConnect(reader, writer, clientManager, clientId)
		case 8:
			handleSubscribe(reader, writer, topicTree, clientId)
		case 12:
			handlePingreq(writer)
		}
	}
}

// handleConnect handles the CONNECT packet
func handleConnect(reader *bufio.Reader, writer *bufio.Writer, clientManager *ClientManager, clientId int) {
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

	// Store the client in the client manager
	clientManager.Add(&Client{fmt.Sprint(clientId)}, writer)
	spew.Dump(clientManager.List())
}

// handleSubscribe handles the SUBSCRIBE packet
// TODO: SUBSCRIBE should store the subscription information in a map
func handleSubscribe(reader *bufio.Reader, writer *bufio.Writer, topicTree *TopicTree, clientId int) {
	// Read the remaining length
	// TODO: Create remainingLengthParser
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

	// Extract the topic from the payload
	topic := payload[2 : remainingLength-1]
	topicTree.Add(string(topic), &Client{fmt.Sprint(clientId)})

	// DEBUG: Print the topic tree
	// TODO: I want to print the topic tree from management http API
	topicTree.Print()

	// TODO: Extract QoS levels from the payload

	// Define the return codes for the subscription (for this example, assuming success for one subscription)
	returnCodes := []byte{0x00}

	// Send the SUBACK
	// TODO: Send the SUBACK with the appropriate return codes using QoS
	sendSubAck(writer, packetID, returnCodes)
}

func handlePingreq(writer *bufio.Writer) {
	log.Println("Received PINGREQ")

	// PINGRESP packet is 0xD0 followed by 0x00
	pingResp := []byte{0xD0, 0x00}

	_, err := writer.Write(pingResp)
	if err != nil {
		log.Println("Error sending PINGRESP:", err)
		return
	}

	err = writer.Flush()
	if err != nil {
		log.Println("Error flushing PINGRESP:", err)
	}
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
