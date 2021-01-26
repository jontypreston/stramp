package stramp

import "strconv"

// Stringer is a function which parses the given value into a string.
type Stringer func(interface{}) (string, bool)

var (
	Stringers = []Stringer{
		StrampScalar,
	}
)

// Stringify parses the given value into a string.
// If no function exists which supports the given type, false is returned.
func Stringify(i interface{}) (string, bool) {
	for _, stringer := range Stringers {
		out, ok := stringer(i)

		if !ok {
			continue
		}

		return out, true
	}

	return "", false
}

// StrampScalar parses a given scalar value into a string.
// Returns false if the given kind of scalar is not supported.
func StrampScalar(i interface{}) (string, bool) {
	switch i := i.(type) {
	case string:
		return i, true
	case int:
		return strconv.FormatInt(int64(i), 10), true
	case float64:
		return strconv.FormatFloat(i, 'E', -1, 64), true
	default:
		return "", false
	}
}