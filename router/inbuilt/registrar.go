package inbuilt

import (
	"fmt"
	"github.com/indigo-web/indigo/http/method"
	"github.com/indigo-web/indigo/router/inbuilt/radix"
	"github.com/indigo-web/indigo/router/inbuilt/rmap"
	"github.com/indigo-web/indigo/router/inbuilt/types"
	"strings"
)

type registrar struct {
	routes    map[string]map[method.Method]types.Handler
	isDynamic bool
}

func newRegistrar() *registrar {
	return &registrar{
		routes: make(map[string]map[method.Method]types.Handler),
	}
}

func (r *registrar) Add(path string, m method.Method, handler types.Handler) error {
	path = stripTrailingSlash(path)
	methodsMap := r.routes[path]
	if methodsMap == nil {
		methodsMap = make(map[method.Method]types.Handler)
	}

	if _, ok := methodsMap[m]; ok {
		return fmt.Errorf("route already registered: %s %s", method.ToString(m), path)
	}

	methodsMap[m] = handler
	r.routes[path] = methodsMap
	r.isDynamic = r.isDynamic || !radix.MustParse(path).IsStatic()

	return nil
}

func (r *registrar) Merge(another *registrar) error {
	r.isDynamic = r.isDynamic || another.isDynamic

	for path, v := range another.routes {
		for method_, handler := range v {
			if r.routes[path] == nil {
				r.routes[path] = make(map[method.Method]types.Handler)
			}

			if r.routes[path][method_] != nil {
				return fmt.Errorf("route already registered: %s %s", method.ToString(method_), path)
			}

			r.routes[path][method_] = handler
		}
	}

	return nil
}

func (r *registrar) Apply(f func(handler types.Handler) types.Handler) {
	for _, v := range r.routes {
		for key, handler := range v {
			v[key] = f(handler)
		}
	}
}

func (r *registrar) IsDynamic() bool {
	return r.isDynamic
}

func (r *registrar) AsRMap() *rmap.Map {
	routesMap := rmap.New()

	for path, v := range r.routes {
		for method_, handler := range v {
			routesMap.Add(path, method_, handler)
		}
	}

	return routesMap
}

func (r *registrar) AsRadixTree() radix.Tree {
	tree := radix.NewTree()

	for path, v := range r.routes {
		var (
			methodsMap types.MethodsMap
			allow      string
		)

		for method_, handler := range v {
			methodsMap[method_] = handler
			allow += method.ToString(method_) + ","
		}

		tree.MustInsert(radix.MustParse(path), radix.Payload{
			MethodsMap: methodsMap,
			Allow:      strings.TrimSuffix(allow, ","),
		})
	}

	return tree
}
