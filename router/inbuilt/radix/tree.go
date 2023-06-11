package radix

import (
	"errors"

	"github.com/indigo-web/indigo/router/inbuilt/types"
)

var ErrNotImplemented = errors.New(
	"different dynamic segment names are not allowed for common path prefix",
)

type Params map[string]string

type Payload struct {
	MethodsMap types.MethodsMap
	Allow      string
}

type Tree interface {
	Insert(Template, Payload) error
	MustInsert(Template, Payload)
	Match(Params, string) *Payload
}

type Node struct {
	staticSegments map[string]*Node
	next           *Node
	payload        *Payload
	dynamicName    string
	isDynamic      bool
}

func NewTree() Tree {
	return newNode(new(Payload), false, "")
}

func newNode(payload *Payload, isDyn bool, dynName string) *Node {
	return &Node{
		staticSegments: make(map[string]*Node),
		isDynamic:      isDyn,
		dynamicName:    dynName,
		payload:        payload,
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

	if segment.IsDynamic {
		if n.isDynamic && segment.Payload != n.dynamicName {
			return ErrNotImplemented
		}

		n.isDynamic = true
		n.dynamicName = segment.Payload

		if n.next == nil {
			n.next = newNode(nil, false, "")
		}

		return n.next.insertRecursively(segments[1:], payload)
	}

	if node, found := n.staticSegments[segment.Payload]; found {
		return node.insertRecursively(segments[1:], payload)
	}

	node := newNode(nil, false, "")
	n.staticSegments[segment.Payload] = node

	return node.insertRecursively(segments[1:], payload)
}

func (n *Node) Match(params Params, path string) *Payload {
	if path[0] != '/' {
		// all http request paths MUST have a leading slash
		return nil
	}

	path = path[1:]

	var (
		offset int
		node   = n
	)

	for i := range path {
		if path[i] == '/' {
			var ok bool
			node, ok = processSegment(params, path[offset:i], node)
			if !ok {
				return nil
			}

			offset = i + 1
		}
	}

	if offset < len(path) {
		var ok bool
		node, ok = processSegment(params, path[offset:], node)
		if !ok {
			return nil
		}
	}

	return node.payload
}

func processSegment(params Params, segment string, node *Node) (*Node, bool) {
	if nextNode, found := node.staticSegments[segment]; found {
		return nextNode, true
	}

	if !node.isDynamic || len(segment) == 0 {
		return nil, false
	}

	if len(node.dynamicName) > 0 {
		params[node.dynamicName] = segment
	}

	return node.next, true
}
