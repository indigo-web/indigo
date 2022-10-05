package main

import (
	"context"
	"fmt"
	"log"

	"github.com/fakefloordiv/indigo"
	"github.com/fakefloordiv/indigo/router/inbuilt"
	"github.com/fakefloordiv/indigo/types"
)

var addr = "localhost:9090"

func MyDynamicHandler(ctx context.Context, _ *types.Request) types.Response {
	worldName := ctx.Value("world-name").(string)

	return types.WithBody("your world-name is " + worldName)
}

func main() {
	r := inbuilt.NewRouter()

	r.Get("/hello/{world-name}", MyDynamicHandler)

	fmt.Println("Listening on", addr)
	app := indigo.NewApp(addr)
	log.Fatal(app.Serve(r))
}
