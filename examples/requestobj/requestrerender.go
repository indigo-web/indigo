package main

import (
	"fmt"
	"github.com/fakefloordiv/indigo"
	headers2 "github.com/fakefloordiv/indigo/http/headers"
	methods "github.com/fakefloordiv/indigo/http/method"
	"github.com/fakefloordiv/indigo/http/proto"
	"github.com/fakefloordiv/indigo/http/status"
	"github.com/fakefloordiv/indigo/router/inbuilt"
	"github.com/fakefloordiv/indigo/types"
	"log"
)

var addr = "localhost:9090"

func prefixIfNotEmpty(prefix, value string) string {
	if len(value) > 0 {
		return prefix + value
	}

	return value
}

func renderHeaders(headers headers2.Headers) (rendered string) {
	for key, value := range headers {
		rendered += key + ": " + string(value) + "\r\n"
	}
	return rendered
}

func ReRenderResponse(request *types.Request) types.Response {
	body, err := request.Body()
	if err != nil {
		return types.WithResponse.WithCode(status.BadRequest)
	}

	return types.WithResponse.
		WithBody(fmt.Sprintf(
			"%s %s%s%s %s\r\n%s\r\n\r\n%s",
			methods.ToString(request.Method), request.Path,
			prefixIfNotEmpty("?", string(request.Query.Raw())),
			prefixIfNotEmpty("#", request.Fragment),
			string(proto.ToBytes(request.Proto)),
			renderHeaders(request.Headers),
			string(body),
		))
}

func main() {
	r := inbuilt.NewRouter()
	r.Get("/", ReRenderResponse)

	app := indigo.NewApp(addr)
	fmt.Println("Listening on", addr)
	log.Fatal(app.Serve(r))
}
