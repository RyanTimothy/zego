package lexer

// Pos represents a byte position in the original input text from which
// this template was parsed.
type Pos struct {
	Index  int // current position in the input
	Line   int // 1+number of newlines seen
	Column int // 1+number of characters seen on line
}

func (p Pos) Position() Pos {
	return p
}
