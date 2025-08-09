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
	ErrCloseConnection = NewError(CloseConnection, "actively closing the connection")

	ErrBadRequest                    = NewError(BadRequest, "bad request")
	ErrTooLongRequestLine            = NewError(BadRequest, "request line is too long")
	ErrURLDecoding                   = NewError(BadRequest, "invalid urlencoded sequence")
	ErrBadParams                     = NewError(BadRequest, "bad URI params")
	ErrBadEncoding                   = NewError(BadRequest, "bad request encoding")
	ErrBadChunk                      = NewError(BadRequest, "malformed chunk-encoded data")
	ErrNotFound                      = NewError(NotFound, "not found")
	ErrInternalServerError           = NewError(InternalServerError, "internal server error")
	ErrNotImplemented                = NewError(NotImplemented, "not implemented")
	ErrMethodNotImplemented          = NewError(NotImplemented, "request method is not supported")
	ErrMethodNotAllowed              = NewError(MethodNotAllowed, "method not allowed")
	ErrBodyTooLarge                  = NewError(RequestEntityTooLarge, "request body is too large")
	ErrRequestEntityTooLarge         = NewError(RequestEntityTooLarge, "request entity too large")
	ErrHeaderFieldsTooLarge          = NewError(HeaderFieldsTooLarge, "too large headers section")
	ErrTooManyEncodingTokens         = NewError(HeaderFieldsTooLarge, "too many encoding tokens specified")
	ErrTooManyHeaders                = NewError(HeaderFieldsTooLarge, "too many headers")
	ErrURITooLong                    = NewError(RequestURITooLong, "request URI too long")
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
	ErrUnsupportedEncoding           = NewError(UnsupportedMediaType, "encoding is not supported")
	ErrRequestedRangeNotSatisfiable  = NewError(RequestedRangeNotSatisfiable, "requested range is not satisfiable")
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
	ErrGatewayTimeout                = NewError(GatewayTimeout, "gateway timeout")
	ErrVariantAlsoNegotiates         = NewError(VariantAlsoNegotiates, "variant also negotiates")
	ErrInsufficientStorage           = NewError(InsufficientStorage, "insufficient storage")
	ErrLoopDetected                  = NewError(LoopDetected, "loop detected")
	ErrNotExtended                   = NewError(NotExtended, "not extended")
	ErrNetworkAuthenticationRequired = NewError(NetworkAuthenticationRequired, "network authentication required")
)
