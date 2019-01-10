package experiment

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestParallelPrefixSum(t *testing.T) {
	assertInstance := assert.New(t)

	testCase := []int{10, 100, 1000, 10000}

	for _, N := range testCase {
		t.Run(fmt.Sprintf("test ParallelReducer for %d", N), func(t *testing.T) {
			input, output := make([]int, N), make([]int, N)

			for i := range input {
				input[i] = i
			}

			parallelPrefixSum(input, output, 100)

			currentVal := 0

			for index, val := range output {
				currentVal += input[index]

				assertInstance.Equal(val, currentVal)
			}
		})
	}
}
