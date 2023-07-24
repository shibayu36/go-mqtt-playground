package main

import (
	"bufio"
	"log"
	"sync"
)

type Handler struct {
	TopicTree     *TopicTree
	ClientManager *ClientManager
	NextClientId  int
	mu            sync.Mutex
}

func NewHandler() *Handler {
	return &Handler{
		TopicTree:     NewTopicTree(),
		ClientManager: NewClientManager(),
		NextClientId:  0,
	}
}

func (h *Handler) Handle(reader *bufio.Reader, writer *bufio.Writer) {
	// TODO: ClientID should be the ID sent by CONNECT
	currentClientId := h.NextClientId

	h.mu.Lock()
	h.NextClientId++
	h.mu.Unlock()

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
			handleConnect(reader, writer, h.ClientManager, currentClientId)
		case 3:
			handlePublish(reader, writer, h.TopicTree, h.ClientManager, currentClientId)
		case 8:
			handleSubscribe(reader, writer, h.TopicTree, currentClientId)
		case 12:
			handlePingreq(writer)
		}
	}
}
