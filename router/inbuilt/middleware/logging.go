package middleware

import (
	"github.com/indigo-web/indigo/http"
	"github.com/indigo-web/indigo/http/status"
	"github.com/indigo-web/indigo/router/inbuilt"
	"log"
)

type Logger interface {
	Printf(fmt string, v ...any)
}

func LogRequests(loggers ...Logger) inbuilt.Middleware {
	if len(loggers) == 0 {
		loggers = append(loggers, log.Default())
	}

	return func(next inbuilt.Handler, request *http.Request) *http.Response {
		response := next(request)
		if response.Reveal().Code == status.CloseConnection {
			return response
		}

		for _, logger := range loggers {
			logger.Printf("%s %s %d", request.Method.String(), http.Escape(request.Path), response.Reveal().Code)
		}

		return response
	}
}
