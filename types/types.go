package types

type (
	ResponseWriter func(b []byte) error
	BodyWriter     func(b []byte)
)
