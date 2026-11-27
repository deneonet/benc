package common_test

import (
	"go/ast"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/banditmoscow1337/benc/cmd/generator/c"
	"github.com/banditmoscow1337/benc/cmd/generator/common"
	"github.com/banditmoscow1337/benc/cmd/generator/golang"
	"github.com/banditmoscow1337/benc/cmd/generator/javascript"
)

// The equivalent source declarations must enter every generator as the same
// named Go AST struct, irrespective of their source syntax.
func TestEquivalentSourceAST(t *testing.T) {
	want := []string{"Record", "ID:int32", "Flag:bool"}
	for _, tc := range []struct {
		name, file string
		parse      func(*common.Context) error
	}{
		{"go", "record.go", golang.Parse},
		{"c", "record.h", c.Parse},
		{"js", "record.js", javascript.Parse},
	} {
		t.Run(tc.name, func(t *testing.T) {
			ctx := common.NewContext(filepath.Join("..", "..", "..", "testdata", "conformance", "sources", tc.file))
			if err := tc.parse(ctx); err != nil {
				t.Fatal(err)
			}
			if len(ctx.Types) != 1 {
				t.Fatalf("types = %d, want one", len(ctx.Types))
			}
			ts := ctx.Types[0]
			st, ok := ts.Type.(*ast.StructType)
			if !ok {
				t.Fatalf("%T is not a struct", ts.Type)
			}
			got := []string{ts.Name.Name}
			for _, field := range st.Fields.List {
				if len(field.Names) != 1 {
					t.Fatalf("field names = %d", len(field.Names))
				}
				got = append(got, field.Names[0].Name+":"+ctx.ExprToString(field.Type))
			}
			if !reflect.DeepEqual(got, want) {
				t.Fatalf("normalized AST = %v; want %v", got, want)
			}
		})
	}
}
