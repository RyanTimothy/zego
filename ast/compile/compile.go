package compile

import (
	"avidbound.com/zego/ast"
	"avidbound.com/zego/ast/internal/parser"
	"avidbound.com/zego/ast/term"
)

type Compiler struct {
	//  package ex
	//  a := "ex.a"
	//  b := "ex.b"
	//
	//  root
	//    └─── zego (no rules)
	//           └─── ex (no rules)
	//                 ├─── a (1 rule)
	//                 └─── b (1 rule)
	RuleTree *TreeNode

	Errors ast.Errors
}

func NewCompiler() *Compiler {
	return &Compiler{}
}

func (c *Compiler) Compile(modules map[string]*ast.Module) {

}

func (c *Compiler) Failed() bool {
	return false
}

type TreeNode struct {
	Key      term.Value // rule path ie: rego.a.b
	Values   []ast.Rule
	Children map[term.Value]*TreeNode
}

func CompileModules(modules map[string]string) (*Compiler, error) {

	parsed := make(map[string]*ast.Module, len(modules))

	for f, module := range modules {
		var pm *ast.Module
		var err error
		if pm, err = parser.ParseModule(f, module); err != nil {
			return nil, err
		}
		parsed[f] = pm
	}

	compiler := NewCompiler()
	compiler.Compile(parsed)

	if compiler.Failed() {
		return nil, compiler.Errors
	}

	return compiler, nil
}
