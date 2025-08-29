package inbuilt

import (
	"fmt"
	"strings"

	"github.com/indigo-web/indigo/http/method"
	"github.com/indigo-web/indigo/router/inbuilt/internal/radix"
	"github.com/indigo-web/indigo/router/inbuilt/uri"
)

type registrar struct {
	endpoints map[string]map[method.Method]Handler
	isDynamic bool
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

	path = uri.Normalize(path)
	methodsMap := r.endpoints[path]
	if methodsMap == nil {
		methodsMap = make(map[method.Method]Handler)
	}

	if _, ok := methodsMap[m]; ok {
		return fmt.Errorf("duplicate endpoint: %s %s", m, path)
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
				return fmt.Errorf("duplicate endpoint: %s %s", method_, path)
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

func (r *registrar) Options(includeTRACE bool) string {
	var (
		totalEndpoints   int
		methodsStatistic [method.Count + 1]int
	)

	for _, ep := range r.endpoints {
		totalEndpoints++

		for m := range ep {
			methodsStatistic[m]++
		}
	}

	if totalEndpoints == 0 {
		// a server with no endpoints at all. Must be rare enough, I guess.
		return ""
	}

	// As OPTIONS is supported, it must appear unconditionally
	methodsStatistic[method.OPTIONS] = totalEndpoints

	if includeTRACE {
		methodsStatistic[method.TRACE] = totalEndpoints
	}

	if methodsStatistic[method.GET] == totalEndpoints {
		// HEAD requests must also be unconditionally enabled, if GET
		// are also supported, as HEAD are automatically redirected to
		// GET handlers if weren't explicitly redefined.
		methodsStatistic[method.HEAD] = totalEndpoints
	}

	options := make([]string, 0, method.Count)

	for m, usage := range methodsStatistic {
		if usage == totalEndpoints {
			// EACH endpoint supports this method
			options = append(options, method.Method(m).String())
		}
	}

	return strings.Join(options, ", ")
}
