package inbuilt

import (
	"fmt"
	"github.com/indigo-web/indigo/http/method"
	"github.com/indigo-web/indigo/router/inbuilt/internal/radix"
	"github.com/indigo-web/indigo/router/inbuilt/uri"
	"strings"
)

type registrar struct {
	endpoints   map[string]map[method.Method]Handler
	usedMethods [method.Count]bool
	isDynamic   bool
}

func newRegistrar() *registrar {
	return &registrar{
		endpoints: make(map[string]map[method.Method]Handler),
	}
}

func (r *registrar) Add(path string, m method.Method, handler Handler) error {
	if len(path) == 0 {
		return fmt.Errorf("empty path")
	}

	// TODO: support urlencoded characters in endpoints.
	path = uri.Normalize(path)
	methodsMap := r.endpoints[path]
	if methodsMap == nil {
		methodsMap = make(map[method.Method]Handler)
	}

	if _, ok := methodsMap[m]; ok {
		return fmt.Errorf("%s %s: already registered", m, path)
	}

	methodsMap[m] = handler
	r.endpoints[path] = methodsMap
	r.isDynamic = r.isDynamic || radix.IsDynamicTemplate(path)

	return nil
}

func (r *registrar) Merge(another *registrar) error {
	r.isDynamic = r.isDynamic || another.isDynamic

	for path, v := range another.endpoints {
		for method_, handler := range v {
			if r.endpoints[path] == nil {
				r.endpoints[path] = make(map[method.Method]Handler)
			}

			if r.endpoints[path][method_] != nil {
				return fmt.Errorf("route already registered: %s %s", method_.String(), path)
			}

			r.endpoints[path][method_] = handler
		}
	}

	return nil
}

func (r *registrar) Apply(f func(handler Handler) Handler) {
	for _, v := range r.endpoints {
		for key, handler := range v {
			v[key] = f(handler)
		}
	}
}

func (r *registrar) IsDynamic() bool {
	return r.isDynamic
}

func (r *registrar) AsMap() routesMap {
	rmap := make(routesMap, len(r.endpoints))

	for path, v := range r.endpoints {
		for method_, handler := range v {
			rmap.Add(path, method_, handler)
		}
	}

	return rmap
}

func (r *registrar) AsRadixTree() radixTree {
	tree := radix.New[endpoint]()

	for path, e := range r.endpoints {
		if len(path) == 0 {
			panic("empty path")
		}

		var (
			mlut  methodLUT
			allow string
		)

		for m, handler := range e {
			mlut[m] = handler
			allow += m.String() + ","
		}

		if err := tree.Insert(path, endpoint{
			methods: mlut,
			allow:   strings.TrimSuffix(allow, ","),
		}); err != nil {
			panic(err)
		}
	}

	return tree
}
