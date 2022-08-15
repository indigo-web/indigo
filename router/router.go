package router

import (
	"indigo/types"
)

type Router interface {
	OnRequest(request *types.Request, writer types.ResponseWriter) error
	OnError(err error)
}
