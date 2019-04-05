package experiment

import (
	"fmt"
	"github.com/hung-phan/apps/lib/parallel"
)

type ITree interface {
	IsTree() bool
	GetVal() int
}

type Tree struct {
	Val int
}

func (tree *Tree) IsTree() bool {
	return true
}

func (tree *Tree) GetVal() int {
	return tree.Val
}

type Leaf struct {
	*Tree

	From  int
	Until int
}

type Node struct {
	*Tree

	Left  ITree
	Right ITree
}

func upsweepSequential(input []int, from, until int) int {
	sum := 0

	for i := from; i < until; i++ {
		sum += input[i]
	}

	return sum
}

// map func
func upsweep(input []int, from, until int, threshold int) ITree {
	if until-from < threshold {
		return Leaf{
			From:  from,
			Until: until,
			Tree: &Tree{
				Val: upsweepSequential(input, from, until),
			},
		}
	} else {
		mid := from + (until-from)/2

		res := parallel.Parallel(
			func() parallel.Val {
				return upsweep(input, from, mid, threshold)
			},
			func() parallel.Val {
				return upsweep(input, mid, until, threshold)
			},
		)

		left, right := res[0].(ITree), res[1].(ITree)

		return Node{
			Left:  left,
			Right: right,
			Tree: &Tree{
				Val: left.GetVal() + right.GetVal(),
			},
		}
	}
}

func downsweepSequential(
	input, output []int,
	startVal int,
	from, until int,
) {
	if from < until {
		currentVal := startVal + input[from]

		output[from] = currentVal

		downsweepSequential(input, output, currentVal, from+1, until)
	}
}

// reduce func
func downsweep(
	input, output []int,
	startVal int,
	tree ITree,
) {
	switch val := tree.(type) {

	case Leaf:
		downsweepSequential(input, output, startVal, val.From, val.Until)

	case Node:
		parallel.Parallel(
			func() parallel.Val {
				downsweep(input, output, startVal, val.Left)

				return nil
			},
			func() parallel.Val {
				downsweep(input, output, startVal+val.Left.GetVal(), val.Right)

				return nil
			},
		)

	default:
		panic(fmt.Sprintf("cannot determine the type of tree: %s", tree))
	}
}

func parallelScan(
	input, output []int,
	startVal int,
	threshold int,
) {
	downsweep(input, output, startVal, upsweep(input, 0, len(input), threshold))
}
