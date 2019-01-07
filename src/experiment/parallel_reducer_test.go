package experiment

import (
	"fmt"
	"github.com/hung-phan/apps/src/lib/parallel"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestParallelReducer(t *testing.T) {
	assertInstance := assert.New(t)

	testCase := []int{10, 100, 1000, 1000, 10000, 100000}

	for _, N := range testCase {
		t.Run(fmt.Sprintf("test ParallelReducer for %d", N), func(t *testing.T) {
			arr := make([]int, N)

			for i := range arr {
				arr[i] = i
			}

			var (
				add = func(a, b parallel.Val) parallel.Val {
					return a.(int) + b.(int)
				}
				res = parallelReduce(arr, add, 0, len(arr)-1)
			)

			assertInstance.Equal(res, N*(N-1)/2)
		})
	}
}
