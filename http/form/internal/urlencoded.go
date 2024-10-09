package internal

import (
	"github.com/indigo-web/indigo/http/form"
	"github.com/indigo-web/indigo/internal/qparams"
)

func ParseURLEncoded(into form.Form, data []byte) (form.Form, error) {
	err := qparams.Parse(data, func(k string, v string) {
		into = append(into, form.Data{
			Name:  k,
			Value: v,
		})
	})

	return into, err
}
