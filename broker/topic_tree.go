package main

import "strings"

type TopicTreeNode struct {
	part     string
	clients  map[*Client]bool
	subnodes map[string]*TopicTreeNode
}

func NewTopicTreeNode(part string) *TopicTreeNode {
	return &TopicTreeNode{
		part:     part,
		clients:  make(map[*Client]bool),
		subnodes: make(map[string]*TopicTreeNode),
	}
}

func (n *TopicTreeNode) IsWildcard() bool {
	return n.part == "#"
}

func (n *TopicTreeNode) Subscribe(topic string, client *Client) {
	parts := strings.Split(topic, "/")

	current := n
	for _, part := range parts {
		if _, exists := current.subnodes[part]; !exists {
			current.subnodes[part] = NewTopicTreeNode(part)
		}
		current = current.subnodes[part]
	}
	current.clients[client] = true
}

func (n *TopicTreeNode) ClientsToPublish(topic string) []*Client {
	parts := strings.Split(topic, "/")

	matchingClients := make([]*Client, 0)

	var traverse func(*TopicTreeNode, []string)
	traverse = func(node *TopicTreeNode, parts []string) {
		if len(parts) == 0 || node.IsWildcard() {
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
	traverse(n, parts)

	return matchingClients
}
