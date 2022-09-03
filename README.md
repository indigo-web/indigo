<img src="logo.png" alt="drawing" width="300" align="top" title="What are you looking for?"/>

Indigo is non-idiomatic, but focusing on simplicity and performance web-server

It provides such a features:
- Streaming-based body processing
- Server-wide settings
- Response object constructor
- Middlewares
- Endpoint groups
- Connection hijacking

See examples in `examples/` folder

# Documentation
## Routers
### `indigo/router/inbuilt`
Default router that supports:
- Groups
- Middlewares
- Error handlers

#### Router interface
Instantiating with `inbuilt.NewRouter()`. Includes following methods:
- `Route(path string, method methods.Method, handler HandlerFunc, mwares ...Middleware)`
  - Basic method for registering method. Actually should not be used, preferring method predicates instead
  - Passed middlewares are called "point applied", and they are applied only to the handler
- `<Method>(path string, handler HandlerFunc, mwares ...Middleware)`
  - Where Method is a method name in camelcase
  - Is just a shorthand to `Route(path, methods.METHOD, handler, mwares...)`
- `Group(prefix string) Router`
  - Instantiates a new separated router, routes of which will be merged into the root router later
  - Prefix works simply as a prefix (wow). Each route will have a prefix concatenated to its path
- `Use(middleware Middleware)`
  - Applies a group-wide middleware. It will be applied to every registered handler (even if it was registered before calling this method)
  - Each group inherited from current one will also inherit all the local middlewares
    - Local middleware is a group-wide middleware
    - Global middleware is a middleware applied in the root router
  - All the child groups (in case they already exist) will not inherit newly applied middlewares. So they must be applied before inheriting
- `SetDefaultHeaders(headers headers.Headers)`
  - Set headers by default. Headers by default are headers that will be added to response headers in case user has not specified them by himself
  - If not specified, using next default headers:
    - Server: indigo
    - Connection: keep-alive
    - Accept-Encoding: \<list of accepted encodings separated by comma>

#### Handling HTTP requests
To handle HTTP request, you need to register a corresponding handler with corresponding method and
path. Handler is simply a function that takes a pointer to `types.Request` struct, and returns `types.Response`
struct. 
`types.Request` includes following fields:
- `Method methods.Method`
  - Request method
- `Path string`
  - Request path
- `Query url.Query`
  - Path query
  - Is a structure with following methods:
    - `Get(key string) (value []byte, err error)`
      - In case called for a first time, path query will be parsed because it will not be parsed until user will need it
      - In case key is not found, `ErrNoSuchKey` will be returned
    - `Raw() []byte`
      - Returns a raw value of query
- `Fragment string`
  - Path fragment; usually should be avoided because is used mostly by browsers
- `Proto proto.Proto`
  - Request protocol; may be one of the following values:
    - `HTTP09`
    - `HTTP10`
    - `HTTP11`
- `Headers headers.Headers`
  - Just a `map[string][]byte`
- `OnBody(onBody onBodyCb, onComplete onCompleteCb) error`
  - `onBodyCb` is simply `func([]byte) error`
    - In case returned error is not nil, `onComplete` will be called with it and all the same error will be returned from method
  - `onCompleteCb` is simply `func(error)`
- `Body() ([]byte, error)`
  - Is just a wrapper for `OnBody`, but saving incoming request body into some local buffer that is kept across requests, so it will be kept until client will disconnect
- `Hijack`
  - Not actually a method, but an attribute, that takes no arguments, but returns `net.Conn` and `error` (if occurred)
  - Calling this method automatically erases body (waits until will be completely received, then reading into nowhere, if was not read by user before), and returns error if occurred while reading body
  - After hijacking, connection must be utilized by user, because this process is permanent (for example, server core is shutting down, but without closing the connection)

`types.Response` includes following fields:
- `WithCode(code status.Code) Response`
  - Set response code, returns new `Response` object with updated `Code` field
- `WithStatus(status status.Status) Response`
  - Set a custom status. Not recommended to use because `WithCode` automatically sets a corresponding status
- `WithHeader(key, value string) Response`
  - Set a header and receive a new `Response` object
    - Actually, this is the only method that is not clear. It uses an underlying `headers.Headers` map, so this call modifies the map and is absolutely not compulsory to save a returned value, it is just _trying_ to behave like a clear method
- `WithHeaders(headers map[string]string) Response`
  - Does the same as `WithHeader` does, but headers are in map now
- `Headers() headers.Headers`
  - Returns response headers
- `WithBodyByte(body []byte) Response`
  - Sets a response body. Sets - not appends
- `WithBody(body string) Response`
  - Does all the same as `WithBodyByte` does, but takes a string as an argument
- `WithFile(path string, errHandler FileErrHandler) Response`
  - In case called, request body will be ignored
  - Responds to client with a body that is a content of file
    - Current implementation does not use chunked encoding or smth. It simply takes a size of the file, sets Content-Length value equal to it, and just writes file content into the socket by 64kb blocks
  - If file not found, `errHandler` (that is simply `func(err error) types.Response`) will be called
    - Yes, it is a fundamental design problem - user cannot receive an error from the outside. But the way net/http does (or smth), I don't like at all. So looking for a balance
- `Code status.Code`
  - Attribute with response code. Default value is 200 (if instantiating using `NewResponse` or using `types.WithResponse`)
- `Status status.Status`
  - Attribute with response status. Default value is a corresponding status to 200
- `Body []byte`
  - Attribute with response body. Can be used to modify response body by middlewares

Example:
```golang
return types.WithResponse.
  WithCode(status.OK).
  WithHeader("Hello", "World!").
  WithBody("How are you doing, fellow kids?")
```

Note: you can `return types.WithResponse` from handler. This will return a simple 200 OK response

### `indigo/router/simple`
Simple router with no routing at all. All it does is just encapsulates router-specific structure.
It ignores all the errors (the only thing it does is writing 400 Bad Request back), in case of
new request calls a single handler that is specified from `NewRouter()` call. Btw handler is still
a handler from inbuilt, so it takes a request and returns a response builder

The only use case I found is file server
