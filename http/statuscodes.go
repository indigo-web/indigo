package http

import (
	"indigo/internal"
	"net/http"
	"strconv"
)

type StatusCode uint16

const (
	StatusContinue   StatusCode = 100
	StatusSwitching  StatusCode = 101
	StatusProcessing StatusCode = 102
	StatusEarlyHints StatusCode = 103

	StatusOk                          StatusCode = 200
	StatusCreated                     StatusCode = 201
	StatusAccepted                    StatusCode = 202
	StatusNonAuthoritativeInformation StatusCode = 203
	StatusNoContent                   StatusCode = 204
	StatusResetContent                StatusCode = 205
	StatusPartialContent              StatusCode = 206
	StatusMultiStatus                 StatusCode = 207
	StatusAlreadyReported             StatusCode = 208
	StatusIMUsed                      StatusCode = 226

	StatusMultipleChoices   StatusCode = 300
	StatusMovedPermanently  StatusCode = 301
	StatusMovedTemporarily  StatusCode = 302
	StatusSeeOther          StatusCode = 303
	StatusNotModified       StatusCode = 304
	StatusUseProxy          StatusCode = 305
	StatusTemporaryRedirect StatusCode = 307
	StatusPermanentRedirect StatusCode = 308

	StatusBadRequest                  StatusCode = 400
	StatusUnauthorized                StatusCode = 401
	StatusPaymentRequired             StatusCode = 402
	StatusForbidden                   StatusCode = 403
	StatusNotFound                    StatusCode = 404
	StatusMethodNotAllowed            StatusCode = 405
	StatusNotAcceptable               StatusCode = 406
	StatusProxyAuthenticationRequired StatusCode = 407
	StatusRequestTimeout              StatusCode = 408
	StatusConflict                    StatusCode = 409
	StatusGone                        StatusCode = 410
	StatusLengthRequired              StatusCode = 411
	StatusPreconditionFailed          StatusCode = 412
	StatusPayloadTooLarge             StatusCode = 413
	StatusURITooLong                  StatusCode = 414
	StatusUnsupportedMediaType        StatusCode = 415
	StatusRangeNotSatisfiable         StatusCode = 416
	StatusExpectationFailed           StatusCode = 417
	StatusImATeapot                   StatusCode = 418
	StatusAuthenticationTimeout       StatusCode = 419
	StatusMisdirectedRequest          StatusCode = 421
	StatusUnprocessableEntity         StatusCode = 422
	StatusLocked                      StatusCode = 423
	StatusFailedDependency            StatusCode = 424
	StatusTooEarly                    StatusCode = 425
	StatusUpgradeRequired             StatusCode = 426
	StatusPreconditionRequired        StatusCode = 428
	StatusTooManyRequests             StatusCode = 429
	StatusRequestHeaderFieldsTooLarge StatusCode = 431
	StatusRetryWith                   StatusCode = 449
	StatusUnavailableForLegalReasons  StatusCode = 451 // wow, who's gonna use this code in my webserver?
	StatusClientClosedConnection      StatusCode = 499

	StatusInternalServerError           StatusCode = 500
	StatusNotImplemented                StatusCode = 501
	StatusBadGateway                    StatusCode = 502
	StatusServiceUnavailable            StatusCode = 503
	StatusGatewayTimeout                StatusCode = 504
	StatusHTTPVersionNotSupported       StatusCode = 505
	StatusVariantAlsoNegotiates         StatusCode = 506
	StatusInsufficientStorage           StatusCode = 507
	StatusLoopDetected                  StatusCode = 508
	StatusBandwidthLimitExceeded        StatusCode = 509
	StatusNotExtended                   StatusCode = 510
	StatusNetworkAuthenticationRequired StatusCode = 511
	StatusUnknownError                  StatusCode = 520
	StatusWebServerIsDown               StatusCode = 521
	StatusConnectionTimedOut            StatusCode = 522
	StatusOriginIsUnreachable           StatusCode = 523
	StatusTimeoutOccurred               StatusCode = 524
	StatusSSLHandshakeFailed            StatusCode = 525
	StatusInvalidSSLCertificate         StatusCode = 526
)

var statuscodes = []StatusCode{
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

var byteStatusCodes = genStatuscodesBytesTrailingSpaces(statuscodes)

func genStatuscodesBytesTrailingSpaces(statusCodes []StatusCode) map[StatusCode][]byte {
	statusCodesBytesMap := make(map[StatusCode][]byte, len(statuscodes))

	for _, statuscode := range statusCodes {
		statusCodesBytesMap[statuscode] = internal.S2B(strconv.Itoa(int(statuscode)) + " ")
	}

	return statusCodesBytesMap
}

// GetStatusTrailingCRLF TODO: add memoization
func GetStatusTrailingCRLF(code StatusCode) []byte {
	return internal.S2B(http.StatusText(int(code)) + "\r\n")
}

// GetByteCodeTrailingSpace TODO: add memoization
func GetByteCodeTrailingSpace(code StatusCode) []byte {
	return byteStatusCodes[code]
}
