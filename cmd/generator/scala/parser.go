package scala

import (
	"bufio"
	"bytes"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"regexp"
	"strings"

	"github.com/banditmoscow1337/benc/cmd/generator/common"
)

var structHeader = regexp.MustCompile(`^(?:(?:final|case)\s+)*case\s+class\s+([A-Za-z_][A-Za-z_0-9]*)\s*\(\s*$`)
var mapHeader = regexp.MustCompile(`^(?:(?:final|case)\s+)*case\s+class\s+([A-Za-z_][A-Za-z_0-9]*)\s*\(value:\s*Map\[.*\](?:\s*=\s*Map\.empty)?\)\s*(?://\s*(.*))?$`)
var field = regexp.MustCompile(`^(?:val\s+)?([A-Za-z_][A-Za-z_0-9]*)\s*:\s*(.+?)\s*=\s*(.*?),?\s*//\s*(.+)$`)

// Parse translates Scala 3 case-class data into Go AST; width/fixed-length
// annotations are part of the declaration, never a separate schema.
func Parse(ctx *common.Context) error {
	data, err := os.ReadFile(ctx.InputFile)
	if err != nil {
		return err
	}
	ctx.FileSet = token.NewFileSet()
	f := ctx.FileSet.AddFile(ctx.InputFile, -1, len(data))
	ctx.PkgName = regexp.MustCompile(`[^a-z0-9_]`).ReplaceAllString(strings.ToLower(ctx.BaseName), "_")
	var current *ast.TypeSpec
	var fields []*ast.Field
	line, offset := 0, 0
	scan := bufio.NewScanner(bytes.NewReader(data))
	for scan.Scan() {
		line++
		raw := scan.Text()
		s := strings.TrimSpace(raw)
		pos := f.Pos(offset)
		offset += len(raw) + 1
		if offset < len(data) {
			f.AddLine(offset)
		}
		if current == nil {
			if strings.HasPrefix(s, "package ") {
				ctx.PkgName = strings.ReplaceAll(strings.TrimPrefix(s, "package "), ".", "_")
			}
			if m := mapHeader.FindStringSubmatch(s); m != nil {
				if !strings.HasPrefix(strings.TrimSpace(m[2]), "benc:type ") {
					return fmt.Errorf("%s:%d: named map requires // benc:type map[K]V", ctx.InputFile, line)
				}
				e, eErr := parser.ParseExpr(strings.TrimSpace(strings.TrimPrefix(strings.TrimSpace(m[2]), "benc:type ")))
				if eErr != nil {
					return fmt.Errorf("%s:%d: %w", ctx.InputFile, line, eErr)
				}
				id := ast.NewIdent(m[1])
				id.NamePos = pos
				ctx.Types = append(ctx.Types, &ast.TypeSpec{Name: id, Type: e})
				continue
			}
			if m := structHeader.FindStringSubmatch(s); m != nil {
				id := ast.NewIdent(m[1])
				id.NamePos = pos
				current = &ast.TypeSpec{Name: id}
				fields = nil
				continue
			}
			if strings.Contains(s, "case class ") {
				return fmt.Errorf("%s:%d: unsupported Scala case class", ctx.InputFile, line)
			}
			continue
		}
		if s == ")" {
			current.Type = &ast.StructType{Fields: &ast.FieldList{List: fields}}
			ctx.Types = append(ctx.Types, current)
			current = nil
			continue
		}
		if s == "" || strings.HasPrefix(s, "//") {
			continue
		}
		m := field.FindStringSubmatch(s)
		if m == nil {
			return fmt.Errorf("%s:%d: unsupported Scala field in %s", ctx.InputFile, line, current.Name.Name)
		}
		if strings.TrimSpace(m[4]) == "benc:ignore" {
			continue
		}
		e, eErr := parser.ParseExpr(strings.TrimSpace(m[4]))
		if eErr != nil {
			return fmt.Errorf("%s:%d: %s.%s: %w", ctx.InputFile, line, current.Name.Name, m[1], eErr)
		}
		id := ast.NewIdent(m[1])
		id.NamePos = pos
		fields = append(fields, &ast.Field{Names: []*ast.Ident{id}, Type: e})
	}
	if err := scan.Err(); err != nil {
		return err
	}
	if current != nil {
		return fmt.Errorf("%s: unterminated case class %s", ctx.InputFile, current.Name.Name)
	}
	return nil
}
