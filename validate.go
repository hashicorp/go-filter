package bexpr

import (
	"fmt"
	"regexp"
)

type selectorProcessFn func(idx int, field string, cfg *FieldConfiguration) bool

func processSelector(sel Selector, fields FieldConfigurations, process selectorProcessFn) (*FieldConfiguration, error) {
	configs := fields
	var lastConfig *FieldConfiguration

	if len(sel) < 1 {
		return nil, fmt.Errorf("Invalid selector: %q", sel)
	}

	for idx, field := range sel {
		if fcfg, ok := configs[FieldName(field)]; ok {
			lastConfig = fcfg
			configs = fcfg.SubFields
		} else if fcfg, ok := configs[FieldNameAny]; ok {
			lastConfig = fcfg
			configs = fcfg.SubFields
		} else {
			return nil, fmt.Errorf("Selector %q is not valid", sel[:idx+1])
		}

		// this just verifies that the FieldConfigurations we are using was created properly
		if lastConfig == nil {
			return nil, fmt.Errorf("FieldConfiguration for selector %q is nil", sel[:idx+1])
		}

		if process != nil && !process(idx, field, lastConfig) {
			break
		}
	}

	return lastConfig, nil
}

func expandImplicitCollectionOp(exp SelectorExpression, fields FieldConfigurations) (Expression, error) {
	sel := exp.GetSelector()
	internalMaxThresh := len(sel) - 1

	_, err := processSelector(sel, fields, func(idx int, field string, cfg *FieldConfiguration) bool {
		if idx < internalMaxThresh && cfg.CollectionType == CollectionTypeList {
			exp.SetSelector(append(Selector{"v"}, sel[idx+1:]...))

			exp = &CollectionExpression{
				Operator: CollectionOpAny,
				Selector: sel[:idx+1], // idx + 1 because the range is exclusive of the end index
				NameBinding: CollectionNameBinding{
					Mode:  CollectionBindValue,
					Value: "v",
				},
				Expression: exp.(Expression),
			}
			return false
		}
		return true
	})

	return exp.(Expression), err
}

func validateSelector(sel Selector, fields FieldConfigurations) (*FieldConfiguration, error) {
	return processSelector(sel, fields, nil)
}

func validateSelectorExpression(selectorExp SelectorExpression, fields FieldConfigurations, maxRawValueLength int) (int, Expression, error) {
	cfg, err := validateSelector(selectorExp.GetSelector(), fields)
	if err != nil {
		return 0, selectorExp.(Expression), err
	}

	switch exp := selectorExp.(type) {
	case *CollectionExpression:
		// need to validate that it supports collection ops
		if cfg.CollectionType == CollectionTypeNone {
			return 0, selectorExp.(Expression), fmt.Errorf("%s expression not supported for selector %q", exp.Operator, exp.Selector)
		}

		expFields := make(FieldConfigurations)
		ctype := cfg.CollectionType

		// synthesize the fields config for the key/value pair that can be used within the
		// collection expression
		switch exp.NameBinding.Mode {
		case CollectionBindDefault:
			if ctype == CollectionTypeMap {
				if cfg.IndexConfiguration != nil {
					expFields[FieldName(exp.NameBinding.Index)] = cfg.IndexConfiguration
				}
			} else {
				if cfg.ValueConfiguration != nil {
					expFields[FieldName(exp.NameBinding.Value)] = cfg.ValueConfiguration
				}
			}
		case CollectionBindIndex:
			if cfg.IndexConfiguration != nil {
				expFields[FieldName(exp.NameBinding.Index)] = cfg.IndexConfiguration
			}
		case CollectionBindValue:
			if cfg.ValueConfiguration != nil {
				expFields[FieldName(exp.NameBinding.Value)] = cfg.ValueConfiguration
			}
		case CollectionBindIndexAndValue:
			if cfg.IndexConfiguration != nil {
				expFields[FieldName(exp.NameBinding.Index)] = cfg.IndexConfiguration
			}
			if cfg.ValueConfiguration != nil {
				expFields[FieldName(exp.NameBinding.Value)] = cfg.ValueConfiguration
			}
		default:
			return 0, selectorExp.(Expression), fmt.Errorf("Invalid name binding mode: %d", exp.NameBinding.Mode)
		}

		matches, subExp, err := validateRecurse(exp.Expression, expFields, maxRawValueLength)
		if err == nil {
			exp.Expression = subExp
		}

		newExp, err := expandImplicitCollectionOp(selectorExp, fields)
		return matches, newExp, err
	case *MatchExpression:
		if exp.Value != nil && maxRawValueLength != 0 && len(exp.Value.Raw) > maxRawValueLength {
			return 1, selectorExp.(Expression), fmt.Errorf("Value in expression with length %d for selector %q exceeds maximum length of", len(exp.Value.Raw), maxRawValueLength)
		}

		// validate the operator
		found := false
		for _, op := range cfg.SupportedOperations {
			if op == exp.Operator {
				found = true
				break
			}
		}

		if !found {
			// TODO actually output the selector that was given by the users - this Selector is potentially
			// one that was altered during the expansion
			return 1, selectorExp.(Expression), fmt.Errorf("Invalid match operator %q for selector %q", exp.Operator, exp.Selector)
		}

		// check the operator
		// coerce/validate the value
		if exp.Value != nil {
			if cfg.CoerceFn != nil {
				coerced, err := cfg.CoerceFn(exp.Value.Raw)
				if err != nil {
					return 1, selectorExp.(Expression), fmt.Errorf("Failed to coerce value %q for selector %q: %v", exp.Value.Raw, exp.Selector, err)
				}

				exp.Value.Converted = coerced
			}

			if exp.Operator == MatchMatches || exp.Operator == MatchNotMatches {
				var regRaw string
				if strVal, ok := exp.Value.Converted.(string); ok {
					regRaw = strVal
				} else if exp.Value.Converted == nil {
					regRaw = exp.Value.Raw
				} else {
					return 1, selectorExp.(Expression), fmt.Errorf("Match operator %q cannot be used with fields whose coercion functions return non string values", exp.Operator)
				}

				re, err := regexp.Compile(regRaw)
				if err != nil {
					return 1, selectorExp.(Expression), fmt.Errorf("Failed to compile regular expression %q: %v", regRaw, err)
				}

				exp.Value.Converted = re
			}
		} else {
			switch exp.Operator {
			case MatchIsEmpty, MatchIsNotEmpty:
				// these don't require values
			default:
				return 1, selectorExp.(Expression), fmt.Errorf("Match operator %q requires a non-nil value", exp.Operator)
			}
		}

		newExp, err := expandImplicitCollectionOp(selectorExp, fields)
		return 1, newExp, err
	}

	return 0, selectorExp.(Expression), fmt.Errorf("Invalid AST type: %T", selectorExp)
}

func validateRecurse(ast Expression, fields FieldConfigurations, maxRawValueLength int) (int, Expression, error) {
	switch node := ast.(type) {
	case *UnaryExpression:
		switch node.Operator {
		case UnaryOpNot:
			// this is fine
		default:
			return 0, ast, fmt.Errorf("Invalid unary expression operator: %d", node.Operator)
		}

		if node.Operand == nil {
			return 0, ast, fmt.Errorf("Invalid unary expression operand: nil")
		}
		matches, newExp, err := validateRecurse(node.Operand, fields, maxRawValueLength)
		if err == nil {
			node.Operand = newExp
		}
		return matches, ast, err
	case *BinaryExpression:
		switch node.Operator {
		case BinaryOpAnd, BinaryOpOr:
			// this is fine
		default:
			return 0, ast, fmt.Errorf("Invalid binary expression operator: %d", node.Operator)
		}

		if node.Left == nil {
			return 0, ast, fmt.Errorf("Invalid left hand side of binary expression: nil")
		} else if node.Right == nil {
			return 0, ast, fmt.Errorf("Invalid right hand side of binary expression: nil")
		}

		leftMatches, leftExp, err := validateRecurse(node.Left, fields, maxRawValueLength)
		if err != nil {
			return leftMatches, ast, err
		}

		rightMatches, rightExp, err := validateRecurse(node.Right, fields, maxRawValueLength)
		if err == nil {
			node.Left = leftExp
			node.Right = rightExp
		}
		return leftMatches + rightMatches, ast, err
	case *MatchExpression:
		if len(node.Selector) < 1 {
			return 1, ast, fmt.Errorf("Invalid selector: %q", node.Selector)
		}

		if node.Value != nil && maxRawValueLength != 0 && len(node.Value.Raw) > maxRawValueLength {
			return 1, ast, fmt.Errorf("Value in expression with length %d for selector %q exceeds maximum length of", len(node.Value.Raw), maxRawValueLength)
		}

		// exit early if we have no fields to check against
		if len(fields) < 1 {
			return 1, ast, nil
		}

		return validateSelectorExpression(node, fields, maxRawValueLength)
	case *CollectionExpression:
		if len(node.Selector) < 1 {
			return 1, ast, fmt.Errorf("Invalid selector: %q", node.Selector)
		}

		// exit early if we have no fields to check against
		if len(fields) < 1 {
			return 1, ast, nil
		}

		return validateSelectorExpression(node, fields, maxRawValueLength)
	}

	return 0, ast, fmt.Errorf("Cannot validate: Invalid AST node: %T", ast)
}

func validate(ast Expression, fields FieldConfigurations, maxMatches, maxRawValueLength int) (Expression, error) {
	matches, ast, err := validateRecurse(ast, fields, maxRawValueLength)
	if err != nil {
		return ast, err
	}

	if maxMatches != 0 && matches > maxMatches {
		return ast, fmt.Errorf("Number of match expressions (%d) exceeds the limit (%d)", matches, maxMatches)
	}

	return ast, nil
}
