package settings

const (
	DefaultMaxHeaders       = 255
	DefaultMaxBodyPieceSize = 2048
)

type Settings struct {
	// MaxHeaders is a max number of headers allowed to keep, in case of exceeding this value
	// connection will be closed with StatusBadRequest code. By default, the value is 255,
	// and it cannot be more. IMHO nobody even needs more as 255 is already a hell
	MaxHeaders uint8

	// MaxBodyPieceSize is a max number of bytes can be read
	MaxBodyPieceSize int
}

func CookSettings(settings Settings) Settings {
	maxHeaders := settings.MaxHeaders

	if maxHeaders == 0 {
		maxHeaders = DefaultMaxHeaders
	}

	return Settings{
		MaxHeaders: maxHeaders,
	}
}
