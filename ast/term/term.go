package term

import (
	"math/rand"
	"time"
)

var hashSeed = rand.New(rand.NewSource(time.Now().UnixNano()))
var hashSeed0 = (uint64(hashSeed.Uint32()) << 32) | uint64(hashSeed.Uint32())

type (
	Term struct {
		Value    Value     `json:"value"` // the value of the Term as represented in Go
		Location *Location `json:"-"`     // the location of the Term in the source
	}

	Value interface {
		Equal(other Value) bool
		Compare(other Value) int // Compare returns <0, 0, or >0 if this Value is less than, equal to, or greater than other, respectively.
		String() string          // String returns a human readable string representation of the value.
		Hash() int               // Returns hash code of the value.
		SortOrder() int          // Returns the sort order of the value.
	}

	Location struct {
		File   string `json:"file"`
		Line   int    `json:"line"`
		Column int    `json:"column"`
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

func (t *Term) Compare(other *Term) int {
	switch v := t.Value.(type) {
	case Boolean, Call, Number, Op, Ref, String, Var:
		return v.Compare(other.Value)
	}
	return 0
}

func (t *Term) String() string {
	return t.Value.String()
}

func compareSortOrder(a, b Value) int {
	ao := a.SortOrder()
	bo := b.SortOrder()

	if ao < bo {
		return -1
	} else if bo < ao {
		return 1
	}

	return 0
}

func TermSliceCompare(a, b []*Term) int {
	minLen := len(a)
	if len(b) < minLen {
		minLen = len(b)
	}
	for i := 0; i < minLen; i++ {
		if cmp := a[i].Value.Compare(b[i].Value); cmp != 0 {
			return cmp
		}
	}
	if len(a) < len(b) {
		return -1
	} else if len(b) < len(a) {
		return 1
	}
	return 0
}

func termSliceHash(a []*Term) int {
	var hash int
	for _, v := range a {
		hash += v.Value.Hash()
	}
	return hash
}
