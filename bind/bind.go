package bind

import (
	"flag"
	"reflect"
	"strconv"
	"strings"
)

// Parameters map a field name to one or more values.
type Parameters map[string][]string

// convertibleType is a specification for converting from a string to a
// particular type, and for adding a flag of that type to a flag.FlagSet.
type convertibleType struct {
	Type        reflect.Type
	ConvertFrom func(string) (reflect.Value, error)
	FlagFunc    func(*flag.FlagSet) reflect.Value
}

type convertibleTypes []convertibleType

// AssignableTo returns a pointer to a convertibleType that knows how to
// convert a string to a type assignable to targetType, or it returns nil if it
// cannot find such a type.
func (types convertibleTypes) AssignableTo(targetType reflect.Type) *convertibleType {
	for _, convertible := range allConvertibleTypes {
		if convertible.Type.AssignableTo(targetType) {
			return &convertible
		}
	}
	return nil
}

// convertibleTypes declares the types that this package understands.
var allConvertibleTypes = convertibleTypes{
	// Conversion info for int.
	convertibleType{
		Type: reflect.TypeOf((*int)(nil)).Elem(),
		FlagFunc: func(f *flag.FlagSet) reflect.Value {
			return reflect.ValueOf(f.IntVar)
		},
		ConvertFrom: func(val string) (reflect.Value, error) {
			converted, err := strconv.Atoi(val)
			return reflect.ValueOf(converted), err
		}},
	// Conversion info for floats.
	convertibleType{
		Type: reflect.TypeOf((*float64)(nil)).Elem(),
		FlagFunc: func(f *flag.FlagSet) reflect.Value {
			return reflect.ValueOf(f.Float64Var)
		},
		ConvertFrom: func(val string) (reflect.Value, error) {
			converted, err := strconv.ParseFloat(val, 64)
			return reflect.ValueOf(converted), err
		}},
	// Conversion info for bools.
	convertibleType{
		Type: reflect.TypeOf((*bool)(nil)).Elem(),
		FlagFunc: func(f *flag.FlagSet) reflect.Value {
			return reflect.ValueOf(f.BoolVar)
		},
		ConvertFrom: func(val string) (reflect.Value, error) {
			converted, err := strconv.ParseBool(val)
			return reflect.ValueOf(converted), err
		}},
	// Conversion info for strings.
	convertibleType{
		Type: reflect.TypeOf((*string)(nil)).Elem(),
		FlagFunc: func(f *flag.FlagSet) reflect.Value {
			return reflect.ValueOf(f.StringVar)
		},
		ConvertFrom: func(val string) (reflect.Value, error) {
			return reflect.ValueOf(val), nil
		}},
}

// Given an arbitrary object, return its struct type, struct value, and a
// boolean indicating whether the object was a pointer to a struct.
func inspect(bindable interface{}) (reflect.Type, reflect.Value, bool) {
	maybePointer := reflect.ValueOf(bindable)
	if maybePointer.Kind() != reflect.Ptr {
		return reflect.TypeOf(nil), reflect.ValueOf(nil), false
	}

	structValue := maybePointer.Elem()
	if structValue.Kind() != reflect.Struct {
		return reflect.TypeOf(nil), reflect.ValueOf(nil), false
	}

	structType := structValue.Type()
	return structType, structValue, true
}

// GenerateFlags builds and returns a flag.FlagSet from a struct's fields that,
// when used to parse arguments, sets the field values accordingly.
func GenerateFlags(bindable interface{}, name string) (*flag.FlagSet, bool) {
	bindType, bindValue, ok := inspect(bindable)
	if !ok {
		return nil, false
	}

	result := flag.NewFlagSet(name, flag.ContinueOnError)

	for i := 0; i < bindType.NumField(); i++ {
		field := bindType.Field(i)
		fieldValue := bindValue.Field(i)

		// Skip fields that don't have a tag.
		if len(field.Tag) == 0 {
			continue
		}

		// Get the flag name from the tag, falling back to the field name.
		flagName := field.Tag.Get("name")
		if len(flagName) == 0 {
			flagName = strings.ToLower(field.Name)
		}

		defaultStr := field.Tag.Get("default")
		help := field.Tag.Get("help")
		ptr := fieldValue.Addr()

		convertible := allConvertibleTypes.AssignableTo(field.Type)
		if convertible == nil {
			continue
		}

		// Convert the default (extracted from the tag) to a Value.
		defaultVal, err := convertible.ConvertFrom(defaultStr)
		if err != nil {
			defaultVal = reflect.Zero(convertible.Type)
		}

		// Build arguments and create a flag via reflection. A pointer to
		// the struct field is passed to the flag function so parsing
		// arguments can set it directly.
		args := []reflect.Value{
			ptr,
			reflect.ValueOf(flagName),
			defaultVal,
			reflect.ValueOf(help)}
		convertible.FlagFunc(result).Call(args)
	}
	return result, true
}

// Bind sets the fields of an arbitary struct value (from a pointer) from
// a map of string values.
func Bind(bindable interface{}, params Parameters) bool {
	structType, structValue, ok := inspect(bindable)
	if !ok {
		return false
	}

	for i := 0; i < structType.NumField(); i++ {
		field := structType.Field(i)

		// Skip fields that don't have a value in params.
		paramValues, ok := params[strings.ToLower(field.Name)]
		if !ok {
			continue
		}

		fieldValue := structValue.Field(i)

		// Assign a slice if we can. Otherwise, skip anything that's not a
		// slice with exactly one value.
		// TODO Support slices of other types we know how to convert
		if reflect.TypeOf(paramValues).AssignableTo(field.Type) {
			fieldValue.Set(reflect.ValueOf(paramValues))
		} else if len(paramValues) != 1 {
			continue
		}

		convertible := allConvertibleTypes.AssignableTo(field.Type)
		if convertible == nil {
			continue
		}

		// Convert the string into the struct field's type, and set it.
		result, err := convertible.ConvertFrom(paramValues[0])
		if err != nil {
			continue
		}
		fieldValue.Set(result)
	}
	return true
}
