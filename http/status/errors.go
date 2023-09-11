package status

type HTTPError struct {
	Message string
	Code    Code
}

func NewError(code Code, message string) error {
	return HTTPError{
		Code:    code,
		Message: message,
	}
}

func (h HTTPError) Error() string {
	return h.Message
}

var (
	ErrConnectionTimeout = NewError(RequestTimeout, "connection timed out")
	ErrCloseConnection   = NewError(CloseConnection, "internal error as a signal")
	ErrShutdown          = NewError(CloseConnection, "graceful shutdown")

	ErrBadRequest                    = NewError(BadRequest, "bad request")
	ErrTooLongRequestLine            = NewError(BadRequest, "request line is too long")
	ErrTooLongResponseLine           = NewError(BadRequest, "response line is too long")
	ErrURIDecoding                   = NewError(BadRequest, "invalid URI encoding")
	ErrBadQuery                      = NewError(BadRequest, "bad URL query")
	ErrNotFound                      = NewError(NotFound, "not found")
	ErrInternalServerError           = NewError(InternalServerError, "internal server error")
	ErrNotImplemented                = NewError(NotImplemented, "not implemented")
	ErrMethodNotImplemented          = NewError(NotImplemented, "request method is not supported")
	ErrMethodNotAllowed              = NewError(MethodNotAllowed, "MethodNotAllowed")
	ErrTooLarge                      = NewError(RequestEntityTooLarge, "too large")
	ErrRequestEntityTooLarge         = NewError(RequestEntityTooLarge, "request entity too large")
	ErrHeaderFieldsTooLarge          = NewError(HeaderFieldsTooLarge, "too large headers section")
	ErrHeaderKeyTooLarge             = NewError(HeaderFieldsTooLarge, "too large header key")
	ErrHeaderValueTooLarge           = NewError(HeaderFieldsTooLarge, "too large header value")
	ErrTooManyHeaders                = NewError(HeaderFieldsTooLarge, "too many headers")
	ErrRequestHeaderFieldsTooLarge   = NewError(HeaderFieldsTooLarge, "request header fields too large")
	ErrURITooLong                    = NewError(RequestURITooLong, "request URI too long")
	ErrRequestURITooLong             = NewError(RequestURITooLong, "request URI too long")
	ErrUnsupportedProtocol           = NewError(HTTPVersionNotSupported, "protocol is not supported")
	ErrHTTPVersionNotSupported       = NewError(HTTPVersionNotSupported, "HTTP version not supported")
	ErrUnauthorized                  = NewError(Unauthorized, "unauthorized")
	ErrPaymentRequired               = NewError(PaymentRequired, "payment required")
	ErrForbidden                     = NewError(Forbidden, "forbidden")
	ErrNotAcceptable                 = NewError(NotAcceptable, "not acceptable")
	ErrProxyAuthRequired             = NewError(ProxyAuthRequired, "proxy auth required")
	ErrRequestTimeout                = NewError(RequestTimeout, "request timeout")
	ErrConflict                      = NewError(Conflict, "conflict")
	ErrGone                          = NewError(Gone, "gone")
	ErrLengthRequired                = NewError(LengthRequired, "length required")
	ErrPreconditionFailed            = NewError(PreconditionFailed, "precondition failed")
	ErrUnsupportedMediaType          = NewError(UnsupportedMediaType, "unsupported media type")
	ErrUnsupportedEncoding           = NewError(UnsupportedMediaType, "content encoding is not supported")
	ErrRequestedRangeNotSatisfiable  = NewError(RequestedRangeNotSatisfiable, "requested range not satisfiable")
	ErrExpectationFailed             = NewError(ExpectationFailed, "expectation failed")
	ErrTeapot                        = NewError(Teapot, "i'm a teapot")
	ErrMisdirectedRequest            = NewError(MisdirectedRequest, "misdirected request")
	ErrUnprocessableEntity           = NewError(UnprocessableEntity, "unprocessable entity")
	ErrLocked                        = NewError(Locked, "locked")
	ErrFailedDependency              = NewError(FailedDependency, "failed dependency")
	ErrTooEarly                      = NewError(TooEarly, "too early")
	ErrUpgradeRequired               = NewError(UpgradeRequired, "upgrade required")
	ErrPreconditionRequired          = NewError(PreconditionRequired, "precondition required")
	ErrTooManyRequests               = NewError(TooManyRequests, "too many requests")
	ErrUnavailableForLegalReasons    = NewError(UnavailableForLegalReasons, "unavailable for legal reasons")
	ErrBadGateway                    = NewError(BadGateway, "bad gateway")
	ErrServiceUnavailable            = NewError(ServiceUnavailable, "service unavailable")
	ErrGatewayTimeout                = NewError(GatewayTimeout, "GatewayTimeout")
	ErrVariantAlsoNegotiates         = NewError(VariantAlsoNegotiates, "variant also negotiates")
	ErrInsufficientStorage           = NewError(InsufficientStorage, "insufficient storage")
	ErrLoopDetected                  = NewError(LoopDetected, "loop detected")
	ErrNotExtended                   = NewError(NotExtended, "not extended")
	ErrNetworkAuthenticationRequired = NewError(NetworkAuthenticationRequired, "network authentication required")
)
