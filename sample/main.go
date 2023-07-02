package main

import (
	"fmt"
	"strings"
)

type Client struct {
	ID string
}

type Node struct {
	clients    map[*Client]bool
	subnodes   map[string]*Node
	isWildcard bool
}

func NewNode() *Node {
	return &Node{
		clients:  make(map[*Client]bool),
		subnodes: make(map[string]*Node),
	}
}

func (n *Node) Subscribe(topic string, client *Client) {
	parts := strings.Split(topic, "/")

	current := n
	for _, part := range parts {
		if _, exists := current.subnodes[part]; !exists {
			current.subnodes[part] = NewNode()
		}
		current = current.subnodes[part]

		if part == "#" {
			current.isWildcard = true
		}
	}
	current.clients[client] = true
}

func (n *Node) Publish(topic string, message string) {
	parts := strings.Split(topic, "/")

	var publish func(*Node, []string)
	publish = func(node *Node, parts []string) {
		if len(parts) == 0 || node.isWildcard {
			for client := range node.clients {
				fmt.Printf("Message '%s' delivered to client %s\n", message, client.ID)
			}
		}

		if len(parts) > 0 {
			part := parts[0]
			if nextNode, exists := node.subnodes[part]; exists {
				publish(nextNode, parts[1:])
			}
			if nextNode, exists := node.subnodes["+"]; exists {
				publish(nextNode, parts[1:])
			}
			if nextNode, exists := node.subnodes["#"]; exists {
				publish(nextNode, parts)
			}
		}
	}

	publish(n, parts)
}

func main() {
	root := NewNode()

	client1 := &Client{ID: "Client1"}
	client2 := &Client{ID: "Client2"}

	root.Subscribe("home/kitchen/temperature", client1)
	root.Subscribe("home/+/temperature", client2)
	root.Subscribe("home/#", client2)

	root.Publish("home/kitchen/temperature", "22C")
	root.Publish("home/living_room/temperature", "23C")
	root.Publish("office/room1/temperature", "24C")
}
