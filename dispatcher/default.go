package dispatcher

import (
	"indigo/webserver"
)

type Dispatcher struct {
}

func (d Dispatcher) ProcessRequest(client webserver.Client, completed webserver.RequestCompleted) {
	// TODO: there is must be a table with handlers, from which one we'll take a handler and call it
	//       currently this is just a dummy, so fuck it

	// we actually don't care because dispatcher must handle all the errors by his own
	// the only case is when this function returns error. In this case there is must be completion <- false,
	// then server will close the connection
	err := client.WriteResponse([]byte("HTTP/1.1 200 OK\r\nServer: indigo-alpha\r\nContent-Length: 20\r\n\r\n" +
		"Looks like it works!"))

	completed <- err == nil
}
