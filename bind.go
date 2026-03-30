package bind

import (
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"sync"
	"time"
)

const (
	// the tag in the struct definition. ex: `query:"customer_id"`
	queryTagKey = "query"

	// the tag in the struct definition. ex: `env:"PORT"`
	envTagKey = "env"

	// the tag in the struct definition. ex: `default:"8080"`
	defaultTagKey = "default"

	// the tag in the struct definition. ex: `header:"X-Product-Id"`
	headerTagKey = "header"

	// the tag in the struct definition. ex: `form:"customer_id"`
	formTagKey = "form"

	// the tag in the struct definition. ex: `flag:"configDir"`
	flagTagKey = "flag"

	// time format expected in  date field TODO: name better? make configurable with default?
	tagTimeFormat = "2006-01-02"
)

type fieldMeta struct {
	index int
	field reflect.StructField
}

var (
	// cached time type
	timeType = reflect.TypeOf(time.Time{})

	// cached struct field metadata by type
	structFieldCache sync.Map // map[reflect.Type][]fieldMeta
)

// Query binds query-string values into receiver using `query` tags.
//
// Nil receiver is treated as a no-op. Non-pointer or unsupported receiver
// types return ErrReceiverUnsupportedType.
func Query(receiver any, data map[string][]string) error {
	return parse(receiver, queryTagKey, data)
}

// Form binds form values into receiver using `form` tags.
//
// Nil receiver is treated as a no-op. Non-pointer or unsupported receiver
// types return ErrReceiverUnsupportedType.
func Form(receiver any, data map[string][]string) error {
	return parse(receiver, formTagKey, data)
}

// Header binds header values into receiver using `header` tags.
//
// Nil receiver is treated as a no-op. Non-pointer or unsupported receiver
// types return ErrReceiverUnsupportedType.
func Header(receiver any, data map[string][]string) error {
	return parse(receiver, headerTagKey, data)
}

func receiverElem(receiver any, allowMap bool) (reflect.Type, reflect.Value, error) {
	typ := reflect.TypeOf(receiver)
	if typ == nil {
		return nil, reflect.Value{}, fmt.Errorf("%w; got <nil>", ErrReceiverUnsupportedType)
	}

	val := reflect.ValueOf(receiver)
	if typ.Kind() != reflect.Ptr {
		return nil, reflect.Value{}, fmt.Errorf("%w; got %s", ErrReceiverUnsupportedType, typ.Kind().String())
	}

	if val.IsNil() {
		return nil, reflect.Value{}, fmt.Errorf("%w; got nil pointer", ErrReceiverUnsupportedType)
	}

	elemType := typ.Elem()
	elemVal := val.Elem()

	if allowMap && elemType.Kind() == reflect.Map {
		return elemType, elemVal, nil
	}

	if elemType.Kind() != reflect.Struct {
		return nil, reflect.Value{}, fmt.Errorf("%w; got %s", ErrReceiverUnsupportedType, elemType.Kind().String())
	}

	return elemType, elemVal, nil
}

func parse(receiver any, tagKey string, data map[string][]string) error {
	if receiver == nil || data == nil {
		return nil
	}

	typ, val, err := receiverElem(receiver, true)
	if err != nil {
		return err
	}

	// Receiver is a map, so copy data to receiver.
	if typ.Kind() == reflect.Map {
		if val.IsNil() {
			val.Set(reflect.MakeMap(typ))
		}

		for k, v := range data {
			if len(v) == 0 {
				continue
			}
			val.SetMapIndex(reflect.ValueOf(k), reflect.ValueOf(v[0]))
		}

		return nil
	}

	fields := cachedFields(typ)
	for _, meta := range fields {
		field := meta.field
		value := val.Field(meta.index)

		if field.Anonymous && value.Kind() == reflect.Ptr {
			if value.IsNil() {
				continue
			}
			value = value.Elem()
		}

		if !value.CanSet() {
			continue
		}

		if err := bindValueToField(field, value, tagKey, data); err != nil {
			return err
		}
	}

	return nil
}

func cachedFields(typ reflect.Type) []fieldMeta {
	if cached, ok := structFieldCache.Load(typ); ok {
		return cached.([]fieldMeta)
	}

	fields := make([]fieldMeta, 0, typ.NumField())
	for i := 0; i < typ.NumField(); i++ {
		fields = append(fields, fieldMeta{index: i, field: typ.Field(i)})
	}

	actual, _ := structFieldCache.LoadOrStore(typ, fields)
	return actual.([]fieldMeta)
}

func bindValueToField(field reflect.StructField, value reflect.Value, tagKey string, data map[string][]string) error {
	valueKind := value.Kind()
	tag := field.Tag.Get(tagKey)

	// handle anonymous struct fields
	if field.Anonymous && valueKind == reflect.Struct && tag != "" {
		return fmt.Errorf("%s %w", tagKey, ErrFieldAnonymousStruct)
	}

	if tag == "" {
		return nil // This property of the struct isn't bindable.
	}

	inputValue, exists := lookupInputValue(tag, data)
	if !exists {
		// check for default tag
		def := field.Tag.Get(defaultTagKey)
		if def != "" {
			exists = true
			inputValue = []string{def}
		}
	}

	if !exists {
		return nil // This property doesn't exist in the input data.
	}

	// Slice should populate from the data slice (converted to correct base type)
	numElems := len(inputValue)
	if valueKind == reflect.Slice && numElems > 0 {
		sliceOf := value.Type().Elem().Kind()
		slice := reflect.MakeSlice(value.Type(), numElems, numElems)

		for j := 0; j < numElems; j++ {
			err := setWithProperType(sliceOf, inputValue[j], slice.Index(j))
			if err != nil { // TODO: test un-parsable type in slice
				return fmt.Errorf("%s is an %w", field.Name, err) // <Type> is an unsupported type
			}
		}

		value.Set(slice)
		return nil
	}

	if len(inputValue) == 0 {
		return nil
	}

	// handle time.Time specifically
	if value.Type() == timeType {
		t, err := time.Parse(tagTimeFormat, inputValue[0])
		if err != nil {
			return fmt.Errorf("%w for %s: %v", ErrFieldTimeFormat, inputValue[0], err)
		}

		value.Set(reflect.ValueOf(t))
		return nil
	}

	// Not a slice, add first string in data for this struct field to the struct.
	err := setWithProperType(field.Type.Kind(), inputValue[0], value)
	if err != nil {
		return fmt.Errorf("%s is an %w", field.Name, err) // <Type> is an unsupported type
	}

	return nil
}

func lookupInputValue(tag string, data map[string][]string) ([]string, bool) {
	inputValue, exists := data[tag]
	if exists {
		return inputValue, true
	}

	// The data didn't have an exact match on the key so let's make sure that the
	// URL parameter isn't masked by case sensitivity.
	for k, v := range data {
		if strings.EqualFold(k, tag) {
			return v, true
		}
	}

	return nil, false
}

func setWithProperType(valueKind reflect.Kind, val string, structField reflect.Value) error {
	switch valueKind {
	case reflect.Ptr:
		if structField.IsNil() {
			structField.Set(reflect.New(structField.Type().Elem()))
		}
		return setWithProperType(structField.Elem().Kind(), val, structField.Elem())
	case reflect.Int:
		return setIntField(val, 0, structField)
	case reflect.Int8:
		return setIntField(val, 8, structField)
	case reflect.Int16:
		return setIntField(val, 16, structField)
	case reflect.Int32:
		return setIntField(val, 32, structField)
	case reflect.Int64:
		return setIntField(val, 64, structField)
	case reflect.Uint:
		return setUintField(val, 0, structField)
	case reflect.Uint8:
		return setUintField(val, 8, structField)
	case reflect.Uint16:
		return setUintField(val, 16, structField)
	case reflect.Uint32:
		return setUintField(val, 32, structField)
	case reflect.Uint64:
		return setUintField(val, 64, structField)
	case reflect.Bool:
		return setBoolField(val, structField)
	case reflect.Float32:
		return setFloatField(val, 32, structField)
	case reflect.Float64:
		return setFloatField(val, 64, structField)
	case reflect.String:
		structField.SetString(val)
	default:
		return fmt.Errorf("%w: %s", ErrFieldUnsupportedType, valueKind.String())
	}
	return nil
}

func setIntField(value string, bitSize int, field reflect.Value) error {
	if value == "" {
		value = "0"
	}
	intVal, err := strconv.ParseInt(value, 10, bitSize)
	if err == nil {
		field.SetInt(intVal)
	}
	return err
}

func setUintField(value string, bitSize int, field reflect.Value) error {
	if value == "" {
		value = "0"
	}
	uintVal, err := strconv.ParseUint(value, 10, bitSize)
	if err == nil {
		field.SetUint(uintVal)
	}
	return err
}

func setBoolField(value string, field reflect.Value) error {
	if value == "" {
		value = "false"
	}
	boolVal, err := strconv.ParseBool(value)
	if err == nil {
		field.SetBool(boolVal)
	}
	return err
}

func setFloatField(value string, bitSize int, field reflect.Value) error {
	if value == "" {
		value = "0.0"
	}
	floatVal, err := strconv.ParseFloat(value, bitSize)
	if err == nil {
		field.SetFloat(floatVal)
	}
	return err
}
