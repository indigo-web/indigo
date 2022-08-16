package router

import (
	"indigo/types"
)

type Router interface {
	OnRequest(request *types.Request, writer types.ResponseWriter) error
	OnError(request *types.Request, writer types.ResponseWriter, err error)
}
