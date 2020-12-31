package ast

import (
	"fmt"
	"strings"

	"avidbound.com/zego/ast/term"
)

type (
	// Node represents a node in an AST. Nodes may be statements in a policy module
	// or elements of an ad-hoc query, expression, etc.
	Node interface {
		fmt.Stringer
		Loc() *term.Location
		SetLoc(*term.Location)
	}

	// Statement represents a single statement in a policy module.
	Statement interface {
		Node
	}
)

type (
	// Args represents zero or more arguments to a rule.
	Args []*term.Term

	// Body represents one or more expressions contained inside a rule or user
	// function.
	Body []*Expr

	// Expr represents a single expression contained inside the body of a rule.
	Expr struct {
		Location  *term.Location `json:"-"`
		Generated bool           `json:"generated,omitempty"`
		Index     int            `json:"index"`
		Terms     interface{}    `json:"terms"`
	}

	// Module represents a collection of policies (defined by rules)
	// within a namespace (defined by the package) and optional
	// dependencies on external documents (defined by imports).
	Module struct {
		Package *Package `json:"package"`
		Rules   []*Rule  `json:"rules,omitempty"`
	}

	// Package represents the namespace of the documents produced
	// by rules inside the module.
	Package struct {
		Location *term.Location `json:"-"`
		Path     term.Ref       `json:"path"`
	}

	// Rule represents a rule as defined in the language. Rules define the
	// content of documents that represent policy decisions.
	Rule struct {
		Location *term.Location `json:"-"`
		Name     term.Var       `json:"name,omitempty"`
		Value    *term.Term     `json:"value,omitempty"`
		Body     Body           `json:"body"`

		// Module is a pointer to the module containing this rule. If the rule
		// was NOT created while parsing/constructing a module, this should be
		// left unset. The pointer is not included in any standard operations
		// on the rule (e.g., printing, comparison, visiting, etc.)
		Module *Module `json:"-"`
	}
)

// NewExpr returns a new Expr object.
func NewExpr(terms interface{}) *Expr {
	return &Expr{
		Terms: terms,
		Index: 0,
	}
}

func (expr *Expr) String() string {
	switch t := expr.Terms.(type) {
	case []*term.Term:
		return term.Call(t).String()
	case *term.Term:
		return t.String()
	}
	return ""
}

func (p *Package) Loc() *term.Location {
	return p.Location
}

func (p *Package) SetLoc(l *term.Location) {
	p.Location = l
}

func (p *Package) String() string {
	path := p.Path.String()
	return fmt.Sprintf("package %v", path)
}

func (r *Rule) Loc() *term.Location {
	return r.Location
}

func (r *Rule) SetLoc(l *term.Location) {
	r.Location = l
}

func (r *Rule) String() string {
	if r.Value == nil {
		return r.Name.String() + " {\n" + r.Body.String() + "\n}\n"
	}
	return r.Name.String() + " := " + r.Value.String() + " {\n" + r.Body.String() + "\n}\n"
}

// Append adds the expr to the body and updates the expr's index accordingly.
func (body *Body) Append(expr *Expr) {
	n := len(*body)
	expr.Index = n
	*body = append(*body, expr)
}

func (body Body) String() string {
	var buf []string
	for _, v := range body {
		buf = append(buf, v.String())
	}
	return " " + strings.Join(buf, "\n ")
}
