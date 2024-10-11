package formdata

import (
	"github.com/indigo-web/indigo/http/form"
	"github.com/indigo-web/indigo/internal/qparams"
	"github.com/indigo-web/indigo/internal/urlencoded"
)

func ParseURLEncoded(into form.Form, data []byte, buff []byte) (form.Form, error) {
	err := qparams.Parse(data,
		func(k string, v string) {
			into = append(into, form.Data{
				Name:  k,
				Value: v,
			})
		},
		func(bytes []byte) (data []byte, err error) {
			data, buff, err = urlencoded.LazyDecode(bytes, buff)
			return data, err
		},
	)

	return into, err
}
