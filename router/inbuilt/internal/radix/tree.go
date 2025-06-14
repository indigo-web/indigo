package radix

import (
	"errors"
	"github.com/indigo-web/indigo/kv"
	"github.com/indigo-web/indigo/router/inbuilt/internal/types"
	"strings"
)

var ErrNotImplemented = errors.New(
	"different dynamic segment names are not allowed for common path prefix",
)

type Params = *kv.Storage

type Payload struct {
	MethodsMap types.MethodsMap
	Allow      string
}

type Tree = *Node

type Node struct {
	statics    arrMap
	next       *Node
	payload    *Payload
	wildcard   string
	isWildcard bool
}

func New() *Node {
	return newNode(new(Payload), false, "")
}

func newNode(payload *Payload, isDyn bool, dynName string) *Node {
	return &Node{
		isWildcard: isDyn,
		wildcard:   dynName,
		payload:    payload,
	}
}

func (n *Node) Insert(template Template, payload Payload) error {
	return n.insertRecursively(template.segments, &payload)
}

func (n *Node) MustInsert(template Template, payload Payload) {
	if err := n.Insert(template, payload); err != nil {
		panic(err.Error())
	}
}

func (n *Node) insertRecursively(segments []Segment, payload *Payload) error {
	if len(segments) == 0 {
		n.payload = payload

		return nil
	}

	segment := segments[0]

	if segment.IsWildcard {
		if n.isWildcard && segment.Payload != n.wildcard {
			return ErrNotImplemented
		}

		n.isWildcard = true
		n.wildcard = segment.Payload

		if n.next == nil {
			n.next = newNode(nil, false, "")
		}

		return n.next.insertRecursively(segments[1:], payload)
	}

	if node := n.statics.Lookup(segment.Payload); node != nil {
		return node.insertRecursively(segments[1:], payload)
	}

	node := newNode(nil, false, "")
	n.statics.Add(segment.Payload, node)

	return node.insertRecursively(segments[1:], payload)
}

func (n *Node) Match(path string, params Params) *Payload {
	if path[0] != '/' {
		// all http request paths MUST have a leading slash
		return nil
	}

	path = path[1:]
	node := n

	for len(path) > 0 {
		slash := strings.IndexByte(path, '/')
		var segment string
		if slash == -1 {
			segment, path = path, ""
		} else {
			segment, path = path[:slash], path[slash+1:]
		}

		next, ok := processSegment(params, segment, node)
		if !ok {
			return nil
		}

		node = next
	}

	return node.payload
}

func processSegment(params Params, segment string, node *Node) (*Node, bool) {
	// manually inlined arrMap.Lookup(segment)
	if !node.statics.arrOverflow {
		for _, entry := range node.statics.arr {
			if entry.Key == segment {
				return entry.Node, true
			}
		}
	} else if n := node.statics.m[segment]; n != nil {
		return n, true
	}

	if !node.isWildcard || len(segment) == 0 {
		return nil, false
	}

	if len(node.wildcard) > 0 {
		params.Add(node.wildcard, segment)
	}

	return node.next, true
}
