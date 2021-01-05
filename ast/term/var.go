package term

import "github.com/OneOfOne/xxhash"

type Var string

// VarTerm creates a new Term with a Variable value.
func VarTerm(s string) *Term {
	return &Term{Value: Var(s)}
}

// Equal returns true if the other Value is a Variable and has the same value (name).
func (v Var) Equal(other Value) bool {
	switch other := other.(type) {
	case Var:
		return v == other
	default:
		return false
	}
}

// Compare compares str to other, return <0, 0, or >0 if it is less than, equal to,
// or greater than other.
func (v Var) Compare(other Value) int {
	if sort := compareSortOrder(v, other); sort != 0 {
		return sort
	}

	o := other.(Var)
	if v.Equal(o) {
		return 0
	}
	if v < o {
		return -1
	}
	return 1
}

func (v Var) String() string {
	return string(v)
}

// Hash returns the hash code for the Value.
func (v Var) Hash() int {
	h := xxhash.ChecksumString64S(string(v), hashSeed0)
	return int(h)
}

func (v Var) SortOrder() int {
	return 4
}
