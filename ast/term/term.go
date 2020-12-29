package term

import (
	"fmt"
)

type (
	Term struct {
		Value    Value     `json:"value"` // the value of the Term as represented in Go
		Location *Location `json:"-"`     // the location of the Term in the source
	}

	Value interface {
		Compare(other Value) int // Compare returns <0, 0, or >0 if this Value is less than, equal to, or greater than other, respectively.
		String() string          // String returns a human readable string representation of the value.
	}

	Location struct {
	}
)

// NewTerm returns a new Term object.
func NewTerm(v Value) *Term {
	return &Term{
		Value: v,
	}
}

func (t *Term) SetLoc(l *Location) *Term {
	t.Location = l // TODO: set location
	return t
}

func (term *Term) String() string {
	return term.Value.String()
}

func Compare(a, b interface{}) int {

	if t, ok := a.(*Term); ok {
		if t == nil {
			a = nil
		} else {
			a = t.Value
		}
	}

	if t, ok := b.(*Term); ok {
		if t == nil {
			b = nil
		} else {
			b = t.Value
		}
	}

	if a == nil {
		if b == nil {
			return 0
		}
		return -1
	}
	if b == nil {
		return 1
	}

	sortA := sortOrder(a)
	sortB := sortOrder(b)

	if sortA < sortB {
		return -1
	} else if sortB < sortA {
		return 1
	}

	switch a := a.(type) {
	// case Null:
	// 	return 0
	// case Boolean:
	// 	b := b.(Boolean)
	// 	if a.Equal(b) {
	// 		return 0
	// 	}
	// 	if !a {
	// 		return -1
	// 	}
	// 	return 1
	// case Number:
	// 	if ai, err := json.Number(a).Int64(); err == nil {
	// 		if bi, err := json.Number(b.(Number)).Int64(); err == nil {
	// 			if ai == bi {
	// 				return 0
	// 			}
	// 			if ai < bi {
	// 				return -1
	// 			}
	// 			return 1
	// 		}
	// 	}

	// 	bigA, ok := new(big.Float).SetString(string(a))
	// 	if !ok {
	// 		panic("illegal value")
	// 	}
	// 	bigB, ok := new(big.Float).SetString(string(b.(Number)))
	// 	if !ok {
	// 		panic("illegal value")
	// 	}
	// 	return bigA.Cmp(bigB)
	case String:
		b := b.(String)
		if a.Equal(b) {
			return 0
		}
		if a < b {
			return -1
		}
		return 1
	case Var:
		b := b.(Var)
		if a.Equal(b) {
			return 0
		}
		if a < b {
			return -1
		}
		return 1
		// case Ref:
		// 	b := b.(Ref)
		// 	return termSliceCompare(a, b)
		// case *Array:
		// 	b := b.(*Array)
		// 	return termSliceCompare(a.elems, b.elems)
		// case *object:
		// 	b := b.(*object)
		// 	return a.Compare(b)
		// case Set:
		// 	b := b.(Set)
		// 	return a.Compare(b)
		// case *ArrayComprehension:
		// 	b := b.(*ArrayComprehension)
		// 	if cmp := Compare(a.Term, b.Term); cmp != 0 {
		// 		return cmp
		// 	}
		// 	return Compare(a.Body, b.Body)
		// case *ObjectComprehension:
		// 	b := b.(*ObjectComprehension)
		// 	if cmp := Compare(a.Key, b.Key); cmp != 0 {
		// 		return cmp
		// 	}
		// 	if cmp := Compare(a.Value, b.Value); cmp != 0 {
		// 		return cmp
		// 	}
		// 	return Compare(a.Body, b.Body)
		// case *SetComprehension:
		// 	b := b.(*SetComprehension)
		// 	if cmp := Compare(a.Term, b.Term); cmp != 0 {
		// 		return cmp
		// 	}
		// 	return Compare(a.Body, b.Body)
		// case Call:
		// 	b := b.(Call)
		// 	return termSliceCompare(a, b)
		// case *Expr:
		// 	b := b.(*Expr)
		// 	return a.Compare(b)
		// case *SomeDecl:
		// 	b := b.(*SomeDecl)
		// 	return a.Compare(b)
		// case *With:
		// 	b := b.(*With)
		// 	return a.Compare(b)
		// case Body:
		// 	b := b.(Body)
		// 	return a.Compare(b)
		// case *Head:
		// 	b := b.(*Head)
		// 	return a.Compare(b)
		// case *Rule:
		// 	b := b.(*Rule)
		// 	return a.Compare(b)
		// case Args:
		// 	b := b.(Args)
		// 	return termSliceCompare(a, b)
		// case *Import:
		// 	b := b.(*Import)
		// 	return a.Compare(b)
		// case *Package:
		// 	b := b.(*Package)
		// 	return a.Compare(b)
		// case *Module:
		// 	b := b.(*Module)
		// 	return a.Compare(b)
	}
	panic(fmt.Sprintf("illegal value: %T", a))
}

func sortOrder(x interface{}) int {
	switch x.(type) {
	// case Null:
	// 	return 0
	// case Boolean:
	// 	return 1
	// case Number:
	// 	return 2
	case String:
		return 3
	case Var:
		return 4
	case Ref:
		return 5
		// case *Array:
		// 	return 6
		// case Object:
		// 	return 7
		// case Set:
		// 	return 8
		// case *ArrayComprehension:
		// 	return 9
		// case *ObjectComprehension:
		// 	return 10
		// case *SetComprehension:
		// 	return 11
		// case Call:
		// 	return 12
		// case Args:
		// 	return 13
		// case *Expr:
		// 	return 100
		// case *SomeDecl:
		// 	return 101
		// case *With:
		// 	return 110
		// case *Head:
		// 	return 120
		// case Body:
		// 	return 200
		// case *Rule:
		// 	return 1000
		// case *Import:
		// 	return 1001
		// case *Package:
		// 	return 1002
		// case *Module:
		// 	return 10000
	}
	panic(fmt.Sprintf("illegal value: %T", x))
}
