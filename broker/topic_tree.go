package main

import (
	"fmt"
	"strings"
	"sync"
)

type TopicTree struct {
	root *topicTreeNode
	mu   sync.RWMutex
}

func NewTopicTree() *TopicTree {
	return &TopicTree{
		root: newTopicTreeNode(""),
	}
}

func (t *TopicTree) Add(topic string, client *Client) {
	t.mu.Lock()
	defer t.mu.Unlock()

	parts := strings.Split(topic, "/")

	current := t.root
	for _, part := range parts {
		if _, exists := current.subnodes[part]; !exists {
			current.subnodes[part] = newTopicTreeNode(part)
		}
		current = current.subnodes[part]
	}
	current.clients[client] = true
}

func (t *TopicTree) Get(topic string) []*Client {
	t.mu.RLock()
	defer t.mu.RUnlock()

	parts := strings.Split(topic, "/")

	matchingClients := make([]*Client, 0)

	var traverse func(*topicTreeNode, []string)
	traverse = func(node *topicTreeNode, parts []string) {
		if len(parts) == 0 || node.isWildcard() {
			for client := range node.clients {
				matchingClients = append(matchingClients, client)
			}
		}

		if len(parts) > 0 {
			part := parts[0]
			if nextNode, exists := node.subnodes[part]; exists {
				traverse(nextNode, parts[1:])
			}
			if nextNode, exists := node.subnodes["+"]; exists {
				traverse(nextNode, parts[1:])
			}
			if nextNode, exists := node.subnodes["#"]; exists {
				traverse(nextNode, parts)
			}
		}
	}
	traverse(t.root, parts)

	return matchingClients
}

// Dump prints the topic tree to stdout for debug.
func (t *TopicTree) Print() {
	t.mu.RLock()
	defer t.mu.RUnlock()

	var traverse func(node *topicTreeNode, prefix string)
	traverse = func(node *topicTreeNode, prefix string) {
		clientIDs := make([]string, 0)
		for client := range node.clients {
			clientIDs = append(clientIDs, string(client.ID))
		}
		fmt.Printf("%s%s clients: [%s]\n", prefix, node.part, strings.Join(clientIDs, ", "))

		for _, subnode := range node.subnodes {
			traverse(subnode, prefix+"  ")
		}
	}
	traverse(t.root, "")
}

type topicTreeNode struct {
	part     string
	clients  map[*Client]bool
	subnodes map[string]*topicTreeNode
}

func newTopicTreeNode(part string) *topicTreeNode {
	return &topicTreeNode{
		part:     part,
		clients:  make(map[*Client]bool),
		subnodes: make(map[string]*topicTreeNode),
	}
}

func (n *topicTreeNode) isWildcard() bool {
	return n.part == "#"
}
