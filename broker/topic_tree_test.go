package main

import (
	"fmt"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestTopicTreeSubscribeAndClientsToPublish(t *testing.T) {
	t.Run("empty topic tree", func(t *testing.T) {
		tree := NewTopicTree()
		assert.Equal(t, tree.Get(("foo")), []*Client{})
		assert.Equal(t, tree.Get(("foo/bar")), []*Client{})
		assert.Equal(t, tree.Get(("foo/bar/#")), []*Client{})
		assert.Equal(t, tree.Get(("#")), []*Client{})
		assert.Equal(t, tree.Get(("foo/+/baz")), []*Client{})
	})

	t.Run("simple topic tree", func(t *testing.T) {
		tree := NewTopicTree()
		client1 := &Client{ID: "client1"}
		client2 := &Client{ID: "client2"}
		client3 := &Client{ID: "client3"}

		tree.Add("foo/bar", client1)
		tree.Add("foo/bar/baz", client1)

		tree.Add("foo/bar", client2)
		tree.Add("hoge", client2)

		tree.Add("foo/bar", client3)
		tree.Add("hoge", client3)
		tree.Add("hoge/fuga", client3)

		assert.ElementsMatch(t, tree.Get(("foo/bar")), []*Client{client1, client2, client3})
		assert.ElementsMatch(t, tree.Get(("foo/bar/baz")), []*Client{client1})
		assert.ElementsMatch(t, tree.Get(("hoge")), []*Client{client2, client3})
		assert.ElementsMatch(t, tree.Get(("hoge/fuga")), []*Client{client3})
		assert.ElementsMatch(t, tree.Get(("notexists/1")), []*Client{})
	})

	t.Run("wildcard topic tree", func(t *testing.T) {
		tree := NewTopicTree()
		client1 := &Client{ID: "client1"}
		client2 := &Client{ID: "client2"}
		client3 := &Client{ID: "client3"}
		client4 := &Client{ID: "client4"}

		tree.Add("#", client1)
		tree.Add("a/b/c", client2)
		tree.Add("a/+/c", client3)
		tree.Add("a/#", client4)

		assert.ElementsMatch(t, tree.Get(("a")), []*Client{client1})
		assert.ElementsMatch(t, tree.Get(("a/b")), []*Client{client1, client4})
		assert.ElementsMatch(t, tree.Get(("a/b/c")), []*Client{client1, client2, client3, client4})
		assert.ElementsMatch(t, tree.Get(("a/b/c/d")), []*Client{client1, client4})

		assert.ElementsMatch(t, tree.Get(("b")), []*Client{client1})
	})
}

func TestTopicTreeConcurrency(t *testing.T) {
	tree := NewTopicTree()

	var wg sync.WaitGroup
	const numGoroutines = 100

	// 同時にクライアントを追加するgoroutineを起動
	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			client := &Client{ID: fmt.Sprint(id)}
			tree.Add("topic", client)
		}(i)
	}

	// 同時にトピックを取得するgoroutineを起動
	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			tree.Get("topic")
		}()
	}

	// すべてのgoroutineが終了するのを待つ
	wg.Wait()

	// 最終的な結果を確認
	clients := tree.Get("topic")
	if len(clients) != numGoroutines {
		t.Errorf("Expected %d clients, but got %d", numGoroutines, len(clients))
	}
}

func TestTopicTreeNodeIsWildcard(t *testing.T) {
	t.Run("returns true if the part is #", func(t *testing.T) {
		node := &topicTreeNode{part: "#"}
		assert.True(t, node.isWildcard())
	})

	t.Run("returns false if the part is normal string", func(t *testing.T) {
		node := &topicTreeNode{part: "foo"}
		assert.False(t, node.isWildcard())
	})

	t.Run("returns false if the part is +", func(t *testing.T) {
		node := &topicTreeNode{part: "+"}
		assert.False(t, node.isWildcard())
	})
}
