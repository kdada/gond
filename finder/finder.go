package finder

import (
	"go/ast"
	"go/parser"
	"go/token"

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

type identDesc struct {
	ident  *ast.Ident
	isFunc bool
	isStar bool
}

func (id *identDesc) String() string {
	if id.isFunc {
		return id.ident.Name + "()"
	}
	if id.isStar {
		return "*" + id.ident.Name
	}
	return id.ident.Name
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
