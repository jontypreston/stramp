package stramp

import (
	"reflect"
	"strconv"
)

type Typeifier func(string, reflect.Kind) (interface{}, bool)

var (
	Typeifiers = []Typeifier{
		DeStrampScalar,
	}
)

func Typeify(i string, t reflect.Kind) (interface{}, bool) {
	for _, typeifier := range Typeifiers {
		out, ok := typeifier(i, t)

		if !ok {
			continue
		}

		return out, true
	}

	return nil, false
}

func DeStrampScalar(i string, t reflect.Kind) (interface{}, bool) {
	switch t {
	case reflect.String:
		return i, true
	case reflect.Int:
		x, err := strconv.ParseInt(i, 10, 64)

		if err != nil {
			return nil, false
		}

		return int(x), true
	case reflect.Float64:
		x, err := strconv.ParseFloat(i, 64)

		if err != nil {
			return nil, false
		}

		return x, true
	default:
		return nil, false
	}
}
