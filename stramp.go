package stramp

import (
	"errors"
	"fmt"
	"reflect"
	"strconv"
	"strings"
)

// KV is a string -> string mapping
type KV map[string]string

// Merge merges in the key-value pairs from the other KV.
// In case of duplicate keys, those from the other KV take precedence.
func (kv KV) Merge(other KV) {
	for k, v := range other {
		kv[k] = v
	}
}

// Prefixed returns a new KV where all of the keys have been prefixed.
func (kv KV) Prefixed(prefix string) KV {
	prefixed := make(KV)

	for k, v := range kv {
		prefixed[Key(prefix, k)] = v
	}

	return prefixed
}

// Pop returns a new KV containing only the keys which have the prefix given.
// The prefix is stripped from the new KV.
func (kv KV) Pop(prefix string) KV {
	popped := make(KV)

	for k, v := range kv {
		parts := strings.Split(k, Sep)

		if len(parts) > 0 && parts[0] == prefix {
			popped[Key(parts[1:]...)] = v
		}
	}

	return popped
}

// IndexKeyFn is a function that translates a slice index into a string key.
type IndexKeyFn func(int) string

var (
	// Sep is the key separator for nested keys
	Sep = "."

	// TagKey is the tag prefix for struct fields
	TagKey = "etcd"

	// IndexKey translates a slice index into a string key
	IndexKey = func(i int) string { return strconv.FormatInt(int64(i), 10) }
)

// Key joins all non-empty strings given into a single string key using the Sep separator.
func Key(parts ...string) string {
	var clean []string

	for _, part := range parts {
		if part != "" {
			clean = append(clean, part)
		}
	}

	return strings.Join(clean, Sep)
}

// Stramp converts a struct into a KV map.
func Stramp(i interface{}) (KV, error) {
	kv := make(KV)

	source := reflect.ValueOf(i)
	blueprint := source.Type()

	if source.Kind() != reflect.Struct {
		return nil, errors.New("non-struct type given")
	}

	for i := 0; i < blueprint.NumField(); i++ {
		field := blueprint.Field(i)
		value := source.Field(i)

		name, ok := field.Tag.Lookup(TagKey)

		if !ok {
			continue
		}

		switch value.Kind() {
		case reflect.Struct:
			nested, err := Stramp(value.Interface())

			if err != nil {
				return nil, err
			}

			kv.Merge(nested.Prefixed(name))

		case reflect.Slice:
			for j := 0; j < value.Len(); j++ {
				item := value.Index(j)

				conv, ok := Stringify(item.Interface())

				if !ok {
					return nil, fmt.Errorf("unsupported type %s", field.Type)
				}

				kv[Key(name, IndexKey(j))] = conv
			}

		default:
			conv, ok := Stringify(value.Interface())

			if !ok {
				return nil, fmt.Errorf("unsupported type %s", field.Type)
			}

			kv[name] = conv
		}
	}

	return kv, nil
}

// DeStramp populates a given struct using the KV provided.
// The struct must be mutable (passed as a pointer).
func DeStramp(i interface{}, kv KV) error {
	ptr := reflect.ValueOf(i)

	if ptr.Kind() != reflect.Ptr {
		return errors.New("non-pointer type given")
	}

	dest := ptr.Elem()
	blueprint := dest.Type()

	for i := 0; i < blueprint.NumField(); i++ {
		field := blueprint.Field(i)
		value := dest.Field(i)

		name, ok := field.Tag.Lookup(TagKey)

		if !ok {
			continue
		}

		switch field.Type.Kind() {
		case reflect.Struct:
			if err := DeStramp(value.Addr().Interface(), kv.Pop(name)); err != nil {
				return err
			}

		case reflect.Slice:
			for j := 0; true; j++ {
				repr, ok := kv[Key(name, IndexKey(j))]

				if !ok {
					break
				}

				conv, ok := Typeify(repr, field.Type.Elem().Kind())

				if !ok {
					continue
				}

				value.Set(reflect.Append(value, reflect.ValueOf(conv)))
			}

		default:
			repr, ok := kv[name]

			if !ok {
				continue
			}

			conv, ok := Typeify(repr, field.Type.Kind())

			if !ok {
				continue
			}

			if value.IsValid() && value.CanSet() {
				value.Set(reflect.ValueOf(conv))
			}
		}
	}

	return nil
}
