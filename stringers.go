package stramp

import "strconv"

type Stringer func(interface{}) (string, bool)

var (
	Stringers = []Stringer{
		StrampScalar,
	}
)

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