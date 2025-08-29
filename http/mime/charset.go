package mime

type Charset = string

const (
	UTF8   Charset = "utf8"
	UTF16  Charset = "utf16"
	UTF32  Charset = "utf32"
	ASCII  Charset = "ascii"
	CP1251 Charset = "cp1251"
	CP1252 Charset = "cp1252"
	// feel free to add more widespread charsets!
)
