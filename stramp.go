package stramp

import (
	"errors"
	"fmt"
	"reflect"
	"strconv"
	"strings"
)

type KV map[string]string

func (kv KV) Merge(other KV) {
	for k, v := range other {
		kv[k] = v
	}
}

func (kv KV) Prefixed(prefix string) KV {
	prefixed := make(KV)

	for k, v := range kv {
		prefixed[Key(prefix, k)] = v
	}

	return prefixed
}

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

type IndexKeyFn func(int) string

var (
	Sep = "."
	TagKey = "etcd"
	IndexKey = func(i int) string {return strconv.FormatInt(int64(i), 10)}
)

func Key(parts ...string) string {
	var clean []string

	for _, part := range parts {
		if part != "" {
			clean = append(clean, part)
		}
	}

	return strings.Join(clean, Sep)
}

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