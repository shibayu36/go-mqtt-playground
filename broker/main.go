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

	handler := NewHandler()

	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Println("Error accepting connection: ", err)
			continue
		}

		log.Println("New connection accepted")
		go handleConn(conn, handler)
	}
}

func handleConn(conn net.Conn, handler *Handler) {
	defer conn.Close()

	reader := bufio.NewReader(conn)
	writer := bufio.NewWriter(conn)

	handler.Handle(reader, writer)
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
	clientManager.Add(&Client{ClientID(fmt.Sprint(clientId))}, writer)
	spew.Dump(clientManager.List())
}

// handlePublish handles the PUBLISH packet
func handlePublish(reader *bufio.Reader, writer *bufio.Writer, topicTree *TopicTree, clientManager *ClientManager, clientId int) {
	remainingLength, err := readRemainingLength(reader)
	if err != nil {
		log.Println("Error reading remaining length:", err)
		return
	}
	log.Println("Remaining Length:", remainingLength)

	// Read the topic name
	topicLengthBytes := make([]byte, 2)
	reader.Read(topicLengthBytes)
	topicLen := int(topicLengthBytes[0])<<8 | int(topicLengthBytes[1])
	topicBytes := make([]byte, topicLen)
	reader.Read(topicBytes)
	topic := string(topicBytes)

	// Read the message payload
	payloadLen := int(remainingLength) - 2 - topicLen
	payload := make([]byte, payloadLen)
	reader.Read(payload)

	log.Printf("Received PUBLISH (topic: %s, message: %s)\n", topic, string(payload))

	clients := topicTree.Get(topic)
	spew.Dump(clients)
	for _, client := range clients {
		log.Printf("Sending message to client %s\n", client.ID)
		writer := clientManager.Get(client)
		sendPublish(writer, topic, string(payload))
	}

	// TODO: Handle QoS

	// when QoS == 0, no response is required
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

	// Extract the topic length from the payload
	topicLength := int(payload[2])<<8 | int(payload[3])

	// Extract the topic from the payload
	topicStart := 4
	topicEnd := topicStart + topicLength
	topic := payload[topicStart:topicEnd]
	log.Printf("Topic: %s\n", string(topic))
	topicTree.Add(string(topic), &Client{ClientID(fmt.Sprint(clientId))})

	// DEBUG: Print the topic tree
	// TODO: I want to print the topic tree from management http API
	topicTree.Print()

	// TODO: Extract QoS levels from the payload

	// TODO: Handle multiple topics in the payload

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

// sendPublish sends a PUBLISH packet to the client
func sendPublish(writer *bufio.Writer, topic string, payload string) {
	// Packet Type for PUBLISH is 0011 0000 (0x30)
	packetType := byte(0x30)

	// Remaining Length = topic length + 2 bytes for topic length + payload length
	remainingLength := len(topic) + 2 + len(payload)
	remainingLengthBytes := encodeRemainingLength(remainingLength)

	// Write Fixed Header
	writer.WriteByte(packetType)
	writer.Write(remainingLengthBytes)

	// Write Variable Header
	writer.WriteByte(byte(len(topic) >> 8))
	writer.WriteByte(byte(len(topic)))
	writer.WriteString(topic)

	// Write Payload
	writer.WriteString(payload)

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
