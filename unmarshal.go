package stramp

import (
	"errors"
	"fmt"
	"reflect"
	"strconv"
)

type UnMarshaler func (prefix string, kv KV, i interface{}) bool

var (
	UnMarshalChain []UnMarshaler
	ErrImmutableType = errors.New("immutable type")
)

func init() {
	UnMarshalChain = []UnMarshaler{
		UnMarshalInt,
		UnMarshalString,
		UnMarshalStruct,
		UnMarshalSlice,
	}
}

func UnMarshal(prefix string, kv KV, i interface{}) error {
	if reflect.ValueOf(i).Kind() != reflect.Ptr {
		return fmt.Errorf("%w: %s", ErrImmutableType, reflect.TypeOf(i))
	}

	for _, fn := range UnMarshalChain {
		ok := fn(prefix, kv, i)

		if ok {
			return nil
		}
	}

	return fmt.Errorf("%w: %s (or one of the types within)", ErrUnsupportedType, reflect.TypeOf(i))
}

func UnMarshalInt(prefix string, kv KV, i interface{}) bool {
	repr, ok := kv[prefix]

	if !ok || reflect.ValueOf(i).Kind() != reflect.Ptr {
		return false
	}

	raw, err := strconv.ParseInt(repr, 10, 64)

	if err != nil {
		return false
	}

	var value reflect.Value

	switch i.(type) {
	case *int:
		value = reflect.ValueOf(int(raw))
	case *int8:
		value = reflect.ValueOf(int8(raw))
	default:
		return false
	}

	reflect.ValueOf(i).Elem().Set(value)
	return true
}

func UnMarshalString(prefix string, kv KV, i interface{}) bool {
	repr, ok := kv[prefix]

	if !ok || reflect.ValueOf(i).Kind() != reflect.Ptr {
		return false
	}

	if _, ok := i.(*string); !ok {
		return false
	}

	reflect.ValueOf(i).Elem().Set(reflect.ValueOf(repr))
	return true
}

func UnMarshalStruct(prefix string, kv KV, i interface{}) bool {
	v := reflect.ValueOf(i)

	if v.Kind() != reflect.Ptr {
		return false
	}

	if v.Elem().Kind() != reflect.Struct {
		return false
	}

	v = v.Elem()

	for i := 0; i < v.Type().NumField(); i++ {
		value := v.Field(i)
		field := v.Type().Field(i)

		tag, ok := field.Tag.Lookup(TagKey)

		if !ok {
			if !RequireTag {
				continue
			}

			tag = field.Name
		}

		key := Key(prefix, tag)

		if err := UnMarshal(key, kv, value.Addr().Interface()); err != nil {
			return false
		}
	}

	return true
}

func UnMarshalSlice(prefix string, kv KV, i interface{}) bool {
	v := reflect.ValueOf(i)

	if v.Kind() != reflect.Ptr {
		return false
	}

	v = v.Elem()

	if v.Kind() != reflect.Slice {
		return false
	}

	for i := 0; true; i++ {
		key := Key(prefix, IndexKey(i))

		if _, ok := kv[key]; !ok {
			break
		}

		val := reflect.New(v.Type().Elem())

		if err := UnMarshal(key, kv, val.Interface()); err != nil {
			return false
		}

		v.Set(reflect.Append(v, val.Elem()))
	}

	return true
}