package stramp

import (
	"errors"
	"fmt"
	"reflect"
	"strconv"
)

type UnMarshaler func (prefix string, kv KV, i interface{}) error

var (
	UnMarshalChain []UnMarshaler

	ErrImmutableType = errors.New("destination type is immutable")
	ErrFailedUnMarshal = errors.New("failed to unmarshal")
	ErrMissingKey = errors.New("missing key")
)

func init() {
	UnMarshalChain = []UnMarshaler{
		UnMarshalString,
		UnMarshalInt,
		UnMarshalFloat,
		UnMarshalStruct,
		UnMarshalSlice,
	}
}

func UnMarshal(prefix string, kv KV, i interface{}) error {
	for _, fn := range UnMarshalChain {
		switch err := fn(prefix, kv, i); err {
		case nil:
			return nil
		case ErrTypeMismatch:
			continue
		case ErrMissingKey:
			continue
		default:
			return err
		}
	}

	return fmt.Errorf("%w %s", ErrUnsupportedType, reflect.TypeOf(i))
}

func UnMarshalString(prefix string, kv KV, i interface{}) error {
	if reflect.ValueOf(i).Kind() != reflect.Ptr {
		return ErrImmutableType
	}

	repr, ok := kv[prefix]

	if !ok {
		return ErrMissingKey
	}

	if _, ok := i.(*string); !ok {
		return ErrTypeMismatch
	}

	reflect.ValueOf(i).Elem().Set(reflect.ValueOf(repr))
	return nil
}

func UnMarshalInt(prefix string, kv KV, i interface{}) error {
	if reflect.ValueOf(i).Kind() != reflect.Ptr {
		return ErrImmutableType
	}

	repr, ok := kv[prefix]

	if !ok {
		return ErrMissingKey
	}

	raw, err := strconv.ParseInt(repr, 10, 64)

	if err != nil {
		return ErrTypeMismatch
	}

	var value reflect.Value

	switch i.(type) {
	case *int:
		value = reflect.ValueOf(int(raw))
	case *int8:
		value = reflect.ValueOf(int8(raw))
	case *int16:
		value = reflect.ValueOf(int16(raw))
	case *int32:
		value = reflect.ValueOf(int32(raw))
	case *int64:
		value = reflect.ValueOf(int64(raw))
	default:
		return ErrTypeMismatch
	}

	reflect.ValueOf(i).Elem().Set(value)
	return nil
}

func UnMarshalFloat(prefix string, kv KV, i interface{}) error {
	if reflect.ValueOf(i).Kind() != reflect.Ptr {
		return ErrImmutableType
	}

	repr, ok := kv[prefix]

	if !ok {
		return ErrMissingKey
	}

	raw, err := strconv.ParseFloat(repr, 64)

	if err != nil {
		return ErrTypeMismatch
	}

	var value reflect.Value

	switch i.(type) {
	case *float32:
		value = reflect.ValueOf(float32(raw))
	case *float64:
		value = reflect.ValueOf(raw)
	default:
		return ErrTypeMismatch
	}

	reflect.ValueOf(i).Elem().Set(value)
	return nil
}

func UnMarshalStruct(prefix string, kv KV, i interface{}) error {
	v := reflect.ValueOf(i)

	if v.Kind() != reflect.Ptr {
		return ErrImmutableType
	}

	if v.Elem().Kind() != reflect.Struct {
		return ErrTypeMismatch
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
			return fmt.Errorf("%w %s.%s: %s", ErrFailedUnMarshal, v.Type(), field.Name, err.Error())
		}
	}

	return nil
}

func UnMarshalSlice(prefix string, kv KV, i interface{}) error {
	v := reflect.ValueOf(i)

	if v.Kind() != reflect.Ptr {
		return ErrImmutableType
	}

	v = v.Elem()

	if v.Kind() != reflect.Slice {
		return ErrTypeMismatch
	}

	for i := 0; true; i++ {
		key := Key(prefix, IndexKey(i))

		if _, ok := kv[key]; !ok {
			break
		}

		val := reflect.New(v.Type().Elem())

		if err := UnMarshal(key, kv, val.Interface()); err != nil {
			return fmt.Errorf("%w %s[%d]: %s", ErrFailedUnMarshal, v.Type(), i, err.Error())
		}

		v.Set(reflect.Append(v, val.Elem()))
	}

	return nil
}