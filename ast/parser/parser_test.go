package parser

import (
	"testing"

	"avidbound.com/zego/ast"
	"avidbound.com/zego/ast/internal/lexer"
	"avidbound.com/zego/ast/term"
)

func TestParseTermRelation(t *testing.T) {
	assertParseTermRelation(t, "relation nested brace", `a < input[input["a"]].b`,
		term.CallTerm(term.OpTerm("lt"),
			term.VarTerm("a"),
			term.RefTerm(term.VarTerm("input"),
				term.RefTerm(term.VarTerm("input"), term.StringTerm("a")), term.StringTerm("b"))))

	assertParseTermRelation(t, "relation nested parenthises", `a == (b + (c - d)) * e`,
		term.CallTerm(term.OpTerm("equal"),
			term.VarTerm("a"),
			term.CallTerm(term.OpTerm("multiply"),
				term.CallTerm(term.OpTerm("add"),
					term.VarTerm("b"),
					term.CallTerm(term.OpTerm("minus"),
						term.VarTerm("c"),
						term.VarTerm("d"))),
				term.VarTerm("e"))))
}

func TestPackage(t *testing.T) {
	assertParsePackage(t, "single", `package foo`, modulePackage(1, 1, refTerm(1, 9, stringTerm(1, 9, "foo"))))
	assertParsePackage(t, "multiple", `package foo.bar`, modulePackage(1, 1, refTerm(1, 9, stringTerm(1, 9, "foo"), stringTerm(1, 12, "bar"))))
	assertParsePackage(t, "space", `package foo["bar bizz"]`, modulePackage(1, 1, refTerm(1, 9, stringTerm(1, 9, "foo"), stringTerm(1, 13, "bar bizz"))))
}

func TestRule(t *testing.T) {
	assertParseRule(t, "constant",
		`test := "abc" {
			b := true
		}`,
		&ast.Rule{
			Name:  term.Var("test"),
			Value: term.StringTerm("abc"),
			Body: ast.NewBody(
				ast.NewExpr(term.CallTerm(term.OpTerm("declare"), term.VarTerm("b"), term.BooleanTerm(true))),
			),
		})

	assertParseRule(t, "constant",
		`test := false`,
		&ast.Rule{
			Name:  term.Var("test"),
			Value: term.BooleanTerm(false),
		})

	assertParseRule(t, "dynamic",
		`test := input["a"] {
			input.b[ 1 ] == 12.34
		}`,
		&ast.Rule{
			Name:  term.Var("test"),
			Value: term.RefTerm(term.VarTerm("input"), term.StringTerm("a")),
			Body: ast.NewBody(
				ast.NewExpr(term.CallTerm(
					term.OpTerm("equal"),
					term.RefTerm(term.VarTerm("input"), term.StringTerm("b"), term.NumberTerm("1")),
					term.NumberTerm("12.34"))),
			),
		})

	assertParseRule(t, "call",
		`test := "abc" {
			b := a.b( 1 )
		}`,
		&ast.Rule{
			Name:  term.Var("test"),
			Value: term.StringTerm("abc"),
			Body: ast.NewBody(
				ast.NewExpr(term.CallTerm(term.OpTerm("declare"), term.VarTerm("b"), term.CallTerm(term.RefTerm(term.VarTerm("a"), term.StringTerm("b")), term.NumberTerm("1")))),
			),
		})
}

func assertParseTermRelation(t *testing.T, msg, input string, expected *term.Term) {
	t.Helper()

	a := (&parser{items: lexer.Lex("", input)}).parseTermRelation(nil)

	if !a.Value.Equal(expected.Value) {
		t.Errorf("Error on test \"%s\": relation not equal: %v (parsed), %v (expected)", msg, a, expected)
	}
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

func assertParseRule(t *testing.T, msg string, input string, expected *ast.Rule) {
	t.Helper()

	assertParseOne(t, msg, input, func(parsed interface{}) {
		rul := parsed.(*ast.Rule)
		if !rul.Equal(expected) {
			t.Errorf("Error on test \"%s\": packages not equal: %v (parsed), %v (expected)", msg, rul, expected)
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

func assertTermLocation(t *testing.T, msg string, trm *term.Term, line, column int) *term.Term {
	t.Helper()

	if trm.Location.Line != line {
		t.Errorf("Error on test %s: expected line: %d got: %v", msg, line, trm.Location.Line)
	}
	if trm.Location.Column != column {
		t.Errorf("Error on test %s: expected column: %d got: %v", msg, column, trm.Location.Column)
	}
	return trm
}

func modulePackage(line, column int, path *term.Term) *ast.Package {
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

func refTerm(line, column int, t ...*term.Term) *term.Term {
	return term.RefTerm(t...).SetLoc(&term.Location{Line: line, Column: column})
}

func stringTerm(line, column int, value string) *term.Term {
	return term.StringTerm(value).SetLoc(&term.Location{Line: line, Column: column})
}
