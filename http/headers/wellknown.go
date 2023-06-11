package headers

type TransferEncoding struct {
	Chunked, HasTrailer bool
	Token               string
}
