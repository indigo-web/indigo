package types

type Client struct {
	Request  *Request
	Response *Response
}

type (
	ResponseWriter func(b []byte) error
	BodyWriter     func(b []byte)
)
