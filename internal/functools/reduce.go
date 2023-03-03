package functools

import "github.com/indigo-web/indigo/internal/constraints"

func Reduce[T constraints.Addable](f func(T, T) T, input []T, initial ...T) (result T) {
	switch len(initial) {
	case 0:
	case 1:
		result = initial[0]
	default:
		panic("reduce: cannot pass more than one initial value")
	}

	switch len(input) {
	case 0:
		return result
	case 1:
		return input[0]
	}

	for _, element := range input {
		result = f(result, element)
	}

	return result
}
