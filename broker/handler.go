package main

import (
	"bufio"
	"fmt"
	"log"
	"sync"

	"github.com/davecgh/go-spew/spew"
)

type Handler struct {
	topicTree     *TopicTree
	clientManager *ClientManager
	nextClientId  int
	mu            sync.Mutex
}

func NewHandler() *Handler {
	return &Handler{
		topicTree:     NewTopicTree(),
		clientManager: NewClientManager(),
		nextClientId:  0,
	}
}

func (h *Handler) Handle(reader *bufio.Reader, writer *bufio.Writer) {
	// TODO: ClientID should be the ID sent by CONNECT
	currentClientId := h.nextClientId

	h.mu.Lock()
	h.nextClientId++
	h.mu.Unlock()

	// First packet must be CONNECT
	bs, err := reader.Peek(1)
	if err != nil {
		log.Println("Error reading packet type:", err)
		return
	}
	packetType := bs[0]
	if packetType>>4 != 1 {
		log.Println("First packet must be CONNECT")
		return
	}
	h.handleConnect(reader, writer, currentClientId)

	for {
		// Read the first byte (this should be the packet type)
		bs, err := reader.Peek(1)
		if err != nil {
			log.Println("Error reading packet type:", err)
			return
		}
		packetType := bs[0]

		log.Println("Packet Type:", packetType)

		switch packetType >> 4 {
		case 1:
			log.Println("Received CONNECT packet twice")
			return
		case 3:
			h.handlePublish(reader, writer, currentClientId)
		case 8:
			h.handleSubscribe(reader, writer, currentClientId)
		case 12:
			h.handlePingreq(reader, writer)
		default:
			log.Println("Unsupported packet type:", packetType)
			reader.ReadByte() // Read the byte to advance the reader
		}
	}
}

func (h *Handler) handleConnect(reader *bufio.Reader, writer *bufio.Writer, clientId int) {
	// Read the first byte (this should be the packet type)
	reader.ReadByte()

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
	h.clientManager.Add(&Client{ClientID(fmt.Sprint(clientId))}, writer)
	spew.Dump(h.clientManager.List())
}

// handlePublish handles the PUBLISH packet
func (h *Handler) handlePublish(reader *bufio.Reader, writer *bufio.Writer, clientId int) {
	// Read the first byte (this should be the packet type)
	reader.ReadByte()

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

	clients := h.topicTree.Get(topic)
	spew.Dump(clients)
	for _, client := range clients {
		log.Printf("Sending message to client %s\n", client.ID)
		writer := h.clientManager.Get(client)
		h.sendPublish(writer, topic, string(payload))
	}

	// TODO: Handle QoS

	// when QoS == 0, no response is required
}

// handleSubscribe handles the SUBSCRIBE packet
// TODO: SUBSCRIBE should store the subscription information in a map
func (h *Handler) handleSubscribe(reader *bufio.Reader, writer *bufio.Writer, clientId int) {
	// Read the first byte (this should be the packet type)
	reader.ReadByte()

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
	h.topicTree.Add(string(topic), &Client{ClientID(fmt.Sprint(clientId))})

	// DEBUG: Print the topic tree
	// TODO: I want to print the topic tree from management http API
	h.topicTree.Print()

	// TODO: Extract QoS levels from the payload

	// TODO: Handle multiple topics in the payload

	// Define the return codes for the subscription (for this example, assuming success for one subscription)
	returnCodes := []byte{0x00}

	// Send the SUBACK
	// TODO: Send the SUBACK with the appropriate return codes using QoS
	h.sendSubAck(writer, packetID, returnCodes)
}

func (h *Handler) handlePingreq(reader *bufio.Reader, writer *bufio.Writer) {
	// Read the first byte (this should be the packet type)
	reader.ReadByte()

	log.Println("Received PINGREQ")

	// Pingreq has no payload, so read the remaining length and ignore it
	_, err := readRemainingLength(reader)
	if err != nil {
		log.Println("Error reading remaining length:", err)
		return
	}

	// PINGRESP packet is 0xD0 followed by 0x00
	pingResp := []byte{0xD0, 0x00}

	_, err = writer.Write(pingResp)
	if err != nil {
		log.Println("Error sending PINGRESP:", err)
		return
	}

	err = writer.Flush()
	if err != nil {
		log.Println("Error flushing PINGRESP:", err)
	}
}

func (h *Handler) sendSubAck(writer *bufio.Writer, packetID []byte, returnCodes []byte) {
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
func (h *Handler) sendPublish(writer *bufio.Writer, topic string, payload string) {
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
