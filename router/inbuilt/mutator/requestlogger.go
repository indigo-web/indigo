package mutator

import (
	"log"

	"github.com/indigo-web/indigo/http"
	"github.com/indigo-web/indigo/router/inbuilt/internal"
)

type Logger interface {
	Printf(fmt string, v ...any)
}

func LogRequest(loggers ...Logger) internal.Mutator {
	if len(loggers) == 0 {
		loggers = append(loggers, log.Default())
	}

	return func(request *http.Request) {
		for _, logger := range loggers {
			logger.Printf("%s %s", request.Method.String(), request.Path)
		}
	}
}
