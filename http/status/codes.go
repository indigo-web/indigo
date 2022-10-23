package status

/*
INFO: this is copy-paste from net/http/status.go. Added because
of unwanted name collisions between indigo/http and net/http
(also goland does not propose me status codes in autocomplete)
*/

type (
	Code   uint16
	Status string
)

// HTTP status codes as registered with IANA.
// See: https://www.iana.org/assignments/http-status-codes/http-status-codes.xhtml
const (
	Continue           Code = 100 // RFC 9110, 15.2.1
	SwitchingProtocols Code = 101 // RFC 9110, 15.2.2
	Processing         Code = 102 // RFC 2518, 10.1
	EarlyHints         Code = 103 // RFC 8297

	OK                   Code = 200 // RFC 9110, 15.3.1
	Created              Code = 201 // RFC 9110, 15.3.2
	Accepted             Code = 202 // RFC 9110, 15.3.3
	NonAuthoritativeInfo Code = 203 // RFC 9110, 15.3.4
	NoContent            Code = 204 // RFC 9110, 15.3.5
	ResetContent         Code = 205 // RFC 9110, 15.3.6
	PartialContent       Code = 206 // RFC 9110, 15.3.7
	MultiStatus          Code = 207 // RFC 4918, 11.1
	AlreadyReported      Code = 208 // RFC 5842, 7.1
	IMUsed               Code = 226 // RFC 3229, 10.4.1

	MultipleChoices   Code = 300 // RFC 9110, 15.4.1
	MovedPermanently  Code = 301 // RFC 9110, 15.4.2
	Found             Code = 302 // RFC 9110, 15.4.3
	SeeOther          Code = 303 // RFC 9110, 15.4.4
	NotModified       Code = 304 // RFC 9110, 15.4.5
	UseProxy          Code = 305 // RFC 9110, 15.4.6
	_                 Code = 306 // RFC 9110, 15.4.7 (Unused)
	TemporaryRedirect Code = 307 // RFC 9110, 15.4.8
	PermanentRedirect Code = 308 // RFC 9110, 15.4.9

	BadRequest                   Code = 400 // RFC 9110, 15.5.1
	Unauthorized                 Code = 401 // RFC 9110, 15.5.2
	PaymentRequired              Code = 402 // RFC 9110, 15.5.3
	Forbidden                    Code = 403 // RFC 9110, 15.5.4
	NotFound                     Code = 404 // RFC 9110, 15.5.5
	MethodNotAllowed             Code = 405 // RFC 9110, 15.5.6
	NotAcceptable                Code = 406 // RFC 9110, 15.5.7
	ProxyAuthRequired            Code = 407 // RFC 9110, 15.5.8
	RequestTimeout               Code = 408 // RFC 9110, 15.5.9
	Conflict                     Code = 409 // RFC 9110, 15.5.10
	Gone                         Code = 410 // RFC 9110, 15.5.11
	LengthRequired               Code = 411 // RFC 9110, 15.5.12
	PreconditionFailed           Code = 412 // RFC 9110, 15.5.13
	RequestEntityTooLarge        Code = 413 // RFC 9110, 15.5.14
	RequestURITooLong            Code = 414 // RFC 9110, 15.5.15
	UnsupportedMediaType         Code = 415 // RFC 9110, 15.5.16
	RequestedRangeNotSatisfiable Code = 416 // RFC 9110, 15.5.17
	ExpectationFailed            Code = 417 // RFC 9110, 15.5.18
	Teapot                       Code = 418 // RFC 9110, 15.5.19 (Unused)
	MisdirectedRequest           Code = 421 // RFC 9110, 15.5.20
	UnprocessableEntity          Code = 422 // RFC 9110, 15.5.21
	Locked                       Code = 423 // RFC 4918, 11.3
	FailedDependency             Code = 424 // RFC 4918, 11.4
	TooEarly                     Code = 425 // RFC 8470, 5.2.
	UpgradeRequired              Code = 426 // RFC 9110, 15.5.22
	PreconditionRequired         Code = 428 // RFC 6585, 3
	TooManyRequests              Code = 429 // RFC 6585, 4
	RequestHeaderFieldsTooLarge  Code = 431 // RFC 6585, 5
	UnavailableForLegalReasons   Code = 451 // RFC 7725, 3

	InternalServerError           Code = 500 // RFC 9110, 15.6.1
	NotImplemented                Code = 501 // RFC 9110, 15.6.2
	BadGateway                    Code = 502 // RFC 9110, 15.6.3
	ServiceUnavailable            Code = 503 // RFC 9110, 15.6.4
	GatewayTimeout                Code = 504 // RFC 9110, 15.6.5
	HTTPVersionNotSupported       Code = 505 // RFC 9110, 15.6.6
	VariantAlsoNegotiates         Code = 506 // RFC 2295, 8.1
	InsufficientStorage           Code = 507 // RFC 4918, 11.5
	LoopDetected                  Code = 508 // RFC 5842, 7.2
	NotExtended                   Code = 510 // RFC 2774, 7
	NetworkAuthenticationRequired Code = 511 // RFC 6585, 6
)

// Text returns a text for the HTTP status code. It returns the empty
// string if the code is unknown.
func Text(code Code) Status {
	switch code {
	case Continue:
		return "Continue"
	case SwitchingProtocols:
		return "Switching Protocols"
	case Processing:
		return "Processing"
	case EarlyHints:
		return "Early Hints"
	case OK:
		return "OK"
	case Created:
		return "Created"
	case Accepted:
		return "Accepted"
	case NonAuthoritativeInfo:
		return "Non-Authoritative Information"
	case NoContent:
		return "No Content"
	case ResetContent:
		return "Reset Content"
	case PartialContent:
		return "Partial Content"
	case MultiStatus:
		return "Multi-Status"
	case AlreadyReported:
		return "Already Reported"
	case IMUsed:
		return "IM Used"
	case MultipleChoices:
		return "Multiple Choices"
	case MovedPermanently:
		return "Moved Permanently"
	case Found:
		return "Found"
	case SeeOther:
		return "See Other"
	case NotModified:
		return "Not Modified"
	case UseProxy:
		return "Use Proxy"
	case TemporaryRedirect:
		return "Temporary Redirect"
	case PermanentRedirect:
		return "Permanent Redirect"
	case BadRequest:
		return "Bad Request"
	case Unauthorized:
		return "Unauthorized"
	case PaymentRequired:
		return "Payment Required"
	case Forbidden:
		return "Forbidden"
	case NotFound:
		return "Not Found"
	case MethodNotAllowed:
		return "Method Not Allowed"
	case NotAcceptable:
		return "Not Acceptable"
	case ProxyAuthRequired:
		return "Proxy Authentication Required"
	case RequestTimeout:
		return "Request Timeout"
	case Conflict:
		return "Conflict"
	case Gone:
		return "Gone"
	case LengthRequired:
		return "Length Required"
	case PreconditionFailed:
		return "Precondition Failed"
	case RequestEntityTooLarge:
		return "Request Entity Too Large"
	case RequestURITooLong:
		return "Request URI Too Long"
	case UnsupportedMediaType:
		return "Unsupported Media Type"
	case RequestedRangeNotSatisfiable:
		return "Requested Range Not Satisfiable"
	case ExpectationFailed:
		return "Expectation Failed"
	case Teapot:
		return "I'm a teapot"
	case MisdirectedRequest:
		return "Misdirected Request"
	case UnprocessableEntity:
		return "Unprocessable Entity"
	case Locked:
		return "Locked"
	case FailedDependency:
		return "Failed Dependency"
	case TooEarly:
		return "Too Early"
	case UpgradeRequired:
		return "Upgrade Required"
	case PreconditionRequired:
		return "Precondition Required"
	case TooManyRequests:
		return "Too Many Requests"
	case RequestHeaderFieldsTooLarge:
		return "Request Header Fields Too Large"
	case UnavailableForLegalReasons:
		return "Unavailable For Legal Reasons"
	case InternalServerError:
		return "Internal Server Error"
	case NotImplemented:
		return "Not Implemented"
	case BadGateway:
		return "Bad Gateway"
	case ServiceUnavailable:
		return "Service Unavailable"
	case GatewayTimeout:
		return "Gateway Timeout"
	case HTTPVersionNotSupported:
		return "HTTP Version Not Supported"
	case VariantAlsoNegotiates:
		return "Variant Also Negotiates"
	case InsufficientStorage:
		return "Insufficient Storage"
	case LoopDetected:
		return "Loop Detected"
	case NotExtended:
		return "Not Extended"
	case NetworkAuthenticationRequired:
		return "Network Authentication Required"
	default:
		return "Unknown Status Code"
	}
}

// CodeStatus returns a pre-defined line with code and status text (including
// terminating CRLF sequence) in case code is known to server, otherwise empty
// line is returned
func CodeStatus(code Code) string {
	switch code {
	case Continue:
		return "100 Continue\r\n"
	case SwitchingProtocols:
		return "101 Switching Protocol\r\n"
	case Processing:
		return "102 Processing\r\n"
	case EarlyHints:
		return "103 Early Hints\r\n"
	case OK:
		return "200 OK\r\n"
	case Created:
		return "201 Created\r\n"
	case Accepted:
		return "202 Accepted\r\n"
	case NonAuthoritativeInfo:
		return "203 Non-Authoritative Information\r\n"
	case NoContent:
		return "204 No Content\r\n"
	case ResetContent:
		return "205 Reset Content\r\n"
	case PartialContent:
		return "206 Partial Content\r\n"
	case MultiStatus:
		return "207 Multi-Status\r\n"
	case AlreadyReported:
		return "208 Already Reported\r\n"
	case IMUsed:
		return "226 IM Used\r\n"
	case MultipleChoices:
		return "300 Multiple Choices\r\n"
	case MovedPermanently:
		return "301 Moved Permanently\r\n"
	case Found:
		return "302 Found\r\n"
	case SeeOther:
		return "303 See Other\r\n"
	case NotModified:
		return "304 Not Modified\r\n"
	case UseProxy:
		return "305 Use Proxy\r\n"
	case TemporaryRedirect:
		return "307 Temporary Redirect\r\n"
	case PermanentRedirect:
		return "308 Permanent Redirect\r\n"
	case BadRequest:
		return "400 Bad Request\r\n"
	case Unauthorized:
		return "401 Unauthorized\r\n"
	case PaymentRequired:
		return "402 Payment Required\r\n"
	case Forbidden:
		return "403 Forbidden\r\n"
	case NotFound:
		return "404 Not Found\r\n"
	case MethodNotAllowed:
		return "405 Method Not Allowed\r\n"
	case NotAcceptable:
		return "406 Not Acceptable\r\n"
	case ProxyAuthRequired:
		return "407 Proxy Authentication Required\r\n"
	case RequestTimeout:
		return "408 Request Timeout\r\n"
	case Conflict:
		return "409 Conflict\r\n"
	case Gone:
		return "410 Gone\r\n"
	case LengthRequired:
		return "411 Length Required\r\n"
	case PreconditionFailed:
		return "412 Precondition Failed\r\n"
	case RequestEntityTooLarge:
		return "413 Request Entity Too Large\r\n"
	case RequestURITooLong:
		return "414 Request URI Too Long\r\n"
	case UnsupportedMediaType:
		return "415 Unsupported Media Type\r\n"
	case RequestedRangeNotSatisfiable:
		return "416 Requested Range Not Satisfiable\r\n"
	case ExpectationFailed:
		return "417 Expectation Failed\r\n"
	case Teapot:
		return "418 I'm a teapot\r\n"
	case MisdirectedRequest:
		return "421 Misdirected Request\r\n"
	case UnprocessableEntity:
		return "422 Unprocessable Entity\r\n"
	case Locked:
		return "423 Locked\r\n"
	case FailedDependency:
		return "424 Failed Dependency\r\n"
	case TooEarly:
		return "425 Too Early\r\n"
	case UpgradeRequired:
		return "426 Upgrade Required\r\n"
	case PreconditionRequired:
		return "428 Precondition Required\r\n"
	case TooManyRequests:
		return "429 Too Many Requests\r\n"
	case RequestHeaderFieldsTooLarge:
		return "431 Request Header Fields Too Large\r\n"
	case UnavailableForLegalReasons:
		return "451 Unavailable For Legal Reasons\r\n"
	case InternalServerError:
		return "500 Internal Server Error\r\n"
	case NotImplemented:
		return "501 Not Implemented\r\n"
	case BadGateway:
		return "502 Bad Gateway\r\n"
	case ServiceUnavailable:
		return "503 Service Unavailable\r\n"
	case GatewayTimeout:
		return "504 Gateway Timeout\r\n"
	case HTTPVersionNotSupported:
		return "505 HTTP Version Not Supported\r\n"
	case VariantAlsoNegotiates:
		return "506 Variant Also Negotiates\r\n"
	case InsufficientStorage:
		return "507 Insufficient Storage\r\n"
	case LoopDetected:
		return "508 Loop Detected\r\n"
	case NotExtended:
		return "510 Not Extended\r\n"
	case NetworkAuthenticationRequired:
		return "511 Network Authentication Required\r\n"
	default:
		return ""
	}
}
