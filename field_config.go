package bexpr

import (
	"fmt"
	"reflect"
	"strings"
)

type FieldCollectionType int

const (
	// Indicates that the collection operators any/all should not be supported
	CollectionTypeNone FieldCollectionType = iota
	// Indicates that the collection operators should be supported with map semantics
	CollectionTypeMap
	// Indicates that the collection operators should be supported with list semantics
	CollectionTypeList
)

func (ctype FieldCollectionType) String() string {
	switch ctype {
	case CollectionTypeNone:
		return "None"
	case CollectionTypeMap:
		return "Map"
	case CollectionTypeList:
		return "List"
	default:
		return "UNKNOWN"
	}
}

// Function type for usage with a SelectorConfiguration
type FieldValueCoercionFn func(value string) (interface{}, error)

// Strongly typed name of a field
type FieldName string

// Used to represent an arbitrary field name
const FieldNameAny FieldName = ""

type FieldPath []FieldName

func (path FieldPath) String() string {
	var parts []string

	for _, part := range path {
		if part == FieldNameAny {
			parts = append(parts, "<any>")
		} else {
			parts = append(parts, string(part))
		}
	}

	return strings.Join(parts, ".")
}

// The FieldConfiguration struct represents how boolean expression
// validation and preparation should work for the given field. A field
// in this case is a single element of a selector.
//
// Example: foo.bar.baz has 3 fields separate by '.' characters.
type FieldConfiguration struct {
	// Name to use when looking up fields within a struct. This is useful when
	// the name(s) you want to expose to users writing the expressions does not
	// exactly match the Field name of the structure. If this is empty then the
	// user provided name will be used
	StructFieldName string

	// Either CollectionTypeNone, CollectionTypeMap or CollectionTypeList
	CollectionType FieldCollectionType
	// IndexConfiguration and ValueConfiguration contain the unbound field configuration
	// for the index and value of a collection when the CollectionType is set to something
	// other than CollectionTypeNone
	IndexConfiguration *FieldConfiguration
	ValueConfiguration *FieldConfiguration

	// Nested field configurations
	SubFields FieldConfigurations

	// Function to run on the raw string value present in the expression
	// syntax to coerce into whatever form is needed during evaluation
	// The coercion happens only once and will then be passed as the `value`
	// parameter to all evaluations using the same expression.
	CoerceFn FieldValueCoercionFn

	// List of MatchOperators supported for this field. This configuration
	// is used to pre-validate an expressions fields before execution.
	SupportedOperations []MatchOperator
}

// Represents all the valid fields and their corresponding configuration
type FieldConfigurations map[FieldName]*FieldConfiguration

var primitiveMatchOps = map[reflect.Kind][]MatchOperator{
	reflect.Bool:    []MatchOperator{MatchEqual, MatchNotEqual},
	reflect.Int:     []MatchOperator{MatchEqual, MatchNotEqual},
	reflect.Int8:    []MatchOperator{MatchEqual, MatchNotEqual},
	reflect.Int16:   []MatchOperator{MatchEqual, MatchNotEqual},
	reflect.Int32:   []MatchOperator{MatchEqual, MatchNotEqual},
	reflect.Int64:   []MatchOperator{MatchEqual, MatchNotEqual},
	reflect.Uint:    []MatchOperator{MatchEqual, MatchNotEqual},
	reflect.Uint8:   []MatchOperator{MatchEqual, MatchNotEqual},
	reflect.Uint16:  []MatchOperator{MatchEqual, MatchNotEqual},
	reflect.Uint32:  []MatchOperator{MatchEqual, MatchNotEqual},
	reflect.Uint64:  []MatchOperator{MatchEqual, MatchNotEqual},
	reflect.Float32: []MatchOperator{MatchEqual, MatchNotEqual},
	reflect.Float64: []MatchOperator{MatchEqual, MatchNotEqual},
	reflect.String:  []MatchOperator{MatchEqual, MatchNotEqual, MatchIn, MatchNotIn, MatchMatches, MatchNotMatches},
}

func primitiveFieldConfiguration(rkind reflect.Kind) *FieldConfiguration {
	coerceFn, ok := primitiveCoercionFns[rkind]
	if !ok {
		return nil
	}
	ops, ok := primitiveMatchOps[rkind]
	if !ok {
		return nil
	}

	return &FieldConfiguration{
		CoerceFn:            coerceFn,
		SupportedOperations: ops,
	}
}

func generateFieldConfigurationInternal(rtype reflect.Type) (*FieldConfiguration, error) {
	// must be done after checking for interface implementing
	rtype = derefType(rtype)

	// Handle primitive types
	if cfg := primitiveFieldConfiguration(rtype.Kind()); cfg != nil {
		ops := []MatchOperator{MatchEqual, MatchNotEqual}

		if rtype.Kind() == reflect.String {
			ops = append(ops, MatchIn, MatchNotIn, MatchMatches, MatchNotMatches)
		}

		return cfg, nil
	}

	// Handle compound types
	switch rtype.Kind() {
	case reflect.Map:
		return generateMapFieldConfiguration(derefType(rtype.Key()), rtype.Elem())
	case reflect.Array, reflect.Slice:
		return generateSliceFieldConfiguration(rtype.Elem())
	case reflect.Struct:
		subfields, err := generateStructFieldConfigurations(rtype)
		if err != nil {
			return nil, err
		}

		return &FieldConfiguration{
			SubFields: subfields,
		}, nil

	default: // unsupported types are just not filterable
		return nil, nil
	}
}

func generateSliceFieldConfiguration(elemType reflect.Type) (*FieldConfiguration, error) {
	cfg := &FieldConfiguration{
		CollectionType:     CollectionTypeList,
		IndexConfiguration: primitiveFieldConfiguration(reflect.Int),
	}

	if elemCfg := primitiveFieldConfiguration(elemType.Kind()); elemCfg != nil {
		// slices of primitives have somewhat different supported operations
		cfg.ValueConfiguration = elemCfg
		cfg.CoerceFn = elemCfg.CoerceFn
		cfg.SupportedOperations = []MatchOperator{MatchIn, MatchNotIn, MatchIsEmpty, MatchIsNotEmpty}
		return cfg, nil
	}

	elemCfg, err := generateFieldConfigurationInternal(elemType)
	if err != nil {
		return nil, err
	}

	cfg.ValueConfiguration = elemCfg
	cfg.SupportedOperations = []MatchOperator{MatchIsEmpty, MatchIsNotEmpty}

	if elemCfg != nil && len(elemCfg.SubFields) > 0 {
		cfg.SubFields = elemCfg.SubFields
	}

	return cfg, nil
}

func generateMapFieldConfiguration(keyType, valueType reflect.Type) (*FieldConfiguration, error) {
	switch keyType.Kind() {
	case reflect.String:
		subfield, err := generateFieldConfigurationInternal(valueType)
		if err != nil {
			return nil, err
		}

		cfg := &FieldConfiguration{
			CollectionType:      CollectionTypeMap,
			IndexConfiguration:  primitiveFieldConfiguration(reflect.String),
			ValueConfiguration:  subfield,
			CoerceFn:            CoerceString,
			SupportedOperations: []MatchOperator{MatchIsEmpty, MatchIsNotEmpty, MatchIn, MatchNotIn},
		}

		if subfield != nil {
			cfg.SubFields = FieldConfigurations{
				FieldNameAny: subfield,
			}
		}

		return cfg, nil

	default:
		// TODO (any/all expressions) - We might be able to allow performing any/all expressions
		// against maps without string indexes but just prevent binding the index so long as the
		// value is supported. Do we need to is the question?

		// For maps with non-string keys we can really only do emptiness checks
		// and cannot index into them at all
		return &FieldConfiguration{
			SupportedOperations: []MatchOperator{MatchIsEmpty, MatchIsNotEmpty},
		}, nil
	}
}

func generateStructFieldConfigurations(rtype reflect.Type) (FieldConfigurations, error) {
	fieldConfigs := make(FieldConfigurations)

	for i := 0; i < rtype.NumField(); i++ {
		field := rtype.Field(i)

		fieldTag := field.Tag.Get("bexpr")

		var fieldNames []string

		if field.PkgPath != "" {
			// we cant handle unexported fields using reflection
			continue
		}

		if fieldTag != "" {
			parts := strings.Split(fieldTag, ",")

			if len(parts) > 0 {
				if parts[0] == "-" {
					continue
				}

				fieldNames = parts
			} else {
				fieldNames = append(fieldNames, field.Name)
			}
		} else {
			fieldNames = append(fieldNames, field.Name)
		}

		cfg, err := generateFieldConfigurationInternal(field.Type)
		if err != nil {
			return nil, err
		}
		cfg.StructFieldName = field.Name

		// link the config to all the correct names
		for _, name := range fieldNames {
			fieldConfigs[FieldName(name)] = cfg
		}
	}

	return fieldConfigs, nil
}

// `generateFieldConfigurations` can be used to generate the `FieldConfigurations` map
// It supports generating configurations for either a `map[string]*` or a `struct` as the `topLevelType`
//
// Internally within the top level type the following is supported:
//
// Primitive Types:
//    strings
//    integers (all width types and signedness)
//    floats (32 and 64 bit)
//    bool
//
// Compound Types
//   `map[*]*`
//       - Supports emptiness checking. Does not support further selector nesting.
//   `map[string]*`
//       - Supports in/contains operations on the keys.
//   `map[string]<supported type>`
//       - Will have a single subfield with name `FieldNameAny` (wildcard) and the rest of
//         the field configuration will come from the `<supported type>`
//   `[]*`
//       - Supports emptiness checking only. Does not support further selector nesting.
//   `[]<supported primitive type>`
//       - Supports in/contains operations against the primitive values.
//   `[]<supported compund type>`
//       - Will have subfields with the configuration of whatever the supported
//         compound type is.
//       - Does not support indexing of individual values like a map does currently
//         and with the current evaluation logic slices of slices will mostly be
//         handled as if they were flattened. One thing that cannot be done is
//         to be able to perform emptiness/contains checking against the internal
//         slice.
//   structs
//       - No operations are supported on the struct itself
//       - Will have subfield configurations generated for the fields of the struct.
//       - A struct tag like `bexpr:"<name>"` allows changing the name that allows indexing
//         into the subfield.
//       - By default unexported fields of a struct are not selectable. If The struct tag is
//         present then this behavior is overridden.
//       - Exported fields can be made unselectable by adding a tag to the field like `bexpr:"-"`
func GenerateFieldConfigurations(topLevelType interface{}) (FieldConfigurations, error) {
	return generateFieldConfigurations(reflect.TypeOf(topLevelType))
}

func generateFieldConfigurations(rtype reflect.Type) (FieldConfigurations, error) {
	// Do this after we check for interface implementation
	rtype = derefType(rtype)

	switch rtype.Kind() {
	case reflect.Struct:
		fields, err := generateStructFieldConfigurations(rtype)
		return fields, err
	case reflect.Map:
		if rtype.Key().Kind() != reflect.String {
			return nil, fmt.Errorf("Cannot generate FieldConfigurations for maps with keys that are not strings")
		}

		elemType := rtype.Elem()

		field, err := generateFieldConfigurationInternal(elemType)
		if err != nil {
			return nil, err
		}

		if field == nil {
			return nil, nil
		}

		return FieldConfigurations{
			FieldNameAny: field,
		}, nil
	}

	return nil, fmt.Errorf("Invalid top level type - can only use structs or an map[string]*")
}

func (config *FieldConfiguration) stringInternal(builder *strings.Builder, level int, path string) {
	fmt.Fprintf(builder, "%sPath: %s, StructFieldName: %s, CoerceFn: %p, SupportedOperations: %v\n", strings.Repeat("   ", level), path, config.StructFieldName, config.CoerceFn, config.SupportedOperations)
	if len(config.SubFields) > 0 {
		config.SubFields.stringInternal(builder, level+1, path)
	}
}

func (config *FieldConfiguration) String() string {
	var builder strings.Builder
	config.stringInternal(&builder, 0, "")
	return builder.String()
}

func (configs FieldConfigurations) stringInternal(builder *strings.Builder, level int, path string) {
	for fieldName, cfg := range configs {
		newPath := string(fieldName)
		if level > 0 {
			newPath = fmt.Sprintf("%s.%s", path, fieldName)
		}
		cfg.stringInternal(builder, level, newPath)
	}
}

func (configs FieldConfigurations) String() string {
	var builder strings.Builder
	configs.stringInternal(&builder, 0, "")
	return builder.String()
}

type FieldConfigurationWalkFn func(path FieldPath, config *FieldConfiguration) bool

func (configs FieldConfigurations) walk(path FieldPath, walkFn FieldConfigurationWalkFn) bool {
	for fieldName, fieldConfig := range configs {
		newPath := append(path, fieldName)

		if !walkFn(newPath, fieldConfig) {
			return false
		}

		if !fieldConfig.SubFields.walk(newPath, walkFn) {
			return false
		}
	}

	return true
}

func (configs FieldConfigurations) Walk(walkFn FieldConfigurationWalkFn) bool {
	return configs.walk(nil, walkFn)
}
