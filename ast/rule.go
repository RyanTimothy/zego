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

func (e *Expr) SetLoc(l *term.Location) {
	e.Location = l
}

// Compare returns an integer indicating whether expr is less than, equal to,
// or greater than other.
//
// Expressions are compared as follows:
//
// 1. Declarations are always less than other expressions.
// 2. Preceding expression (by Index) is always less than the other expression.
// 3. Non-negated expressions are always less than than negated expressions.
// 4. Single term expressions are always less than built-in expressions.
//
// Otherwise, the expression terms are compared normally. If both expressions
// have the same terms, the modifiers are compared.
func (e *Expr) Compare(other *Expr) int {

	if e == nil {
		if other == nil {
			return 0
		}
		return -1
	} else if other == nil {
		return 1
	}

	o1 := e.sortOrder()
	o2 := other.sortOrder()
	if o1 < o2 {
		return -1
	} else if o2 < o1 {
		return 1
	}

	switch {
	case e.Index < other.Index:
		return -1
	case e.Index > other.Index:
		return 1
	}

	switch t := e.Terms.(type) {
	case *term.Term:
		if cmp := t.Value.Compare(other.Terms.(*term.Term).Value); cmp != 0 {
			return cmp
		}
	case []*term.Term:
		if cmp := term.TermSliceCompare(t, other.Terms.([]*term.Term)); cmp != 0 {
			return cmp
		}
	}
	return 0
}

func (e *Expr) sortOrder() int {
	switch e.Terms.(type) {
	case *term.Term:
		return 0
	case []*term.Term:
		return 1
	}
	return -1
}

func (e *Expr) String() string {
	switch t := e.Terms.(type) {
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

// Equal returns true if pkg is equal to other.
func (p *Package) Equal(other *Package) bool {
	return p.Compare(other) == 0
}

// Compare returns an integer indicating whether pkg is less than, equal to,
// or greater than other.
func (pkg *Package) Compare(other *Package) int {
	return pkg.Path.Compare(other.Path)
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

// Equal returns true if rule is equal to other.
func (rule *Rule) Equal(other *Rule) bool {
	return rule.Compare(other) == 0
}

// Compare returns an integer indicating whether rule is less than, equal to,
// or greater than other.
func (rule *Rule) Compare(other *Rule) int {
	if rule == nil {
		if other == nil {
			return 0
		}
		return -1
	} else if other == nil {
		return 1
	}
	if cmp := rule.Name.Compare(other.Name); cmp != 0 {
		return cmp
	}
	if cmp := rule.Value.Compare(other.Value); cmp != 0 {
		return cmp
	}
	return rule.Body.Compare(other.Body)
}

func (r *Rule) String() string {
	if r.Value == nil {
		return r.Name.String() + " {\n" + r.Body.String() + "\n}\n"
	}
	return r.Name.String() + " := " + r.Value.String() + " {\n" + r.Body.String() + "\n}\n"
}

// NewBody returns a new Body containing the given expressions. The indices of
// the immediate expressions will be reset.
func NewBody(exprs ...*Expr) Body {
	for i, expr := range exprs {
		expr.Index = i
	}
	return Body(exprs)
}

// Loc returns the location of the Body in the definition.
func (body Body) Loc() *term.Location {
	if len(body) == 0 {
		return nil
	}
	return body[0].Location
}

// SetLoc sets the location on body.
func (body Body) SetLoc(loc *term.Location) {
	if len(body) != 0 {
		body[0].SetLoc(loc)
	}
}

func (body Body) Compare(other Body) int {
	minLen := len(body)
	if len(other) < minLen {
		minLen = len(other)
	}
	for i := 0; i < minLen; i++ {
		if cmp := body[i].Compare(other[i]); cmp != 0 {
			return cmp
		}
	}
	if len(body) < len(other) {
		return -1
	}
	if len(other) < len(body) {
		return 1
	}
	return 0
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
