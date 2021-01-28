package stramp

import (
	"errors"
	"fmt"
	"reflect"
	"strconv"
)

type Marshaler func (prefix string, i interface{}) (KV, bool)

var (
	MarshalChain []Marshaler
	ErrUnsupportedType = errors.New("unsupported type")
)

func init() {
	MarshalChain = []Marshaler{
		MarshalInt,
		MarshalString,
		MarshalFloat,
		MarshalStruct,
		MarshalSlice,
	}
}

func Marshal(prefix string, i interface{}) (KV, error) {
	for _, fn := range MarshalChain {
		kv, ok := fn(prefix, i)

		if ok {
			return kv, nil
		}
	}

	return nil, fmt.Errorf("%w: %s (or one of the types within)", ErrUnsupportedType, reflect.TypeOf(i))
}

func MarshalInt(prefix string, i interface{}) (KV, bool) {
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
		return nil, false
	}

	kv := make(KV)
	kv[prefix] = strconv.FormatInt(x, 10)

	return kv, true
}

func MarshalString(prefix string, i interface{}) (KV, bool) {
	x, ok := i.(string)

	if !ok {
		return nil, false
	}

	kv := make(KV)
	kv[prefix] = x

	return kv, true
}

func MarshalFloat(prefix string, i interface{}) (KV, bool) {
	var x float64

	switch i := i.(type) {
	case float32:
		x = float64(i)
	case float64:
		x = i
	default:
		return nil, false
	}

	kv := make(KV)
	kv[prefix] = strconv.FormatFloat(x, 'E', -1, 64)

	return kv, true
}

func MarshalStruct(prefix string, i interface{}) (KV, bool) {
	v := reflect.ValueOf(i)

	if v.Kind() != reflect.Struct {
		return nil, false
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
			return nil, false
		}

		kv.Merge(nest)
	}

	return kv, true
}

func MarshalSlice(prefix string, i interface{}) (KV, bool) {
	v := reflect.ValueOf(i)

	if v.Kind() != reflect.Slice {
		return nil, false
	}

	kv := make(KV)

	for i := 0; i < v.Len(); i++ {
		key := Key(prefix, IndexKey(i))
		x := v.Index(i).Interface()

		nest, err := Marshal(key, x)

		if err != nil {
			return nil, false
		}

		kv.Merge(nest)
	}

	return kv, true
}