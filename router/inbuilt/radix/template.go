package radix

import (
	"context"
	"errors"
	context2 "github.com/fakefloordiv/indigo/valuectx"
	"strings"
)

type templateParserState uint8

const (
	eStatic templateParserState = iota + 1
	ePartName
	eFinishPartName
)

var (
	ErrNeedLeadingSlash = errors.New(
		"leading slash is compulsory",
	)
	ErrInvalidPartName = errors.New(
		"slashes or figure braces are not allowed inside of the template part name",
	)
	ErrEmptyPath = errors.New(
		"template cannot be empty",
	)
	ErrMustEndWithSlash = errors.New(
		"a following slash is compulsory after the end of dynamic part",
	)
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
				return template, ErrNeedLeadingSlash
			}

			// skip leading slash
			continue
		}

		switch state {
		case eStatic:
			switch char {
			case '{':
				template.staticParts = append(template.staticParts, tmpl[offset:i])

				offset = i + 1
				state = ePartName
			}
		case ePartName:
			switch char {
			case '}':
				template.markerNames = append(template.markerNames, tmpl[offset:i])
				offset = i + 1
				state = eFinishPartName
			case '/', '{':
				return template, ErrInvalidPartName
			}
		case eFinishPartName:
			switch char {
			case '/':
				state = eStatic
			default:
				return template, ErrMustEndWithSlash
			}
		}
	}

	if state == eStatic && offset < len(tmpl)-1 {
		template.staticParts = append(template.staticParts, tmpl[offset:])
	}

	return template, nil
}

func MustParse(tmpl string) Template {
	template, err := Parse(tmpl)
	if err != nil {
		panic(err.Error())
	}

	return template
}

func (t Template) Match(ctx context.Context, path string) (context.Context, bool) {
	if len(t.staticParts) == 1 {
		return ctx, path == t.staticParts[0]
	}

	var (
		staticIndex int
	)

	if len(t.staticParts) > len(t.markerNames) {
		if newPath := strings.TrimSuffix(path, t.staticParts[len(t.staticParts)-1]); newPath != path {
			path = newPath
		} else {
			return ctx, false
		}
	}

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
