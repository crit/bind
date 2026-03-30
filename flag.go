package bind

import (
	"flag"
	"fmt"
	"reflect"
	"strings"
	"sync"
)

var flagsMux = sync.Mutex{}
var registeredFlags = make(map[string]map[string]struct{})

func parseTypeKey(receiver any) string {
	return strings.TrimLeft(fmt.Sprintf("%T", receiver), "*")
}

// RegisterFlags registers all `flag` tags for the provided receiver types on the
// default command line FlagSet.
func RegisterFlags(receivers ...any) error {
	return RegisterFlagsWithSet(flag.CommandLine, receivers...)
}

// RegisterFlagsWithSet registers all `flag` tags for the provided receiver types
// on the given FlagSet.
func RegisterFlagsWithSet(fs *flag.FlagSet, receivers ...any) error {
	if fs == nil {
		return fmt.Errorf("%w: nil FlagSet", ErrUnknown)
	}

	flagsMux.Lock()
	defer flagsMux.Unlock()

	for _, receiver := range receivers {
		typ := reflect.TypeOf(receiver)
		if typ == nil {
			return fmt.Errorf("%w; got <nil>", ErrReceiverUnsupportedType)
		}

		if typ.Kind() == reflect.Ptr {
			typ = typ.Elem()
		}

		if typ.Kind() != reflect.Struct {
			return fmt.Errorf("%w; got %s", ErrReceiverUnsupportedType, typ.Kind().String())
		}

		// Register the receiver type in the registeredFlags map
		key := parseTypeKey(receiver)
		if _, exists := registeredFlags[key]; exists {
			// Skip already registered types
			continue
		}

		registeredFlags[key] = make(map[string]struct{})

		for i := 0; i < typ.NumField(); i++ {
			field := typ.Field(i)
			tag := field.Tag.Get(flagTagKey)

			if tag == "" {
				continue
			}

			if fs.Lookup(tag) == nil {
				_ = fs.String(tag, field.Tag.Get(defaultTagKey), field.Name)
			}
			registeredFlags[key][tag] = struct{}{}
		}
	}

	// Keep old convenience behavior for the default command line only.
	if fs == flag.CommandLine && !flag.Parsed() {
		flag.Parse()
	}

	return nil
}

// Flag binds values from the default command-line FlagSet into receiver.
//
// Nil receiver is a no-op. Non-pointer, typed nil pointer, or unsupported
// receiver kinds return ErrReceiverUnsupportedType.
func Flag(receiver any) error {
	return FlagWithSet(flag.CommandLine, receiver)
}

func FlagWithSet(fs *flag.FlagSet, receiver any) error {
	if receiver == nil {
		return nil
	}
	if fs == nil {
		return fmt.Errorf("%w: nil FlagSet", ErrUnknown)
	}

	receiverType, _, err := receiverElem(receiver, false)
	if err != nil {
		return err
	}

	for i := 0; i < receiverType.NumField(); i++ {
		field := receiverType.Field(i)
		if field.Tag.Get(flagTagKey) == "" {
			continue
		}
		if field.Type.Kind() == reflect.Slice {
			continue // supported via CSV parsing in parse path
		}
	}

	flagsMux.Lock()
	tags, ok := registeredFlags[parseTypeKey(receiver)]
	flagsMux.Unlock()
	if !ok {
		return ErrFlagNotRegistered
	}

	data := make(map[string][]string, len(tags))
	for tag := range tags {
		if f := fs.Lookup(tag); f != nil {
			data[tag] = []string{f.Value.String()}
		}
	}

	return parse(receiver, flagTagKey, data)
}
