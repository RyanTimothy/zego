package parser

import (
	"encoding/json"
	"fmt"
	"math/big"

	"avidbound.com/zego/ast"
	"avidbound.com/zego/ast/internal/lexer"
	"avidbound.com/zego/ast/internal/tokens"
	"avidbound.com/zego/ast/term"
)

type state struct {
	parser *parser
	index  int
}

type parser struct {
	file   string
	items  []lexer.Item
	errors ast.Errors
	index  int
}

// parse runs the state machine for the parser.
func (p *parser) parse() ([]ast.Statement, error) {
	var statements []ast.Statement

	if p.token() == tokens.Whitespace || p.token() == tokens.EOL {
		p.nextNonSpace()
	}

Loop:
	for {
		tok := p.token()
		switch tok {
		case tokens.Illegal:
			p.errorf(p.loc(), p.items[p.index].Value)
		case tokens.Package:
			if pkg := p.parsePackage(); pkg != nil {
				statements = append(statements, pkg)
			}
		case tokens.Identifier:
			if rule := p.parseRule(); rule != nil {
				statements = append(statements, rule)
			}
		case tokens.EOF:
			break Loop
		}

		if len(p.errors) > 0 {
			return nil, p.errors
		}
	}

	return statements, nil
}

func (p *parser) parsePackage() *ast.Package {
	loc := p.loc()
	tok := p.nextNonSpace()
	if tok != tokens.Identifier {
		p.errorf(p.loc(), "expected identifier")
	}

	pkg := &ast.Package{}
	pkg.SetLoc(loc)

	// string term for first ident - parse reference
	ident := p.items[p.index].Value
	loc = p.loc()
	p.next()
	t := p.parseRef(term.StringTerm(ident).SetLoc(loc))

	switch v := t.Value.(type) {
	case term.Var:
		pkg.Path = term.Ref{term.StringTerm(string(v)).SetLoc(loc)}
		break
	case term.Ref:
		pkg.Path = v
		break
	}

	return pkg
}

func (p *parser) parseRule() *ast.Rule {
	rule := &ast.Rule{}

	if name := p.parseVar(); name != nil {
		if v, ok := name.Value.(term.Var); ok {
			rule.Name = v
		}
	}
	if rule.Name == "" {
		p.errorf(p.loc(), "expected rule head name")
	}

	p.nextNonSpace()

	if p.token() != tokens.Declare {
		p.errorf(p.loc(), "rules must use := operator")
		return nil
	}

	p.nextNonSpace()

	rule.Value = p.parseTermRelation(nil) // example: rule := a+b {}

	if p.token() == tokens.LBrace {
		p.nextNonSpace()
		if rule.Body = p.parseQuery(tokens.RBrace); rule.Body == nil {
			return nil
		}
		p.nextNonSpace()
	}

	return rule
}

func (p *parser) parseQuery(end tokens.Token) ast.Body {
	body := ast.Body{}

	if p.token() == end {
		p.errorf(p.loc(), "found empty body")
		return nil
	}

	for {
		expr := p.parseExpr()
		if expr == nil {
			return nil
		}

		body.Append(expr)

		if p.token() == end {
			return body
		}
	}
}

func (p *parser) parseExpr() *ast.Expr {
	lhs := p.parseTermRelation(nil)
	if lhs == nil {
		return nil
	}

	tok := p.token()
	if tok == tokens.Declare {
		loc := p.loc()
		p.nextNonSpace()
		if rhs := p.parseTermRelation(nil); rhs != nil {
			op := term.OpTerm(tok.String()).SetLoc(loc)
			return ast.NewExpr([]*term.Term{op, lhs, rhs})
		}
		return nil
	}

	// if call, ok := lhs.Value.(term.Call); ok {
	// 	return ast.NewExpr([]*term.Term(call))
	// }

	return ast.NewExpr(lhs)
}

func (p *parser) parseTerm() *term.Term {
	switch p.token() {
	case tokens.True:
		term := term.BooleanTerm(true).SetLoc(p.loc())
		p.nextNonSpace()
		return term
	case tokens.False:
		term := term.BooleanTerm(false).SetLoc(p.loc())
		p.nextNonSpace()
		return term
	case tokens.String:
		term := p.parseString()
		p.nextNonSpace()
		return term
	case tokens.Identifier:
		term := p.parseVar()
		// check if next is ident.field or ident.field[_] or ident.call(_)
		if tok := p.next(); tok == tokens.Field || tok == tokens.LParenthesis || tok == tokens.LBracket {
			return p.parseRef(term)
		}
		return term
	case tokens.Number:
		return p.parseNumber()
	case tokens.LParenthesis:
		p.nextNonSpace()
		if term := p.parseTermRelation(nil); term != nil {
			if p.token() != tokens.RParenthesis {
				p.errorf(p.loc(), "non-terminated expression")
				return nil
			}
			if tok := p.next(); tok == tokens.Field || tok == tokens.LParenthesis || tok == tokens.LBracket {
				return p.parseRef(term)
			}
			return term
		} else {
			return nil
		}
	default:
		tok := p.token()
		tokType := "token"
		if tok >= tokens.Package && tok <= tokens.False {
			tokType = "keyword"
		}
		p.errorf(p.loc(), "unexpected %s %s", tok.String(), tokType)
		return nil
	}
}

func (p *parser) parseString() *term.Term {
	var s string
	err := json.Unmarshal([]byte(p.items[p.index].Value), &s)
	if err != nil {
		p.errorf(p.loc(), "illegal string literal: %s", p.items[p.index].Value)
		return nil
	}
	return term.StringTerm(s).SetLoc(p.loc())
}

func (p *parser) parseNumber() *term.Term {
	loc := p.loc()

	// Ensure that the number is valid
	s := p.items[p.index].Value
	f, ok := new(big.Float).SetString(s)
	if !ok {
		p.errorf(loc, "expected number")
		return nil
	}

	// Put limit on size of exponent to prevent non-linear cost of String()
	// function on big.Float from causing denial of service: https://github.com/golang/go/issues/11068
	//
	// n == sign * mantissa * 2^exp
	// 0.5 <= mantissa < 1.0
	//
	// The limit is arbitrary.
	exp := f.MantExp(nil)
	if exp > 1e5 || exp < -1e5 {
		p.errorf(loc, "number too big")
		return nil
	}

	// Note: Use the original string, do *not* round trip from
	// the big.Float as it can cause precision loss.
	r := term.NumberTerm(json.Number(s)).SetLoc(loc)
	p.nextNonSpace()
	return r
}

func (p *parser) parseVar() *term.Term {
	v := p.items[p.index].Value // TODO: generate wildcard '_'
	return term.VarTerm(v)
}

func (p *parser) parseRef(head *term.Term) *term.Term {
	ref := []*term.Term{head}

	loc := p.loc()

	for {
		switch p.token() {
		case tokens.Field:
			field := p.items[p.index].Value[1:] // .field <-- remove dot from lexer value
			ref = append(ref, term.StringTerm(field).SetLoc(p.loc()))
			p.next()
		case tokens.LParenthesis:
			term := p.parseCall(term.RefTerm(ref...).SetLoc(loc))
			if term != nil {
				if tok := p.token(); tok == tokens.Field || tok == tokens.LBracket {
					p.parseRef(term) // with 'method(x).something' OR 'method(x)[_]'
				}
				p.next()
			}
			break
		case tokens.LBracket:
			p.nextNonSpace()
			if term := p.parseTermRelation(nil); term != nil {
				if p.token() != tokens.RBracket {
					p.errorf(p.loc(), "expected %v", tokens.RBracket)
					return nil
				}
				ref = append(ref, term)
				p.next()
			} else {
				return nil
			}
			break
		default:
			if p.token() == tokens.Whitespace || p.token() == tokens.EOL {
				p.nextNonSpace()
			}
			return term.RefTerm(ref...)
		}
	}
}

func (p *parser) parseCall(operator *term.Term) *term.Term {
	p.nextNonSpace()

	if p.token() == tokens.RParenthesis {
		return term.CallTerm(operator)
	}

	if r := p.parseTermList(tokens.RParenthesis, []*term.Term{operator}); r != nil {
		p.next()
		return term.CallTerm(r...)
	}

	return nil
}

func (p *parser) parseTermList(end tokens.Token, r []*term.Term) []*term.Term {
	for {
		term := p.parseTermRelation(nil)
		if term != nil {
			r = append(r, term)
			switch p.token() {
			case end:
				return r
			case tokens.Comma:
				p.nextNonSpace()
				if p.token() == end {
					return r
				}
				continue
			default:
				p.errorf(p.loc(), "expected %q or %q", tokens.Comma, end)
			}
		}
		return nil
	}
}

func (p *parser) parseTermRelation(lhs *term.Term) *term.Term {
	if lhs == nil {
		lhs = p.parseTerm()
	}

	if lhs != nil {
		tok := p.token()
		if tok == tokens.Whitespace {
			tok = p.nextNonSpace()
		}
		loc := p.loc()

		if tok == tokens.Multiply || tok == tokens.Divide || tok == tokens.Modulus {
			p.nextNonSpace()
			if rhs := p.parseTermRelation(nil); rhs != nil {
				op := term.OpTerm(tok.String()).SetLoc(loc)
				call := term.CallTerm(op, lhs, rhs).SetLoc(lhs.Location)

				tok = p.token()
				if tok == tokens.Multiply || tok == tokens.Divide || tok == tokens.Modulus {
					return p.parseTermRelation(call)
				}
				return call
			}
		} else if tok == tokens.Add || tok == tokens.Subtract {
			p.nextNonSpace()
			if rhs := p.parseTermRelation(nil); rhs != nil {
				op := term.OpTerm(tok.String()).SetLoc(loc)
				call := term.CallTerm(op, lhs, rhs).SetLoc(lhs.Location)

				tok = p.token()
				if tok == tokens.Add || tok == tokens.Subtract {
					return p.parseTermRelation(call)
				}
				return call
			}
		} else if tok == tokens.And {
			p.nextNonSpace()
			if rhs := p.parseTermRelation(nil); rhs != nil {
				op := term.OpTerm(tok.String()).SetLoc(loc)
				call := term.CallTerm(op, lhs, rhs).SetLoc(lhs.Location)

				tok = p.token()
				if tok == tokens.And {
					return p.parseTermRelation(call)
				}
				return call
			}
		} else if tok == tokens.Or {
			p.nextNonSpace()
			if rhs := p.parseTermRelation(nil); rhs != nil {
				op := term.OpTerm(tok.String()).SetLoc(loc)
				call := term.CallTerm(op, lhs, rhs).SetLoc(lhs.Location)

				tok = p.token()
				if tok == tokens.Or {
					return p.parseTermRelation(call)
				}
				return call
			}
		} else if tok == tokens.Equal || tok == tokens.NEqual || tok == tokens.LT || tok == tokens.GT || tok == tokens.LTE || tok == tokens.GTE {
			p.nextNonSpace()
			if rhs := p.parseTermRelation(nil); rhs != nil {
				op := term.OpTerm(tok.String()).SetLoc(loc)
				call := term.CallTerm(op, lhs, rhs).SetLoc(lhs.Location)

				tok = p.token()
				if tok == tokens.Equal || tok == tokens.NEqual || tok == tokens.LT || tok == tokens.GT || tok == tokens.LTE || tok == tokens.GTE {
					return p.parseTermRelation(call)
				}
				return call
			}
		}
	}

	return lhs
}

func (p *parser) errorf(l *term.Location, f string, a ...interface{}) {
	p.errors = append(p.errors, &ast.Error{
		Message:  fmt.Sprintf(f, a...),
		Location: l,
	})
}

func (p *parser) nextNonSpace() tokens.Token {
Loop:
	for {
		switch p.next() {
		case tokens.Whitespace, tokens.EOL:
			continue Loop
		default:
			break Loop
		}
	}
	return p.token()
}

func (p *parser) next() tokens.Token {
	p.index++
	return p.token()
}

func (p *parser) token() tokens.Token {
	if p.index < len(p.items) {
		return p.items[p.index].Token
	}
	return tokens.EOF
}

func (p *parser) loc() *term.Location {
	if p.index < len(p.items) {
		return &term.Location{
			File:   p.file,
			Line:   p.items[p.index].Pos.Line,
			Column: p.items[p.index].Pos.Column,
		}
	}
	return &term.Location{
		File: p.file,
	}
}
