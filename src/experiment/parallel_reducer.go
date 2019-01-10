package experiment

import (
	"github.com/hung-phan/apps/src/lib/parallel"
)

const (
	THRESHOLD = 5000
)

func sequentialReduce(arr []int, from, until int) int {
	sum := 0

	for i := from; i < until; i++ {
		sum += arr[i]
	}

	return sum
}

func parallelReduce(
	arr []int,
	fn func(a, b parallel.Val) parallel.Val,
	from, until int,
) parallel.Val {
	switch {

	case from >= until:
		return 0

	case until-from < THRESHOLD:
		return sequentialReduce(arr, from, until)

	default:
		mid := from + (until-from)/2

		res := parallel.Parallel(
			func() parallel.Val {
				return parallelReduce(arr, fn, from, mid)
			},
			func() parallel.Val {
				return parallelReduce(arr, fn, mid, until)
			},
		)

		return fn(res[0], res[1])
	}
}
