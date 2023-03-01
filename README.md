<img src="indigo.svg" alt="This is just a logo" title="What are you looking for?"/>

Indigo is non-idiomatic, but focusing on simplicity and performance web-server

It provides such features:
- Streaming-based body processing
- Server-wide settings
- Response object constructor
- Middlewares
- Endpoint groups
- Connection hijacking

# Hello, world!

```golang
package main

import (
  "log"
  
  "github.com/indigo-web/indigo/v2"
  "github.com/indigo-web/indigo/v2/http"
  "github.com/indigo-web/indigo/v2/router/inbuilt"
)

var addr = "localhost:9090"

func MyHandler(request *http.Request) http.Response {
  return http.RespondTo(request).WithBody("Hello, world!")
}

func main() {
  router := inbuilt.NewRouter()
  router.Get("/", MyHandler)

  app := indigo.NewApp(addr)
  log.Fatal(app.Serve(router))
}
```

More examples in [examples/](https://github.com/indigo-web/indigo/tree/master/examples) folder.

Project workspace (TODO list included): trello.com/w/indigowebserver
