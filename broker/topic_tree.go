package main

import "strings"

type TopicTree struct {
	root *topicTreeNode
}

func NewTopicTree() *TopicTree {
	return &TopicTree{
		root: newTopicTreeNode(""),
	}
}

func (t *TopicTree) Add(topic string, client *Client) {
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
