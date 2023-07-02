package main

type TopicTreeNode struct {
	part     string
	clients  map[*Client]bool
	subnodes map[string]*TopicTreeNode
}

func NewTopicTreeNode() *TopicTreeNode {
	return &TopicTreeNode{
		clients:  make(map[*Client]bool),
		subnodes: make(map[string]*TopicTreeNode),
	}
}

func (n *TopicTreeNode) IsWildcard() bool {
	return n.part == "#"
}
