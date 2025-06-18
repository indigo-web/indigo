package radix

import (
	"fmt"
	"github.com/indigo-web/indigo/kv"
	"github.com/stretchr/testify/require"
	"strconv"
	"strings"
	"testing"
)

func BenchmarkTreeMatch(b *testing.B) {
	tree := New[int]()
	tree.Insert("/hello/world", 1)
	tree.Insert("/hello/whopper", 2)
	tree.Insert("/henry/world", 3)
	tree.Insert("/hello/world/somewhere", 4)
	b.ResetTimer()

	for range b.N {
		_, _ = tree.Lookup("/hello/world/somewhere", nil)
	}
	//
	//b.Run("10 static", func(b *testing.B) {
	//	paths := make([]string, 0, 10)
	//	for i := range 10 {
	//
	//	}
	//})
}

func TestTree(t *testing.T) {
	t.Run("static", func(t *testing.T) {
		tree := New[int]()
		keys := []string{"hello", "hell", "henry", "aboba"}

		for i, key := range keys {
			require.NoError(t, tree.Insert(key, i+1))
		}

		for i, key := range keys {
			value, found := tree.Lookup(key, nil)
			require.True(t, found)
			require.Equal(t, i+1, value)
		}
	})

	t.Run("basic dynamic at the end", func(t *testing.T) {
		tree := New[int]()
		wildcards := kv.New()
		require.NoError(t, tree.Insert("/user/:id", 1))
		_, found := tree.Lookup("/user", nil)
		require.False(t, found)
		value, found := tree.Lookup("/user/wow", wildcards)
		require.True(t, found)
		require.Equal(t, 1, value)
		require.Equal(t, "wow", wildcards.Value("id"))
	})

	t.Run("dynamic in the middle", func(t *testing.T) {
		tree := New[int]()
		require.NoError(t, tree.Insert("/user/:id", 1))
		require.NoError(t, tree.Insert("/user/:id/name", 2))
		require.NoError(t, tree.Insert("/user/:id/naked", 3))

		_, found := tree.Lookup("/user/", nil)
		require.False(t, found)

		_, found = tree.Lookup("/user/", nil)
		require.False(t, found)

		_, found = tree.Lookup("/user/42/na", kv.New())
		require.False(t, found)

		_, found = tree.Lookup("/user//name", nil)
		require.False(t, found)

		test(t, tree, "/user/42", 1, "id", "42")
		test(t, tree, "/user/42/name", 2, "id", "42")
		test(t, tree, "/user/42/naked", 3, "id", "42")
	})

	t.Run("overriding static", func(t *testing.T) {
		tree := New[int]()
		require.NoError(t, tree.Insert("hello/world", 1))
		require.NoError(t, tree.Insert("hello/pavlo", 2))
		require.NoError(t, tree.Insert("hello/:name", 3))
		require.NoError(t, tree.Insert("hello/:name/hi", 4))
		require.NoError(t, tree.Insert("hello/pavlo/hi", 5))
		//printTree(tree, 1)

		test(t, tree, "hello/world", 1, "", "")
		test(t, tree, "hello/pavlo", 2, "", "")
		test(t, tree, "hello/pavlo/hi", 5, "", "")
		test(t, tree, "hello/henry", 3, "name", "henry")
		test(t, tree, "hello/jimmy/hi", 4, "name", "jimmy")
	})
}

func test(t *testing.T, tree *Node[int], path string, value int, wKey, wVal string) {
	w := kv.New()
	val, found := tree.Lookup(path, w)
	require.True(t, found)
	require.Equal(t, value, val)
	require.Equal(t, wVal, w.Value(wKey))
}

func printTree(node *Node[int], depth int) {
	for _, p := range node.predecessors {
		fmt.Print(strings.Repeat("-", depth))

		fmt.Print(" ", strconv.Quote(p.value))

		if p.dyn != nil {
			fmt.Printf(" [%s", strconv.Quote(p.dyn.wildcard))
			if p.dyn.isLeaf {
				fmt.Printf(" {%d}", p.dyn.payload)
			}

			fmt.Print("]")

			if p.dyn.next != nil {
				fmt.Println()
				fmt.Println(strings.Repeat("-", depth), "dyn:")
				printTree(p.dyn.next, depth+1)
			}
		}

		if p.isLeaf {
			fmt.Printf(" {%d} ", p.payload)
		}

		fmt.Println()
		printTree(p, depth+1)
	}
}
