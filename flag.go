package bind

import (
	"flag"
	"fmt"
	"reflect"
	"time"
)

func RegisterFlags(receivers ...any) {
	for _, receiver := range receivers {
		// get struct tag for all properties of receiver
		typ := reflect.TypeOf(receiver)
		if typ.Kind() == reflect.Ptr {
			typ = typ.Elem()
		}

		if typ.Kind() != reflect.Struct {
			// receiver must be a struct or a pointer to a struct
			return
		}

		for i := 0; i < typ.NumField(); i++ {
			field := typ.Field(i)
			tag := field.Tag.Get(flagTagKey)

			if tag == "" {
				continue
			}

			flag.String(tag, "", field.Name)
		}
	}

	flag.Parse()
}

func Flag(receiver any) error {
	if receiver == nil {
		return nil
	}

	typ := reflect.TypeOf(receiver).Elem()
	val := reflect.ValueOf(receiver).Elem()

	if typ.Kind() != reflect.Struct {
		return fmt.Errorf("%w; got %s", ErrReceiverUnsupportedType, typ.Kind().String())
	}

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
		tag := field.Tag.Get(flagTagKey)

		if field.Anonymous && valueKind == reflect.Struct && tag != "" {
			return ErrFieldAnonymousStruct
		}

		if tag == "" {
			continue // This property of the struct isn't bindable.
		}

		var flagValue string
		flagPtr := flag.Lookup(tag)

		if flagPtr == nil {
			// get default value from struct tag `default`
			def := field.Tag.Get(defaultTagKey)

			if def == "" {
				continue // This property doesn't exist in the flags or have a default value
			}

			flagValue = def
		} else {
			flagValue = flagPtr.Value.String()
		}

		// Slice isn't yet supported; TODO: parse flag value as csv?
		if valueKind == reflect.Slice {
			return fmt.Errorf("%w: %s", ErrFieldSliceType, field.Name)
		}

		// handle time.Time specifically
		if value.Type() == timeType {
			t, err := time.Parse(tagTimeFormat, flagValue)
			if err != nil {
				return fmt.Errorf("%w for %s: %v", ErrFieldTimeFormat, flagValue, err)
			}

			value.Set(reflect.ValueOf(t))
			continue
		}

		// Not a slice, add first string in data for this struct field to the struct.
		err := setWithProperType(field.Type.Kind(), flagValue, value)
		if err != nil {
			return fmt.Errorf("%s is an %w", field.Name, err) // <Type> is an unsupported type
		}
	}

	return nil
}
