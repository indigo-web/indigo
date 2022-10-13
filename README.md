<img src="logo.png" alt="drawing" width="300" align="top" title="What are you looking for?"/>

Indigo is non-idiomatic, but focusing on simplicity and performance web-server

It provides such a features:
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
  "context"
  "log"
  
  "github.com/fakefloordiv/indigo"
  "github.com/fakefloordiv/indigo/router/inbuilt"
  "github.com/fakefloordiv/indigo/types"
)

var addr = "localhost:9090"

func MyHandler(_ context.Context, req *types.Request) types.Response {
  return types.WithBody("Hello, world!")
}

func main() {
  router := inbuilt.NewRouter()
  router.Get("/", MyHandler)

  app := indigo.NewApp(addr)
  log.Fatal(app.Serve(router))
}
```

More examples in [examples/](https://github.com/fakefloordiv/indigo/tree/master/examples) folder.


Project workspace (TODO list included): trello.com/w/indigowebserver
