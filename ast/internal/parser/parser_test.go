package parser

import (
	"fmt"
	"testing"

	"avidbound.com/zego/ast"
	"avidbound.com/zego/ast/internal/lexer"
	"avidbound.com/zego/ast/term"
)

func TestParseTermRelation(t *testing.T) {
	// Arrange
	input := `input.c == (input.a[input.b[0]] + c) * d`
	p := parser{
		items: lexer.Lex("test.zego", input),
	}

	// Act
	term := p.parseTermRelation(nil)

	// Assert
	fmt.Println(term)
	//t.Fail() // TODO : build assert methods for Term call
}

func TestPackageParseStatements(t *testing.T) {
	assertParsePackage(t, "single", `package foo`, Package(1, 1, RefTerm(1, 9, StringTerm(1, 9, "foo"))))
	assertParsePackage(t, "single", `package foo.bar`, Package(1, 1, RefTerm(1, 9, StringTerm(1, 9, "foo"), StringTerm(1, 12, "bar"))))
}

func Package(line, column int, path *term.Term) *ast.Package {
	pkg := &ast.Package{
		Location: &term.Location{Line: line, Column: column},
	}
	switch v := path.Value.(type) {
	case term.Var:
		pkg.Path = term.Ref{term.StringTerm(string(v)).SetLoc(path.Location)}
	case term.Ref:
		pkg.Path = v
	}
	return pkg
}

func RefTerm(line, column int, t ...*term.Term) *term.Term {
	return term.RefTerm(t...).SetLoc(&term.Location{Line: line, Column: column})
}

func StringTerm(line, column int, value string) *term.Term {
	return term.StringTerm(value).SetLoc(&term.Location{Line: line, Column: column})
}

func assertParsePackage(t *testing.T, msg string, input string, expected *ast.Package) {
	t.Helper()

	assertParseOne(t, msg, input, func(parsed interface{}) {
		pkg := parsed.(*ast.Package)
		if !pkg.Equal(expected) {
			t.Errorf("Error on test \"%s\": packages not equal: %v (parsed), %v (expected)", msg, pkg, expected)
		}
		// assert location
		if expected.Location != nil {
			if pkg.Location.Line != expected.Location.Line {
				t.Errorf("Error on test %s: expected line: %d got: %v", msg, expected.Location.Line, pkg.Location.Line)
			}
			if pkg.Location.Column != expected.Location.Column {
				t.Errorf("Error on test %s: expected column: %d got: %v", msg, expected.Location.Column, pkg.Location.Column)
			}
			for i, p := range pkg.Path {
				if i < len(expected.Path) && expected.Path[i].Location != nil {
					if p.Location.Line != expected.Path[i].Location.Line {
						t.Errorf("Error on test %s: expected line: %d got: %v", msg, expected.Path[i].Location.Line, p.Location.Line)
					}
					if p.Location.Column != expected.Path[i].Location.Column {
						t.Errorf("Error on test %s: expected column: %d got: %v", msg, expected.Path[i].Location.Column, p.Location.Column)
					}
				}
			}
		}
	})
}

func assertParseOne(t *testing.T, msg string, input string, correct func(interface{})) {
	t.Helper()

	p, err := ParseStatement(input)
	if err != nil {
		t.Errorf("Error on test \"%s\": parse error on %s: %s", msg, input, err)
		return
	}
	correct(p)
}

func AssertTermLocation(t *testing.T, msg string, trm *term.Term, line, column int) *term.Term {
	t.Helper()

	if trm.Location.Line != line {
		t.Errorf("Error on test %s: expected line: %d got: %v", msg, line, trm.Location.Line)
	}
	if trm.Location.Column != column {
		t.Errorf("Error on test %s: expected column: %d got: %v", msg, column, trm.Location.Column)
	}
	return trm
}
