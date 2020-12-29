package parser

import (
	"fmt"
	"testing"

	"avidbound.com/zego/ast/internal/lexer"
)

func TestParseTermRelation(t *testing.T) {
	// Arrange
	input := `input.c == (input.a[input.b[0]] + c) * d`
	p := parser{
		items: lexer.Lex("test.zego", input),
	}

	p.state = &state{
		parser: &p,
		index:  0,
	}

	// Act
	term := p.parseTermRelation(nil)

	// Assert
	fmt.Println(term)
	t.Fail()
}
