package stramp

import (
	"errors"
	"reflect"
	"strings"
)

func Lookup(s interface{}, key string) (interface{}, error) {
	keys := strings.Split(key, Sep)

	if len(keys) == 0 {
		return s, nil
	}

	cur := reflect.ValueOf(s)

	for _, x := range keys {
		switch cur.Kind() {
		case reflect.Struct:
			for j := 0; j < cur.Type().NumField(); j++ {
				name, ok := cur.Type().Field(j).Tag.Lookup(TagKey)

				if !ok {
					continue
				}

				if name == x {
					cur = cur.Field(j)
				}
			}

		case reflect.Slice:
			idx := KeyIndex(x)

			if idx >= cur.Len() {
				return nil, errors.New("out-of-bounds slice index")
			}

			cur = cur.Index(idx)

		default:
			return nil, errors.New("nested key on fundamental type")
		}
	}

	return cur.Interface(), nil
}