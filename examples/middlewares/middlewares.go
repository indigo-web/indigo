package main

import (
	"fmt"
	"indigo/router"
	"indigo/types"
)

func HelloWorldMiddleware(next router.HandlerFunc, request *types.Request) types.Response {
	fmt.Println("running middleware before handler")
	response := next(request)
	fmt.Println("running middleware after handler")

	return response
}

func SecondMiddleware(next router.HandlerFunc, request *types.Request) types.Response {
	fmt.Println("running second middleware before first one")
	response := next(request)
	fmt.Println("running second middleware after first one")

	return response
}
