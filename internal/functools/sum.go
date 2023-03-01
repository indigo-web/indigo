package functools

import "github.com/indigo-web/indigo/v2/internal/constraints"

func accumulator[T constraints.Addable](prev T, curr T) T {
	return prev + curr
}

func Sum[T constraints.Addable](input []T) T {
	return Reduce(accumulator[T], input)
}
