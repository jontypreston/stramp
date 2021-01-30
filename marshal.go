package stramp

import (
	"errors"
	"fmt"
	"reflect"
	"strconv"
)

type Marshaler func (prefix string, i interface{}) (KV, error)

var (
	MarshalChain []Marshaler

	ErrTypeMismatch = errors.New("type mismatch")
	ErrUnsupportedType = errors.New("unsupported type")
	ErrFailedMarshal = errors.New("failed to marshal")
)

func init() {
	MarshalChain = []Marshaler{
		MarshalString,
		MarshalInt,
		MarshalFloat,
		MarshalStruct,
		MarshalSlice,
	}
}

func Marshal(prefix string, i interface{}) (KV, error) {
	for _, fn := range MarshalChain {
		kv, err := fn(prefix, i)

		switch err {
		case nil:
			return kv, nil

		case ErrTypeMismatch:
			continue

		default:
			return nil, err
		}
	}

	return nil, fmt.Errorf("%w %s", ErrUnsupportedType, reflect.TypeOf(i))
}

func MarshalString(prefix string, i interface{}) (KV, error) {
	x, ok := i.(string)

	if !ok {
		return nil, ErrTypeMismatch
	}

	kv := make(KV)
	kv[prefix] = x

	return kv, nil
}

func MarshalInt(prefix string, i interface{}) (KV, error) {
	var x int64

	switch i := i.(type) {
	case int:
		x = int64(i)
	case int8:
		x = int64(i)
	case int16:
		x = int64(i)
	case int32:
		x = int64(i)
	case int64:
		x = i
	default:
		return nil, ErrTypeMismatch
	}

	kv := make(KV)
	kv[prefix] = strconv.FormatInt(x, 10)

	return kv, nil
}

func MarshalFloat(prefix string, i interface{}) (KV, error) {
	var x float64

	switch i := i.(type) {
	case float32:
		x = float64(i)
	case float64:
		x = i
	default:
		return nil, ErrTypeMismatch
	}

	kv := make(KV)
	kv[prefix] = strconv.FormatFloat(x, 'E', -1, 64)

	return kv, nil
}

func MarshalStruct(prefix string, i interface{}) (KV, error) {
	v := reflect.ValueOf(i)

	if v.Kind() != reflect.Struct {
		return nil, ErrTypeMismatch
	}

	kv := make(KV)

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

		nest, err := Marshal(key, value.Interface())

		if err != nil {
			return nil, fmt.Errorf("%w %s.%s: %s", ErrFailedMarshal, v.Type(), field.Name, err.Error())
		}

		kv.Merge(nest)
	}

	return kv, nil
}

func MarshalSlice(prefix string, i interface{}) (KV, error) {
	v := reflect.ValueOf(i)

	if v.Kind() != reflect.Slice {
		return nil, ErrTypeMismatch
	}

	kv := make(KV)

	for i := 0; i < v.Len(); i++ {
		key := Key(prefix, IndexKey(i))
		x := v.Index(i).Interface()

		nest, err := Marshal(key, x)

		if err != nil {
			return nil, fmt.Errorf("%w %s[%d]: %s", ErrFailedMarshal, v.Type(), i, err.Error())
		}

		kv.Merge(nest)
	}

	return kv, nil
}