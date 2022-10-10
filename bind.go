package bind

import (
	"errors"
	"fmt"
	"os"
	"reflect"
	"strconv"
	"strings"
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

	// time format expected in  date field TODO: name better? make configurable with default?
	tagTimeFormat = "2006-01-02"
)

var (
	// cached time type
	timeType = reflect.TypeOf(time.Time{})
)

// TODO: look at github.com/go-playground/validator/v10 struct caching tech to make improvements here.

func Query(receiver any, data map[string][]string) error {
	return parse(receiver, queryTagKey, data)
}

func Form(receiver any, data map[string][]string) error {
	return parse(receiver, formTagKey, data)
}

func Header(receiver any, data map[string][]string) error {
	return parse(receiver, headerTagKey, data)
}

func Env(receiver any) error {
	if receiver == nil { // TODO: test nil
		return nil
	}

	typ := reflect.TypeOf(receiver).Elem()
	val := reflect.ValueOf(receiver).Elem()

	// We do not have a struct, error.
	if typ.Kind() != reflect.Struct { // TODO: test non-struct
		return fmt.Errorf("got %s, expected struct", typ.Kind().String())
	}

	// Loop over all the fields in the struct.
	for i := 0; i < typ.NumField(); i++ {
		field := typ.Field(i)
		value := val.Field(i)

		if field.Anonymous {
			if value.Kind() == reflect.Ptr {
				value = value.Elem()
			}
		}

		if !value.CanSet() {
			continue
		}

		valueKind := value.Kind()
		tag := field.Tag.Get(envTagKey)

		// handle anonymous struct fields
		if field.Anonymous && valueKind == reflect.Struct && tag != "" { // TODO: test anonymous field
			return fmt.Errorf("env tags are not allowed with anonymous struct fields")
		}

		if tag == "" {
			continue // This property of the struct isn't bindable.
		}

		envValue := os.Getenv(tag)
		if envValue == "" {
			envValue = field.Tag.Get(defaultTagKey)
		}

		if envValue == "" {
			continue // This property doesn't exist in the env and default isn't set.
		}

		// Slice isn't yet supported; TODO: parse env value as csv to populate slice
		if valueKind == reflect.Slice { // TODO: test slice
			return fmt.Errorf("slice is not supported on field: %s", field.Name)
		}

		// handle time.Time specifically
		if value.Type() == timeType {
			t, err := time.Parse(tagTimeFormat, envValue)
			if err != nil { // TODO: test incorrect time value
				return fmt.Errorf("unable to parse time for %s: %v", envValue, err)
			}

			value.Set(reflect.ValueOf(t))
			continue
		}

		// Not a slice, add first string in data for this struct field to the struct.
		err := setWithProperType(field.Type.Kind(), envValue, value)
		if err != nil { // todo: test erring prop type
			return fmt.Errorf("%v", err)
		}
	}

	return nil

}

func parse(receiver any, tagKey string, data map[string][]string) error {
	if receiver == nil || data == nil {
		return nil
	}

	typ := reflect.TypeOf(receiver).Elem()
	val := reflect.ValueOf(receiver).Elem()

	// Receiver is a map, so copy data to receiver.
	if typ.Kind() == reflect.Map {
		for k, v := range data {
			val.SetMapIndex(reflect.ValueOf(k), reflect.ValueOf(v[0]))
		}

		return nil
	}

	// We do not have a struct, error.
	if typ.Kind() != reflect.Struct {
		return fmt.Errorf("got %s, expected struct", typ.Kind().String())
	}

	// Loop over all the fields in the struct.
	for i := 0; i < typ.NumField(); i++ {
		field := typ.Field(i)
		value := val.Field(i)

		if field.Anonymous {
			if value.Kind() == reflect.Ptr {
				value = value.Elem()
			}
		}

		if !value.CanSet() {
			continue
		}

		valueKind := value.Kind()
		tag := field.Tag.Get(tagKey)

		// handle anonymous struct fields
		if field.Anonymous && valueKind == reflect.Struct && tag != "" { // TODO: test anonymous field
			return fmt.Errorf("%s tags are not allowed with anonymous struct fields", tagKey)
		}

		if tag == "" {
			continue // This property of the struct isn't bindable.
		}

		inputValue, exists := data[tag]
		if !exists { // TODO: test messed up key casing in data
			// The data didn't have an exact match on the key so let's make sure that the
			// URL parameter isn't masked by case sensitivity.
			for k, v := range data {
				if strings.EqualFold(k, tag) {
					inputValue = v
					exists = true
					break
				}
			}
		}

		if !exists {
			continue // This property doesn't exist in the query data.
		}

		// Slice should populate from the data slice (converted to correct base type)
		numElems := len(inputValue)
		if valueKind == reflect.Slice && numElems > 0 {
			sliceOf := value.Type().Elem().Kind()
			slice := reflect.MakeSlice(value.Type(), numElems, numElems)

			for j := 0; j < numElems; j++ {
				err := setWithProperType(sliceOf, inputValue[j], slice.Index(j))
				if err != nil { // TODO: test un-parsable type in slice
					return fmt.Errorf("%v", err)
				}
			}

			val.Field(i).Set(slice)

			continue
		}

		// handle time.Time specifically
		if value.Type() == timeType {
			t, err := time.Parse(tagTimeFormat, inputValue[0])
			if err != nil { // TODO: test un-parsable time value in data
				return fmt.Errorf("unable to parse time for %s: %v", inputValue[0], err)
			}

			value.Set(reflect.ValueOf(t))
			continue
		}

		// Not a slice, add first string in data for this struct field to the struct.
		err := setWithProperType(field.Type.Kind(), inputValue[0], value)
		if err != nil { // TODO: test un-parsable type in struct
			return fmt.Errorf("%v", err)
		}
	}

	return nil
}

func setWithProperType(valueKind reflect.Kind, val string, structField reflect.Value) error {
	switch valueKind {
	case reflect.Ptr:
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
		return errors.New("unknown type")
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
