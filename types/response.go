package types

type Response struct {
	code int16
}

func (r *Response) WithCode(code int16) *Response {
	r.code = code
	return r
}
