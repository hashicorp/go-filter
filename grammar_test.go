package bexpr

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestExpressionParsing(t *testing.T) {
	t.Parallel()
	type testCase struct {
		input    string
		expected Expression
		err      string
	}

	tests := map[string]testCase{
		"Match Equality": {
			input:    "foo == 3",
			expected: &MatchExpression{Selector: Selector{"foo"}, Operator: MatchEqual, Value: &MatchValue{Raw: "3"}},
			err:      "",
		},
		"Match Inequality": {
			input:    "foo != xyz",
			expected: &MatchExpression{Selector: Selector{"foo"}, Operator: MatchNotEqual, Value: &MatchValue{Raw: "xyz"}},
			err:      "",
		},
		"Match Is Empty": {
			input:    "list is empty",
			expected: &MatchExpression{Selector: Selector{"list"}, Operator: MatchIsEmpty, Value: nil},
			err:      "",
		},
		"Match Is Not Empty": {
			input:    "list is not empty",
			expected: &MatchExpression{Selector: Selector{"list"}, Operator: MatchIsNotEmpty, Value: nil},
			err:      "",
		},
		"Match In": {
			input:    "foo in bar",
			expected: &MatchExpression{Selector: Selector{"bar"}, Operator: MatchIn, Value: &MatchValue{Raw: "foo"}},
			err:      "",
		},
		"Match Not In": {
			input:    "foo not in bar",
			expected: &MatchExpression{Selector: Selector{"bar"}, Operator: MatchNotIn, Value: &MatchValue{Raw: "foo"}},
			err:      "",
		},
		"Match Contains": {
			input:    "bar contains foo",
			expected: &MatchExpression{Selector: Selector{"bar"}, Operator: MatchIn, Value: &MatchValue{Raw: "foo"}},
			err:      "",
		},
		"Match Not Contains": {
			input:    "bar not contains foo",
			expected: &MatchExpression{Selector: Selector{"bar"}, Operator: MatchNotIn, Value: &MatchValue{Raw: "foo"}},
			err:      "",
		},
		"Match Matches": {
			input:    "foo matches bar",
			expected: &MatchExpression{Selector: Selector{"foo"}, Operator: MatchMatches, Value: &MatchValue{Raw: "bar"}},
			err:      "",
		},
		"Match Not Matches": {
			input:    "foo not matches bar",
			expected: &MatchExpression{Selector: Selector{"foo"}, Operator: MatchNotMatches, Value: &MatchValue{Raw: "bar"}},
			err:      "",
		},
		"Logical Not": {
			input: "not prod in tags",
			expected: &UnaryExpression{
				Operator: UnaryOpNot,
				Operand:  &MatchExpression{Selector: Selector{"tags"}, Operator: MatchIn, Value: &MatchValue{Raw: "prod"}},
			},
			err: "",
		},
		"Logical And": {
			input: "port != 80 and port != 8080",
			expected: &BinaryExpression{
				Operator: BinaryOpAnd,
				Left:     &MatchExpression{Selector: Selector{"port"}, Operator: MatchNotEqual, Value: &MatchValue{Raw: "80"}},
				Right:    &MatchExpression{Selector: Selector{"port"}, Operator: MatchNotEqual, Value: &MatchValue{Raw: "8080"}},
			},
			err: "",
		},
		"Logical Or": {
			input: "port == 80 or port == 443",
			expected: &BinaryExpression{
				Operator: BinaryOpOr,
				Left:     &MatchExpression{Selector: Selector{"port"}, Operator: MatchEqual, Value: &MatchValue{Raw: "80"}},
				Right:    &MatchExpression{Selector: Selector{"port"}, Operator: MatchEqual, Value: &MatchValue{Raw: "443"}},
			},
			err: "",
		},
		"Any Expression": {
			input: "any foo.bar as elem { `proxy` in elem }",
			expected: &CollectionExpression{
				Operator: CollectionOpAny,
				Selector: Selector{"foo", "bar"},
				NameBinding: CollectionNameBinding{
					Mode:    CollectionBindDefault,
					Default: "elem",
				},
				Expression: &MatchExpression{
					Selector: Selector{"elem"},
					Operator: MatchIn,
					Value:    &MatchValue{Raw: "proxy"},
				},
			},
			err: "",
		},
		"Any Expression - 2 idents": {
			input: "any foo.bar as index, value { index contains proxy }",
			expected: &CollectionExpression{
				Operator: CollectionOpAny,
				Selector: Selector{"foo", "bar"},
				NameBinding: CollectionNameBinding{
					Mode:  CollectionBindIndexAndValue,
					Index: "index",
					Value: "value",
				},
				Expression: &MatchExpression{
					Selector: Selector{"index"},
					Operator: MatchIn,
					Value:    &MatchValue{Raw: "proxy"},
				},
			},
			err: "",
		},
		"Any Expression - 2 idents first underscore": {
			input: "any foo.bar as _, value { value contains proxy }",
			expected: &CollectionExpression{
				Operator: CollectionOpAny,
				Selector: Selector{"foo", "bar"},
				NameBinding: CollectionNameBinding{
					Mode:  CollectionBindValue,
					Value: "value",
				},
				Expression: &MatchExpression{
					Selector: Selector{"value"},
					Operator: MatchIn,
					Value:    &MatchValue{Raw: "proxy"},
				},
			},
			err: "",
		},
		"Any Expression - 2 idents second underscore": {
			input: "any foo.bar as index, _ { index contains proxy }",
			expected: &CollectionExpression{
				Operator: CollectionOpAny,
				Selector: Selector{"foo", "bar"},
				NameBinding: CollectionNameBinding{
					Mode:  CollectionBindIndex,
					Index: "index",
				},
				Expression: &MatchExpression{
					Selector: Selector{"index"},
					Operator: MatchIn,
					Value:    &MatchValue{Raw: "proxy"},
				},
			},
			err: "",
		},
		"All Expression": {
			input: "all foo.bar as elem { elem.baz contains something}",
			expected: &CollectionExpression{
				Operator: CollectionOpAll,
				Selector: Selector{"foo", "bar"},
				NameBinding: CollectionNameBinding{
					Mode:    CollectionBindDefault,
					Default: "elem",
				},
				Expression: &MatchExpression{
					Selector: Selector{"elem", "baz"},
					Operator: MatchIn,
					Value:    &MatchValue{Raw: "something"},
				},
			},
			err: "",
		},
		"All Expression - 2 idents": {
			input: "all foo.bar as index, elem { elem.baz contains something}",
			expected: &CollectionExpression{
				Operator: CollectionOpAll,
				Selector: Selector{"foo", "bar"},
				NameBinding: CollectionNameBinding{
					Mode:  CollectionBindIndexAndValue,
					Index: "index",
					Value: "elem",
				},
				Expression: &MatchExpression{
					Selector: Selector{"elem", "baz"},
					Operator: MatchIn,
					Value:    &MatchValue{Raw: "something"},
				},
			},
			err: "",
		},
		"All Expression - 2 idents first underscore": {
			input: "all foo.bar as _, elem { elem.baz contains something}",
			expected: &CollectionExpression{
				Operator: CollectionOpAll,
				Selector: Selector{"foo", "bar"},
				NameBinding: CollectionNameBinding{
					Mode:  CollectionBindValue,
					Value: "elem",
				},
				Expression: &MatchExpression{
					Selector: Selector{"elem", "baz"},
					Operator: MatchIn,
					Value:    &MatchValue{Raw: "something"},
				},
			},
			err: "",
		},
		"All Expression - 2 idents second underscore": {
			input: "all foo.bar as index, _ { index contains something}",
			expected: &CollectionExpression{
				Operator: CollectionOpAll,
				Selector: Selector{"foo", "bar"},
				NameBinding: CollectionNameBinding{
					Mode:  CollectionBindIndex,
					Index: "index",
				},
				Expression: &MatchExpression{
					Selector: Selector{"index"},
					Operator: MatchIn,
					Value:    &MatchValue{Raw: "something"},
				},
			},
			err: "",
		},
		"Nested Collection Expression": {
			input: "any foo.bar as elem { all elem.Checks as chk { chk.Status == passing} }",
			expected: &CollectionExpression{
				Operator: CollectionOpAny,
				Selector: Selector{"foo", "bar"},
				NameBinding: CollectionNameBinding{
					Mode:    CollectionBindDefault,
					Default: "elem",
				},
				Expression: &CollectionExpression{
					Operator: CollectionOpAll,
					Selector: Selector{"elem", "Checks"},
					NameBinding: CollectionNameBinding{
						Mode:    CollectionBindDefault,
						Default: "chk",
					},
					Expression: &MatchExpression{
						Selector: Selector{"chk", "Status"},
						Operator: MatchEqual,
						Value:    &MatchValue{Raw: "passing"},
					},
				},
			},
			err: "",
		},
		"Double Quoted Value (Equal)": {
			input:    "foo == \"bar\"",
			expected: &MatchExpression{Selector: Selector{"foo"}, Operator: MatchEqual, Value: &MatchValue{Raw: "bar"}},
			err:      "",
		},
		"Double Quoted Value (Not Equal)": {
			input:    "foo != \"bar\"",
			expected: &MatchExpression{Selector: Selector{"foo"}, Operator: MatchNotEqual, Value: &MatchValue{Raw: "bar"}},
			err:      "",
		},
		"Double Quoted Value (In)": {
			input:    "\"foo\" in bar",
			expected: &MatchExpression{Selector: Selector{"bar"}, Operator: MatchIn, Value: &MatchValue{Raw: "foo"}},
			err:      "",
		},
		"Double Quoted Value (Not In)": {
			input:    "\"foo\" not in bar",
			expected: &MatchExpression{Selector: Selector{"bar"}, Operator: MatchNotIn, Value: &MatchValue{Raw: "foo"}},
			err:      "",
		},
		"Backtick Quoted Value (Equal)": {
			input:    "foo == `bar`",
			expected: &MatchExpression{Selector: Selector{"foo"}, Operator: MatchEqual, Value: &MatchValue{Raw: "bar"}},
			err:      "",
		},
		"Backtick Quoted Value (Not Equal)": {
			input:    "foo != `bar`",
			expected: &MatchExpression{Selector: Selector{"foo"}, Operator: MatchNotEqual, Value: &MatchValue{Raw: "bar"}},
			err:      "",
		},
		"Backtick Quoted Value (In)": {
			input:    "`foo` in bar",
			expected: &MatchExpression{Selector: Selector{"bar"}, Operator: MatchIn, Value: &MatchValue{Raw: "foo"}},
			err:      "",
		},
		"Backtick Quoted Value (Not In)": {
			input:    "`foo` not in bar",
			expected: &MatchExpression{Selector: Selector{"bar"}, Operator: MatchNotIn, Value: &MatchValue{Raw: "foo"}},
			err:      "",
		},
		// This is standard boolean expression precedence
		//
		// 1 - not
		// 2 - and
		// 3 - or
		"Logical Operator Precedence": {
			input: "x in foo and not str == something or list is empty",
			expected: &BinaryExpression{
				Operator: BinaryOpOr,
				Left: &BinaryExpression{
					Operator: BinaryOpAnd,
					Left:     &MatchExpression{Selector: Selector{"foo"}, Operator: MatchIn, Value: &MatchValue{Raw: "x"}},
					Right: &UnaryExpression{
						Operator: UnaryOpNot,
						Operand:  &MatchExpression{Selector: Selector{"str"}, Operator: MatchEqual, Value: &MatchValue{Raw: "something"}},
					},
				},
				Right: &MatchExpression{Selector: Selector{"list"}, Operator: MatchIsEmpty, Value: nil},
			},
			err: "",
		},
		// not in the absence of parentheses would normally get applied
		// to a MatchExpr
		//
		// or operators normal are the last to be applied but here they
		// happen earlier
		"Logical Operator Precedence (Parenthesized)": {
			input: "x in foo and not (str == something or list is empty)",
			expected: &BinaryExpression{
				Operator: BinaryOpAnd,
				Left:     &MatchExpression{Selector: Selector{"foo"}, Operator: MatchIn, Value: &MatchValue{Raw: "x"}},
				Right: &UnaryExpression{
					Operator: UnaryOpNot,
					Operand: &BinaryExpression{
						Operator: BinaryOpOr,
						Left:     &MatchExpression{Selector: Selector{"str"}, Operator: MatchEqual, Value: &MatchValue{Raw: "something"}},
						Right:    &MatchExpression{Selector: Selector{"list"}, Operator: MatchIsEmpty, Value: nil},
					},
				},
			},
			err: "",
		},
		"Extra Whitespace (Equal)": {
			input:    "\t\r\n  foo \t\r\n == \t\r\n x \t\r\n",
			expected: &MatchExpression{Selector: Selector{"foo"}, Operator: MatchEqual, Value: &MatchValue{Raw: "x"}},
			err:      "",
		},
		"Extra Whitespace (Not Equal)": {
			input:    "\t\r\n  foo \t\r\n != \t\r\n x \t\r\n",
			expected: &MatchExpression{Selector: Selector{"foo"}, Operator: MatchNotEqual, Value: &MatchValue{Raw: "x"}},
			err:      "",
		},
		"Extra Whitespace (In)": {
			input:    "\t\r\n  foo \t\r\n in \t\r\n x \t\r\n",
			expected: &MatchExpression{Selector: Selector{"x"}, Operator: MatchIn, Value: &MatchValue{Raw: "foo"}},
			err:      "",
		},
		"Extra Whitespace (Not In)": {
			input:    "\t\r\n  foo \t\r\n not \t\r\n in \t\r\n x \t\r\n",
			expected: &MatchExpression{Selector: Selector{"x"}, Operator: MatchNotIn, Value: &MatchValue{Raw: "foo"}},
			err:      "",
		},
		"Extra Whitespace (Is Empty)": {
			input:    "\t\r\n  foo \t\r\n is \t\r\n empty \t\r\n",
			expected: &MatchExpression{Selector: Selector{"foo"}, Operator: MatchIsEmpty, Value: nil},
			err:      "",
		},
		"Extra Whitespace (Is Not Empty)": {
			input:    "\t\r\n  foo \t\r\n is \t\r\n not \t\r\n empty \t\r\n",
			expected: &MatchExpression{Selector: Selector{"foo"}, Operator: MatchIsNotEmpty, Value: nil},
			err:      "",
		},
		"Extra Whitespace (Not)": {
			input: "\t\r\n not \t\r\n  foo \t\r\n in \t\r\n x \t\r\n",
			expected: &UnaryExpression{
				Operator: UnaryOpNot,
				Operand:  &MatchExpression{Selector: Selector{"x"}, Operator: MatchIn, Value: &MatchValue{Raw: "foo"}},
			},
			err: "",
		},
		"Extra Whitespace (And)": {
			input: "\t\r\n foo \t\r\n == \t\r\n x \t\r\n and \t\r\n y \t\r\n is \t\r\n empty \t\r\n",
			expected: &BinaryExpression{
				Operator: BinaryOpAnd,
				Left:     &MatchExpression{Selector: Selector{"foo"}, Operator: MatchEqual, Value: &MatchValue{Raw: "x"}},
				Right:    &MatchExpression{Selector: Selector{"y"}, Operator: MatchIsEmpty, Value: nil},
			},
			err: "",
		},
		"Extra Whitespace (Or)": {
			input: "\t\r\n foo \t\r\n == \t\r\n x \t\r\n or \t\r\n y \t\r\n is \t\r\n empty \t\r\n",
			expected: &BinaryExpression{
				Operator: BinaryOpOr,
				Left:     &MatchExpression{Selector: Selector{"foo"}, Operator: MatchEqual, Value: &MatchValue{Raw: "x"}},
				Right:    &MatchExpression{Selector: Selector{"y"}, Operator: MatchIsEmpty, Value: nil},
			},
			err: "",
		},
		"Extra Whitespace (Parentheses)": {
			input:    "\t\r\n ( \t\r\n foo \t\r\n == \t\r\n x \t\r\n ) \t\r\n",
			expected: &MatchExpression{Selector: Selector{"foo"}, Operator: MatchEqual, Value: &MatchValue{Raw: "x"}},
			err:      "",
		},
		"Selector Path": {
			input:    "`environment` in foo.bar[\"meta\"].tags[\t`ENV` ]",
			expected: &MatchExpression{Selector: Selector{"foo", "bar", "meta", "tags", "ENV"}, Operator: MatchIn, Value: &MatchValue{Raw: "environment"}},
			err:      "",
		},
		"Selector All Indexes": {
			input:    `environment in foo["bar"]["meta"]["tags"]["ENV"]`,
			expected: &MatchExpression{Selector: Selector{"foo", "bar", "meta", "tags", "ENV"}, Operator: MatchIn, Value: &MatchValue{Raw: "environment"}},
			err:      "",
		},
		"Selector All Dotted": {
			input:    "environment in foo.bar.meta.tags.ENV",
			expected: &MatchExpression{Selector: Selector{"foo", "bar", "meta", "tags", "ENV"}, Operator: MatchIn, Value: &MatchValue{Raw: "environment"}},
			err:      "",
		},
		// selectors can contain almost any character set when index expressions are used
		// This includes whitespace, hyphens, unicode, etc.
		"Selector Index Chars": {
			input:    "environment in foo[\"abc-def ghi åß∂ƒ\"]",
			expected: &MatchExpression{Selector: Selector{"foo", "abc-def ghi åß∂ƒ"}, Operator: MatchIn, Value: &MatchValue{Raw: "environment"}},
			err:      "",
		},
		"Unterminated String Literal 1": {
			input:    "foo == \"12x",
			expected: nil,
			err:      "1:12 (11): rule \"string\": Unterminated string literal",
		},
		"Unterminated String Literal 2": {
			input:    "foo == `12x",
			expected: nil,
			err:      "1:12 (11): rule \"string\": Unterminated string literal",
		},
		"Invalid Number": {
			input:    "foo == 3x",
			expected: nil,
			err:      "1:9 (8): rule \"number\": Invalid number literal",
		},
		"Invalid Index Key": {
			input:    "foo[3] == abc",
			expected: nil,
			err:      "1:5 (4): rule \"index\": Invalid index",
		},
		"Unclosed Index Expression 1": {
			input:    "x in foo[\"abc\"",
			expected: nil,
			err:      "1:15 (14): rule \"index\": Unclosed index expression",
		},
		"Unclosed Index Expression 2": {
			input:    "foo[\"abc\" == 3",
			expected: nil,
			err:      "1:11 (10): rule \"index\": Unclosed index expression",
		},
		"Invalid Selector 1": {
			input:    "x in 32",
			expected: nil,
			err:      "1:6 (5): rule \"match\": Invalid selector",
		},
		"Invalid Selector 2": {
			input:    "32 == 32",
			expected: nil,
			err:      `1:4 (3): no match found, expected: "in", "not" or [ \t\r\n]`,
		},
		"Invalid Selector 3": {
			input:    "32 is empty",
			expected: nil,
			err:      `1:4 (3): no match found, expected: "in", "not" or [ \t\r\n]`,
		},
		"Junk at the end 1": {
			input:    "x in foo abc",
			expected: nil,
			err:      `1:10 (9): no match found, expected: "and", "or", [ \t\r\n] or EOF`,
		},
		"Junk at the end 2": {
			input:    "x in foo and ",
			expected: nil,
			err:      "1:14 (13): no match found, expected: \"(\", \"-\", \"0\", \"\\\"\", \"`\", \"all\", \"any\", \"not\", [ \\t\\r\\n], [1-9] or [a-zA-Z]",
		},
		"Junk at the end 3": {
			input:    "x in foo or ",
			expected: nil,
			err:      "1:13 (12): no match found, expected: \"(\", \"-\", \"0\", \"\\\"\", \"`\", \"all\", \"any\", \"not\", [ \\t\\r\\n], [1-9] or [a-zA-Z]",
		},
		"Junk at the end 4": {
			input:    "x in foo or not ",
			expected: nil,
			err:      "1:17 (16): no match found, expected: \"!=\", \"(\", \"-\", \"0\", \"==\", \"\\\"\", \"`\", \"all\", \"any\", \"contains\", \"in\", \"is\", \"matches\", \"not\", [ \\t\\r\\n], [1-9] or [a-zA-Z]",
		},
		"Float Literal 1": {
			input:    "foo == 0.2",
			expected: &MatchExpression{Selector: Selector{"foo"}, Operator: MatchEqual, Value: &MatchValue{Raw: "0.2"}},
			err:      "",
		},
		"Float Literal 2": {
			input:    "foo == 11.11",
			expected: &MatchExpression{Selector: Selector{"foo"}, Operator: MatchEqual, Value: &MatchValue{Raw: "11.11"}},
			err:      "",
		},
		"Negative Float": {
			input:    "foo == -0.2",
			expected: &MatchExpression{Selector: Selector{"foo"}, Operator: MatchEqual, Value: &MatchValue{Raw: "-0.2"}},
			err:      "",
		},
		"Unmatched Parentheses": {
			input:    "(foo == 4",
			expected: nil,
			err:      "1:10 (9): rule \"grouping\": Unmatched parentheses",
		},
		"Double Not": {
			input:    "not not foo == 3",
			expected: &MatchExpression{Selector: Selector{"foo"}, Operator: MatchEqual, Value: &MatchValue{Raw: "3"}},
			err:      "",
		},
		"Complex": {
			input: "(((foo == 3) and (not ((bar in baz) and (not (one != two))))) or (((next is empty) and (not (foo is not empty))) and (bar not in foo)))",
			expected: &BinaryExpression{
				Operator: BinaryOpOr,
				Left: &BinaryExpression{
					Operator: BinaryOpAnd,
					Left:     &MatchExpression{Selector: Selector{"foo"}, Operator: MatchEqual, Value: &MatchValue{Raw: "3"}},
					Right: &UnaryExpression{
						Operator: UnaryOpNot,
						Operand: &BinaryExpression{
							Operator: BinaryOpAnd,
							Left:     &MatchExpression{Selector: Selector{"baz"}, Operator: MatchIn, Value: &MatchValue{Raw: "bar"}},
							Right: &UnaryExpression{
								Operator: UnaryOpNot,
								Operand:  &MatchExpression{Selector: Selector{"one"}, Operator: MatchNotEqual, Value: &MatchValue{Raw: "two"}},
							},
						},
					},
				},
				Right: &BinaryExpression{
					Operator: BinaryOpAnd,
					Left: &BinaryExpression{
						Operator: BinaryOpAnd,
						Left:     &MatchExpression{Selector: Selector{"next"}, Operator: MatchIsEmpty, Value: nil},
						Right: &UnaryExpression{
							Operator: UnaryOpNot,
							Operand:  &MatchExpression{Selector: Selector{"foo"}, Operator: MatchIsNotEmpty, Value: nil},
						},
					},
					Right: &MatchExpression{Selector: Selector{"foo"}, Operator: MatchNotIn, Value: &MatchValue{Raw: "bar"}},
				},
			},
			err: "",
		},
	}

	for name, tcase := range tests {
		tcase := tcase
		name := name
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			raw, err := Parse("", []byte(tcase.input))
			if tcase.err != "" {
				require.Error(t, err)
				require.EqualError(t, err, tcase.err)
				require.Nil(t, raw)
			} else {
				require.NoError(t, err)
				require.NotNil(t, raw)
				require.Equal(t, tcase.expected, raw)
			}
		})
	}
}

func BenchmarkExpressionParsing(b *testing.B) {
	benchmarks := map[string]string{
		"Equals":                "foo == 3",
		"Not Equals":            "foo != 3",
		"In":                    "foo in bar",
		"Not In":                "foo not in bar",
		"Contains":              "bar not contains foo",
		"Not Contains":          "bar not contains foo",
		"Is Empty":              "foo is empty",
		"Is Not Empty":          "foo is not empty",
		"Not In Or Equals":      "foo not in bar or bar.foo == 3",
		"In And Not Equals":     "foo in bar and bar.foo != \"\"",
		"Not Equals And Equals": "not (foo == 3 and bar == 4)",
		"Matches":               "foo matches bar",
		"Not Matches":           "foo not matches bar",
		"Big Selectors":         "abcdefghijklmnopqrstuvwxyz.foo.bar.baz.one.two.three.four.five.six.seven.eight.nine.ten == 42",
		"Many Ors":              "foo == 3 or bar in baz or one != two or next is empty or other is not empty or name == \"\"",
		"Lots of Ops":           "foo == 3 and not bar in baz and not one != two or next is empty and not foo is not empty and bar not in foo",
		"Lots of Parens":        "(((foo == 3) and (not ((bar in baz) and (not (one != two))))) or (((next is empty) and (not (foo is not empty))) and (bar not in foo)))",
	}
	for name, bm := range benchmarks {
		b.Run(name, func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				expr, err := Parse("", []byte(bm))
				// this extra verification does add roughly 2k-3k ns/op to each iteration
				// we could disable it but its good to also have these checks to ensure the parser is working for some of the crazier inputs here
				require.NoError(b, err)
				require.NotNil(b, expr)
			}
		})
	}
}
