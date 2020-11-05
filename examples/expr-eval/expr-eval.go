package main

import (
	"fmt"

	bexpr "github.com/hashicorp/go-bexpr"
)

type Internal struct {
	Name string

	// Use an alternative name for referencing this field in expressions
	Values []int `bexpr:"fields"`

	// Hides this field so it cannot be used in expression evaluation
	Hidden int `bexpr:"-"`

	// Unexported fields are not available for use by the evaluator
	unexported int
}

type Matchable struct {
	Map      map[string]string
	X        int
	Internal Internal

	// Slices are handled specially. If any value within the slice
	// evaluates to true the overall evaluation for the slice will
	// be true.
	SliceInternal []Internal
	MapInternal   map[string]Internal
}

var data Matchable = Matchable{
	Map: map[string]string{
		"abc": "def",
		"ghi": "jkl",
		"mno": "pqr",
		"stu": "vwx",
		"y":   "z",
	},
	X: 5,
	Internal: Internal{
		Name:   "main",
		Values: []int{1, 2, 3, 4, 5},
	},
	SliceInternal: []Internal{
		Internal{
			Name:   "odd",
			Values: []int{1, 3, 5, 7, 9},
		},
		Internal{
			Name:   "even",
			Values: []int{2, 4, 6, 8, 10},
		},
		Internal{
			Name:   "fib",
			Values: []int{0, 1, 1, 2, 3, 5},
		},
	},
	MapInternal: map[string]Internal{
		"odd": Internal{
			Name:   "odd",
			Values: []int{1, 3, 5, 7, 9},
		},
		"even": Internal{
			Name:   "even",
			Values: []int{2, 4, 6, 8, 10},
		},
		"fib": Internal{
			Name:   "fib",
			Values: []int{0, 1, 1, 2, 3, 5},
		},
	},
}

var expressions []string = []string{
	// should error out in creating the evaluator as Foo is not a valid selector
	"Foo == 3",
	// should error out because the field is hidden
	"Internal.Hidden == 5",
	// should error out because the field is not exported
	"Internal.unexported == 3",
	// should evaluate to true
	"Map[`abc`] == `def`",
	// should evaluate to false
	"X == 3",
	// should evaluate to true
	"Internal.fields is not empty",
	// should evaluate to true
	"0 in SliceInternal.fields",
	// should evaluate to false
	"4 not in SliceInternal.fields",
	// should evaluate to false
	"MapInternal.fib.Name != fib",
	// should evaluate to true
	"odd in MapInternal",
}

func main() {
	for _, expression := range expressions {
		eval, err := bexpr.CreateEvaluatorForType(expression, nil, (*Matchable)(nil))

		if err != nil {
			fmt.Printf("Failed to create evaluator for expression %q: %v\n", expression, err)
			continue
		}

		result, err := eval.Evaluate(&data)
		if err != nil {
			fmt.Printf("Failed to run evaluation of expression %q: %v\n", expression, err)
			continue
		}

		fmt.Printf("Result of expression %q evaluation: %t\n", expression, result)
	}
}
