package rmap

import (
	"github.com/indigo-web/indigo/http/method"
	"github.com/indigo-web/indigo/router/inbuilt/types"
	"strings"
)

type mapEntry struct {
	methodsMap types.MethodsMap
	allow      string
}

type Map struct {
	entries map[string]mapEntry
}

func New() *Map {
	return &Map{entries: map[string]mapEntry{}}
}

func (m *Map) Add(path string, method_ method.Method, handler types.Handler) {
	entry := m.entries[path]
	entry.methodsMap[method_] = handler
	entry.allow = getAllowString(entry.methodsMap)
	m.entries[path] = entry
}

func (m *Map) Get(path string) (methodsMap types.MethodsMap, allow string, ok bool) {
	entry, ok := m.entries[path]
	return entry.methodsMap, entry.allow, ok
}

func getAllowString(methodsMap types.MethodsMap) (allowed string) {
	for i, handler := range methodsMap {
		if handler == nil {
			continue
		}

		allowed += method.Method(i).String() + ","
	}

	return strings.TrimRight(allowed, ",")
}
