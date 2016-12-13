package main

import (
	"container/list"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"

	"log"

	"golang.org/x/tools/go/ast/astutil"
)

// Definition describes definition of an identifier
type Definition struct {
	Name        string `json:"name"`
	Package     string `json:"package"`
	Declaration string `json:"declaration"`
	Path        string `json:"path"`
	Document    string `json:"document"`
}

func findDefinition(goRoot, goPath, filepath string, pos int) (*Definition, error) {
	set := token.NewFileSet()
	f, err := parser.ParseFile(set, filepath, nil, parser.ParseComments)
	if err != nil {
		return nil, err
	}
	nodes, _ := astutil.PathEnclosingInterval(f, token.Pos(pos), token.Pos(pos))
	if len(nodes) > 0 {
		ident, ok := nodes[0].(*ast.Ident)
		if ok {
			stack := list.New()
			if _, ok = nodes[1].(*ast.SelectorExpr); ok {
				findSelectorStack(stack, nodes[1])
			} else {
				stack.PushBack(ident)
			}
			return findDefinitionByStack(set, stack)
		}
	}
	return nil, fmt.Errorf("can't find definition")
}

func findSelectorStack(stack *list.List, node ast.Node) {
	switch n := node.(type) {
	case *ast.SelectorExpr:
		stack.PushBack(n.Sel)
		findSelectorStack(stack, n.X)
	case *ast.TypeAssertExpr:
		findSelectorStack(stack, n.Type)
	case *ast.StarExpr:
		findSelectorStack(stack, n.X)
	case *ast.CallExpr:
		findSelectorStack(stack, n.Fun)
	case *ast.Ident:
		stack.PushBack(n)
	}
}

func findDefinitionByStack(set *token.FileSet, stack *list.List) (*Definition, error) {
	// for stack.Len() > 0 {
	// 	value := stack.Remove(stack.Back())
	// 	ast.Print(set, value)
	// 	log.Println("=======")
	// }

	ident := stack.Remove(stack.Front()).(*ast.Ident)
	ast.Print(set, ident)
	f := set.File(ident.Pos())
	if ident.Obj != nil {
		switch decl := ident.Obj.Decl.(type) {
		case (*ast.AssignStmt):
			for _, expr := range decl.Lhs {
				if e, ok := expr.(*ast.Ident); ok && e.Name == ident.Name {
					return &Definition{
						Name:        ident.Name,
						Package:     f.Name(),
						Declaration: "",
						Path:        set.Position(e.Pos()).String(),
						Document:    "",
					}, nil
				}
			}
		case (*ast.ValueSpec):
			for _, name := range decl.Names {
				if name.Name == ident.Name {
					return &Definition{
						Name:        ident.Name,
						Package:     f.Name(),
						Declaration: "",
						Path:        set.Position(name.Pos()).String(),
						Document:    "",
					}, nil
				}
			}
			log.Println(decl)
		}
	} else {
		log.Println("external", ident)
	}
	return nil, fmt.Errorf("can't find definition")
}
