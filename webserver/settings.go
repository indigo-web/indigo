package webserver

const (
	DefaultMaxHeaders = 255
)

type Settings struct {
	// TODO: duplicate settings from snowdrop.Settings to pass them directly

	// MaxHeaders is a max number of headers allowed to keep, in case of exceeding this value
	// connection will be closed with StatusBadRequest code. By default, the value is 255,
	// and it cannot be more. IMHO nobody even needs more as 255 is already a hell
	MaxHeaders uint8
}

// TODO: write a function that "prepares" Settings object by replacing zero-values by default ones
