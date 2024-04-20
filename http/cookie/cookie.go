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
}

type SameSite = string

const (
	SameSiteLax    SameSite = "Lax"
	SameSiteStrict SameSite = "Strict"
	SameSiteNone   SameSite = "None"
)
