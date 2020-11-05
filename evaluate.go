package bexpr

import (
	"fmt"
	"reflect"
	"regexp"
	"strings"
)

var byteSliceTyp reflect.Type = reflect.TypeOf([]byte{})

type evalMode int

const (
	evalModeDefault evalMode = iota
	evalModeCollection
)

type evalStackItem struct {
	mode evalMode

	// used for modes evalModeCollection
	indexName   string
	valueName   string
	indexFields FieldConfigurations
	valueFields FieldConfigurations
	index       reflect.Value

	// used for all modes
	value reflect.Value
}

const defaultEvalContextStackSize int = 4

type evalContext struct {
	stack []evalStackItem
	depth int
}

func (c *evalContext) push(si evalStackItem) {
	if c.depth >= cap(c.stack) {
		c.stack = append(c.stack, si)
	} else {
		c.stack[c.depth] = si
	}

	c.depth += 1
}

func (c *evalContext) pop() {
	if c.depth > 0 {
		c.depth -= 1
	}
}

func (c *evalContext) head() *evalStackItem {
	if c.depth == 0 {
		return nil
	}

	return &c.stack[c.depth-1]
}

var primitiveEqualityFns = map[reflect.Kind]func(first interface{}, second reflect.Value) bool{
	reflect.Bool:    doEqualBool,
	reflect.Int:     doEqualInt,
	reflect.Int8:    doEqualInt8,
	reflect.Int16:   doEqualInt16,
	reflect.Int32:   doEqualInt32,
	reflect.Int64:   doEqualInt64,
	reflect.Uint:    doEqualUint,
	reflect.Uint8:   doEqualUint8,
	reflect.Uint16:  doEqualUint16,
	reflect.Uint32:  doEqualUint32,
	reflect.Uint64:  doEqualUint64,
	reflect.Float32: doEqualFloat32,
	reflect.Float64: doEqualFloat64,
	reflect.String:  doEqualString,
}

func doEqualBool(first interface{}, second reflect.Value) bool {
	return first.(bool) == second.Bool()
}

func doEqualInt(first interface{}, second reflect.Value) bool {
	return first.(int) == int(second.Int())
}

func doEqualInt8(first interface{}, second reflect.Value) bool {
	return first.(int8) == int8(second.Int())
}

func doEqualInt16(first interface{}, second reflect.Value) bool {
	return first.(int16) == int16(second.Int())
}

func doEqualInt32(first interface{}, second reflect.Value) bool {
	return first.(int32) == int32(second.Int())
}

func doEqualInt64(first interface{}, second reflect.Value) bool {
	return first.(int64) == second.Int()
}

func doEqualUint(first interface{}, second reflect.Value) bool {
	return first.(uint) == uint(second.Uint())
}

func doEqualUint8(first interface{}, second reflect.Value) bool {
	return first.(uint8) == uint8(second.Uint())
}

func doEqualUint16(first interface{}, second reflect.Value) bool {
	return first.(uint16) == uint16(second.Uint())
}

func doEqualUint32(first interface{}, second reflect.Value) bool {
	return first.(uint32) == uint32(second.Uint())
}

func doEqualUint64(first interface{}, second reflect.Value) bool {
	return first.(uint64) == second.Uint()
}

func doEqualFloat32(first interface{}, second reflect.Value) bool {
	return first.(float32) == float32(second.Float())
}

func doEqualFloat64(first interface{}, second reflect.Value) bool {
	return first.(float64) == second.Float()
}

func doEqualString(first interface{}, second reflect.Value) bool {
	return first.(string) == second.String()
}

// Get rid of 0 to many levels of pointers to get at the real type
func derefType(rtype reflect.Type) reflect.Type {
	for rtype.Kind() == reflect.Ptr {
		rtype = rtype.Elem()
	}
	return rtype
}

func doMatchMatches(expression *MatchExpression, value reflect.Value) (bool, error) {
	if !value.Type().ConvertibleTo(byteSliceTyp) {
		return false, fmt.Errorf("Value of type %s is not convertible to []byte", value.Type())
	}

	re := expression.Value.Converted.(*regexp.Regexp)

	return re.Match(value.Convert(byteSliceTyp).Interface().([]byte)), nil
}

func doMatchEqual(expression *MatchExpression, value reflect.Value) (bool, error) {
	// NOTE: see preconditions in evaluateMatchExpressionRecurse
	eqFn := primitiveEqualityFns[value.Kind()]
	matchValue := getMatchExprValue(expression)
	return eqFn(matchValue, value), nil
}

func doMatchIn(expression *MatchExpression, value reflect.Value) (bool, error) {
	// NOTE: see preconditions in evaluateMatchExpressionRecurse
	matchValue := getMatchExprValue(expression)

	switch kind := value.Kind(); kind {
	case reflect.Map:
		found := value.MapIndex(reflect.ValueOf(matchValue))
		return found.IsValid(), nil
	case reflect.Slice, reflect.Array:
		itemType := derefType(value.Type().Elem())
		eqFn := primitiveEqualityFns[itemType.Kind()]

		for i := 0; i < value.Len(); i++ {
			item := value.Index(i)

			// the value will be the correct type as we verified the itemType
			if eqFn(matchValue, reflect.Indirect(item)) {
				return true, nil
			}
		}

		return false, nil
	case reflect.String:
		return strings.Contains(value.String(), matchValue.(string)), nil
	default:
		// this shouldn't be possible but we have to have something to return to keep the compiler happy
		return false, fmt.Errorf("Cannot perform in/contains operations on type %s for selector: %q", kind, expression.Selector)
	}
}

func doMatchIsEmpty(matcher *MatchExpression, value reflect.Value) (bool, error) {
	// NOTE: see preconditions in evaluateMatchExpressionRecurse
	return value.Len() == 0, nil
}

func getMatchExprValue(expression *MatchExpression) interface{} {
	// NOTE: see preconditions in evaluateMatchExpressionRecurse
	if expression.Value == nil {
		return nil
	}

	if expression.Value.Converted != nil {
		return expression.Value.Converted
	}

	return expression.Value.Raw
}

func getValueForSelector(sel Selector, rvalue reflect.Value, fields FieldConfigurations) (*reflect.Value, *FieldConfiguration, error) {
	value := rvalue
	var cfg *FieldConfiguration

	for idx, fieldName := range sel {
		switch value.Kind() {
		case reflect.Struct:
			cfg = fields[FieldName(fieldName)]

			if cfg.StructFieldName != "" {
				fieldName = cfg.StructFieldName
			}

			value = reflect.Indirect(value.FieldByName(fieldName))
			fields = cfg.SubFields
		case reflect.Slice, reflect.Array:
			return nil, nil, fmt.Errorf("Invalid AST: intermediate selector references a list type")
		case reflect.Map:
			cfg = fields[FieldNameAny]
			value = reflect.Indirect(value.MapIndex(reflect.ValueOf(fieldName)))

			if !value.IsValid() {
				return nil, cfg, nil
			}

			fields = cfg.SubFields
		default:
			return nil, nil, fmt.Errorf("Value at selector %q with type %s does not support nested field selection", sel[:idx+1], value.Kind())
		}
	}

	return &value, cfg, nil
}

func evaluateMatchExpression(exp *MatchExpression, ctx *evalContext, fields FieldConfigurations) (bool, error) {
	var sel Selector
	var value reflect.Value

	curCtx := ctx.head()

	if curCtx == nil {
		return false, fmt.Errorf("Invalid evaluation context: nil stack")
	}

	switch curCtx.mode {
	case evalModeDefault:
		sel = exp.Selector
		value = curCtx.value
	case evalModeCollection:
		sel = exp.Selector[1:]
		if curCtx.indexName != "" && exp.Selector[0] == curCtx.indexName {
			value = curCtx.index
			fields = curCtx.indexFields
		} else if curCtx.valueName != "" && exp.Selector[0] == curCtx.valueName {
			value = curCtx.value
			fields = curCtx.valueFields
		} else {
			return false, fmt.Errorf("Invalid selector: %q", exp.Selector[0])
		}
	default:
		return false, fmt.Errorf("Invalid evaluation context mode: %d", curCtx.mode)
	}

	matchValue, _, err := getValueForSelector(sel, value, fields)
	if err != nil {
		return false, err
	}

	if matchValue == nil {
		// when the key doesn't exist in the map
		switch exp.Operator {
		// MatchEqual, MatchIsNotEmpty and MatchIn cannot possible be true for a not found value
		case MatchEqual, MatchIsNotEmpty, MatchIn:
			return false, nil
		default:
			// MatchNotEqual, MatchIsEmpty, MatchNotIn
			// Whatever you were looking for cannot be equal because it doesn't exist
			// Similarly it cannot be in some other container and every other container
			// is always empty.
			return true, nil
		}
	}

	switch exp.Operator {
	case MatchEqual:
		return doMatchEqual(exp, *matchValue)
	case MatchNotEqual:
		result, err := doMatchEqual(exp, *matchValue)
		if err == nil {
			return !result, nil
		}
		return false, err
	case MatchIn:
		return doMatchIn(exp, *matchValue)
	case MatchNotIn:
		result, err := doMatchIn(exp, *matchValue)
		if err == nil {
			return !result, nil
		}
		return false, err
	case MatchIsEmpty:
		return doMatchIsEmpty(exp, *matchValue)
	case MatchIsNotEmpty:
		result, err := doMatchIsEmpty(exp, *matchValue)
		if err == nil {
			return !result, nil
		}
		return false, err
	case MatchMatches:
		return doMatchMatches(exp, *matchValue)
	case MatchNotMatches:
		result, err := doMatchMatches(exp, *matchValue)
		if err == nil {
			return !result, nil
		}
		return false, err
	default:
		return false, fmt.Errorf("Invalid match operation: %d", exp.Operator)
	}
}

func evaluateCollectionExpression(exp *CollectionExpression, ctx *evalContext, fields FieldConfigurations) (bool, error) {
	var sel Selector
	var value reflect.Value

	curCtx := ctx.head()

	if curCtx == nil {
		return false, fmt.Errorf("Invalid evaluation context: nil stack")
	}

	switch curCtx.mode {
	case evalModeDefault:
		sel = exp.Selector
		value = curCtx.value
	case evalModeCollection:
		sel = exp.Selector[1:]
		// we don't need to care about the index/map key here as they cannot be used as the main selector of
		// a collection expression because we cannot iterate over strings (map) and ints (list)
		if curCtx.valueName != "" && exp.Selector[0] == curCtx.valueName {
			value = curCtx.value
			fields = curCtx.valueFields
		} else {
			return false, fmt.Errorf("Invalid selector in collection expression: %s", exp.Selector[0])
		}
	default:
		return false, fmt.Errorf("Invalid evaluation context mode: %d", curCtx.mode)
	}

	collectionValue, fieldConfig, err := getValueForSelector(sel, value, fields)
	if err != nil {
		return false, err
	}

	if collectionValue == nil {
		return false, nil
	}

	var indexName string
	var valueName string
	switch exp.NameBinding.Mode {
	case CollectionBindDefault:
		if fieldConfig.CollectionType == CollectionTypeMap {
			indexName = exp.NameBinding.Default
		} else {
			valueName = exp.NameBinding.Default
		}
	case CollectionBindIndex:
		indexName = exp.NameBinding.Index
	case CollectionBindValue:
		valueName = exp.NameBinding.Value
	case CollectionBindIndexAndValue:
		indexName = exp.NameBinding.Index
		valueName = exp.NameBinding.Value
	}

	switch fieldConfig.CollectionType {
	case CollectionTypeMap:
		iter := (*collectionValue).MapRange()
		for iter.Next() {
			ctx.push(evalStackItem{
				mode:        evalModeCollection,
				indexName:   indexName,
				valueName:   valueName,
				index:       reflect.Indirect(iter.Key()),
				indexFields: fieldConfig.IndexConfiguration.SubFields,
				value:       reflect.Indirect(iter.Value()),
				valueFields: fieldConfig.ValueConfiguration.SubFields,
			})
			result, err := evaluateRecurse(exp.Expression, ctx, fieldConfig.SubFields)
			ctx.pop()

			if err != nil {
				return false, err
			}

			if result && exp.Operator == CollectionOpAny {
				return true, nil
			} else if !result && exp.Operator == CollectionOpAll {
				return false, nil
			}
		}

		// if we got here then we were executing an all expression and all matched
		// or executing an any expression and none matched or there were no map elements
		// If there were no map elements then an "all" expression will return true and an
		// "any" expression will return false. This seems like sane behavior.
		return exp.Operator == CollectionOpAll, nil

	case CollectionTypeList:
		for i := 0; i < (*collectionValue).Len(); i++ {
			ctx.push(evalStackItem{
				mode:        evalModeCollection,
				indexName:   indexName,
				valueName:   valueName,
				index:       reflect.ValueOf(i),
				indexFields: fieldConfig.IndexConfiguration.SubFields,
				value:       reflect.Indirect((*collectionValue).Index(i)),
				valueFields: fieldConfig.ValueConfiguration.SubFields,
			})
			result, err := evaluateRecurse(exp.Expression, ctx, fieldConfig.SubFields)
			ctx.pop()

			if err != nil {
				return false, err
			}

			if result && exp.Operator == CollectionOpAny {
				return true, nil
			} else if !result && exp.Operator == CollectionOpAll {
				return false, nil
			}

		}

		// if we got here then we were executing an all expression and all matched
		// or executing an any expression and none matched or there were no map elements
		// If there were no map elements then an "all" expression will return true and an
		// "any" expression will return false. This seems like sane behavior.
		return exp.Operator == CollectionOpAll, nil
	default:
		return false, fmt.Errorf("Invalid collection type: %d", fieldConfig.CollectionType)
	}
}

func evaluateRecurse(ast Expression, ctx *evalContext, fields FieldConfigurations) (bool, error) {
	switch node := ast.(type) {
	case *UnaryExpression:
		switch node.Operator {
		case UnaryOpNot:
			result, err := evaluateRecurse(node.Operand, ctx, fields)
			return !result, err
		}
	case *BinaryExpression:
		switch node.Operator {
		case BinaryOpAnd:
			result, err := evaluateRecurse(node.Left, ctx, fields)
			if err != nil || !result {
				return result, err
			}

			return evaluateRecurse(node.Right, ctx, fields)

		case BinaryOpOr:
			result, err := evaluateRecurse(node.Left, ctx, fields)
			if err != nil || result {
				return result, err
			}

			return evaluateRecurse(node.Right, ctx, fields)
		}
	case *CollectionExpression:
		return evaluateCollectionExpression(node, ctx, fields)
	case *MatchExpression:
		return evaluateMatchExpression(node, ctx, fields)
	}
	return false, fmt.Errorf("Invalid AST node: %T", ast)
}

func evaluate(ast Expression, datum interface{}, fields FieldConfigurations) (bool, error) {
	ctx := evalContext{
		// 4 should be sufficiently large unless we have more than 3 nested collection ops
		stack: make([]evalStackItem, defaultEvalContextStackSize),
		depth: 0,
	}

	ctx.push(evalStackItem{mode: evalModeDefault, value: reflect.ValueOf(datum)})
	return evaluateRecurse(ast, &ctx, fields)
}
