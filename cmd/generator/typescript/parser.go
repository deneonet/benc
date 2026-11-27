package typescript

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

var tsDeclaration = regexp.MustCompile(`^(?:(?:export|declare)\s+)*(class|interface)\s+([A-Za-z_][A-Za-z_0-9]*)(?:\s+extends\s+[^{}]+)?\s*\{\s*$`)
var tsMap = regexp.MustCompile(`^(?:(?:export|declare)\s+)*type\s+([A-Za-z_][A-Za-z_0-9]*)\s*=\s*Map<.*>;\s*(?://\s*(.*))?$`)
var tsField = regexp.MustCompile(`^(?:readonly\s+)?([A-Za-z_][A-Za-z_0-9]*)\s*:\s*(.+?)\s*;\s*(?://\s*(.*))?$`)

// Parse reads TypeScript declarations and converts field types directly to Go
// AST. Ambiguous number, bigint and container types require a source comment
// with their Go type (e.g. `Count: number; // uint32`). This annotation is in
// the declaration itself, not in a separate schema or intermediate model.
func Parse(ctx *common.Context) error {
	data, err := os.ReadFile(ctx.InputFile)
	if err != nil {
		return err
	}
	ctx.FileSet = token.NewFileSet()
	f := ctx.FileSet.AddFile(ctx.InputFile, -1, len(data))
	ctx.PkgName = regexp.MustCompile(`[^a-z0-9_]`).ReplaceAllString(strings.ToLower(ctx.BaseName), "_")
	if ctx.PkgName == "" || ctx.PkgName[0] >= '0' && ctx.PkgName[0] <= '9' {
		ctx.PkgName = "ts_" + ctx.PkgName
	}
	var current *ast.TypeSpec
	var fields []*ast.Field
	var hint string
	line, lineOffset := 0, 0
	scan := bufio.NewScanner(bytes.NewReader(data))
	for scan.Scan() {
		line++
		raw := scan.Text()
		trimmed := strings.TrimSpace(raw)
		pos := f.Pos(lineOffset)
		lineOffset += len(raw) + 1
		if lineOffset < len(data) {
			f.AddLine(lineOffset)
		}
		if trimmed == "" || strings.HasPrefix(trimmed, "import ") || strings.HasPrefix(trimmed, "// Code generated") {
			continue
		}
		if current == nil {
			if strings.HasPrefix(trimmed, "//benc:type ") {
				hint = strings.TrimSpace(strings.TrimPrefix(trimmed, "//benc:type "))
				continue
			}
			if m := tsMap.FindStringSubmatch(trimmed); m != nil {
				t := strings.TrimSpace(m[2])
				if t == "" {
					t = hint
				}
				e, eErr := parser.ParseExpr(t)
				if eErr != nil {
					return fmt.Errorf("%s:%d: invalid map type: %w", ctx.InputFile, line, eErr)
				}
				if _, ok := e.(*ast.MapType); !ok {
					return fmt.Errorf("%s:%d: type %s requires // map[K]V", ctx.InputFile, line, m[1])
				}
				id := ast.NewIdent(m[1])
				id.NamePos = pos
				ctx.Types = append(ctx.Types, &ast.TypeSpec{Name: id, Type: e})
				hint = ""
				continue
			}
			if m := tsDeclaration.FindStringSubmatch(trimmed); m != nil {
				id := ast.NewIdent(m[2])
				id.NamePos = pos
				current = &ast.TypeSpec{Name: id}
				if hint != "" {
					e, eErr := parser.ParseExpr(hint)
					if eErr != nil {
						return fmt.Errorf("%s:%d: named Map requires //benc:type map[K]V", ctx.InputFile, line)
					}
					if _, ok := e.(*ast.MapType); !ok {
						return fmt.Errorf("%s:%d: expected map type", ctx.InputFile, line)
					}
					current.Type = e
				}
				hint = ""
				fields = nil
				continue
			}
			if strings.HasPrefix(trimmed, "export ") || strings.HasPrefix(trimmed, "class ") || strings.HasPrefix(trimmed, "interface ") {
				return fmt.Errorf("%s:%d: unsupported TypeScript declaration", ctx.InputFile, line)
			}
			continue // imports, annotations, type-only helpers
		}
		if trimmed == "}" || trimmed == "};" {
			if current.Type == nil {
				current.Type = &ast.StructType{Fields: &ast.FieldList{List: fields}}
			}
			ctx.Types = append(ctx.Types, current)
			current = nil
			fields = nil
			continue
		}
		if current.Type != nil {
			continue
		} // methods of a named map are not fields
		if strings.HasPrefix(trimmed, "//") || strings.HasPrefix(trimmed, "/*") || strings.HasPrefix(trimmed, "*") {
			continue
		}
		if strings.Contains(trimmed, "(") && strings.HasSuffix(trimmed, ";") {
			if strings.HasPrefix(trimmed, "constructor(") || strings.HasPrefix(trimmed, "size(") || strings.HasPrefix(trimmed, "marshal(") || strings.HasPrefix(trimmed, "unmarshal(") {
				continue
			}
			return fmt.Errorf("%s:%d: unsupported method in %s", ctx.InputFile, line, current.Name.Name)
		}
		m := tsField.FindStringSubmatch(trimmed)
		if m == nil {
			return fmt.Errorf("%s:%d: unsupported TypeScript field in %s", ctx.InputFile, line, current.Name.Name)
		}
		typeHint := strings.TrimSpace(m[3])
		if typeHint == "benc:ignore" {
			continue
		}
		if typeHint == "" {
			switch m[2] {
			case "string":
				typeHint = "string"
			case "boolean":
				typeHint = "bool"
			default:
				if token.IsIdentifier(m[2]) && m[2] != "number" && m[2] != "bigint" {
					typeHint = m[2]
				}
			}
		}
		if typeHint == "" {
			return fmt.Errorf("%s:%d: %s.%s has ambiguous type %q (add // GoType)", ctx.InputFile, line, current.Name.Name, m[1], m[2])
		}
		e, eErr := parser.ParseExpr(typeHint)
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
		return fmt.Errorf("%s: unterminated TypeScript declaration %s", ctx.InputFile, current.Name.Name)
	}
	return nil
}
