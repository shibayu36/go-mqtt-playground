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

func TestTopicTreeSubscribeAndClientsToPublish(t *testing.T) {
	t.Run("empty topic tree", func(t *testing.T) {
		node := NewTopicTreeNode("")
		assert.Equal(t, node.ClientsToPublish(("foo")), []*Client{})
		assert.Equal(t, node.ClientsToPublish(("foo/bar")), []*Client{})
		assert.Equal(t, node.ClientsToPublish(("foo/bar/#")), []*Client{})
		assert.Equal(t, node.ClientsToPublish(("#")), []*Client{})
		assert.Equal(t, node.ClientsToPublish(("foo/+/baz")), []*Client{})
	})

	t.Run("simple topic tree", func(t *testing.T) {
		node := NewTopicTreeNode("")
		client1 := &Client{ID: "client1"}
		client2 := &Client{ID: "client2"}
		client3 := &Client{ID: "client3"}

		node.Subscribe("foo/bar", client1)
		node.Subscribe("foo/bar/baz", client1)

		node.Subscribe("foo/bar", client2)
		node.Subscribe("hoge", client2)

		node.Subscribe("foo/bar", client3)
		node.Subscribe("hoge", client3)
		node.Subscribe("hoge/fuga", client3)

		assert.ElementsMatch(t, node.ClientsToPublish(("foo/bar")), []*Client{client1, client2, client3})
		assert.ElementsMatch(t, node.ClientsToPublish(("foo/bar/baz")), []*Client{client1})
		assert.ElementsMatch(t, node.ClientsToPublish(("hoge")), []*Client{client2, client3})
		assert.ElementsMatch(t, node.ClientsToPublish(("hoge/fuga")), []*Client{client3})
		assert.ElementsMatch(t, node.ClientsToPublish(("notexists/1")), []*Client{})
	})

	t.Run("wildcard topic tree", func(t *testing.T) {
		node := NewTopicTreeNode("")
		client1 := &Client{ID: "client1"}
		client2 := &Client{ID: "client2"}
		client3 := &Client{ID: "client3"}
		client4 := &Client{ID: "client4"}

		node.Subscribe("#", client1)
		node.Subscribe("a/b/c", client2)
		node.Subscribe("a/+/c", client3)
		node.Subscribe("a/#", client4)

		assert.ElementsMatch(t, node.ClientsToPublish(("a")), []*Client{client1})
		assert.ElementsMatch(t, node.ClientsToPublish(("a/b")), []*Client{client1, client4})
		assert.ElementsMatch(t, node.ClientsToPublish(("a/b/c")), []*Client{client1, client2, client3, client4})
		assert.ElementsMatch(t, node.ClientsToPublish(("a/b/c/d")), []*Client{client1, client4})

		assert.ElementsMatch(t, node.ClientsToPublish(("b")), []*Client{client1})
	})
}
