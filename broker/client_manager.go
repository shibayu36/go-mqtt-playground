package main

import (
	"bufio"
	"sync"
)

type ClientManager struct {
	clients map[*Client]*bufio.Writer
	mu      sync.Mutex
}

func NewClientManager() *ClientManager {
	return &ClientManager{
		clients: make(map[*Client]*bufio.Writer),
	}
}

func (cm *ClientManager) Add(client *Client, writer *bufio.Writer) {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	cm.clients[client] = writer
}

func (cm *ClientManager) List() []*Client {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	clients := make([]*Client, 0)
	for client := range cm.clients {
		clients = append(clients, client)
	}

	return clients
}
