package main

import (
	"fmt"
	"indigo"
	"indigo/http"
	"indigo/router"
	"indigo/types"
	"log"
)

func MyPostHandler(request *types.Request) types.Response {
	body, err := request.GetFullBody()

	if err != nil {
		return types.NewResponse().
			WithCode(http.StatusBadRequest).
			WithBody("bad request")
	}

	fmt.Println("somebody said:", string(body))

	return types.NewResponse().WithBody("Received and processed")
}

func main() {
	myRouter := router.NewDefaultRouter()
	myRouter.Route("/tell", MyPostHandler)

	fmt.Println("Listening on localhost:9090")
	app := indigo.NewApp("localhost", 9090)
	log.Fatal(app.Serve(myRouter))
}
