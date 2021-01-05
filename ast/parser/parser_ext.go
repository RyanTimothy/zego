package parser

import (
	"fmt"

	"avidbound.com/zego/ast"
	"avidbound.com/zego/ast/term"
)

func ParseQuery(input string) (ast.Body, error) {
	body, errs := NewParser("", input).parseQuery()

	if len(errs) > 0 {
		return nil, errs
	}

	return body, nil
}

// ParseModule returns a parsed Module object.
// For details on Module objects and their fields, see policy.go.
// Empty input will return nil, nil.
func ParseModule(filename, input string) (*ast.Module, error) {
	stmts, err := ParseStatements(filename, input)
	if err != nil {
		return nil, err
	}
	return parseModule(filename, stmts)
}

func ParseStatement(input string) (ast.Statement, error) {
	stmts, err := ParseStatements("", input)
	if err != nil {
		return nil, err
	}
	if len(stmts) != 1 {
		return nil, fmt.Errorf("expected exactly one statement")
	}
	return stmts[0], nil
}

func ParseStatements(name, input string) ([]ast.Statement, error) {
	return NewParser(name, input).parse()
}

func parseModule(filename string, stmts []ast.Statement) (*ast.Module, error) {

	if len(stmts) == 0 {
		return nil, ast.NewError(&term.Location{File: filename}, "empty module")
	}

	var errs ast.Errors

	pkg, ok := stmts[0].(*ast.Package)
	if !ok {
		loc := stmts[0].(ast.Statement).Loc()
		errs = append(errs, ast.NewError(loc, "package expected"))
	}

	mod := &ast.Module{
		Package: pkg,
	}

	for _, stmt := range stmts[1:] {
		switch stmt := stmt.(type) {
		case *ast.Rule:
			stmt.Module = mod
			mod.Rules = append(mod.Rules, stmt)
		case *ast.Package:
			errs = append(errs, ast.NewError(stmt.Loc(), "unexpected package"))
		default:
			panic("illegal value") // Indicates grammar is out-of-sync with code.
		}
	}

	if len(errs) == 0 {
		return mod, nil
	}

	return nil, errs
}
