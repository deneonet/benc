package golang

import (
	"go/ast"
	"go/parser"
	"go/token"

	"github.com/banditmoscow1337/benc/cmd/generator/common"
)

func Parse(ctx *common.Context) error {
	fset := token.NewFileSet()
	node, err := parser.ParseFile(fset, ctx.InputFile, nil, parser.ParseComments)
	if err != nil {
		return err
	}

	ctx.FileSet = fset
	ctx.PkgName = node.Name.Name
	ctx.Types = collectTypes(node)
	return nil
}

func collectTypes(node *ast.File) []*ast.TypeSpec {
	var types []*ast.TypeSpec
	for _, decl := range node.Decls {
		group, ok := decl.(*ast.GenDecl)
		if !ok {
			continue
		}
		for _, spec := range group.Specs {
			ts, ok := spec.(*ast.TypeSpec)
			if !ok {
				continue
			}
			switch ts.Type.(type) {
			case *ast.StructType, *ast.MapType:
				types = append(types, ts)
			}
		}
	}
	return types
}
