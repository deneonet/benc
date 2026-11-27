package kotlin

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

var class = regexp.MustCompile(`^(?:(?:public|internal|data|open)\s+)*class\s+([A-Za-z_][A-Za-z_0-9]*)\s*(?:\([^)]*\))?\s*\{\s*(?://\s*(.*))?$`)
var field = regexp.MustCompile(`^(?:public\s+)?(?:var|val)\s+([A-Za-z_][A-Za-z_0-9]*)\s*:\s*(.+?)\s*=\s*(.*?)(?:\s*//\s*(.*))?$`)

// Parse maps Kotlin class properties to Go AST; comments retain Go widths and fixed lengths.
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
			if strings.HasPrefix(s, "package ") {
				ctx.PkgName = strings.ReplaceAll(strings.TrimPrefix(s, "package "), ".", "_")
			}
			if m := class.FindStringSubmatch(s); m != nil {
				id := ast.NewIdent(m[1])
				id.NamePos = pos
				current = &ast.TypeSpec{Name: id}
				fields = nil
				depth = 1
				if strings.HasPrefix(m[2], "benc:type ") {
					current.Type, err = parser.ParseExpr(strings.TrimPrefix(m[2], "benc:type "))
					if err != nil {
						return fmt.Errorf("%s:%d: %w", ctx.InputFile, line, err)
					}
				}
				continue
			}
			if strings.Contains(s, " class ") || strings.HasPrefix(s, "class ") {
				return fmt.Errorf("%s:%d: unsupported Kotlin class", ctx.InputFile, line)
			}
			continue
		}
		if depth == 1 && current.Type == nil && (strings.HasPrefix(s, "var ") || strings.HasPrefix(s, "val ") || strings.HasPrefix(s, "public var ") || strings.HasPrefix(s, "public val ")) {
			m := field.FindStringSubmatch(s)
			if m == nil {
				return fmt.Errorf("%s:%d: unsupported Kotlin property", ctx.InputFile, line)
			}
			if strings.TrimSpace(m[4]) != "benc:ignore" {
				var e ast.Expr
				if m[4] != "" {
					e, err = parser.ParseExpr(strings.TrimSpace(m[4]))
				} else {
					e, err = kotlinType(m[2])
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
		return fmt.Errorf("%s: unterminated class %s", ctx.InputFile, current.Name.Name)
	}
	return nil
}

func kotlinType(raw string) (ast.Expr, error) {
	s := strings.TrimSpace(raw)
	if strings.HasSuffix(s, "?") {
		e, err := kotlinType(strings.TrimSuffix(s, "?"))
		return &ast.StarExpr{X: e}, err
	}
	if s == "ByteArray" {
		return &ast.ArrayType{Elt: ast.NewIdent("byte")}, nil
	}
	if i := strings.IndexByte(s, '<'); i >= 0 && strings.HasSuffix(s, ">") {
		args := split(s[i+1 : len(s)-1])
		name := strings.TrimSpace(s[:i])
		switch name {
		case "List", "MutableList", "ArrayList":
			if len(args) == 1 {
				e, err := kotlinType(args[0])
				return &ast.ArrayType{Elt: e}, err
			}
		case "Map", "MutableMap", "HashMap", "LinkedHashMap":
			if len(args) == 2 {
				k, err := kotlinType(args[0])
				if err != nil {
					return nil, err
				}
				v, err := kotlinType(args[1])
				return &ast.MapType{Key: k, Value: v}, err
			}
		}
		return nil, fmt.Errorf("unsupported Kotlin generic %q", s)
	}
	if s == "Instant" {
		return &ast.SelectorExpr{X: ast.NewIdent("time"), Sel: ast.NewIdent("Time")}, nil
	}
	primitives := map[string]string{"String": "string", "Boolean": "bool", "Byte": "int8", "UByte": "uint8", "Short": "int16", "UShort": "uint16", "Int": "int32", "UInt": "uint32", "Long": "int64", "ULong": "uint64", "Float": "float32", "Double": "float64"}
	if v, ok := primitives[s]; ok {
		return ast.NewIdent(v), nil
	}
	if !token.IsIdentifier(s) {
		return nil, fmt.Errorf("unsupported Kotlin type %q", s)
	}
	return ast.NewIdent(s), nil
}

func split(s string) []string {
	start, depth := 0, 0
	var parts []string
	for i, r := range s {
		switch r {
		case '<':
			depth++
		case '>':
			depth--
		case ',':
			if depth == 0 {
				parts = append(parts, strings.TrimSpace(s[start:i]))
				start = i + 1
			}
		}
	}
	return append(parts, strings.TrimSpace(s[start:]))
}
