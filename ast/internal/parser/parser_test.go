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

func TestModuleParse(t *testing.T) {
	// Arrange
	input := `
	package example

	abEquals := input.bob {
		a := input.test[1+2]
		b := 13.5
		a == b
	}
	`

	// Act
	statements := Parse("test.zego", input)

	// Assert
	for _, s := range statements {
		fmt.Println(s)
	}
	t.Fail()
}
