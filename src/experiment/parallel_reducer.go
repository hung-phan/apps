package experiment

import (
	"github.com/hung-phan/apps/src/lib/parallel"
)

const (
	THRESHOLD = 5000
)

func sequentialReduce(arr []int, from, to int) int {
	sum := 0

	for i := from; i <= to; i++ {
		sum += arr[i]
	}

	return sum
}

func parallelReduce(
	arr []int,
	fn func(a, b parallel.Val) parallel.Val,
	from, to int,
) parallel.Val {
	switch {

	case from > to:
		return 0

	case to-from+1 < THRESHOLD:
		return sequentialReduce(arr, from, to)

	default:
		mid := from + (to-from)/2

		res := parallel.Parallel(
			func() parallel.Val {
				return parallelReduce(arr, fn, from, mid-1)
			},
			func() parallel.Val {
				return parallelReduce(arr, fn, mid, to)
			},
		)

		return fn(res[0], res[1])
	}
}
