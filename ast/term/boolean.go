package term

import "strconv"

// Boolean represents a boolean value defined by JSON.
type Boolean bool

// BooleanTerm creates a new Term with a Boolean value.
func BooleanTerm(b bool) *Term {
	return &Term{Value: Boolean(b)}
}

// Compare compares bol to other, return <0, 0, or >0 if it is less than, equal to,
// or greater than other.
func (bol Boolean) Compare(other Value) int {
	return Compare(bol, other)
}

func (bol Boolean) String() string {
	return strconv.FormatBool(bool(bol))
}
