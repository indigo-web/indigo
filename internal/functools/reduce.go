package functools

import (
	"errors"
	"github.com/indigo-web/indigo/internal/constraints"
)

var (
	errTooManyInitialValues = errors.New("reduce: cannot pass more than one initial value")
)

func Reduce[T constraints.Addable](f func(T, T) T, input []T, initial ...T) (result T) {
	switch len(input) {
	case 0:
		return result
	case 1:
		switch len(initial) {
		case 0:
			return input[0]
		case 1:
			return f(initial[0], input[0])
		default:
			panic(errTooManyInitialValues)
		}
	}

	switch len(initial) {
	case 0:
	case 1:
		input[0] = f(initial[0], input[0])
	default:
		panic(errTooManyInitialValues)
	}

	return reduce[T](f, input)
}

func reduce[T any](f func(T, T) T, input []T) (result T) {
	result, input = input[0], input[1:]

	for _, element := range input {
		result = f(result, element)
	}

	return result
}
