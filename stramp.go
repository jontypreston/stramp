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

// IndexKeyFn is a function that translates a slice index into a string key.
type IndexKeyFn func(int) string

// KeyIndexFn is a function that translates a string key into a slice index.
type KeyIndexFn func(string) int

var (
	// Sep is the key separator for nested keys
	Sep = "."

	// TagKey is the tag prefix for struct fields
	TagKey = "stramp"

	// RequireTag defines if struct fields which do not have a tag named TagKey should be ignored.
	// If false, the field's name will be used if no tag has been set.
	RequireTag = false

	// IndexKey translates a slice index into a string key
	IndexKey = func(i int) string { return strconv.FormatInt(int64(i), 10) }

	KeyIndex = func(i string) int {
		x, err := strconv.ParseInt(i, 10, 64)

		if err != nil {
			return -1
		}

		return int(x)
	}
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

var (
	ErrNotStruct = errors.New("attempt to stramp non-struct type")
)

func Stramp(i interface{}) (KV, error) {
	if kind := reflect.TypeOf(i).Kind(); kind != reflect.Struct {
		return nil, fmt.Errorf("%w: %s", ErrNotStruct, kind)
	}

	return Marshal("", i)
}

func DeStramp(kv KV, i interface{}) error {
	if reflect.ValueOf(i).Kind() != reflect.Ptr {
		return fmt.Errorf("%w: %s", ErrImmutableType, reflect.TypeOf(i))
	}

	if kind := reflect.ValueOf(i).Elem().Kind(); kind != reflect.Struct {
		return fmt.Errorf("%w: %s", ErrNotStruct, kind)
	}

	return UnMarshal("", kv, i)
}

var (
	ErrUnknownField = errors.New("unknown field")
)

func Get(key string, i interface{}) (interface{}, error) {
	if kind := reflect.TypeOf(i).Kind(); kind != reflect.Struct {
		return nil, fmt.Errorf("%w: %s", ErrNotStruct, kind)
	}

	parts := strings.Split(key, Sep)
	cur := reflect.ValueOf(i)

	for _, part := range parts {
		switch cur.Kind() {
		case reflect.Struct:
			for i := 0; i < cur.NumField(); i++ {
				field := cur.Field(i)

				tag, ok := cur.Type().Field(i).Tag.Lookup(TagKey)

				if !ok {
					if RequireTag {
						continue
					}

					tag = cur.Type().Field(i).Name
				}

				if tag == part {
					cur = field
					break
				}

				// If we're on the last iteration of loop and still haven't found field, it must not exist
				if i + 1 == cur.NumField() {
					return nil, fmt.Errorf("%w: %s", ErrUnknownField, part)
				}
			}

		case reflect.Slice:
			idx := KeyIndex(part)

			if idx < 0 {
				return nil, fmt.Errorf("failed to parse %s as slice index", part)
			}

			if idx > cur.Len() + 1 {
				return nil, fmt.Errorf("index %d is out-of-bounds for slice", idx)
			}

			cur = cur.Index(idx)

		default:
			return nil, fmt.Errorf("nested key %s on scalar value %v", part, cur.Interface())
		}
	}

	return cur.Interface(), nil
}