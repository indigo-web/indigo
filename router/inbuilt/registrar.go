package inbuilt

import (
	"fmt"
	"github.com/indigo-web/indigo/http/method"
	radix2 "github.com/indigo-web/indigo/router/inbuilt/internal/radix"
	"github.com/indigo-web/indigo/router/inbuilt/internal/types"
	"strings"
)

type registrar struct {
	routes    map[string]map[method.Method]Handler
	isDynamic bool
}

func newRegistrar() *registrar {
	return &registrar{
		routes: make(map[string]map[method.Method]Handler),
	}
}

func (r *registrar) Add(path string, m method.Method, handler Handler) error {
	path = stripTrailingSlash(path)
	methodsMap := r.routes[path]
	if methodsMap == nil {
		methodsMap = make(map[method.Method]Handler)
	}

	if _, ok := methodsMap[m]; ok {
		return fmt.Errorf("route already registered: %s %s", method.ToString(m), path)
	}

	methodsMap[m] = handler
	r.routes[path] = methodsMap
	r.isDynamic = r.isDynamic || !radix2.MustParse(path).IsStatic()

	return nil
}

func (r *registrar) Merge(another *registrar) error {
	r.isDynamic = r.isDynamic || another.isDynamic

	for path, v := range another.routes {
		for method_, handler := range v {
			if r.routes[path] == nil {
				r.routes[path] = make(map[method.Method]Handler)
			}

			if r.routes[path][method_] != nil {
				return fmt.Errorf("route already registered: %s %s", method_.String(), path)
			}

			r.routes[path][method_] = handler
		}
	}

	return nil
}

func (r *registrar) Apply(f func(handler Handler) Handler) {
	for _, v := range r.routes {
		for key, handler := range v {
			v[key] = f(handler)
		}
	}
}

func (r *registrar) IsDynamic() bool {
	return r.isDynamic
}

func (r *registrar) AsMap() routesMap {
	rmap := make(routesMap)

	for path, v := range r.routes {
		for method_, handler := range v {
			rmap.Add(path, method_, handler)
		}
	}

	return rmap
}

func (r *registrar) AsRadixTree() radix2.Tree {
	tree := radix2.NewTree()

	for path, v := range r.routes {
		var (
			methodsMap types.MethodsMap
			allow      string
		)

		for method_, handler := range v {
			methodsMap[method_] = handler
			allow += method.ToString(method_) + ","
		}

		tree.MustInsert(radix2.MustParse(path), radix2.Payload{
			MethodsMap: methodsMap,
			Allow:      strings.TrimSuffix(allow, ","),
		})
	}

	return tree
}
