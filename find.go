package main

import (
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

	id := nodes[0]
	parent := nodes[1]
	switch p := parent.(type) {
	case *ast.SelectorExpr:
		log.Println("SelectorExpr", p)
	default:
		ident := id.(*ast.Ident)
		if ident.Obj != nil {
			switch decl := ident.Obj.Decl.(type) {
			case (*ast.AssignStmt):
				for _, expr := range decl.Lhs {
					if e, ok := expr.(*ast.Ident); ok && e.Name == ident.Name {
						return &Definition{
							Name:        ident.Name,
							Package:     f.Name.String(),
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
							Package:     f.Name.String(),
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
	}
	return nil, fmt.Errorf("can't find definition")
}
