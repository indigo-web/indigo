package formdata

import (
	"github.com/indigo-web/indigo/http/form"
	"github.com/indigo-web/indigo/internal/qparams"
	"github.com/indigo-web/indigo/internal/urlencoded"
)

func ParseURLEncoded(into form.Form, data, buff []byte, defFlagValue string) (form.Form, []byte, error) {
	buff, err := qparams.Parse(data, buff,
		func(k string, v string) {
			into = append(into, form.Data{
				Name:  k,
				Value: v,
			})
		},
		urlencoded.ExtendedDecode,
		defFlagValue,
	)

	return into, buff, err
}
