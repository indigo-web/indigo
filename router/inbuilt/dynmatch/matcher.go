package main

import (
	"context"
	"errors"
	"fmt"
	context2 "github.com/fakefloordiv/indigo/internal/context"
	"strconv"
	"strings"
)

var (
	ErrNotImplemented  = errors.New("current implementation does not allows static text between slashes")
	ErrInvalidTemplate = errors.New("invalid template")
	ErrEmptyPath       = errors.New("path template cannot be empty")
)

// Template is a parsed template. It simply contains static parts, and marker names
// index of each corresponds to index of static part (first static part, then marker)
type Template struct {
	staticParts []string
	markerNames []string
}

func Parse(tmpl string) (Template, error) {
	var (
		offset   int
		template = Template{}
		state    = eStatic
	)

	if len(tmpl) == 0 {
		return template, ErrEmptyPath
	}

	for i, char := range tmpl {
		if i == 0 {
			if char != '/' {
				return template, ErrInvalidTemplate
			}

			// skip leading slash
			continue
		}

		switch state {
		case eStatic:
			switch char {
			case '/':
				state = eSlash
			case '{':
				return template, ErrNotImplemented
			}
		case eSlash:
			switch char {
			case '{':
				template.staticParts = append(template.staticParts, tmpl[offset:i])
				offset = i + 1
				state = ePartName
			default:
				state = eStatic
			}
		case ePartName:
			switch char {
			case '}':
				template.markerNames = append(template.markerNames, tmpl[offset:i])
				offset = i + 1
				state = eEndPartName
			case '/':
				return template, ErrInvalidTemplate
			}
		case eEndPartName:
			switch char {
			case '/':
				state = eSlash
			default:
				return template, ErrNotImplemented
			}
		}
	}

	if state == eStatic || state == eSlash {
		template.staticParts = append(template.staticParts, tmpl)
	}

	return template, nil
}

func (t Template) Match(ctx context.Context, path string) (context.Context, bool) {
	if len(t.staticParts) == 1 {
		return ctx, path == t.staticParts[0]
	}

	var (
		staticIndex int
	)

	for staticIndex < len(t.staticParts) {
		staticPart := t.staticParts[staticIndex]

		if len(path) < len(staticPart) || path[:len(staticPart)] != staticPart {
			return ctx, false
		}

		dynamicPart := path[len(staticPart):]

		if slash := strings.IndexByte(dynamicPart, '/'); slash != -1 {
			if name := t.markerNames[staticIndex]; len(name) > 0 {
				ctx = context2.WithValue(ctx, name, dynamicPart[:slash])
			}

			path = dynamicPart[slash:]
		} else {
			if name := t.markerNames[staticIndex]; len(name) > 0 {
				ctx = context2.WithValue(ctx, name, dynamicPart)
			}

			return ctx, staticIndex+1 == len(t.staticParts)
		}

		staticIndex++
	}

	return ctx, true
}

func main() {
	tmpl := "/hello/{world}/{ok}/good/{name}"
	fmt.Println("parsing:", strconv.Quote(tmpl))

	template, err := Parse(tmpl)
	fmt.Println("template:", template, "err:", err)

	sample := "/hello/world-world/ok-perfect/good/name-Akakiy"
	_, matched := template.Match(context.Background(), sample)
	fmt.Println("Sample:", sample)
	fmt.Println("matches:", matched)
}
