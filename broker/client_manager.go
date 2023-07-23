package main

import (
	"bufio"
	"sync"
)

type ClientManager struct {
	clients map[ClientID]*bufio.Writer
	mu      sync.Mutex
}

func NewClientManager() *ClientManager {
	return &ClientManager{
		clients: make(map[ClientID]*bufio.Writer),
	}
}

func (cm *ClientManager) Add(client *Client, writer *bufio.Writer) {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	cm.clients[client.ID] = writer
}

func (cm *ClientManager) Get(client *Client) *bufio.Writer {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	return cm.clients[client.ID]
}

func (cm *ClientManager) List() []ClientID {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	clientIDs := make([]ClientID, 0)
	for client := range cm.clients {
		clientIDs = append(clientIDs, client)
	}

	return clientIDs
}
