package radix

import (
	"context"
	"errors"
	"github.com/fakefloordiv/indigo/types"
	"github.com/fakefloordiv/indigo/valuectx"
)

var (
	ErrNotImplemented = errors.New(
		"different dynamic segment names are not allowed for common path prefix",
	)
)

type Handler func(ctx context.Context, request *types.Request) types.Response

type Node struct {
	staticSegments map[string]*Node
	isDynamic      bool
	dynamicName    string
	// Next is used only in case current node is dynamic
	next *Node

	handler Handler
}

func NewTree() *Node {
	return newNode(nil, false, "")
}

func newNode(handler Handler, isDyn bool, dynName string) *Node {
	return &Node{
		staticSegments: make(map[string]*Node),
		isDynamic:      isDyn,
		dynamicName:    dynName,
		handler:        handler,
	}
}

func (n *Node) Insert(template Template, handler Handler) error {
	return n.insertRecursively(template.segments, handler)
}

func (n *Node) insertRecursively(segments []Segment, handler Handler) error {
	if len(segments) == 0 {
		n.handler = handler

		return nil
	}

	segment := segments[0]

	if segment.IsDynamic {
		if n.isDynamic && segment.Payload != n.dynamicName {
			return ErrNotImplemented
		}

		n.isDynamic = true
		n.dynamicName = segment.Payload

		if n.next == nil {
			n.next = newNode(nil, false, "")
		}

		return n.next.insertRecursively(segments[1:], handler)
	}

	if node, found := n.staticSegments[segment.Payload]; found {
		return node.insertRecursively(segments[1:], handler)
	}

	node := newNode(nil, false, "")
	n.staticSegments[segment.Payload] = node

	return node.insertRecursively(segments[1:], handler)
}

func (n *Node) Match(ctx context.Context, path string) (context.Context, Handler) {
	if path[0] != '/' {
		// all http request paths MUST have a leading slash
		return ctx, nil
	}

	path = path[1:]

	var (
		offset int
		node   = n
	)

	for i := range path {
		if path[i] == '/' {
			var ok bool

			ctx, node, ok = processSegment(ctx, path[offset:i], node)
			if !ok {
				return ctx, nil
			}

			offset = i + 1
		}
	}

	if offset < len(path) {
		var ok bool
		ctx, node, ok = processSegment(ctx, path[offset:], node)
		if !ok {
			return ctx, nil
		}
	}

	return ctx, node.handler
}

func processSegment(ctx context.Context, segment string, node *Node) (context.Context, *Node, bool) {
	if nextNode, found := node.staticSegments[segment]; found {
		return ctx, nextNode, true
	}

	if !node.isDynamic {
		return ctx, nil, false
	}

	if len(node.dynamicName) > 0 {
		ctx = valuectx.WithValue(ctx, node.dynamicName, segment)
	}

	return ctx, node.next, true
}
