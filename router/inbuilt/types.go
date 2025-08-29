package inbuilt

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/indigo-web/indigo/http"
	"github.com/indigo-web/indigo/http/method"
	"github.com/indigo-web/indigo/http/status"
	"github.com/indigo-web/indigo/internal/strutil"
	"github.com/indigo-web/indigo/router/inbuilt/internal/radix"
)

type (
	Handler       func(*http.Request) *http.Response
	errorHandlers map[status.Code]Handler
	methodLUT     [method.Count + 1]Handler
)

type endpoint struct {
	methods methodLUT
	allow   string
}

type (
	routesMap map[string]endpoint
	radixTree = *radix.Node[endpoint]
)

func (r routesMap) Add(path string, m method.Method, handler Handler) {
	p, ok := strutil.URLDecode(path)
	if !ok {
		panic(fmt.Errorf("poorly encoded path: %s", strconv.Quote(path)))
	}

	entry := r[p]
	entry.methods[m] = handler
	entry.allow = getAllowString(entry.methods)
	r[p] = entry
}

func getAllowString(methods methodLUT) (allowed string) {
	definedMethods := make([]string, 0, method.Count)

	for i, handler := range methods {
		if handler == nil {
			continue
		}

		definedMethods = append(definedMethods, method.Method(i).String())
		if method.Method(i) == method.GET && methods[method.HEAD] == nil {
			// append HEAD automatically even if they aren't explicitly supported.
			// Actually, we could do this after the loop, but then we'd have the HEAD
			// in the end, probably away from the GET, which looks suboptimal.
			// Aesthetics is always important.
			definedMethods = append(definedMethods, method.HEAD.String())
		}
	}

	return strings.Join(definedMethods, ", ")
}
