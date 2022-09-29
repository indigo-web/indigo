package radix

import (
	"context"
	"errors"
	"github.com/fakefloordiv/indigo/types"
	"github.com/fakefloordiv/indigo/valuectx"
	"strings"
)

var (
	ErrNotImplemented = errors.New("different dynamic names for the same static prefix is not allowed")
)

type Handler func(ctx context.Context, request *types.Request) types.Response

type Node struct {
	DynamicName string
	IsDynamic   bool
	Handler     Handler
	Branches    map[string]*Node
}

func NewTree() *Node {
	return newNode(nil, "", false)
}

func newNode(handler Handler, dynName string, isDyn bool) *Node {
	return &Node{
		DynamicName: dynName,
		IsDynamic:   isDyn,
		Handler:     handler,
		Branches:    make(map[string]*Node),
	}
}

func (n *Node) Insert(template Template, handler Handler) error {
	return n.insertRecursively(template, handler)
}

func (n *Node) insertRecursively(template Template, handler Handler) error {
	if len(template.markerNames) == 0 {
		if len(template.staticParts) == 0 {
			n.Handler = handler

			return nil
		}

		static := template.staticParts[0]

		if node, found := n.Branches[static]; found {
			node.Handler = handler

			return nil
		}

		n.Branches[static] = newNode(handler, "", false)

		return nil
	}

	static := template.staticParts[0]
	dynamic := template.markerNames[0]

	if node, found := n.Branches[static]; found {
		if node.IsDynamic {
			if node.DynamicName != dynamic {
				return ErrNotImplemented
			}

			return node.Insert(nextTemplate(template), handler)
		} else {
			node.IsDynamic = true
			node.DynamicName = dynamic

			return node.Insert(nextTemplate(template), handler)
		}
	} else {
		node = newNode(nil, dynamic, true)
		n.Branches[static] = node

		return node.Insert(nextTemplate(template), handler)
	}
}

func (n *Node) Match(ctx context.Context, path string) (context.Context, Handler) {
	var (
		currentNode = n
	)

	for len(path) > 0 {
		if currentNode.IsDynamic {
			var dynPart string
			dynPart, path = getDynPart(path)

			if len(dynPart) == 0 {
				return ctx, nil
			}

			if len(currentNode.DynamicName) > 0 {
				ctx = valuectx.WithValue(ctx, currentNode.DynamicName, dynPart)
			}

			if len(path) == 0 {
				return ctx, currentNode.Handler
			}
		}

		node, newPath, found := getBranch(path, currentNode.Branches)
		if !found || node == nil {
			return ctx, currentNode.Handler
		}

		path = newPath
		currentNode = node
	}

	if currentNode.IsDynamic {
		return ctx, nil
	}

	return ctx, currentNode.Handler
}

func getBranch(str string, branches map[string]*Node) (*Node, string, bool) {
	for prefix, node := range branches {
		if strings.HasPrefix(str, prefix) {
			return node, str[len(prefix):], true
		}
	}

	return nil, str, false
}

func getDynPart(path string) (string, string) {
	for i := range path {
		if path[i] == '/' {
			return path[:i], path[i:]
		}
	}

	return path, ""
}

func nextTemplate(template Template) Template {
	return Template{
		staticParts: template.staticParts[1:],
		markerNames: template.markerNames[1:],
	}
}
