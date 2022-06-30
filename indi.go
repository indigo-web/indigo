package main

import (
	"fmt"
	"indigo/httpparser"
	"indigo/httpserver"
	"indigo/types"
	"net"
)

type parser struct {
	datas int
}

func (p *parser) Parse(requestStruct *types.Request, data []byte) (done bool, extra []byte, err error) {
	fmt.Println("[PARSER] having data:", string(data))
	p.datas++

	if p.datas%2 == 0 {
		fmt.Println("[PARSER] request is done (returning something back)")
		return true, data[len(data)/2:], nil
	}

	return false, nil, nil
}

type router struct {
}

func (r router) OnRequest(request *types.Request, writeResponse types.ResponseWriter) (err error) {
	fmt.Printf("[ROUTER-OnRequest] request=%v, writeResponse=%p\n", request, writeResponse)
	fmt.Printf("[ROUTER-OnRequest] method=%d path=%s proto=%d headers=%s Easter=%s\n", request.Method, string(request.Path),
		request.Protocol, request.Headers, string(request.Headers["Easter"]))

	fmt.Println("[ROUTER-OnRequest] Writing response: Hello, world!")
	if err = writeResponse([]byte("Hello, world!")); err != nil {
		fmt.Println("[ROUTER-OnRequest] Failed to write a response:", err)
	}

	return err
}

func (r router) OnError(err error) {
	fmt.Println("[ROUTER-OnError] got error:", err)
}

func main() {
	sock, err := net.Listen("tcp", "localhost:5000")
	defer sock.Close()

	if err != nil {
		panic(err)
	}

	myRouter := router{}

	fmt.Println("Listening on localhost:5000")

	exitCode := httpserver.StartTCPServer(sock, func(conn net.Conn) {
		request, writeBody := types.NewRequest(make([]byte, 10), make(map[string][]byte), nil)
		parser := httpparser.NewHTTPParser(&request, writeBody, httpparser.Settings{})

		handler := httpserver.NewHTTPHandler(httpserver.HTTPHandlerArgs{
			Router:           myRouter,
			Request:          &request,
			WriteRequestBody: writeBody,
			Parser:           parser,
			RespWriter: func(b []byte) error {
				_, err = conn.Write(b)
				return err
			},
		})

		go handler.Poll()
		httpserver.DefaultConnHandler(conn, handler.OnData)
	})

	fmt.Println("exited with error", exitCode)
}
