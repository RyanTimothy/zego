package term

import "encoding/json"

// Number represents a numeric value as defined by JSON.
type Number json.Number

// NumberTerm creates a new Term with a Number value.
func NumberTerm(n json.Number) *Term {
	return &Term{Value: Number(n)}
}

// Equal returns true if the other Value is a Number and is equal.
func (n Number) Equal(other Value) bool {
	switch other := other.(type) {
	case Number:
		return n.Compare(other) == 0
	default:
		return false
	}
}

// Compare compares num to other, return <0, 0, or >0 if it is less than, equal to,
// or greater than other.
func (n Number) Compare(other Value) int {
	if ai, err := json.Number(n).Int64(); err == nil {
		if bi, err := json.Number(other.(Number)).Int64(); err == nil {
			if ai == bi {
				return 0
			}
			if ai < bi {
				return -1
			}
			return 1
		}
	}
	return 0
}

func (n Number) String() string {
	return string(n)
}
