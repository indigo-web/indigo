package config

import (
	"github.com/stretchr/testify/assert"
	"reflect"
	"testing"
)

func TestNoZeroFields(t *testing.T) {
	cfg := Default()

	for _, field := range visit(newVar(*cfg), "Config", false) {
		assert.Fail(t, "zero-value field", field)
	}
}

type variable struct {
	Type  reflect.Type
	Value reflect.Value
}

func newVar(a any) variable {
	return variable{reflect.TypeOf(a), reflect.ValueOf(a)}
}

func visit(a variable, name string, nullable bool) (fields []string) {
	if a.Type.Kind() == reflect.Struct {
		for field := range a.Value.NumField() {
			v1 := variable{a.Type.Field(field).Type, a.Value.Field(field)}
			fieldname := a.Type.Field(field).Name
			isNullable := a.Type.Field(field).Tag.Get("test") == "nullable"
			fields = append(fields, visit(v1, name+"."+fieldname, isNullable)...)
		}

		return fields
	}

	if a.Value.IsZero() && !nullable {
		return []string{name}
	}

	return nil
}
