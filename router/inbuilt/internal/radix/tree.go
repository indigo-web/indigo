package radix

import (
	"errors"
	"github.com/indigo-web/indigo/kv"
	"strings"
)

var (
	ErrMismatchingWildcards = errors.New(
		"having two different names for wildcards sharing common prefix isn't supported",
	)
	ErrBadGreedyWildcardPosition = errors.New(
		"the greedy wildcard must always stand the last",
	)
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
	isGreedy bool
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
			if strings.HasPrefix(key, p.value) {
				key = key[len(p.value):]
				node = p
				continue loop
			}
		}

		if node.dyn == nil {
			return value, false
		}

		if node.dyn.isGreedy {
			// as greedy wildcard must always stand the last, it is therefore always a leaf
			addWildcard(node.dyn.wildcard, key, wildcards)

			return node.dyn.payload, true
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

	if node.dyn != nil && node.dyn.isGreedy {
		addWildcard(node.dyn.wildcard, "", wildcards)
		return node.dyn.payload, true
	}

	return node.payload, node.isLeaf
}

func addWildcard(wildcard, value string, into *kv.Storage) {
	if len(wildcard) > 0 {
		into.Add(wildcard, value)
	}
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
			n.dyn.isGreedy = seg.IsGreedy
			n.dyn.payload = value

			return nil
		} else {
			if seg.IsGreedy {
				return ErrBadGreedyWildcardPosition
			}

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
			return p.insert(truncCommon(segs, len(common)), value)
		}

		// len(common) < len(p.value)
		// len(common) <= len(seg.Value)

		stays, goes := p.value[:len(common)], p.value[len(common):]
		substitution := &Node[T]{value: stays}
		p.value = goes
		substitution.predecessors = append(substitution.predecessors, p)
		n.predecessors[i] = substitution

		return substitution.insert(truncCommon(segs, len(common)), value)
	}

	newNode := &Node[T]{value: seg.Value}
	n.predecessors = append(n.predecessors, newNode)

	return newNode.insert(segs[1:], value)
}

func IsDynamicTemplate(str string) bool {
	return strings.IndexByte(str, ':') != -1
}

func truncCommon(segs []pathSegment, length int) []pathSegment {
	segs[0].Value = segs[0].Value[length:]
	if len(segs[0].Value) == 0 {
		segs = segs[1:]
	}
	return segs
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
	IsGreedy   bool
	Value      string
}

func splitPath(str string) (result []pathSegment) {
	for len(str) > 0 {
		colon := strings.IndexByte(str, ':')
		if colon == -1 {
			result = append(result, pathSegment{false, false, str})
			break
		}

		result = append(result, pathSegment{false, false, str[:colon]})
		str = str[colon+1:]

		boundary := strings.IndexByte(str, '/')
		if boundary == -1 {
			boundary = len(str)
		}

		wildcard := str[:boundary]
		greedy := false
		if strings.HasSuffix(wildcard, "...") {
			wildcard = wildcard[:len(wildcard)-3]
			greedy = true
		}

		result = append(result, pathSegment{true, greedy, wildcard})
		if boundary < len(str) {
			boundary++
		}

		str = str[boundary:]
	}

	return result
}
