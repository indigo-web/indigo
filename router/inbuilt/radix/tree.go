package radix

import (
	"context"
	"errors"
	"github.com/fakefloordiv/indigo/valuectx"
)

var (
	ErrNotImplemented = errors.New(
		"different dynamic segment names are not allowed for common path prefix",
	)
)

type Tree[V any] interface {
	Insert(Template, V) error
	MustInsert(Template, V)
	Match(context.Context, string) (context.Context, V)
}

type Node[V any] struct {
	staticSegments map[string]*Node[V]
	isDynamic      bool
	dynamicName    string
	// Next is used only in case current node is dynamic
	next *Node[V]

	payload V
}

func NewTree[V any]() Tree[V] {
	return newNode[V](nil, false, "")
}

func newNode[V any](payload V, isDyn bool, dynName string) *Node[V] {
	return &Node[V]{
		staticSegments: make(map[string]*Node[V]),
		isDynamic:      isDyn,
		dynamicName:    dynName,
		payload:        payload,
	}
}

func (n *Node[V]) Insert(template Template, payload V) error {
	return n.insertRecursively(template.segments, payload)
}

func (n *Node[V]) MustInsert(template Template, payload V) {
	if err := n.Insert(template, payload); err != nil {
		panic(err.Error())
	}
}

func (n *Node[V]) insertRecursively(segments []Segment, payload V) error {
	if len(segments) == 0 {
		n.payload = payload

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
			n.next = newNode[V](nil, false, "")
		}

		return n.next.insertRecursively(segments[1:], payload)
	}

	if node, found := n.staticSegments[segment.Payload]; found {
		return node.insertRecursively(segments[1:], payload)
	}

	node := newNode[V](nil, false, "")
	n.staticSegments[segment.Payload] = node

	return node.insertRecursively(segments[1:], payload)
}

func (n *Node[V]) Match(ctx context.Context, path string) (context.Context, V) {
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

	return ctx, node.payload
}

func processSegment[V any](
	ctx context.Context, segment string, node *Node[V],
) (context.Context, *Node[V], bool) {

	if nextNode, found := node.staticSegments[segment]; found {
		return ctx, nextNode, true
	}

	if !node.isDynamic || len(segment) == 0 {
		return ctx, nil, false
	}

	if len(node.dynamicName) > 0 {
		ctx = valuectx.WithValue(ctx, node.dynamicName, segment)
	}

	return ctx, node.next, true
}
