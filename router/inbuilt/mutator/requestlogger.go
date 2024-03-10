package mutator

import (
	"github.com/indigo-web/indigo/http"
	"github.com/indigo-web/indigo/router/inbuilt/internal/types"
	"log"
)

type logger interface {
	Printf(fmt string, v ...any)
}

func RequestsLog(loggers ...logger) types.Mutator {
	if len(loggers) == 0 {
		loggers = append(loggers, log.Default())
	}

	return func(request *http.Request) {
		for _, logger := range loggers {
			logger.Printf("%s %s", request.Method.String(), http.Escape(request.Path))
		}
	}
}
