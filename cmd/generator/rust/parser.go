package rust

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

var rustStruct = regexp.MustCompile(`^(?:pub\s+)?struct\s+([A-Za-z_][A-Za-z_0-9]*)\s*\{\s*$`)
var rustMap = regexp.MustCompile(`^(?:pub\s+)?struct\s+([A-Za-z_][A-Za-z_0-9]*)\s*\(\s*(?:pub\s+)?HashMap<(.+)>\s*\);\s*(?://\s*(.*))?$`)
var rustField = regexp.MustCompile(`^(?:pub\s+)?([A-Za-z_][A-Za-z_0-9]*)\s*:\s*(.+),\s*(?://\s*(.*))?$`)

// Parse handles Rust data declarations. Go type annotations on a field are
// optional when Rust's type itself is unambiguous; the generated code carries
// them to preserve byte/uint8 and named map spelling on round trips.
func Parse(ctx *common.Context) error {
	data, err := os.ReadFile(ctx.InputFile)
	if err != nil {
		return err
	}
	ctx.FileSet = token.NewFileSet()
	f := ctx.FileSet.AddFile(ctx.InputFile, -1, len(data))
	ctx.PkgName = regexp.MustCompile(`[^a-z0-9_]`).ReplaceAllString(strings.ToLower(ctx.BaseName), "_")
	if ctx.PkgName == "" || ctx.PkgName[0] >= '0' && ctx.PkgName[0] <= '9' {
		ctx.PkgName = "rs_" + ctx.PkgName
	}
	var current *ast.TypeSpec
	var fields []*ast.Field
	line, offset := 0, 0
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
		if current == nil {
			if m := rustMap.FindStringSubmatch(trimmed); m != nil {
				e, eErr := rustType("HashMap<" + m[2] + ">")
				if eErr != nil {
					return fmt.Errorf("%s:%d: %w", ctx.InputFile, line, eErr)
				}
				id := ast.NewIdent(m[1])
				id.NamePos = pos
				ctx.Types = append(ctx.Types, &ast.TypeSpec{Name: id, Type: e})
				continue
			}
			if m := rustStruct.FindStringSubmatch(trimmed); m != nil {
				id := ast.NewIdent(m[1])
				id.NamePos = pos
				current = &ast.TypeSpec{Name: id}
				fields = nil
				continue
			}
			if strings.HasPrefix(trimmed, "pub struct ") || strings.HasPrefix(trimmed, "struct ") {
				return fmt.Errorf("%s:%d: unsupported Rust struct declaration", ctx.InputFile, line)
			}
			continue
		}
		if trimmed == "}" || trimmed == "};" {
			current.Type = &ast.StructType{Fields: &ast.FieldList{List: fields}}
			ctx.Types = append(ctx.Types, current)
			current = nil
			fields = nil
			continue
		}
		if trimmed == "" || strings.HasPrefix(trimmed, "//") {
			continue
		}
		m := rustField.FindStringSubmatch(trimmed)
		if m == nil {
			return fmt.Errorf("%s:%d: unsupported Rust field in %s", ctx.InputFile, line, current.Name.Name)
		}
		if strings.TrimSpace(m[3]) == "benc:ignore" {
			continue
		}
		var e ast.Expr
		if strings.TrimSpace(m[3]) != "" {
			e, err = parser.ParseExpr(strings.TrimSpace(m[3]))
		} else {
			e, err = rustType(m[2])
		}
		if err != nil {
			return fmt.Errorf("%s:%d: %s.%s: %w", ctx.InputFile, line, current.Name.Name, m[1], err)
		}
		id := ast.NewIdent(m[1])
		id.NamePos = pos
		fields = append(fields, &ast.Field{Names: []*ast.Ident{id}, Type: e})
	}
	if err := scan.Err(); err != nil {
		return err
	}
	if current != nil {
		return fmt.Errorf("%s: unterminated struct %s", ctx.InputFile, current.Name.Name)
	}
	return nil
}

func rustType(source string) (ast.Expr, error) {
	s := strings.TrimSpace(source)
	if strings.HasPrefix(s, "[") && strings.HasSuffix(s, "]") {
		parts := splitRust(s[1:len(s)-1], ';')
		if len(parts) != 2 {
			return nil, fmt.Errorf("expected [T; N], got %q", s)
		}
		e, err := rustType(parts[0])
		if err != nil {
			return nil, err
		}
		return &ast.ArrayType{Len: &ast.BasicLit{Kind: token.INT, Value: strings.TrimSpace(parts[1])}, Elt: e}, nil
	}
	if i := strings.IndexByte(s, '<'); i >= 0 && strings.HasSuffix(s, ">") {
		name := strings.TrimSpace(s[:i])
		args := splitRust(s[i+1:len(s)-1], ',')
		switch name {
		case "Vec", "Option":
			if len(args) != 1 {
				return nil, fmt.Errorf("%s expects one type", name)
			}
			e, err := rustType(args[0])
			if err != nil {
				return nil, err
			}
			if name == "Vec" {
				return &ast.ArrayType{Elt: e}, nil
			}
			return &ast.StarExpr{X: e}, nil
		case "HashMap", "std::collections::HashMap":
			if len(args) != 2 {
				return nil, fmt.Errorf("HashMap expects two types")
			}
			k, err := rustType(args[0])
			if err != nil {
				return nil, err
			}
			v, err := rustType(args[1])
			if err != nil {
				return nil, err
			}
			return &ast.MapType{Key: k, Value: v}, nil
		case "chrono::DateTime", "DateTime":
			if len(args) == 1 && (strings.TrimSpace(args[0]) == "chrono::Utc" || strings.TrimSpace(args[0]) == "Utc") {
				return &ast.SelectorExpr{X: ast.NewIdent("time"), Sel: ast.NewIdent("Time")}, nil
			}
		}
		return nil, fmt.Errorf("unsupported Rust generic %s", name)
	}
	primitives := map[string]string{"String": "string", "str": "string", "bool": "bool", "u8": "uint8", "i8": "int8", "u16": "uint16", "i16": "int16", "u32": "uint32", "i32": "int32", "u64": "uint64", "i64": "int64", "usize": "uint", "isize": "int", "f32": "float32", "f64": "float64"}
	if name, ok := primitives[s]; ok {
		return ast.NewIdent(name), nil
	}
	if !token.IsIdentifier(s) {
		return nil, fmt.Errorf("unsupported Rust type %q", s)
	}
	return ast.NewIdent(s), nil
}
func splitRust(s string, separator rune) []string {
	depth, start := 0, 0
	var result []string
	for i, r := range s {
		switch r {
		case '<', '[':
			depth++
		case '>', ']':
			depth--
		default:
			if r == separator && depth == 0 {
				result = append(result, strings.TrimSpace(s[start:i]))
				start = i + 1
			}
		}
	}
	return append(result, strings.TrimSpace(s[start:]))
}
