package term

type Op string

// OpTerm creates a new Term with a Operator value.
func OpTerm(s string) *Term {
	return &Term{Value: Op(s)}
}

// Equal returns true if the other Value is a Variable and has the same value (name).
func (o Op) Equal(other Value) bool {
	switch other := other.(type) {
	case Op:
		return o == other
	default:
		return false
	}
}

// Compare compares str to other, return <0, 0, or >0 if it is less than, equal to,
// or greater than other.
func (o Op) Compare(other Value) int {
	return Compare(o, other)
}

func (o Op) String() string {
	return string(o)
}
