package internal

import "github.com/indigo-web/indigo/http"

// Mutator is kind of pre-middleware. It's being called at the moment, when a request arrives
// to the router, but before the routing will be done. So by that, the request may be mutated.
// For example, mutator may normalize requests' paths, log them, transparently redirect, etc.
type Mutator func(request *http.Request)
