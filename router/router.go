package router

import (
	"indigo/types"
)

/*
Router is an interface for every router may be used for indigo
Actually, it's incompatible with http.Handler, so in case
once I'll need to connect something like chi, I'll have to write
a port for it. But looks like there won't be serious problems with that

In case error is returned from OnRequest, OnError will be called with
error was returned from OnRequest. So no need to call it internally
manually, just trust me
*/
type Router interface {
	OnRequest(request *types.Request, writeResponse types.ResponseWriter) error
	OnError(err error)
}
