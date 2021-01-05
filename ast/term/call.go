package term

import (
	"fmt"
	"strings"
)

// Call represents as function call in the language.
type Call []*Term

// CallTerm returns a new Term with a Call value defined by terms. The first
// term is the operator and the rest are operands.
func CallTerm(terms ...*Term) *Term {
	return NewTerm(Call(terms))
}

// Copy returns a deep copy of c.
func (c Call) Copy() Call {
	return nil //termSliceCopy(c)
}

// Equal returns true if the other Value is a Call and is equal.
func (c Call) Equal(other Value) bool {
	switch other := other.(type) {
	case Call:
		return c.Compare(other) == 0
	default:
		return false
	}
}

// Compare compares c to other, return <0, 0, or >0 if it is less than, equal to,
// or greater than other.
func (c Call) Compare(other Value) int {
	if sort := compareSortOrder(c, other); sort != 0 {
		return sort
	}

	o := other.(Call)
	return TermSliceCompare(c, o)
}

func (c Call) String() string {
	args := make([]string, len(c)-1)
	for i := 1; i < len(c); i++ {
		args[i-1] = c[i].String()
	}
	return fmt.Sprintf("%v(%v)", c[0], strings.Join(args, ", "))
}

// Hash returns the hash code for the Value.
func (c Call) Hash() int {
	return termSliceHash(c)
}

func (c Call) SortOrder() int {
	return 6
}
