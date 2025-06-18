package radix

import (
	"errors"
	"github.com/indigo-web/indigo/kv"
	"strings"
)

var ErrMismatchingWildcards = errors.New(
	"having two different names for wildcards sharing common prefix isn't supported",
)

type Node[T any] struct {
	isLeaf       bool
	value        string
	dyn          *dynamicNode[T]
	predecessors []*Node[T]
	payload      T
}

type dynamicNode[T any] struct {
	isLeaf   bool
	next     *Node[T]
	wildcard string
	payload  T
}

func New[T any]() *Node[T] {
	return new(Node[T])
}

func (n *Node[T]) Lookup(key string, wildcards *kv.Storage) (value T, found bool) {
	node := n

loop:
	for len(key) > 0 {
		for _, p := range node.predecessors {
			if startswith(key, p.value) {
				key = key[len(p.value):]
				node = p
				continue loop
			}
		}

		if node.dyn == nil {
			return value, false
		}

		end := strings.IndexByte(key, '/')
		if end == -1 {
			if node.dyn.isLeaf {
				addWildcard(node.dyn.wildcard, key, wildcards)
			}

			return node.dyn.payload, node.dyn.isLeaf
		}

		segment := key[:end]
		if len(segment) == 0 {
			return value, false
		}

		addWildcard(node.dyn.wildcard, segment, wildcards)
		key = key[end+1:]
		if len(key) == 0 {
			return node.dyn.payload, node.dyn.isLeaf
		}

		node = node.dyn.next
		if node == nil {
			return value, false
		}
	}

	return node.payload, node.isLeaf
}

func addWildcard(wildcard, value string, into *kv.Storage) {
	if len(wildcard) > 0 {
		into.Add(wildcard, value)
	}
}

func startswith(str, with string) bool {
	return len(str) >= len(with) && str[:len(with)] == with
}

func (n *Node[T]) Insert(key string, value T) error {
	return n.insert(splitPath(key), value)
}

func (n *Node[T]) insert(segs []pathSegment, value T) error {
	if len(segs) == 0 {
		n.isLeaf = true
		n.payload = value
		return nil
	}

	seg := segs[0]

	if seg.IsWildcard {
		if n.dyn == nil {
			n.dyn = &dynamicNode[T]{wildcard: seg.Value}
		}

		if n.dyn.wildcard != seg.Value {
			return ErrMismatchingWildcards
		}

		if len(segs) == 1 {
			n.dyn.isLeaf = true
			n.dyn.payload = value

			return nil
		} else {
			if n.dyn.next == nil {
				n.dyn.next = New[T]()
			}

			return n.dyn.next.insert(segs[1:], value)
		}
	}

	for i, p := range n.predecessors {
		common := union(p.value, seg.Value)
		if len(common) == 0 {
			continue
		}

		if len(common) == len(p.value) {
			if len(common) == len(seg.Value) {
				// p.value == seg.Value
				return p.insert(segs[1:], value)
			}

			// len(seg.Value) > len(p.value)
			seg.Value = seg.Value[len(common):]
			segs[0] = seg

			return p.insert(segs, value)
		}

		// len(common) < len(p.value)
		// len(common) <= len(seg.Value)

		stays, goes := p.value[:len(common)], p.value[len(common):]
		substitution := &Node[T]{value: stays}
		p.value = goes
		substitution.predecessors = append(substitution.predecessors, p)
		n.predecessors[i] = substitution

		if len(common) == len(seg.Value) {
			substitution.isLeaf = true
			substitution.payload = value
			return nil
		}

		// len(common) < len(seg.Value)
		newNode := &Node[T]{value: seg.Value[len(common):]}
		substitution.predecessors = append(substitution.predecessors, newNode)
		return newNode.insert(segs[1:], value)
	}

	newNode := &Node[T]{value: seg.Value}
	n.predecessors = append(n.predecessors, newNode)

	return newNode.insert(segs[1:], value)
}

func IsDynamicTemplate(str string) bool {
	return strings.IndexByte(str, ':') != -1
}

func union(a, b string) string {
	for i := 0; i < min(len(a), len(b)); i++ {
		if a[i] != b[i] {
			return a[:i]
		}
	}

	return a[:min(len(a), len(b))]
}

type pathSegment struct {
	IsWildcard bool
	Value      string
}

func splitPath(str string) (result []pathSegment) {
	for len(str) > 0 {
		colon := strings.IndexByte(str, ':')
		if colon == -1 {
			result = append(result, pathSegment{false, str})
			break
		}

		result = append(result, pathSegment{false, str[:colon]})
		str = str[colon+1:]

		boundary := strings.IndexByte(str, '/')
		if boundary == -1 {
			result = append(result, pathSegment{true, str})
			break
		}

		result = append(result, pathSegment{true, str[:boundary]})
		str = str[boundary+1:]
	}

	return result
}
