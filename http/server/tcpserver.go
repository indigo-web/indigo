package server

import (
	"net"
	"sync"
	"time"

	"github.com/fakefloordiv/indigo/http"
)

type (
	connHandler func(*sync.WaitGroup, net.Conn)
	dataHandler func([]byte) error
)

const processed = 0

// StartTCPServer just starts an accept-loop, starting a goroutine provided as a
// handleConn callback for each connection
// If a value from sd (ShutDown channel) is received (no matter which one), server
// will wait until all the goroutines will die and only then release an execution
// flow
func StartTCPServer(sock net.Listener, handleConn connHandler, sd chan struct{}) error {
	wg := new(sync.WaitGroup)

	for {
		select {
		case <-sd:
			wg.Wait()
			sd <- struct{}{}

			return http.ErrShutdown
		default:
			conn, err := sock.Accept()
			if err != nil {
				return err
			}

			wg.Add(1)
			go handleConn(wg, conn)
		}
	}
}

// DefaultConnHandler is a core handler. It takes a buffer for reading provided
// by caller, and starts reading from socket. Guarantees:
// 1) connection will be closed and waitgroup released
//    * in case callback returned errors.ErrHijackConn, connection will not
//      be closed
// 2) on client disconnect, handleData will be called with an empty slice as indicator
//    of disconnect (normally it is not possible)
// Errors occurred while reading socket, will be ignored and user will only know that
// client has disconnected, even if server is guilty
//
// In case timeout is -1 (disabled), ordinary tcp server will be started, removing
// overhead from channels
func DefaultConnHandler(
	wg *sync.WaitGroup, conn net.Conn, timeout int, handleData dataHandler, buff []byte,
) {
	defer wg.Done()

	switch timeout {
	case -1:
		noTimeoutConnHandler(conn, handleData, buff)
	default:
		timeoutConnHandler(conn, handleData, timeout, buff)
	}
}

func noTimeoutConnHandler(conn net.Conn, handleData dataHandler, buff []byte) {
	for {
		n, err := conn.Read(buff)
		err2 := handleData(buff[:n])

		if err2 != nil || err != nil || n == 0 {
			if err2 != http.ErrHijackConn {
				_ = conn.Close()
			}

			return
		}
	}
}

// timeoutConnHandler on each read creates a new timer because time.After does not automatically
// release its underlying timer, so this causes memory leaks. In case timeout occurred,
// connection is closing, so both current and readFromConn goroutines are releasing
func timeoutConnHandler(conn net.Conn, handleData dataHandler, timeout int, buff []byte) {
	ch := make(chan int)
	go readFromConn(ch, conn, buff)

	duration := time.Duration(timeout) * time.Second
	timer := time.NewTimer(duration)

	for {
		select {
		case n := <-ch:
			if !timer.Stop() {
				<-timer.C
			}

			timer.Reset(duration)

			if err := handleData(buff[:n]); err != nil || n == 0 {
				if err != http.ErrHijackConn {
					_ = conn.Close()
				}

				return
			}

			ch <- processed
		case <-timer.C:
			if err := handleData(nil); err != nil {
				panic(
					"BUG: http/server/tcpserver.go:timeoutConnHandler(): handleData(nil) " +
						"returned non-nil error: " + err.Error(),
				)
			}

			_ = conn.Close()
			<-ch
			timer.Stop()
			return
		}
	}
}

// readFromConn simply starts a reading loop, sending to a channel number of bytes
// read. This works because we know that our byte-slice buffer has fixed size, so
// it never grows. As a result, we can pass it by value, and it will always modify
// a single underlying array for everyone
func readFromConn(ch chan int, conn net.Conn, buff []byte) {
	for {
		n, err := conn.Read(buff)
		if err != nil || n == 0 {
			ch <- 0
			return
		}

		ch <- n
		<-ch
	}
}
