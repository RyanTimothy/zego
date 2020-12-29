package term

import (
	"strings"

	"avidbound.com/zego/ast/internal/tokens"
)

// Ref represents a reference as defined by the language.
type Ref []*Term

// RefTerm creates a new Term with a Ref value.
func RefTerm(r ...*Term) *Term {
	return &Term{Value: Ref(r)}
}

// Compare compares str to other, return <0, 0, or >0 if it is less than, equal to,
// or greater than other.
func (v Ref) Compare(other Value) int {
	return Compare(v, other)
}

func (r Ref) String() string {
	if len(r) == 0 {
		return ""
	}
	buf := []string{r[0].Value.String()}
	path := r[1:]
	for _, p := range path {
		switch p := p.Value.(type) {
		case String:
			str := string(p)
			if tokens.Keyword(str) == tokens.Identifier {
				buf = append(buf, "."+str)
			} else {
				buf = append(buf, "["+p.String()+"]")
			}
		default:
			buf = append(buf, "["+p.String()+"]")
		}
	}
	return strings.Join(buf, "")
}
