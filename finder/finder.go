package finder

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

// Finder describe a finder for searching source code
type Finder struct {
	GOPATH   string
	GOROOT   string
	tokenSet *token.FileSet
	astFiles map[string]*ast.File
}

// NewFinder creates a Finder
func NewFinder(GOPATH, GOROOT string) *Finder {
	return &Finder{
		GOPATH,
		GOROOT,
		token.NewFileSet(),
		make(map[string]*ast.File, 0),
	}
}

// AddFile append a file to finder
func (f *Finder) AddFile(file string) error {
	_, ok := f.astFiles[file]
	if ok {
		return nil
	}
	astFile, err := parser.ParseFile(f.tokenSet, file, nil, parser.ParseComments)
	if err != nil {
		return err
	}
	f.astFiles[file] = astFile
	return nil
}

func (f *Finder) file(file string) (*ast.File, error) {
	astFile, ok := f.astFiles[file]
	if ok {
		return astFile, nil
	}
	err := f.AddFile(file)
	if err != nil {
		return nil, err
	}
	astFile, _ = f.astFiles[file]
	return astFile, nil
}

func (f *Finder) fileByPos(pos token.Pos) (*ast.File, error) {
	tf := f.tokenSet.File(pos)
	astFile, err := f.file(tf.Name())
	if err != nil {
		return nil, err
	}
	return astFile, nil
}

func (f *Finder) position(pos token.Pos) token.Position {
	return f.tokenSet.Position(pos)
}

func (f *Finder) nodes(file string, start, end int) ([]ast.Node, error) {
	astFile, err := f.file(file)
	if err != nil {
		return nil, err
	}
	nodes, _ := astutil.PathEnclosingInterval(astFile, token.Pos(start), token.Pos(end))
	return nodes, nil
}

func (f *Finder) definition(ident *ast.Ident) (*Definition, error) {
	file, err := f.fileByPos(ident.Pos())
	if err != nil {
		return nil, err
	}
	return &Definition{
		Name:        ident.Name,
		Package:     file.Name.Name,
		Declaration: "",
		Path:        f.position(ident.Pos()).String(),
		Document:    "",
	}, nil
}

// FindDefinition finds the definition of token at file:pos
func (f *Finder) FindDefinition(file string, pos int) (*Definition, error) {
	nodes, _ := f.nodes(file, pos, pos)
	if len(nodes) > 0 {
		ident, ok := nodes[0].(*ast.Ident)
		if ok {
			stack := list.New()
			if _, ok = nodes[1].(*ast.SelectorExpr); ok && ident.Obj == nil {
				f.findSelector(stack, nodes[1])
			} else {
				stack.PushBack(ident)
			}
			return f.translateStack(stack)
		}
	}
	return nil, fmt.Errorf("can't find definition")
}

func (f *Finder) findSelector(stack *list.List, node ast.Node) {
	switch n := node.(type) {
	case *ast.SelectorExpr:
		stack.PushBack(n.Sel)
		f.findSelector(stack, n.X)
	case *ast.TypeAssertExpr:
		f.findSelector(stack, n.Type)
	case *ast.StarExpr:
		f.findSelector(stack, n.X)
	case *ast.CallExpr:
		f.findSelector(stack, n.Fun)
	case *ast.Ident:
		stack.PushBack(n)
	}
}

func (f *Finder) translateStack(stack *list.List) (*Definition, error) {
	var ident *ast.Ident
	for stack.Len() > 0 {
		ident = stack.Remove(stack.Back()).(*ast.Ident)
		ast.Print(f.tokenSet, ident)
		log.Println("=======")
	}
	if ident != nil && ident.Obj != nil {
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
	return nil, fmt.Errorf("can't find definition")
}
