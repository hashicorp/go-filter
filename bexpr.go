// bexpr is an implementation of a generic boolean expression evaluator.
// The general goal is to be able to evaluate some expression against some
// arbitrary data and get back a boolean of whether or not the data
// was matched by the expression
package bexpr

//go:generate pigeon -o grammar/grammar.go -optimize-parser grammar/grammar.peg
//go:generate goimports -w grammar/grammar.go

import (
	"github.com/hashicorp/go-bexpr/grammar"
	"github.com/mitchellh/pointerstructure"
)

// HookFn provides a way to translate 1 reflect.Value to another during
// evaluation by the bexpr evluator.  This facilitate making go structures
// appear in a way that matches the expected jsonpointers used for evaluation.
// This is helpful, for example, when working with protocol buffers' well
// known types.
type HookFn pointerstructure.GetValueHookFn

type Evaluator struct {
	// The syntax tree
	ast     grammar.Expression
	tagName string
	hook    HookFn
}

func CreateEvaluator(expression string, opts ...Option) (*Evaluator, error) {
	parsedOpts := getOpts(opts...)
	var parserOpts []grammar.Option
	if parsedOpts.withMaxExpressions != 0 {
		parserOpts = append(parserOpts, grammar.MaxExpressions(parsedOpts.withMaxExpressions))
	}

	ast, err := grammar.Parse("", []byte(expression), parserOpts...)
	if err != nil {
		return nil, err
	}

	eval := &Evaluator{
		ast:     ast.(grammar.Expression),
		tagName: parsedOpts.withTagName,
		hook:    parsedOpts.withHookFn,
	}

	return eval, nil
}

func (eval *Evaluator) Evaluate(datum interface{}) (bool, error) {
	return evaluate(eval.ast, datum, WithTagName(eval.tagName), WithHookFn(eval.hook))
}
