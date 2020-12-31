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
	items  []lexer.Item
	errors ast.Errors
	state  *state
}

func Parse(name, input string) []ast.Statement {

	p := parser{
		items: lexer.Lex(name, input),
	}

	p.state = &state{
		parser: &p,
		index:  0,
	}

	return p.parse()
}

// parse runs the state machine for the parser.
func (p *parser) parse() []ast.Statement {
	var statements []ast.Statement

	if p.state.token() == tokens.Whitespace || p.state.token() == tokens.EOL {
		p.state.nextNonSpace()
	}

Loop:
	for {
		tok := p.state.token()
		switch tok {
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
			break
		}
	}

	return statements
}

func (p *parser) parsePackage() *ast.Package {
	tok := p.state.nextNonSpace()
	if tok != tokens.Identifier {
		p.errorf(p.state.loc(), "expected identifier")
	}

	pkg := &ast.Package{}
	pkg.SetLoc(p.state.loc())

	// string term for first ident - parse reference
	ident := p.items[p.state.index].Value
	p.state.next()
	t := p.parseRef(term.StringTerm(ident).SetLoc(p.state.loc()))

	switch v := t.Value.(type) {
	case term.Var:
		pkg.Path = term.Ref{term.StringTerm(string(v)).SetLoc(p.state.loc())}
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
		p.errorf(p.state.loc(), "expected rule head name")
	}

	p.state.nextNonSpace()

	if p.state.token() != tokens.Declare {
		p.errorf(p.state.loc(), "rules must use := operator")
		return nil
	}

	p.state.nextNonSpace()

	rule.Value = p.parseTermRelation(nil) // example: rule := a+b {}

	if p.state.token() == tokens.LBrace {
		p.state.nextNonSpace()
		if rule.Body = p.parseQuery(tokens.RBrace); rule.Body == nil {
			return nil
		}
		p.state.nextNonSpace()
	}

	return rule
}

func (p *parser) parseQuery(end tokens.Token) ast.Body {
	body := ast.Body{}

	if p.state.token() == end {
		p.errorf(p.state.loc(), "found empty body")
		return nil
	}

	for {
		expr := p.parseExpr()
		if expr == nil {
			return nil
		}

		body.Append(expr)

		if p.state.token() == end {
			return body
		}
	}
}

func (p *parser) parseExpr() *ast.Expr {
	lhs := p.parseTermRelation(nil)
	if lhs == nil {
		return nil
	}

	tok := p.state.token()
	if tok == tokens.Declare {
		p.state.nextNonSpace()
		if rhs := p.parseTermRelation(nil); rhs != nil {
			op := term.OpTerm(tok.String()).SetLoc(lhs.Location)
			return ast.NewExpr([]*term.Term{op, lhs, rhs})
		}
		return nil
	}

	if call, ok := lhs.Value.(term.Call); ok {
		return ast.NewExpr([]*term.Term(call))
	}

	return nil
}

func (p *parser) parseTerm() *term.Term {
	switch p.state.token() {
	case tokens.True:
		term := term.BooleanTerm(true).SetLoc(p.state.loc())
		p.state.nextNonSpace()
		return term
	case tokens.False:
		term := term.BooleanTerm(false).SetLoc(p.state.loc())
		p.state.nextNonSpace()
		return term
	case tokens.Identifier:
		term := p.parseVar()
		// check if next is ident.field or ident.field[_] or ident.call(_)
		if tok := p.state.next(); tok == tokens.Field || tok == tokens.LParenthesis || tok == tokens.LBracket {
			return p.parseRef(term)
		}
		p.state.nextNonSpace()
		return term
	case tokens.Number:
		return p.parseNumber()
	case tokens.LParenthesis:
		p.state.nextNonSpace()
		if term := p.parseTermRelation(nil); term != nil {
			if p.state.token() != tokens.RParenthesis {
				p.errorf(p.state.loc(), "non-terminated expression")
				return nil
			}
			if tok := p.state.next(); tok == tokens.Field || tok == tokens.LParenthesis || tok == tokens.LBracket {
				return p.parseRef(term)
			}
			return term
		} else {
			return nil
		}
	default:
		tok := p.state.token()
		tokType := "token"
		if tok >= tokens.Package && tok <= tokens.False {
			tokType = "keyword"
		}
		p.errorf(p.state.loc(), "unexpected %s %s", tok.String(), tokType)
		return nil
	}
}

func (p *parser) parseNumber() *term.Term {
	loc := p.state.loc()

	// Ensure that the number is valid
	s := p.items[p.state.index].Value
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
	p.state.nextNonSpace()
	return r
}

func (p *parser) parseVar() *term.Term {
	v := p.items[p.state.index].Value // TODO: generate wildcard '_'
	return term.VarTerm(v)
}

func (p *parser) parseRef(head *term.Term) *term.Term {
	ref := []*term.Term{head}

	loc := p.state.loc()

	for {
		switch p.state.token() {
		case tokens.Field:
			field := p.items[p.state.index].Value[1:] // .field <-- remove dot from lexer value
			ref = append(ref, term.StringTerm(field).SetLoc(p.state.loc()))
			p.state.next()
		case tokens.LParenthesis:
			term := p.parseCall(term.RefTerm(ref...).SetLoc(loc))
			if term != nil {
				if tok := p.state.token(); tok == tokens.Field || tok == tokens.LBracket {
					p.parseRef(term) // with 'method(x).something' OR 'method(x)[_]'
				}
			}
			break
		case tokens.LBracket:
			p.state.nextNonSpace()
			if term := p.parseTermRelation(nil); term != nil {
				if p.state.token() != tokens.RBracket {
					p.errorf(p.state.loc(), "expected %v", tokens.RBracket)
					return nil
				}
				ref = append(ref, term)
				p.state.next()
			} else {
				return nil
			}
			break
		default:
			p.state.nextNonSpace()
			return term.RefTerm(ref...)
		}
	}
}

func (p *parser) parseCall(operator *term.Term) *term.Term {
	p.state.nextNonSpace()

	if p.state.token() == tokens.RParenthesis {
		return term.CallTerm(operator)
	}

	if r := p.parseTermList(tokens.RParenthesis, []*term.Term{operator}); r != nil {
		p.state.next()
		return term.CallTerm(r...)
	}

	return nil
}

func (p *parser) parseTermList(end tokens.Token, r []*term.Term) []*term.Term {
	for {
		term := p.parseTermRelation(nil)
		if term != nil {
			r = append(r, term)
			switch p.state.token() {
			case end:
				return r
			case tokens.Comma:
				p.state.nextNonSpace()
				if p.state.token() == end {
					return r
				}
				continue
			default:
				p.errorf(p.state.loc(), "expected %q or %q", tokens.Comma, end)
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
		tok := p.state.token()
		if tok == tokens.Whitespace {
			tok = p.state.nextNonSpace()
		}
		loc := p.state.loc()

		if tok == tokens.Multiply || tok == tokens.Divide || tok == tokens.Modulus {
			p.state.nextNonSpace()
			if rhs := p.parseTermRelation(nil); rhs != nil {
				op := term.OpTerm(tok.String()).SetLoc(loc)
				call := term.CallTerm(op, lhs, rhs).SetLoc(lhs.Location)

				tok = p.state.token()
				if tok == tokens.Multiply || tok == tokens.Divide || tok == tokens.Modulus {
					return p.parseTermRelation(call)
				}
				return call
			}
		} else if tok == tokens.Add || tok == tokens.Subtract {
			p.state.nextNonSpace()
			if rhs := p.parseTermRelation(nil); rhs != nil {
				op := term.OpTerm(tok.String()).SetLoc(loc)
				call := term.CallTerm(op, lhs, rhs).SetLoc(lhs.Location)

				tok = p.state.token()
				if tok == tokens.Add || tok == tokens.Subtract {
					return p.parseTermRelation(call)
				}
				return call
			}
		} else if tok == tokens.And {
			p.state.nextNonSpace()
			if rhs := p.parseTermRelation(nil); rhs != nil {
				op := term.OpTerm(tok.String()).SetLoc(loc)
				call := term.CallTerm(op, lhs, rhs).SetLoc(lhs.Location)

				tok = p.state.token()
				if tok == tokens.And {
					return p.parseTermRelation(call)
				}
				return call
			}
		} else if tok == tokens.Or {
			p.state.nextNonSpace()
			if rhs := p.parseTermRelation(nil); rhs != nil {
				op := term.OpTerm(tok.String()).SetLoc(loc)
				call := term.CallTerm(op, lhs, rhs).SetLoc(lhs.Location)

				tok = p.state.token()
				if tok == tokens.Or {
					return p.parseTermRelation(call)
				}
				return call
			}
		} else if tok == tokens.Equal || tok == tokens.NEqual || tok == tokens.LT || tok == tokens.GT || tok == tokens.LTE || tok == tokens.GTE {
			p.state.nextNonSpace()
			if rhs := p.parseTermRelation(nil); rhs != nil {
				op := term.OpTerm(tok.String()).SetLoc(loc)
				call := term.CallTerm(op, lhs, rhs).SetLoc(lhs.Location)

				tok = p.state.token()
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

func (s *state) nextNonSpace() tokens.Token {
Loop:
	for {
		switch s.next() {
		case tokens.Whitespace, tokens.EOL:
			continue Loop
		default:
			break Loop
		}
	}
	return s.token()
}

func (s *state) next() tokens.Token {
	s.index++
	return s.token()
}

func (s *state) token() tokens.Token {
	if s.index < len(s.parser.items) {
		return s.parser.items[s.index].Token
	}
	return tokens.EOF
}

func (s *state) loc() *term.Location {
	return &term.Location{} // TODO : get location
}
