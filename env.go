package bind

import (
	"fmt"
	"os"
	"reflect"
	"time"
)

// Env binds environment variables to struct fields tagged with `env`.
//
// Semantics:
// - If an env var is set (including set to an empty string), that value is used.
// - If an env var is not set, the `default` tag is used when present.
func Env(receiver any) error {
	if receiver == nil {
		return nil
	}

	typ, val, err := receiverElem(receiver, false)
	if err != nil {
		return err
	}

	// Loop over all the fields in the struct.
	for i := 0; i < typ.NumField(); i++ {
		field := typ.Field(i)
		value := val.Field(i)

		if field.Anonymous && value.Kind() == reflect.Ptr {
			if value.IsNil() {
				continue
			}
			value = value.Elem()
		}

		if !value.CanSet() {
			continue
		}

		valueKind := value.Kind()
		tag := field.Tag.Get(envTagKey)

		if field.Anonymous && valueKind == reflect.Struct && tag != "" {
			return ErrFieldAnonymousStruct
		}

		if tag == "" {
			continue // This property of the struct isn't bindable.
		}

		envValue, exists := os.LookupEnv(tag)
		if !exists {
			// get default value if defined
			envValue = field.Tag.Get(defaultTagKey)
			if envValue == "" {
				continue // This property doesn't exist in the env and default isn't set.
			}
		}

		// Slice isn't yet supported; TODO: parse env value as csv to populate slice
		if valueKind == reflect.Slice {
			return fmt.Errorf("%w: %s", ErrFieldSliceType, field.Name)
		}

		// handle time.Time specifically
		if value.Type() == timeType {
			t, err := time.Parse(tagTimeFormat, envValue)
			if err != nil {
				return fmt.Errorf("%w for %s: %v", ErrFieldTimeFormat, envValue, err)
			}

			value.Set(reflect.ValueOf(t))
			continue
		}

		// Not a slice, add first string in data for this struct field to the struct.
		err = setWithProperType(field.Type.Kind(), envValue, value)
		if err != nil {
			return fmt.Errorf("%s is an %w", field.Name, err) // <Type> is an unsupported type
		}
	}

	return nil
}
