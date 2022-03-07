package httplib

type StatusCode uint16

const (
	// 1xx: informational
	StatusContinue   StatusCode = 100
	StatusSwitching  StatusCode = 101
	StatusProcessing StatusCode = 102
	StatusEarlyHints StatusCode = 103

	// 2xx: success
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

	// 3xx: redirection
	StatusMultipleChoices   StatusCode = 300
	StatusMovedPermanently  StatusCode = 301
	StatusMovedTemporarily  StatusCode = 302
	StatusSeeOther          StatusCode = 303
	StatusNotModified       StatusCode = 304
	StatusUseProxy          StatusCode = 305
	StatusTemporaryRedirect StatusCode = 307
	StatusPermanentRedirect StatusCode = 308

	// 4xx: client error
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
	// wow, who's gonna use this code in my webserver?
	StatusUnavailableForLegalReasons StatusCode = 451
	StatusClientClosedConnection     StatusCode = 499

	// 5xx: server error
	StatusInternalServerError           = 500
	StatusNotImplemented                = 501
	StatusBadGateway                    = 502
	StatusServiceUnavailable            = 503
	StatusGatewayTimeout                = 504
	StatusHTTPVersionNotSupported       = 505
	StatusVariantAlsoNegotiates         = 506
	StatusInsufficientStorage           = 507
	StatusLoopDetected                  = 508
	StatusBandwidthLimitExceeded        = 509
	StatusNotExtended                   = 510
	StatusNetworkAuthenticationRequired = 511
	StatusUnknownError                  = 520
	StatusWebServerIsDown               = 521
	StatusConnectionTimedOut            = 522
	StatusOriginIsUnreachable           = 523
	StatusTimeoutOccurred               = 524
	StatusSSLHandshakeFailed            = 525
	StatusInvalidSSLCertificate         = 526
)
