package python

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

var pyClass = regexp.MustCompile(`^class\s+([A-Za-z_][A-Za-z_0-9]*)(?:\([^)]*\))?:\s*(?:#\s*(.*))?$`)

// Parse translates annotated Python classes/dataclasses into Go AST. Source
// comments disambiguate types Python cannot express (fixed-width numbers,
// arrays, binary bytes and nanosecond timestamps). No schema is created.
func Parse(ctx *common.Context) error {
	data, err := os.ReadFile(ctx.InputFile)
	if err != nil {
		return err
	}
	ctx.FileSet = token.NewFileSet()
	f := ctx.FileSet.AddFile(ctx.InputFile, -1, len(data))
	ctx.PkgName = regexp.MustCompile(`[^a-z0-9_]`).ReplaceAllString(strings.ToLower(ctx.BaseName), "_")
	if ctx.PkgName == "" || ctx.PkgName[0] >= '0' && ctx.PkgName[0] <= '9' {
		ctx.PkgName = "py_" + ctx.PkgName
	}
	var current *ast.TypeSpec
	var fields []*ast.Field
	var hint string
	line, offset, inMethod := 0, 0, false
	finish := func() {
		if current == nil {
			return
		}
		if current.Type == nil {
			current.Type = &ast.StructType{Fields: &ast.FieldList{List: fields}}
		}
		ctx.Types = append(ctx.Types, current)
		current = nil
		fields = nil
		inMethod = false
	}
	scan := bufio.NewScanner(bytes.NewReader(data))
	for scan.Scan() {
		line++
		raw := scan.Text()
		trimmed := strings.TrimSpace(raw)
		pos := f.Pos(offset)
		offset += len(raw) + 1
		if offset < len(data) {
			f.AddLine(offset)
		}
		if trimmed == "" {
			continue
		}
		if raw == trimmed {
			finish()
			if strings.HasPrefix(trimmed, "#benc:type ") {
				hint = strings.TrimSpace(strings.TrimPrefix(trimmed, "#benc:type "))
				continue
			}
			if m := pyClass.FindStringSubmatch(trimmed); m != nil {
				id := ast.NewIdent(m[1])
				id.NamePos = pos
				current = &ast.TypeSpec{Name: id}
				mapHint := strings.TrimSpace(m[2])
				if mapHint == "" {
					mapHint = hint
				}
				hint = ""
				if strings.Contains(trimmed, "dict[") || strings.HasPrefix(mapHint, "map[") {
					e, eErr := parser.ParseExpr(mapHint)
					if eErr != nil {
						return fmt.Errorf("%s:%d: named dict requires # map[K]V", ctx.InputFile, line)
					}
					if _, ok := e.(*ast.MapType); !ok {
						return fmt.Errorf("%s:%d: expected map type", ctx.InputFile, line)
					}
					current.Type = e
				}
				continue
			}
			if strings.HasPrefix(trimmed, "class ") {
				return fmt.Errorf("%s:%d: unsupported class declaration", ctx.InputFile, line)
			}
			continue
		}
		if current == nil {
			continue
		}
		indent := len(raw) - len(strings.TrimLeft(raw, " \t"))
		if indent != 4 {
			continue
		} // method body or annotation continuation
		if strings.HasPrefix(trimmed, "def ") || strings.HasPrefix(trimmed, "@") {
			inMethod = true
			continue
		}
		if current.Type != nil || inMethod {
			continue
		}
		if strings.HasPrefix(trimmed, "#") {
			continue
		}
		pair := strings.SplitN(trimmed, "#", 2)
		decl := strings.SplitN(strings.TrimSpace(pair[0]), ":", 2)
		if len(decl) != 2 || !token.IsIdentifier(strings.TrimSpace(decl[0])) {
			return fmt.Errorf("%s:%d: %s: expected annotated field", ctx.InputFile, line, current.Name.Name)
		}
		fieldName := strings.TrimSpace(decl[0])
		if len(pair) == 2 && strings.TrimSpace(pair[1]) == "benc:ignore" {
			continue
		}
		typeHint := ""
		if len(pair) == 2 {
			typeHint = strings.TrimSpace(pair[1])
		}
		if typeHint == "" {
			annotation := strings.TrimSpace(strings.SplitN(decl[1], "=", 2)[0])
			switch annotation {
			case "str":
				typeHint = "string"
			case "bool":
				typeHint = "bool"
			case "int":
				typeHint = "int"
			case "float":
				typeHint = "float64"
			case "bytes":
				typeHint = "[]byte"
			default:
				if token.IsIdentifier(annotation) {
					typeHint = annotation
				}
			}
		}
		if typeHint == "" {
			return fmt.Errorf("%s:%d: %s.%s is ambiguous (add # GoType)", ctx.InputFile, line, current.Name.Name, fieldName)
		}
		e, eErr := parser.ParseExpr(typeHint)
		if eErr != nil {
			return fmt.Errorf("%s:%d: %s.%s: %w", ctx.InputFile, line, current.Name.Name, fieldName, eErr)
		}
		id := ast.NewIdent(fieldName)
		id.NamePos = pos
		fields = append(fields, &ast.Field{Names: []*ast.Ident{id}, Type: e})
	}
	if err := scan.Err(); err != nil {
		return err
	}
	finish()
	return nil
}
