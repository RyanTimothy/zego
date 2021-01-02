package term

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
