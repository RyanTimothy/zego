package zego

import (
	"context"
	"fmt"

	"avidbound.com/zego/ast"
	"avidbound.com/zego/ast/compile"
	"avidbound.com/zego/ast/parser"
)

type Zego struct {
	compiler      *compile.Compiler
	query         string
	modules       []rawModule
	parsedModules map[string]*ast.Module
	parsedQuery   ast.Body
}

type Option func(r *Zego)

func New(options ...Option) *Zego {
	z := &Zego{
		compiler:      compile.NewCompiler(),
		parsedModules: map[string]*ast.Module{},
	}

	for _, o := range options {
		o(z)
	}

	return z
}

type rawModule struct {
	filename string
	module   string
}

func (m rawModule) Parse() (*ast.Module, error) {
	return parser.ParseModule(m.filename, m.module)
}

// Module returns an argument that adds a Rego module.
func Module(filename, input string) func(r *Zego) {
	return func(r *Zego) {
		r.modules = append(r.modules, rawModule{
			filename: filename,
			module:   input,
		})
	}
}

// Query returns an argument that sets the Rego query.
func Query(q string) func(r *Zego) {
	return func(r *Zego) {
		r.query = q
	}
}

type PreparedEvalQuery struct {
}

// PrepareForEval will parse inputs, modules, and query arguments in preparation
// of evaluating them.
func (r *Zego) PrepareForEval(ctx context.Context) (PreparedEvalQuery, error) {
	if r.query == "" {
		return PreparedEvalQuery{}, fmt.Errorf("cannot evaluate empty query")
	}

	err := r.prepare(ctx)

	return PreparedEvalQuery{}, err
}

func (r *Zego) prepare(ctx context.Context) error {
	var err error

	err = r.parseModules(ctx)
	if err != nil {
		return err
	}

	err = r.compileModules(ctx)
	if err != nil {
		return err
	}

	r.parsedQuery, err = r.parseQuery()
	if err != nil {
		return err
	}

	err = r.compileAndCacheQuery()
	if err != nil {
		return err
	}

	return nil
}

func (r *Zego) parseModules(ctx context.Context) error {
	if len(r.modules) == 0 {
		return nil
	}

	var errs ast.Errors
	for _, module := range r.modules {
		p, err := module.Parse()
		if err != nil {
			errs = append(errs, err)
		}
		r.parsedModules[module.filename] = p
	}

	return nil
}

func (r *Zego) parseQuery() (ast.Body, error) {
	if r.parsedQuery != nil {
		return r.parsedQuery, nil
	}

	return parser.ParseQuery(r.query)
}

func (r *Zego) compileModules(ctx context.Context) error {

	if len(r.parsedModules) > 0 {
		if r.compiler.Compile(r.parsedModules); r.compiler.Failed() {
			return r.compiler.Errors
		}
	}

	return nil
}

func (r *Zego) compileAndCacheQuery() error {
	_, _, err := r.compileQuery()
	if err != nil {
		return err
	}

	return nil
}

func (r *Zego) compileQuery() (compile.QueryCompiler, ast.Body, error) {
	qc := r.compiler.NewQueryCompiler()

	compiled, err := qc.Compile(r.parsedQuery)

	return qc, compiled, err
}

func (q *PreparedEvalQuery) Eval(ctx context.Context, options ...EvalOption) (ResultSet, error) {
	return ResultSet{}, nil
}

type EvalContext struct {
	rawInput *interface{}
	hasInput bool
}

// EvalOption defines a function to set an option on an EvalConfig
type EvalOption func(*EvalContext)

// EvalInput configures the input for a Prepared Query's evaluation
func EvalInput(input interface{}) EvalOption {
	return func(e *EvalContext) {
		e.rawInput = &input
		e.hasInput = true
	}
}

type ResultSet []interface{}
