package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestTopicTreeNodeIsWildcard(t *testing.T) {
	t.Run("returns true if the part is #", func(t *testing.T) {
		node := &TopicTreeNode{part: "#"}
		assert.True(t, node.IsWildcard())
	})

	t.Run("returns false if the part is normal string", func(t *testing.T) {
		node := &TopicTreeNode{part: "foo"}
		assert.False(t, node.IsWildcard())
	})

	t.Run("returns false if the part is +", func(t *testing.T) {
		node := &TopicTreeNode{part: "+"}
		assert.False(t, node.IsWildcard())
	})
}
