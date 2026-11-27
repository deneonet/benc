package csharp

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

var csClass = regexp.MustCompile(`^(?:public\s+)?(?:sealed\s+)?class\s+([A-Za-z_][A-Za-z_0-9]*)\s*\{\s*(?://\s*(.*))?$`)
var csField = regexp.MustCompile(`^public\s+(.+?)\s+([A-Za-z_][A-Za-z_0-9]*)\s*=\s*.+;\s*(?://\s*(.*))?$`)

// Parse maps public C# data fields to Go AST. An inline Go type is used for
// C# types that lose signed width, fixed array length or time precision.
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
	depth, skip, offset, line := 0, 0, 0, 0
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
		if skip > 0 {
			skip += strings.Count(raw, "{") - strings.Count(raw, "}")
			continue
		}
		if current == nil {
			if m := csClass.FindStringSubmatch(s); m != nil {
				if strings.TrimSpace(m[2]) == "benc:internal" {
					skip = 1
					continue
				}
				id := ast.NewIdent(m[1])
				id.NamePos = pos
				current = &ast.TypeSpec{Name: id}
				fields = nil
				depth = 1
				if strings.HasPrefix(strings.TrimSpace(m[2]), "benc:type ") {
					current.Type, err = parser.ParseExpr(strings.TrimSpace(strings.TrimPrefix(strings.TrimSpace(m[2]), "benc:type ")))
					if err != nil {
						return fmt.Errorf("%s:%d: %w", ctx.InputFile, line, err)
					}
				}
				continue
			}
			if strings.HasPrefix(s, "public class ") || strings.HasPrefix(s, "class ") {
				return fmt.Errorf("%s:%d: unsupported C# class", ctx.InputFile, line)
			}
			continue
		}
		if depth == 1 && current.Type == nil && strings.HasPrefix(s, "public ") && !strings.Contains(strings.SplitN(s, "=", 2)[0], "(") {
			m := csField.FindStringSubmatch(s)
			if m == nil {
				return fmt.Errorf("%s:%d: unsupported C# field", ctx.InputFile, line)
			}
			if strings.TrimSpace(m[3]) != "benc:ignore" {
				var e ast.Expr
				if m[3] != "" {
					e, err = parser.ParseExpr(strings.TrimSpace(m[3]))
				} else {
					e, err = csType(m[1])
				}
				if err != nil {
					return fmt.Errorf("%s:%d: %s.%s: %w", ctx.InputFile, line, current.Name.Name, m[2], err)
				}
				id := ast.NewIdent(m[2])
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

func csType(raw string) (ast.Expr, error) {
	s := strings.TrimSpace(raw)
	if strings.HasSuffix(s, "?") {
		e, err := csType(strings.TrimSuffix(s, "?"))
		return &ast.StarExpr{X: e}, err
	}
	if strings.HasSuffix(s, "[]") {
		e, err := csType(strings.TrimSuffix(s, "[]"))
		return &ast.ArrayType{Elt: e}, err
	}
	if i := strings.IndexByte(s, '<'); i >= 0 && strings.HasSuffix(s, ">") {
		args := split(s[i+1 : len(s)-1])
		name := s[:i]
		switch name {
		case "List", "IList", "ICollection":
			if len(args) == 1 {
				e, err := csType(args[0])
				return &ast.ArrayType{Elt: e}, err
			}
		case "Dictionary", "IDictionary":
			if len(args) == 2 {
				k, err := csType(args[0])
				if err != nil {
					return nil, err
				}
				v, err := csType(args[1])
				return &ast.MapType{Key: k, Value: v}, err
			}
		}
		return nil, fmt.Errorf("unsupported C# generic %q", s)
	}
	primitives := map[string]string{"string": "string", "bool": "bool", "byte": "uint8", "sbyte": "int8", "short": "int16", "ushort": "uint16", "int": "int32", "uint": "uint32", "long": "int64", "ulong": "uint64", "nint": "int", "nuint": "uint", "float": "float32", "double": "float64"}
	if v, ok := primitives[s]; ok {
		return ast.NewIdent(v), nil
	}
	if !token.IsIdentifier(s) {
		return nil, fmt.Errorf("unsupported C# type %q", s)
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
