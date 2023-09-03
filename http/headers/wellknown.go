package headers

type Encoding struct {
	Chunked, HasTrailer bool
	Tokens              []string
}
