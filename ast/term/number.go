package term

import (
	"encoding/json"

	"github.com/OneOfOne/xxhash"
)

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
	if sort := compareSortOrder(n, other); sort != 0 {
		return sort
	}

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

// Hash returns the hash code for the Value.
func (n Number) Hash() int {
	f, err := json.Number(n).Float64()
	if err != nil {
		bs := []byte(n)
		h := xxhash.Checksum64(bs)
		return int(h)
	}
	return int(f)
}

func (n Number) SortOrder() int {
	return 2
}
