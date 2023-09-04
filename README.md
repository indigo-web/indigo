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
  
  "github.com/indigo-web/indigo"
  "github.com/indigo-web/indigo/http"
  "github.com/indigo-web/indigo/router/inbuilt"
)

const addr = "0.0.0.0:9090"

func MyHandler(request *http.Request) http.Response {
  return request.Respond().WithBody("Hello, world!")
}

func main() {
  router := inbuilt.New()
  router.Resource("/").
    Get(MyHandler).
    Post(MyHandler)

  app := indigo.NewApp(addr)
  if err := app.Serve(router); err != nil {
    log.Fatal(err)
  }
}
```

More examples in [examples/](https://github.com/indigo-web/indigo/tree/master/examples) folder.

Project workspace (TODO list included): trello.com/w/indigowebserver
