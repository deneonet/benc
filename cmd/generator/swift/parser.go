package swift

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

var swiftStruct = regexp.MustCompile(`^(?:public\s+)?struct\s+([A-Za-z_][A-Za-z_0-9]*)\s*\{\s*(?://\s*(.*))?$`)
var swiftField = regexp.MustCompile(`^(?:public\s+)?(?:var|let)\s+([A-Za-z_][A-Za-z_0-9]*)\s*:\s*(.+?)\s*=\s*(.*?)(?:\s*//\s*(.*))?$`)

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
	depth, offset, line := 0, 0, 0
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
			if m := swiftStruct.FindStringSubmatch(s); m != nil {
				id := ast.NewIdent(m[1])
				id.NamePos = pos
				current = &ast.TypeSpec{Name: id}
				depth = 1
				fields = nil
				if strings.HasPrefix(strings.TrimSpace(m[2]), "benc:type ") {
					current.Type, err = parser.ParseExpr(strings.TrimSpace(strings.TrimPrefix(strings.TrimSpace(m[2]), "benc:type ")))
					if err != nil {
						return fmt.Errorf("%s:%d: %w", ctx.InputFile, line, err)
					}
				}
				continue
			}
			if strings.HasPrefix(s, "struct ") || strings.HasPrefix(s, "public struct ") {
				return fmt.Errorf("%s:%d: unsupported Swift struct", ctx.InputFile, line)
			}
			continue
		}
		if depth == 1 && current.Type == nil && (strings.HasPrefix(s, "var ") || strings.HasPrefix(s, "let ") || strings.HasPrefix(s, "public var ") || strings.HasPrefix(s, "public let ")) {
			m := swiftField.FindStringSubmatch(s)
			if m == nil {
				return fmt.Errorf("%s:%d: unsupported Swift property", ctx.InputFile, line)
			}
			if strings.TrimSpace(m[4]) != "benc:ignore" {
				var e ast.Expr
				if m[4] != "" {
					e, err = parser.ParseExpr(strings.TrimSpace(m[4]))
				} else {
					e, err = swiftType(m[2])
				}
				if err != nil {
					return fmt.Errorf("%s:%d: %s.%s: %w", ctx.InputFile, line, current.Name.Name, m[1], err)
				}
				id := ast.NewIdent(m[1])
				id.NamePos = pos
				fields = append(fields, &ast.Field{Names: []*ast.Ident{id}, Type: e})
			}
		}
		depth += strings.Count(raw, "{") - strings.Count(raw, "}")
		if depth == 0 {
			if current.Type == nil {
				current.Type = &ast.StructType{Fields: &ast.FieldList{List: fields}}
			}
			ctx.Types = append(ctx.Types, current)
			current = nil
		}
		if depth < 0 {
			return fmt.Errorf("%s:%d: unexpected closing brace", ctx.InputFile, line)
		}
	}
	if err := scan.Err(); err != nil {
		return err
	}
	if current != nil {
		return fmt.Errorf("%s: unterminated struct %s", ctx.InputFile, current.Name.Name)
	}
	return nil
}
func swiftType(raw string) (ast.Expr, error) {
	s := strings.TrimSpace(raw)
	if strings.HasSuffix(s, "?") {
		e, err := swiftType(strings.TrimSuffix(s, "?"))
		return &ast.StarExpr{X: e}, err
	}
	if strings.HasPrefix(s, "[") && strings.HasSuffix(s, "]") {
		inside := s[1 : len(s)-1]
		parts := splitSwift(inside, ':')
		if len(parts) == 2 {
			k, err := swiftType(parts[0])
			if err != nil {
				return nil, err
			}
			v, err := swiftType(parts[1])
			return &ast.MapType{Key: k, Value: v}, err
		}
		e, err := swiftType(inside)
		return &ast.ArrayType{Elt: e}, err
	}
	if s == "Date" {
		return &ast.SelectorExpr{X: ast.NewIdent("time"), Sel: ast.NewIdent("Time")}, nil
	}
	primitives := map[string]string{"String": "string", "Bool": "bool", "Int": "int", "UInt": "uint", "Int8": "int8", "UInt8": "uint8", "Int16": "int16", "UInt16": "uint16", "Int32": "int32", "UInt32": "uint32", "Int64": "int64", "UInt64": "uint64", "Float": "float32", "Double": "float64"}
	if v, ok := primitives[s]; ok {
		return ast.NewIdent(v), nil
	}
	if !token.IsIdentifier(s) {
		return nil, fmt.Errorf("unsupported Swift type %q", s)
	}
	return ast.NewIdent(s), nil
}
func splitSwift(s string, delimiter rune) []string {
	depth, start := 0, 0
	var parts []string
	for i, r := range s {
		switch r {
		case '[':
			depth++
		case ']':
			depth--
		default:
			if r == delimiter && depth == 0 {
				parts = append(parts, strings.TrimSpace(s[start:i]))
				start = i + 1
			}
		}
	}
	return append(parts, strings.TrimSpace(s[start:]))
}
