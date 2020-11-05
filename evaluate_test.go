package bexpr

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
)

type expressionCheck struct {
	expression string
	result     bool
	err        string
	benchQuick bool
}

type expressionTest struct {
	value       interface{}
	expressions []expressionCheck
}

var evaluateTests map[string]expressionTest = map[string]expressionTest{
	"Flat Struct": {
		testFlatStruct{
			Int:        -1,
			Int8:       -2,
			Int16:      -3,
			Int32:      -4,
			Int64:      -5,
			Uint:       6,
			Uint8:      7,
			Uint16:     8,
			Uint32:     9,
			Uint64:     10,
			Float32:    1.1,
			Float64:    1.2,
			Bool:       true,
			String:     "exported",
			unexported: "unexported",
			Hidden:     true,
		},
		[]expressionCheck{
			{expression: "Int == -1", result: true, benchQuick: true},
			{expression: "Int == -99", result: false, benchQuick: true},
			{expression: "Int != -1", result: false},
			{expression: "Int != -99", result: true},
			{expression: "Int8 == -2", result: true},
			{expression: "Int8 == -99", result: false},
			{expression: "Int8 != -2", result: false},
			{expression: "Int8 != -99", result: true},
			{expression: "Int16 == -3", result: true},
			{expression: "Int16 == -99", result: false},
			{expression: "Int16 != -3", result: false},
			{expression: "Int16 != -99", result: true},
			{expression: "Int32 == -4", result: true},
			{expression: "Int32 == -99", result: false},
			{expression: "Int32 != -4", result: false},
			{expression: "Int32 != -99", result: true},
			{expression: "Int64 == -5", result: true},
			{expression: "Int64 == -99", result: false},
			{expression: "Int64 != -5", result: false},
			{expression: "Int64 != -99", result: true},
			{expression: "Uint == 6", result: true},
			{expression: "Uint == 99", result: false},
			{expression: "Uint != 6", result: false},
			{expression: "Uint != 99", result: true},
			{expression: "Uint8 == 7", result: true},
			{expression: "Uint8 == 99", result: false},
			{expression: "Uint8 != 7", result: false},
			{expression: "Uint8 != 99", result: true},
			{expression: "Uint16 == 8", result: true},
			{expression: "Uint16 == 99", result: false},
			{expression: "Uint16 != 8", result: false},
			{expression: "Uint16 != 99", result: true},
			{expression: "Uint32 == 9", result: true},
			{expression: "Uint32 == 99", result: false},
			{expression: "Uint32 != 9", result: false},
			{expression: "Uint32 != 99", result: true},
			{expression: "Uint64 == 10", result: true},
			{expression: "Uint64 == 99", result: false},
			{expression: "Uint64 != 10", result: false, benchQuick: true},
			{expression: "Uint64 != 99", result: true, benchQuick: true},
			{expression: "Float32 == 1.1", result: true},
			{expression: "Float32 == 9.9", result: false},
			{expression: "Float32 != 1.1", result: false},
			{expression: "Float32 != 9.9", result: true},
			{expression: "Float64 == 1.2", result: true},
			{expression: "Float64 == 9.9", result: false},
			{expression: "Float64 != 1.2", result: false, benchQuick: true},
			{expression: "Float64 != 9.9", result: true, benchQuick: true},
			{expression: "Bool == true", result: true},
			{expression: "Bool == false", result: false},
			{expression: "Bool != true", result: false},
			{expression: "Bool != false", result: true},
			{expression: "String == `exported`", result: true, benchQuick: true},
			{expression: "String == `not-it`", result: false, benchQuick: true},
			{expression: "String != `exported`", result: false},
			{expression: "String != `not-it`", result: true},
			{expression: "port in String", result: true, benchQuick: true},
			{expression: "part in String", result: false},
			{expression: "port not in String", result: false},
			{expression: "part not in String", result: true},
			{expression: "unexported == `unexported`", result: false, err: "Selector \"unexported\" is not valid"},
			{expression: "Hidden == false", result: false, err: "Selector \"Hidden\" is not valid"},
			{expression: "String matches 	`^ex.*`", result: true, benchQuick: true},
			{expression: "String not matches `^anchored.*`", result: true, benchQuick: true},
			{expression: "String matches 	`^anchored.*`", result: false},
			{expression: "String not matches `^ex.*`", result: false},
		},
	},
	"Flat Struct Alt Types": {
		testFlatStructAlt{
			Int:        -1,
			Int8:       -2,
			Int16:      -3,
			Int32:      -4,
			Int64:      -5,
			Uint:       6,
			Uint8:      7,
			Uint16:     8,
			Uint32:     9,
			Uint64:     10,
			Float32:    1.1,
			Float64:    1.2,
			Bool:       true,
			String:     "exported",
			unexported: "unexported",
			Hidden:     true,
		},
		[]expressionCheck{
			{expression: "Int == -1", result: true, benchQuick: true},
			{expression: "Int == -99", result: false, benchQuick: true},
			{expression: "Int != -1", result: false},
			{expression: "Int != -99", result: true},
			{expression: "Int8 == -2", result: true},
			{expression: "Int8 == -99", result: false},
			{expression: "Int8 != -2", result: false},
			{expression: "Int8 != -99", result: true},
			{expression: "Int16 == -3", result: true},
			{expression: "Int16 == -99", result: false},
			{expression: "Int16 != -3", result: false},
			{expression: "Int16 != -99", result: true},
			{expression: "Int32 == -4", result: true},
			{expression: "Int32 == -99", result: false},
			{expression: "Int32 != -4", result: false},
			{expression: "Int32 != -99", result: true},
			{expression: "Int64 == -5", result: true},
			{expression: "Int64 == -99", result: false},
			{expression: "Int64 != -5", result: false},
			{expression: "Int64 != -99", result: true},
			{expression: "Uint == 6", result: true},
			{expression: "Uint == 99", result: false},
			{expression: "Uint != 6", result: false},
			{expression: "Uint != 99", result: true},
			{expression: "Uint8 == 7", result: true},
			{expression: "Uint8 == 99", result: false},
			{expression: "Uint8 != 7", result: false},
			{expression: "Uint8 != 99", result: true},
			{expression: "Uint16 == 8", result: true},
			{expression: "Uint16 == 99", result: false},
			{expression: "Uint16 != 8", result: false},
			{expression: "Uint16 != 99", result: true},
			{expression: "Uint32 == 9", result: true},
			{expression: "Uint32 == 99", result: false},
			{expression: "Uint32 != 9", result: false},
			{expression: "Uint32 != 99", result: true},
			{expression: "Uint64 == 10", result: true},
			{expression: "Uint64 == 99", result: false},
			{expression: "Uint64 != 10", result: false, benchQuick: true},
			{expression: "Uint64 != 99", result: true, benchQuick: true},
			{expression: "Float32 == 1.1", result: true},
			{expression: "Float32 == 9.9", result: false},
			{expression: "Float32 != 1.1", result: false},
			{expression: "Float32 != 9.9", result: true},
			{expression: "Float64 == 1.2", result: true},
			{expression: "Float64 == 9.9", result: false},
			{expression: "Float64 != 1.2", result: false, benchQuick: true},
			{expression: "Float64 != 9.9", result: true, benchQuick: true},
			{expression: "Bool == true", result: true},
			{expression: "Bool == false", result: false},
			{expression: "Bool != true", result: false},
			{expression: "Bool != false", result: true},
			{expression: "String == `exported`", result: true, benchQuick: true},
			{expression: "String == `not-it`", result: false, benchQuick: true},
			{expression: "String != `exported`", result: false},
			{expression: "String != `not-it`", result: true},
			{expression: "unexported == `unexported`", result: false, err: "Selector \"unexported\" is not valid"},
			{expression: "Hidden == false", result: false, err: "Selector \"Hidden\" is not valid"},
		},
	},
	"map[string]map[string]bool": {
		map[string]map[string]bool{
			"foo": {
				"bar": true,
				"baz": false,
			},
			"abc": nil,
		},
		[]expressionCheck{
			{expression: "bar in foo", result: true},
			{expression: "arg in foo", result: false},
			{expression: "arg not in foo", result: true},
			{expression: "baz not in foo", result: false},
			{expression: "foo is empty", result: false},
			{expression: "foo is not empty", result: true},
			{expression: "abc is empty", result: true},
			{expression: "abc is not empty", result: false},
			{expression: "foo in abc", result: false, benchQuick: true},
			{expression: "foo not in abc", result: true},
			{expression: "foo.bar == true", result: true},
			{expression: "foo.bar == false", result: false},
			{expression: "foo.baz == false", result: true},
			{expression: "foo.baz == true", result: false, benchQuick: true},
			{expression: "foo.bar != true", result: false},
			{expression: "foo.bar != false", result: true},
			{expression: "foo.baz != false", result: false},
			{expression: "foo.baz != true", result: true},
			{expression: "foo.bar.baz == 3", result: false, err: "Selector \"foo.bar.baz\" is not valid"},
		},
	},
	"Nested Structs and Maps": {
		testNestedTypes{
			Nested: testNestedLevel1{
				Map: map[string]string{
					"foo": "bar",
					"bar": "baz",
					"abc": "123",
				},
				MapOfStructs: map[string]testNestedLevel2_1{
					"one": testNestedLevel2_1{
						Foo: 42,
						bar: "unexported",
						Baz: "exported",
					},
					"two": testNestedLevel2_1{
						Foo: 77,
						bar: "unexported",
						Baz: "consul",
					},
				},
				SliceOfInts: []int{1, 3, 5, 7, 9},
				SliceOfStructs: []testNestedLevel2_2{
					testNestedLevel2_2{
						X: 1,
						Y: 2,
						z: 10,
					},
					testNestedLevel2_2{
						X: 3,
						Y: 5,
						z: 10,
					},
				},
				SliceOfMapInfInf: []map[interface{}]interface{}{
					map[interface{}]interface{}{
						1: 2,
					},
				},
			},
			TopInt: 5,
		},
		[]expressionCheck{
			{expression: "Nested.Map.foo == bar", result: true, benchQuick: true},
			{expression: "Nested.Map.foo contains ba", result: true, benchQuick: true},
			{expression: "Nested.Map.foo == baz", result: false},
			{expression: "Nested.Map is not empty", result: true},
			{expression: "Nested.Map is not empty", result: true},
			{expression: "Nested.Map contains foo and Nested.Map contains bar", result: true, benchQuick: true},
			{expression: "Nested.Map contains nope", result: false},
			{expression: "Nested.Map contains bar", result: true},
			{expression: "Nested.Map.bar == `bazel`", result: false, benchQuick: true},
			{expression: "TopInt != 0", result: true},
			{expression: "Nested.Map contains nope or (Nested.Map contains bar and Nested.Map.bar == `bazel`) or TopInt != 0", result: true, benchQuick: true},
			{expression: "Nested.MapOfStructs.one.Foo == 42", result: true},
			{expression: "Nested.MapOfStructs is empty or (Nested.SliceOfInts contains 7 and 9 in Nested.SliceOfInts)", result: true, benchQuick: true},
			{expression: "Nested.SliceOfStructs.X == 1", result: true},
			{expression: "Nested.SliceOfStructs.Y == 4", result: false},
			{expression: "Nested.Map.notfound == 4", result: false},
			{expression: "Map in Nested", result: false, err: "Invalid match operator \"In\" for selector \"Nested\""},
			{expression: "Nested.MapInfInf.foo == 4", result: false, err: "Selector \"Nested.MapInfInf.foo\" is not valid"},
			{expression: "Nested.SliceOfMapInfInf.foo == 4", result: false, err: "Selector \"Nested.SliceOfMapInfInf.foo\" is not valid"},
			{expression: "any Nested.Map as key { key == foo }", result: true},
			{expression: "all Nested.Map as _, value { value != foo }", result: true},
			{expression: "all Nested.MapOfStructs as key, value { (key == one and value.Foo == 42) or (key == two and value.Foo == 77) }", result: true, benchQuick: true},
			{expression: "any Nested.SliceOfInts as i { i == 7 }", result: true, benchQuick: true},
			{expression: "all Nested.SliceOfStructs as i, v { i == 0 and v.X == 1 or i == 1 and v.X == 3 }", result: true},
		},
	},
}

func TestEvaluate(t *testing.T) {
	for name, tcase := range evaluateTests {
		// capture these values in the closure
		name := name
		tcase := tcase
		t.Run(name, func(t *testing.T) {
			for i, expTest := range tcase.expressions {
				// capture these values in the closure
				expTest := expTest
				t.Run(fmt.Sprintf("#%d", i), func(t *testing.T) {
					expr, err := CreateEvaluator(expTest.expression, nil)
					require.NoError(t, err)

					match, err := expr.Evaluate(tcase.value)
					if expTest.err != "" {
						require.Error(t, err)
						require.EqualError(t, err, expTest.err)
					} else {
						require.NoError(t, err)
					}
					require.Equal(t, expTest.result, match)
				})
			}
		})
	}
}

func BenchmarkEvaluate(b *testing.B) {
	for name, tcase := range evaluateTests {
		// capture these values in the closure
		name := name
		tcase := tcase
		b.Run(name, func(b *testing.B) {
			for i, expTest := range tcase.expressions {
				// capture these values in the closure
				expTest := expTest
				b.Run(fmt.Sprintf("#%d", i), func(b *testing.B) {
					if !expTest.benchQuick && !FullBenchmarks() {
						b.Skip("Skipping benchmark - rerun with -bench-full to enable")
					}

					expr, err := CreateEvaluatorForType(expTest.expression, nil, tcase.value)
					require.NoError(b, err)

					b.ResetTimer()
					for n := 0; n < b.N; n++ {
						_, err = expr.Evaluate(tcase.value)
						if expTest.err != "" {
							require.Error(b, err)
						} else {
							require.NoError(b, err)
						}
					}
				})
			}
		})
	}
}

func BenchmarkValidate(b *testing.B) {
	for name, tcase := range evaluateTests {
		// capture these values in the closure
		name := name
		tcase := tcase
		b.Run(name, func(b *testing.B) {
			for i, expTest := range tcase.expressions {
				// capture these values in the closure
				expTest := expTest
				b.Run(fmt.Sprintf("#%d", i), func(b *testing.B) {
					if !expTest.benchQuick && !FullBenchmarks() {
						b.Skip("Skipping benchmark - rerun with -bench-full to enable")
					}

					expr, err := CreateEvaluator(expTest.expression, nil)
					require.NoError(b, err)

					b.ResetTimer()
					for n := 0; n < b.N; n++ {
						err = expr.Validate(&expr.config, tcase.value)
						if expTest.err != "" {
							require.Error(b, err)
						} else {
							require.NoError(b, err)
						}
					}
				})
			}
		})
	}
}
