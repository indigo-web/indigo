package cookie

import "time"

type Cookie struct {
	Name    string
	Value   string
	Path    string
	Domain  string
	Expires time.Time
	// MaxAge defines a delta in seconds, when the cookie should be dropped.
	// Note, that zero is treated as a zero-value, so will be ignored. In order
	// to be added with a value of zero, it must be negative. -1 is the conventional
	// value for this purpose
	MaxAge   int
	SameSite SameSite
	Secure   bool
	HttpOnly bool
}

func New(name, value string) Cookie {
	return Cookie{Name: name, Value: value}
}

type Builder struct {
	cookie Cookie
}

// Build is a chainable constructor for cookies. A preferred way of instantiation
func Build(name, value string) Builder {
	return Builder{New(name, value)}
}

func (b Builder) Path(path string) Builder {
	b.cookie.Path = path
	return b
}

func (b Builder) Domain(domain string) Builder {
	b.cookie.Domain = domain
	return b
}

func (b Builder) Expires(expires time.Time) Builder {
	b.cookie.Expires = expires
	return b
}

// MaxAge defines a delta in seconds, when the cookie should be dropped.
// Note, that zero is treated as a zero-value, so will be ignored. In order
// to be added with a value of zero, it must be negative. -1 is the conventional
// value for this purpose
func (b Builder) MaxAge(maxAge int) Builder {
	b.cookie.MaxAge = maxAge
	return b
}

func (b Builder) SameSite(sameSite SameSite) Builder {
	b.cookie.SameSite = sameSite
	return b
}

func (b Builder) Secure(secure bool) Builder {
	b.cookie.Secure = secure
	return b
}

func (b Builder) HttpOnly(httpOnly bool) Builder {
	b.cookie.HttpOnly = httpOnly
	return b
}

// Cookie returns the built cookie instance
func (b Builder) Cookie() Cookie {
	return b.cookie
}

type SameSite = string

const (
	SameSiteLax    SameSite = "Lax"
	SameSiteStrict SameSite = "Strict"
	SameSiteNone   SameSite = "None"
)
