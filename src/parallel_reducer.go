package main

import (
	"github.com/hung-phan/apps/src/lib/parallel"
)

const (
	THRESHOLD = 1000
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
	fn func(a, b parallel.Result) parallel.Result,
	low, high int,
) parallel.Result {
	switch {

	case low > high:
		return 0

	case high-low+1 < THRESHOLD:
		return sequentialReduce(arr, low, high)

	default:
		mid := low + (high-low)/2

		res := parallel.Parallel(
			func() parallel.Result {
				return parallelReduce(arr, fn, low, mid-1)
			},
			func() parallel.Result {
				return parallelReduce(arr, fn, mid, high)
			},
		)

		return fn(res[0], res[1])
	}
}

func main() {
	arr := make([]int, 100000)

	for i := range arr {
		arr[i] = i
	}

	add := func(a, b parallel.Result) parallel.Result {
		return a.(int) + b.(int)
	}
	res := parallelReduce(arr, add, 0, len(arr)-1)

	println(res.(int))
}
