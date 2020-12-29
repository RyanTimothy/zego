package term

import "strconv"

// String represents a string value as defined by JSON.
type String string

// StringTerm creates a new Term with a String value.
func StringTerm(s string) *Term {
	return &Term{Value: String(s)}
}

// Equal returns true if the other Value is a String and is equal.
func (s String) Equal(other Value) bool {
	switch other := other.(type) {
	case String:
		return s == other
	default:
		return false
	}
}

// Compare compares str to other, return <0, 0, or >0 if it is less than, equal to,
// or greater than other.
func (s String) Compare(other Value) int {
	return Compare(s, other)
}

func (s String) String() string {
	return strconv.Quote(string(s))
}
