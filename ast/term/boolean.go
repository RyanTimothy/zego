package term

import "strconv"

// Boolean represents a boolean value defined by JSON.
type Boolean bool

// BooleanTerm creates a new Term with a Boolean value.
func BooleanTerm(b bool) *Term {
	return &Term{Value: Boolean(b)}
}

// Equal returns true if the other Value is a Boolean and is equal.
func (b Boolean) Equal(other Value) bool {
	switch other := other.(type) {
	case Boolean:
		return b == other
	default:
		return false
	}
}

// Compare compares bol to other, return <0, 0, or >0 if it is less than, equal to,
// or greater than other.
func (b Boolean) Compare(other Value) int {
	if sort := compareSortOrder(b, other); sort != 0 {
		return sort
	}

	o := other.(Boolean)
	if b.Equal(o) {
		return 0
	}
	if !b {
		return -1
	}
	return 1
}

func (b Boolean) String() string {
	return strconv.FormatBool(bool(b))
}

// Hash returns the hash code for the Value.
func (b Boolean) Hash() int {
	if b {
		return 1
	}
	return 0
}

func (b Boolean) SortOrder() int {
	return 1
}
