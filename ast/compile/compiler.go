package compile

import (
	"sort"

	"avidbound.com/zego/ast"
	"avidbound.com/zego/ast/term"
	"avidbound.com/zego/util"
)

type Compiler struct {

	// ModuleTree organizes the modules into a tree where each node is keyed by
	// an element in the module's package path. E.g., "a", "a.b", "a.c", "a.b"
	//
	//  root
	//    └─── zego (no modules)
	//           └─── a (1 module)
	//                 ├─── b (2 modules)
	//                 └─── c (1 module)
	ModuleTree *ModuleTreeNode

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

	sorted  []string // list of sorted module names
	Modules map[string]*ast.Module

	Errors ast.Errors
}

func NewCompiler() *Compiler {

	c := &Compiler{
		Modules: map[string]*ast.Module{},
	}

	return c
}

func getGlobals(pkg *ast.Package, rules []term.Var) map[term.Var]term.Ref {
	globals := map[term.Var]term.Ref{}

	// Populate globals with exports within the package.
	for _, v := range rules {
		global := append(term.Ref{}, pkg.Path...)
		global = append(global, &term.Term{Value: term.String(v)})
		globals[v] = global
	}

	return globals
}

func (c *Compiler) Compile(modules map[string]*ast.Module) {
	for k, v := range modules {
		c.Modules[k] = v // TODO : Copy()
		c.sorted = append(c.sorted, k)
	}

	sort.Strings(c.sorted)

	c.compile()
}

func (c *Compiler) compile() {

	c.resolveAllRefs()
	c.setModuleTree()

}

func (c *Compiler) Failed() bool {
	return false
}

func (c *Compiler) NewQueryCompiler() QueryCompiler {
	qc := queryCompiler{
		compiler: c,
	}

	return qc
}

func (c *Compiler) getExports() *util.HashMap {

	rules := util.NewHashMap(func(a, b util.T) bool {
		r1 := a.(term.Ref)
		r2 := a.(term.Ref)
		return r1.Equal(r2)
	}, func(v util.T) int {
		return v.(term.Ref).Hash()
	})

	for _, name := range c.sorted {
		mod := c.Modules[name]
		rv, ok := rules.Get(mod.Package.Path)
		if !ok {
			rv = []term.Var{}
		}
		rvs := rv.([]term.Var)

		for _, rule := range mod.Rules {
			rvs = append(rvs, rule.Name)
		}
		rules.Put(mod.Package.Path, rvs)
	}

	return rules
}

// resolveAllRefs resolves references in expressions to their fully qualified values.
//
// For instance, given the following module:
//
// package a.b
// import data.foo.bar
// p[x] { bar[_] = x }
//
// The reference "bar[_]" would be resolved to "data.foo.bar[_]".
func (c *Compiler) resolveAllRefs() {

	rules := c.getExports()

	for _, name := range c.sorted {
		mod := c.Modules[name]

		var ruleExports []term.Var
		if x, ok := rules.Get(mod.Package.Path); ok {
			ruleExports = x.([]term.Var)
		}

		globals := getGlobals(mod.Package, ruleExports)

		if globals == nil { // TEMPORARY

		}
	}
}

func (c *Compiler) setModuleTree() {
	c.ModuleTree = NewModuleTree(c.Modules)
}

type ModuleTreeNode struct {
	Key      term.Value // package path ie: rego.a.b
	Modules  []*ast.Module
	Children map[term.Value]*ModuleTreeNode
}

// NewModuleTree returns a new ModuleTreeNode that represents the root
// of the module tree populated with the given modules.
func NewModuleTree(mods map[string]*ast.Module) *ModuleTreeNode {
	root := &ModuleTreeNode{
		Children: map[term.Value]*ModuleTreeNode{},
	}
	for _, m := range mods {
		node := root
		for _, x := range m.Package.Path {
			c, ok := node.Children[x.Value]
			if !ok {
				c = &ModuleTreeNode{
					Key:      x.Value,
					Children: map[term.Value]*ModuleTreeNode{},
				}
				node.Children[x.Value] = c
			}
			node = c
		}
		node.Modules = append(node.Modules, m)
	}
	return root
}

type TreeNode struct {
	Key      term.Value // rule path ie: rego.a.b
	Values   []ast.Rule
	Children map[term.Value]*TreeNode
}
