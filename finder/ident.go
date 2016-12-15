package finder

import (
	"container/list"
	"fmt"
	"go/ast"

	"golang.org/x/tools/go/ast/astutil"
)

// FindDefinition finds definition
func (f *Finder) FindDefinition(file string, pos int) (*Definition, error) {
	ident, err := f.FindIdent(file, pos)
	if err != nil {
		return nil, err
	}
	decl, err := f.FindIdentDecl(ident)
	if err != nil {
		return nil, err
	}
	return f.ToDefinition(decl)
}

// FindIdent finds ident at file:pos
func (f *Finder) FindIdent(file string, pos int) (*ast.Ident, error) {
	nodes, err := f.nodes(file, pos, pos)
	if err == nil && len(nodes) > 0 {
		ident, ok := nodes[0].(*ast.Ident)
		if ok {
			return ident, nil
		}
	}
	return nil, fmt.Errorf("can't find identifier")
}

// Chain find parent chain of node
func (f *Finder) Chain(node ast.Node) ([]ast.Node, error) {
	tf := f.tokenSet.File(node.Pos())
	if tf != nil {
		af := f.astFiles[tf.Name()]
		nodes, _ := astutil.PathEnclosingInterval(af, node.Pos(), node.End())
		return nodes, nil
	}
	return nil, fmt.Errorf("can't find node")
}

// FindIdentDecl finds ident decl
func (f *Finder) FindIdentDecl(ident *ast.Ident) (ast.Node, error) {
	stack := list.New()
	nodes, err := f.Chain(ident)
	if err != nil {
		return nil, err
	}
	if len(nodes) < 2 {
		return nil, fmt.Errorf("ident is not a valid node")
	}
	if _, ok := nodes[1].(*ast.SelectorExpr); ok && ident.Obj == nil {
		f.AnalyseSelector(stack, nodes[1])
	} else {
		stack.PushBack(ident)
	}
	node, err := f.AnalyseStack(stack)
	if err != nil {
		return nil, err
	}
	return node, nil
}

// AnalyseSelector analyses selector
func (f *Finder) AnalyseSelector(stack *list.List, node ast.Node) {
	switch n := node.(type) {
	case *ast.SelectorExpr:
		stack.PushBack(n.Sel)
		f.AnalyseSelector(stack, n.X)
	case *ast.TypeAssertExpr:
		f.AnalyseSelector(stack, n.Type)
	case *ast.StarExpr:
		f.AnalyseSelector(stack, n.X)
	case *ast.CallExpr:
		f.AnalyseSelector(stack, n.Fun)
	case *ast.Ident:
		stack.PushBack(n)
	case *ast.CompositeLit:
		f.AnalyseSelector(stack, n.Type)
	}
}

// AnalyseStack analyse selector stack
func (f *Finder) AnalyseStack(stack *list.List) (ast.Node, error) {
	for stack.Len() > 1 {
		ident := stack.Remove(stack.Back()).(*ast.Ident)
		// selector := stack.Remove(stack.Back()).(*ast.Ident)
		if ident.Obj != nil {
			switch decl := ident.Obj.Decl.(type) {
			case (*ast.AssignStmt):
				// x := -1
				// for i, expr := range decl.Lhs {
				// 	if e, ok := expr.(*ast.Ident); ok && e.Name == ident.Name {
				// 		x = i
				// 		break
				// 	}
				// }
				// if len(decl.Lhs) == len(decl.Rhs) {
				// 	v := decl.Rhs[x]
				// 	f.AnalyseSelector(stack)
				// } else {
				// 	// TODO: find func result
				// }
			case (*ast.ValueSpec):
				for i, name := range decl.Names {
					if name.Name == ident.Name {
						if decl.Type != nil {
							f.AnalyseSelector(stack, decl.Type)
						} else {
							if len(decl.Names) == len(decl.Values) {
								f.AnalyseSelector(stack, decl.Values[i])
							} else {
								// the i-th result of the function
								fs := list.New()
								f.AnalyseSelector(fs, decl.Values[0].(*ast.CallExpr).Fun)
								if fs.Len() <= 0 {
									return nil, fmt.Errorf("can't find any function")
								}
								funcDecl, err := f.FindIdentDecl(fs.Front().Value.(*ast.Ident))
								if err != nil {
									return nil, err
								}
								f.AnalyseSelector(stack, funcDecl.(*ast.FuncDecl).Type.Results.List[i].Type)
							}

						}
					}
				}
			case (*ast.TypeSpec):
				f.AnalyseSelector(stack, decl.Type)
			case (*ast.FuncDecl):
				f.AnalyseSelector(stack, decl.Type.Results.List[0].Type)
			}
		} else {
			// TODO
		}
	}
	return stack.Back().Value.(ast.Node), nil
}

// ToDefinition transforms node to difinition
func (f *Finder) ToDefinition(node ast.Node) (*Definition, error) {
	ident, ok := node.(*ast.Ident)
	if ok && ident.Obj != nil {
		switch decl := ident.Obj.Decl.(type) {
		case (*ast.AssignStmt):
			for _, expr := range decl.Lhs {
				if e, ok := expr.(*ast.Ident); ok && e.Name == ident.Name {
					return f.definition(e)
				}
			}
		case (*ast.ValueSpec):
			for _, name := range decl.Names {
				if name.Name == ident.Name {
					return f.definition(name)
				}
			}
		case (*ast.TypeSpec):
			return f.definition(decl.Name)
		case (*ast.FuncDecl):
			return f.definition(decl.Name)
		}
		ast.Print(f.tokenSet, ident)
	}
	return nil, fmt.Errorf("node is not a declaration")
}
