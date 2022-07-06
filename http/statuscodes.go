package http

import (
	"indigo/internal"
	"net/http"
	"strconv"
)

type StatusCode uint16

const (
	StatusContinue StatusCode = iota + 100
	StatusSwitching
	StatusProcessing
	StatusEarlyHints

	StatusOk StatusCode = iota + 200
	StatusCreated
	StatusAccepted
	StatusNonAuthoritativeInformation
	StatusNoContent
	StatusResetContent
	StatusPartialContent
	StatusMultiStatus
	StatusAlreadyReported
	StatusIMUsed StatusCode = 226

	StatusMultipleChoices StatusCode = iota + 300
	StatusMovedPermanently
	StatusMovedTemporarily
	StatusSeeOther
	StatusNotModified
	StatusUseProxy
	StatusTemporaryRedirect StatusCode = iota + 307
	StatusPermanentRedirect

	StatusBadRequest StatusCode = iota + 400
	StatusUnauthorized
	StatusPaymentRequired
	StatusForbidden
	StatusNotFound
	StatusMethodNotAllowed
	StatusNotAcceptable
	StatusProxyAuthenticationRequired
	StatusRequestTimeout
	StatusConflict
	StatusGone
	StatusLengthRequired
	StatusPreconditionFailed
	StatusPayloadTooLarge
	StatusURITooLong
	StatusUnsupportedMediaType
	StatusRangeNotSatisfiable
	StatusExpectationFailed
	StatusImATeapot
	StatusAuthenticationTimeout

	StatusMisdirectedRequest StatusCode = iota + 421
	StatusUnprocessableEntity
	StatusLocked
	StatusFailedDependency
	StatusTooEarly
	StatusUpgradeRequired

	StatusPreconditionRequired        StatusCode = 428
	StatusTooManyRequests             StatusCode = 429
	StatusRequestHeaderFieldsTooLarge StatusCode = 431
	StatusRetryWith                   StatusCode = 449
	StatusUnavailableForLegalReasons  StatusCode = 451 // wow, who's gonna use this code in my webserver?
	StatusClientClosedConnection      StatusCode = 499

	StatusInternalServerError StatusCode = iota + 500
	StatusNotImplemented
	StatusBadGateway
	StatusServiceUnavailable
	StatusGatewayTimeout
	StatusHTTPVersionNotSupported
	StatusVariantAlsoNegotiates
	StatusInsufficientStorage
	StatusLoopDetected
	StatusBandwidthLimitExceeded
	StatusNotExtended
	StatusNetworkAuthenticationRequired

	StatusUnknownError StatusCode = iota + 520
	StatusWebServerIsDown
	StatusConnectionTimedOut
	StatusOriginIsUnreachable
	StatusTimeoutOccurred
	StatusSSLHandshakeFailed
	StatusInvalidSSLCertificate
)

var statuscodes = [...]StatusCode{
	100, 101, 102, 103,
	200, 201, 202, 203, 204, 205, 206, 207, 208, 226,
	300, 301, 302, 303, 304, 305, 307, 308,
	400, 401, 402, 403, 404, 405, 406, 407, 408, 410,
	411, 412, 413, 414, 415, 416, 417, 418, 419, 421,
	422, 423, 424, 425, 426, 427, 428, 429, 431, 449,
	451, 499,
	500, 501, 502, 503, 504, 505, 506, 507, 508, 509,
	510, 511, 520, 521, 522, 523, 524, 525, 526,
}

func genByteStatusCodes() map[StatusCode][]byte {
	byteStatusMap := make(map[StatusCode][]byte, len(statuscodes))

	for _, code := range statuscodes {
		byteStatusMap[code] = append(internal.S2B(strconv.Itoa(int(code))), ' ')
	}

	return byteStatusMap
}

var ByteStatusCodes = genByteStatusCodes()

// GetStatus TODO: add memoization
func GetStatus(code StatusCode) []byte {
	return internal.S2B(http.StatusText(int(code)) + "\r\n")
}
