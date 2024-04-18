package cookie

type Cookie struct {
	Name  string
	Value string

	Path    string
	Domain  string
	Expires int

	// MaxAge specifies, when the cookie should be destroyed. 0 is treated as a zero-value,
	// so in order to really set MaxAge to 0 in order to destroy it immediately, it must
	// be negative.
	MaxAge   int
	Secure   bool
	HttpOnly bool
	SameSite SameSite
	Raw      string
}

type SameSite int

const (
	SameSiteDefault SameSite = iota + 1
	SameSiteLaxM
	SameSiteStrict
	SameSiteNone
)
