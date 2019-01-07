package experiment

import (
	"github.com/hung-phan/apps/src/lib/parallel"
)

const (
	THRESHOLD = 5000
)

func sequentialReduce(arr []int, low, high int) int {
	sum := 0

	for i := low; i <= high; i++ {
		sum += arr[i]
	}

	return sum
}

func parallelReduce(
	arr []int,
	fn func(a, b parallel.Val) parallel.Val,
	low, high int,
) parallel.Val {
	switch {

	case low > high:
		return 0

	case high-low+1 < THRESHOLD:
		return sequentialReduce(arr, low, high)

	default:
		mid := low + (high-low)/2

		res := parallel.Parallel(
			func() parallel.Val {
				return parallelReduce(arr, fn, low, mid-1)
			},
			func() parallel.Val {
				return parallelReduce(arr, fn, mid, high)
			},
		)

		return fn(res[0], res[1])
	}
}
