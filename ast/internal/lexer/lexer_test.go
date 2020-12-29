package lexer

import (
	"testing"

	"avidbound.com/zego/ast/internal/tokens"
)

func TestLexModule(t *testing.T) {
	// Arrange
	input := `
	package test

	abEquals {
		a := input.test[1]
		b := 13.5
		a == b
	}
	`

	tokens := []tokens.Token{
		tokens.EOL,
		tokens.Package, tokens.Whitespace, tokens.Identifier, tokens.EOL, // package test
		tokens.Identifier, tokens.Whitespace, tokens.LBrace, tokens.EOL, // abEquals {
		tokens.Identifier, tokens.Whitespace, tokens.Declare, tokens.Whitespace, tokens.Identifier, tokens.Field, tokens.LBracket, tokens.Number, tokens.RBracket, tokens.EOL, //   a := input.test[1]
		tokens.Identifier, tokens.Whitespace, tokens.Declare, tokens.Whitespace, tokens.Number, tokens.EOL, //   b := 13.5
		tokens.Identifier, tokens.Whitespace, tokens.Equal, tokens.Whitespace, tokens.Identifier, tokens.EOL, //   a == b
		tokens.RBrace, tokens.EOL, tokens.EOF, // }
	}

	// Act
	items := Lex("test", input)

	// Assert
	if len(tokens) != len(items) {
		t.Errorf("want tokens %d but got %d", len(tokens), len(items))
	}

	for i, item := range items {
		if i < len(tokens) && tokens[i] != item.Token {
			t.Errorf("token %d: want token %s but got %s", i, tokens[i], item.Token)
		}
	}
}

func TestLexExpression(t *testing.T) {
	// Arrange
	input := `x == a+b*c*d`

	tokens := []tokens.Token{
		tokens.Identifier,
		tokens.Whitespace,
		tokens.Equal,
		tokens.Whitespace,
		tokens.Identifier,
		tokens.Add,
		tokens.Identifier,
		tokens.Multiply,
		tokens.Identifier,
		tokens.Multiply,
		tokens.Identifier,
		tokens.EOF,
	}

	// Act
	items := Lex("test", input)

	// Assert
	if len(tokens) != len(items) {
		t.Errorf("want tokens %d but got %d", len(tokens), len(items))
	}

	for i, item := range items {
		if i < len(tokens) && tokens[i] != item.Token {
			t.Errorf("token %d: want token %s but got %s", i, tokens[i], item.Token)
		}
	}
}
