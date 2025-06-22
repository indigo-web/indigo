package mime

var Extension = map[string]MIME{
	".avif": AVIF,
	".css":  CSS,
	".gif":  GIF,
	".htm":  HTML,
	".html": HTML,
	".jpeg": JPEG,
	".jpg":  JPEG,
	".js":   JAVASCRIPT,
	".mjs":  JAVASCRIPT,
	".json": JSON,
	".pdf":  PDF,
	".png":  PNG,
	".svg":  SVG,
	".wasm": WASM,
	".webp": WEBP,
	".xml":  XML,
	".gz":   GZIP,
	".sql":  SQL,
	".tzif": TZIF,
	".yaml": YAML,
	".xfdf": XFDF,
	".zip":  ZIP,
	".zlib": ZLIB,
	".zstd": ZSTD,
	".ico":  ICO,
}

// DefaultCharset defines charsets, used by default for MIMEs unless explicitly set.
var DefaultCharset = map[MIME]Charset{
	CSS:        UTF8,
	HTML:       UTF8,
	JAVASCRIPT: UTF8,
	XML:        UTF8,
}
