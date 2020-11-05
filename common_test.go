package bexpr

import (
	"flag"
	"fmt"
	"reflect"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var benchFull *bool = flag.Bool("bench-full", false, "Run all benchmarks rather than a subset")

func FullBenchmarks() bool {
	return benchFull != nil && *benchFull
}

type testFlatStruct struct {
	Int        int
	Int8       int8
	Int16      int16
	Int32      int32
	Int64      int64
	Uint       uint
	Uint8      uint8
	Uint16     uint16
	Uint32     uint32
	Uint64     uint64
	Float32    float32
	Float64    float64
	Bool       bool
	String     string
	unexported string
	Hidden     bool `bexpr:"-"`
}

type CustomInt int
type CustomInt8 int8
type CustomInt16 int16
type CustomInt32 int32
type CustomInt64 int64
type CustomUint uint
type CustomUint8 uint8
type CustomUint16 uint16
type CustomUint32 uint32
type CustomUint64 uint64
type CustomFloat32 float32
type CustomFloat64 float64
type CustomBool bool
type CustomString string

type testFlatStructAlt struct {
	Int        CustomInt
	Int8       CustomInt8
	Int16      CustomInt16
	Int32      CustomInt32
	Int64      CustomInt64
	Uint       CustomUint
	Uint8      CustomUint8
	Uint16     CustomUint16
	Uint32     CustomUint32
	Uint64     CustomUint64
	Float32    CustomFloat32
	Float64    CustomFloat64
	Bool       CustomBool
	String     CustomString
	unexported CustomString
	Hidden     CustomBool `bexpr:"-"`
}

var testFlatStructKindMap map[string]reflect.Kind = map[string]reflect.Kind{
	"Int":     reflect.Int,
	"Int8":    reflect.Int8,
	"Int16":   reflect.Int16,
	"Int32":   reflect.Int32,
	"Int64":   reflect.Int64,
	"Uint":    reflect.Uint,
	"Uint8":   reflect.Uint8,
	"Uint16":  reflect.Uint16,
	"Uint32":  reflect.Uint32,
	"Uint64":  reflect.Uint64,
	"Float32": reflect.Float32,
	"Float64": reflect.Float64,
	"Bool":    reflect.Bool,
	"String":  reflect.String,
}

type testNestedLevel2_1 struct {
	Foo int
	bar string
	Baz string
}

type testNestedLevel2_2 struct {
	X int
	Y int
	z int
}

type testNestedLevel1 struct {
	Map              map[string]string
	MapOfStructs     map[string]testNestedLevel2_1
	MapInfInf        map[interface{}]interface{}
	SliceOfInts      []int
	SliceOfStructs   []testNestedLevel2_2
	SliceOfMapInfInf []map[interface{}]interface{}
}

type testNestedTypes struct {
	Nested testNestedLevel1
	TopInt int
}

func validateFieldConfiguration(t *testing.T, expected, actual *FieldConfiguration, name, path string) bool {
	if expected == nil {
		return assert.Nil(t, actual, "Field %q on path %q should be nil and isn't", name, path)
	}

	ok := assert.NotNil(t, actual, "Field %q on path %q should not be nil and is", name, path)
	if !ok {
		return ok
	}

	ok = ok && assert.Equal(t, expected.StructFieldName, actual.StructFieldName, "Field %q on path %q have different StructFieldNames - Expected: %q, Actual: %q", name, path, expected.StructFieldName, actual.StructFieldName)
	ok = ok && assert.ElementsMatch(t, expected.SupportedOperations, actual.SupportedOperations, "Fields %q on path %q have different SupportedOperations - Expected: %v, Actual: %v", name, path, expected.SupportedOperations, actual.SupportedOperations)
	ok = ok && assert.Equal(t, expected.CollectionType, actual.CollectionType, "Field %q on path %q has different CollectionType - Expected: %q, Actual: %q", name, path, expected.CollectionType, actual.CollectionType)

	newPath := name
	if newPath == "" {
		newPath = "*"
	}
	if path != "" {
		newPath = fmt.Sprintf("%s.%s", path, newPath)
	}

	ok = ok && validateFieldConfiguration(t, expected.IndexConfiguration, actual.IndexConfiguration, "<index>", newPath)
	ok = ok && validateFieldConfiguration(t, expected.ValueConfiguration, actual.ValueConfiguration, "<value>", newPath)
	ok = ok && validateFieldConfigurationsRecurse(t, expected.SubFields, actual.SubFields, newPath)

	return ok
}

func validateFieldConfigurationsRecurse(t *testing.T, expected, actual FieldConfigurations, path string) bool {
	t.Helper()

	ok := assert.Len(t, actual, len(expected), "Actual FieldConfigurations length of %d != expected length of %d for path %q", len(actual), len(expected), path)

	for fieldName, expectedConfig := range expected {
		actualConfig, ok := actual[fieldName]
		ok = ok && assert.True(t, ok, "Actual configuration is missing field %q", fieldName)
		ok = ok && validateFieldConfiguration(t, expectedConfig, actualConfig, string(fieldName), path)

		if !ok {
			break
		}
	}

	return ok
}

func validateFieldConfigurations(t *testing.T, expected, actual FieldConfigurations) {
	t.Helper()
	require.True(t, validateFieldConfigurationsRecurse(t, expected, actual, ""))
}

func dumpFieldConfigurationsRecurse(fields FieldConfigurations, level int, path string) {
	for fieldName, cfg := range fields {
		fmt.Printf("%sPath: %s Field: %s, StructFieldName: %s, CoerceFn: %p, SupportedOperations: %v\n", strings.Repeat("   ", level), path, fieldName, cfg.StructFieldName, cfg.CoerceFn, cfg.SupportedOperations)
		newPath := string(fieldName)
		if path != "" {
			newPath = fmt.Sprintf("%s.%s", path, fieldName)
		}
		dumpFieldConfigurationsRecurse(cfg.SubFields, level+1, newPath)
	}
}

func dumpFieldConfigurations(name string, fields FieldConfigurations) {
	fmt.Printf("===== %s =====\n", name)
	dumpFieldConfigurationsRecurse(fields, 1, "")
}
