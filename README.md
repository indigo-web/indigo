# indigo

<img align="right" width="159" alt="image" src="https://gist.githubusercontent.com/flrdv/e610c73096af43b11ba3db5ec52b4194/raw/eb1b50202c8d4c7ce4709dd0af997df03dfee63d/indigo-logo-mini.svg" />

Elegance, conciseness, flexibility and extensibility — that’s what Indigo is about. The goal is to create a mini ecosystem in which you can focus on actual tasks rather than reimplementing the same things over and over again. Everything you need to get started is already included, and whatever isn’t is trivially pluggable.

- FastHTTP-grade performance
- Stream-based body processing
- Fine-tuning of parameters like memory consumption and timeouts
- A full-fledged built-in router
- Method chaining at its finest

# Documentation

Documentation is available [here](https://floordiv.gitbook.io/indigo/). It might be incomplete however, feel free to open issues.

# Hello, world!

```golang
package main

import (
	"log"

	"github.com/indigo-web/indigo"
	"github.com/indigo-web/indigo/http"
	"github.com/indigo-web/indigo/router/inbuilt"
)

func HelloWorld(request *http.Request) *http.Response {
	return http.String(request, "Hello, world!")
}

func Log(request *http.Request) *http.Response {
	text, err := request.Body.String()
	if err != nil {
		return http.Error(request, err)
	}

	log.Printf("%s says: %s", request.Remote, text)
	return http.String(request, text)
}

func main() {
	r := inbuilt.New()
	r.Resource("/").
		Get(HelloWorld).
		Post(Log)

	err := indigo.New(":8080").Serve(r)
	if err != nil {
		log.Fatal(err)
	}
}

```

You can find more examples in [examples/](https://github.com/indigo-web/indigo/tree/master/examples).
