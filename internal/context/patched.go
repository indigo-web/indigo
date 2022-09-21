package context

import (
	"context"
	"fmt"
)

// Copyright 2014 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// THIS PACKAGE IS SIMPLY COPY-PASTED WithValue FROM context.go
// With the only difference - using directly comparable-generic
// instead of any and reflection. This increases performance a
// lot

// WithValue returns a copy of parent in which the value associated with key is
// val.
//
// Use context Values only for request-scoped data that transits processes and
// APIs, not for passing optional parameters to functions.
//
// The provided key must be comparable and should not be of type
// string or any other built-in type to avoid collisions between
// packages using context. Users of WithValue should define their own
// types for keys. To avoid allocating when assigning to an
// interface{}, context keys often have concrete type
// struct{}. Alternatively, exported context key variables' static
// type should be a pointer or interface.
func WithValue[K comparable, V any](parent context.Context, key K, val V) context.Context {
	if parent == nil {
		panic("cannot create context from nil parent")
	}

	return valueCtx[K, V]{parent, key, val}
}

// A valueCtx carries a key-value pair. It implements Value for that key and
// delegates all other calls to the embedded Context.
type valueCtx[K comparable, V any] struct {
	context.Context
	key K
	val V
}

func (c valueCtx[K, V]) String() string {
	return fmt.Sprintf(
		"github.com/fakefloordiv/indigo/internal/context:valueCtx{key: %s, value: %s}\n",
		fmt.Sprint(c.key), fmt.Sprint(c.val),
	)
}

func (c valueCtx[K, V]) Value(key any) any {
	if c.key == key {
		return c.val
	}
	return c.Context.Value(key)
}
