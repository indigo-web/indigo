package status

type HTTPError struct {
	Message string
	Code    Code
}

func newErr(code Code, message string) HTTPError {
	return HTTPError{
		Code:    code,
		Message: message,
	}
}

func (h HTTPError) Error() string {
	return h.Message
}

var (
	ErrConnectionTimeout = newErr(RequestTimeout, "connection timed out")
	ErrCloseConnection   = newErr(CloseConnection, "internal error as a signal")
	ErrShutdown          = newErr(CloseConnection, "graceful shutdown")

	ErrBadRequest          = newErr(BadRequest, "bad request")
	ErrNotFound            = newErr(NotFound, "not found")
	ErrInternalServerError = newErr(InternalServerError, "internal server error")
	// ErrMethodNotImplemented is actually uses the same error code as just ErrNotImplemented, but used
	// to explain the problem more preciously
	ErrMethodNotImplemented          = newErr(NotImplemented, "request method is not supported")
	ErrMethodNotAllowed              = newErr(MethodNotAllowed, "MethodNotAllowed")
	ErrTooLarge                      = newErr(RequestEntityTooLarge, "too large")
	ErrHeaderFieldsTooLarge          = newErr(RequestHeaderFieldsTooLarge, "header fields too large")
	ErrURITooLong                    = newErr(RequestURITooLong, "request URI too long")
	ErrURIDecoding                   = newErr(BadRequest, "invalid URI encoding")
	ErrBadQuery                      = newErr(BadRequest, "bad URL query")
	ErrUnsupportedProtocol           = newErr(HTTPVersionNotSupported, "protocol is not supported")
	ErrUnsupportedEncoding           = newErr(NotAcceptable, "content encoding is not supported")
	ErrTooManyHeaders                = newErr(RequestHeaderFieldsTooLarge, "too many headers")
	ErrUnauthorized                  = newErr(Unauthorized, "unauthorized")
	ErrPaymentRequired               = newErr(PaymentRequired, "payment required")
	ErrForbidden                     = newErr(Forbidden, "forbidden")
	ErrNotAcceptable                 = newErr(NotAcceptable, "not acceptable")
	ErrProxyAuthRequired             = newErr(ProxyAuthRequired, "proxy auth required")
	ErrRequestTimeout                = newErr(RequestTimeout, "request timeout")
	ErrConflict                      = newErr(Conflict, "conflict")
	ErrGone                          = newErr(Gone, "gone")
	ErrLengthRequired                = newErr(LengthRequired, "length required")
	ErrPreconditionFailed            = newErr(PreconditionFailed, "precondition failed")
	ErrRequestEntityTooLarge         = newErr(RequestEntityTooLarge, "request entity too large")
	ErrRequestURITooLong             = newErr(RequestURITooLong, "request URI too long")
	ErrUnsupportedMediaType          = newErr(UnsupportedMediaType, "unsupported media type")
	ErrRequestedRangeNotSatisfiable  = newErr(RequestedRangeNotSatisfiable, "requested range not satisfiable")
	ErrExpectationFailed             = newErr(ExpectationFailed, "expectation failed")
	ErrTeapot                        = newErr(Teapot, "i'm a teapot")
	ErrMisdirectedRequest            = newErr(MisdirectedRequest, "misdirected request")
	ErrUnprocessableEntity           = newErr(UnprocessableEntity, "unprocessable entity")
	ErrLocked                        = newErr(Locked, "locked")
	ErrFailedDependency              = newErr(FailedDependency, "failed dependency")
	ErrTooEarly                      = newErr(TooEarly, "too early")
	ErrUpgradeRequired               = newErr(UpgradeRequired, "upgrade required")
	ErrPreconditionRequired          = newErr(PreconditionRequired, "precondition required")
	ErrTooManyRequests               = newErr(TooManyRequests, "too many requests")
	ErrRequestHeaderFieldsTooLarge   = newErr(RequestHeaderFieldsTooLarge, "request header fields too large")
	ErrUnavailableForLegalReasons    = newErr(UnavailableForLegalReasons, "unavailable for legal reasons")
	ErrNotImplemented                = newErr(NotImplemented, "not implemented")
	ErrBadGateway                    = newErr(BadGateway, "bad gateway")
	ErrServiceUnavailable            = newErr(ServiceUnavailable, "service unavailable")
	ErrGatewayTimeout                = newErr(GatewayTimeout, "GatewayTimeout")
	ErrHTTPVersionNotSupported       = newErr(HTTPVersionNotSupported, "HTTP version not supported")
	ErrVariantAlsoNegotiates         = newErr(VariantAlsoNegotiates, "variant also negotiates")
	ErrInsufficientStorage           = newErr(InsufficientStorage, "insufficient storage")
	ErrLoopDetected                  = newErr(LoopDetected, "loop detected")
	ErrNotExtended                   = newErr(NotExtended, "not extended")
	ErrNetworkAuthenticationRequired = newErr(NetworkAuthenticationRequired, "network authentication required")
)
