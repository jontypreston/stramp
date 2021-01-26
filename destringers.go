package stramp

import (
	"reflect"
	"strconv"
)

// Typeifier is a function which parses a string into the kind given.
type Typeifier func(string, reflect.Kind) (interface{}, bool)

var (
	Typeifiers = []Typeifier{
		DeStrampScalar,
	}
)

// Typeify parses a given string into a given type.
// If no function exists which supports the given type, false is returned.
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

// DeStrampScalar parses a given string into a built-in scalar type.
// Returns false if the given kind is not supported.
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
