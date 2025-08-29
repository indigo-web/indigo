package radix

import (
	"fmt"
	"iter"
	"slices"
	"strconv"
	"strings"
	"testing"

	"github.com/indigo-web/indigo/kv"
	"github.com/stretchr/testify/require"
)

func nextstr(str []byte) {
	for i := len(str); i > 0; i-- {
		str[i-1]++
		if str[i-1] <= 'Z' {
			return
		}

		str[i-1] = 'A'
	}
}

func produceSegments(n, seglen int) []string {
	segments := make([]string, n)
	str := []byte(strings.Repeat("A", seglen))

	for i := range n {
		segments[i] = string(str)
		nextstr(str)
	}

	return segments
}

func produceStrings(width, depth, strlen int) (iter.Seq[string], string) {
	allSegs := produceSegments(width, strlen)
	segments, carry := allSegs[:width-1], allSegs[width-1]

	return func(yield func(string) bool) {
		var base string

		for i := 0; i < depth; i++ {
			for _, segment := range segments {
				if !yield(base + segment) {
					break
				}
			}

			base += carry
		}

		yield(base)
	}, strings.Repeat(carry, depth)
}

func TestBench(t *testing.T) {
	it, lastKey := produceStrings(3, 3, 2)
	require.Equal(t, "ACACAC", lastKey)
	require.Equal(t,
		[]string{"AA", "AB", "ACAA", "ACAB", "ACACAA", "ACACAB", "ACACAC"},
		slices.Collect(it),
	)
}

func BenchmarkTree(b *testing.B) {
	noError := func(err error) {
		if err != nil {
			panic(err.Error())
		}
	}

	runBench := func(width, depth int) func(b *testing.B) {
		return func(b *testing.B) {
			const seglen = 8
			tree := New[int]()
			it, key := produceStrings(width, depth, seglen)
			for str := range it {
				noError(tree.Insert(str, 1))
			}

			b.ResetTimer()

			for range b.N {
				_, _ = tree.Lookup(key, nil)
			}
		}
	}

	b.Run("deep", func(b *testing.B) {
		b.Run("1x128", runBench(1, 128))
		b.Run("8x8", runBench(8, 8))
		b.Run("8x64", runBench(8, 64))
		b.Run("8x128", runBench(8, 128))
		b.Run("8x256", runBench(68, 128))
	})

	b.Run("wide", func(b *testing.B) {
		b.Run("128x1", runBench(128, 1))
		b.Run("32x8", runBench(32, 8))
		b.Run("64x8", runBench(64, 8))
		b.Run("128x8", runBench(128, 8))
		b.Run("256x8", runBench(256, 8))
	})

	b.Run("quadratic", func(b *testing.B) {
		b.Run("16x16", runBench(16, 16))
		b.Run("32x32", runBench(32, 32))
		b.Run("64x64", runBench(64, 64))
		b.Run("128x128", runBench(128, 64))
	})
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

	test := func(t *testing.T, tree *Node[int], path string, value int, wKey, wVal string) {
		w := kv.New()
		val, found := tree.Lookup(path, w)
		require.True(t, found)
		require.Equal(t, value, val)
		require.Equal(t, wVal, w.Value(wKey))
	}

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

	t.Run("named greedy wildcard", func(t *testing.T) {
		tree := New[int]()
		require.NoError(t, tree.Insert("/:path...", 1))

		test(t, tree, "/", 1, "path", "")
		test(t, tree, "/hello/world", 1, "path", "hello/world")
	})

	t.Run("anonymous greedy wildcard", func(t *testing.T) {
		tree := New[int]()
		require.NoError(t, tree.Insert("/:...", 1))

		test(t, tree, "/", 1, "", "")
		test(t, tree, "/hello/world", 1, "", "")
	})

	t.Run("dynamic breaks down static segment", func(t *testing.T) {
		tree := New[int]()
		require.NoError(t, tree.Insert("/file/by-path/:path...", 2))
		require.NoError(t, tree.Insert("/file/:name", 1))

		test(t, tree, "/file/photo.jpg", 1, "name", "photo.jpg")
		test(t, tree, "/file/by-path/images/photo.jpg", 2, "path", "images/photo.jpg")
	})

	t.Run("no static section", func(t *testing.T) {
		tree := New[int]()
		require.NoError(t, tree.Insert(":path...", 1))

		test(t, tree, "/", 1, "path", "/")
		test(t, tree, "/hello/world", 1, "path", "/hello/world")
	})

	t.Run("wildcard in the middle of a segment", func(t *testing.T) {
		tree := New[int]()
		require.NoError(t, tree.Insert("/prefix:path...", 1))

		test(t, tree, "/prefix42", 1, "path", "42")
		test(t, tree, "/prefixnowhere/like/this", 1, "path", "nowhere/like/this")
	})
}

// isn't used anymore. Left just in case the tree needs to be debugged.
func printTree(node *Node[int], depth int) {
	for _, p := range node.predecessors {
		fmt.Print(strings.Repeat("-", depth))

		fmt.Print(" ", strconv.Quote(p.value))

		if p.dyn != nil {
			fmt.Printf(" [%s", strconv.Quote(p.dyn.wildcard))
			if p.dyn.isLeaf {
				fmt.Printf(" {%d}", p.dyn.payload)
			}

			if p.dyn.isGreedy {
				fmt.Print(" #")
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
