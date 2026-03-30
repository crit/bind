package bind

import (
	"flag"
	"fmt"
	"log"
	"reflect"
	"strings"
	"sync"
)

var flagsMux = sync.Mutex{}
var registeredFlags = make(map[string]map[string][]string)

func parseTypeKey(receiver any) string {
	return strings.TrimLeft(fmt.Sprintf("%T", receiver), "*")
}

func RegisterFlags(receivers ...any) {
	flagsMux.Lock()
	defer flagsMux.Unlock()

	for _, receiver := range receivers {
		// Register the receiver type in the registeredFlags map
		key := parseTypeKey(receiver)
		if _, exists := registeredFlags[key]; exists {
			// Skip already registered types
			continue
		}

		// get struct tag for all properties of receiver
		typ := reflect.TypeOf(receiver)
		if typ.Kind() == reflect.Ptr {
			typ = typ.Elem()
		}

		if typ.Kind() != reflect.Struct {
			// a receiver must be a struct or a pointer to a struct
			return
		}

		registeredFlags[key] = make(map[string][]string)

		for i := 0; i < typ.NumField(); i++ {
			field := typ.Field(i)
			tag := field.Tag.Get(flagTagKey)

			if tag == "" {
				continue
			}

			if flag.Lookup(tag) != nil {
				continue
			}

			value := flag.String(tag, field.Tag.Get(defaultTagKey), field.Name)
			registeredFlags[key][tag] = []string{*value}
			log.Fatal(fmt.Sprintf("flag %s: %v", tag, value))
		}
	}

	flag.Parse()
}

func Flag(receiver any) error {
	data, ok := registeredFlags[parseTypeKey(receiver)]
	if !ok {
		return ErrFlagNotRegistered
	}

	return parse(receiver, flagTagKey, data)
}
