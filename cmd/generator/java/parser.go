package java

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

var javaClass = regexp.MustCompile(`^(?:(?:public|final|abstract)\s+)*class\s+([A-Za-z_][A-Za-z_0-9]*)(?:\s+extends\s+(.+?))?\s*\{\s*(?://\s*(.*))?$`)

// Parse converts public Java data fields and named map classes into Go AST.
// An inline Go type comment disambiguates unsigned widths and fixed arrays.
func Parse(ctx *common.Context) error {
	data, err := os.ReadFile(ctx.InputFile)
	if err != nil {
		return err
	}
	ctx.FileSet = token.NewFileSet()
	f := ctx.FileSet.AddFile(ctx.InputFile, -1, len(data))
	ctx.PkgName = regexp.MustCompile(`[^a-z0-9_]`).ReplaceAllString(strings.ToLower(ctx.BaseName), "_")
	if ctx.PkgName == "" || ctx.PkgName[0] >= '0' && ctx.PkgName[0] <= '9' {
		ctx.PkgName = "java_" + ctx.PkgName
	}
	var current *ast.TypeSpec
	var fields []*ast.Field
	depth, skipDepth, line, offset := 0, 0, 0, 0
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
		if skipDepth > 0 {
			skipDepth += strings.Count(raw, "{") - strings.Count(raw, "}")
			if skipDepth < 0 {
				return fmt.Errorf("%s:%d: extra closing brace", ctx.InputFile, line)
			}
			continue
		}
		if current == nil {
			if strings.HasPrefix(trimmed, "package ") {
				pkg := strings.TrimSuffix(strings.TrimPrefix(trimmed, "package "), ";")
				ctx.PkgName = strings.ReplaceAll(pkg, ".", "_")
			}
			if m := javaClass.FindStringSubmatch(trimmed); m != nil {
				if strings.TrimSpace(m[3]) == "benc:internal" {
					skipDepth = 1
					continue
				}
				id := ast.NewIdent(m[1])
				id.NamePos = pos
				current = &ast.TypeSpec{Name: id}
				fields = nil
				depth = 1
				if m[2] != "" || strings.HasPrefix(strings.TrimSpace(m[3]), "benc:type ") {
					var e ast.Expr
					if strings.HasPrefix(strings.TrimSpace(m[3]), "benc:type ") {
						e, err = parser.ParseExpr(strings.TrimSpace(strings.TrimPrefix(strings.TrimSpace(m[3]), "benc:type ")))
					} else {
						e, err = javaType(m[2])
					}
					if err != nil {
						return fmt.Errorf("%s:%d: %s: %w", ctx.InputFile, line, m[1], err)
					}
					if _, ok := e.(*ast.MapType); !ok {
						return fmt.Errorf("%s:%d: only Map may be a named base", ctx.InputFile, line)
					}
					current.Type = e
				}
				continue
			}
			if strings.Contains(trimmed, " class ") || strings.HasPrefix(trimmed, "class ") {
				return fmt.Errorf("%s:%d: unsupported Java class", ctx.InputFile, line)
			}
			continue
		}
		if depth == 1 && current.Type == nil && strings.HasPrefix(trimmed, "public ") && strings.Contains(trimmed, ";") && !strings.Contains(strings.SplitN(trimmed, "=", 2)[0], "(") {
			parts := strings.SplitN(trimmed, "//", 2)
			decl := strings.TrimSpace(strings.TrimPrefix(parts[0], "public "))
			decl = strings.TrimSpace(strings.TrimSuffix(decl, ";"))
			decl = strings.TrimSpace(strings.SplitN(decl, " = ", 2)[0])
			i := strings.LastIndexByte(decl, ' ')
			if i < 1 || !token.IsIdentifier(decl[i+1:]) {
				return fmt.Errorf("%s:%d: unsupported Java field", ctx.InputFile, line)
			}
			name := decl[i+1:]
			var e ast.Expr
			if len(parts) == 2 && strings.TrimSpace(parts[1]) == "benc:ignore" {
				continue
			}
			if len(parts) == 2 && strings.TrimSpace(parts[1]) != "" {
				e, err = parser.ParseExpr(strings.TrimSpace(parts[1]))
			} else {
				e, err = javaType(decl[:i])
			}
			if err != nil {
				return fmt.Errorf("%s:%d: %s.%s: %w", ctx.InputFile, line, current.Name.Name, name, err)
			}
			id := ast.NewIdent(name)
			id.NamePos = pos
			fields = append(fields, &ast.Field{Names: []*ast.Ident{id}, Type: e})
		}
		depth += strings.Count(raw, "{") - strings.Count(raw, "}")
		if depth == 0 {
			if current.Type == nil {
				current.Type = &ast.StructType{Fields: &ast.FieldList{List: fields}}
			}
			ctx.Types = append(ctx.Types, current)
			current = nil
			fields = nil
		}
		if depth < 0 {
			return fmt.Errorf("%s:%d: extra class closing brace", ctx.InputFile, line)
		}
	}
	if err := scan.Err(); err != nil {
		return err
	}
	if current != nil {
		return fmt.Errorf("%s: unterminated Java class %s", ctx.InputFile, current.Name.Name)
	}
	return nil
}

func javaType(source string) (ast.Expr, error) {
	s := strings.TrimSpace(source)
	if strings.HasSuffix(s, "[]") {
		e, err := javaType(strings.TrimSpace(strings.TrimSuffix(s, "[]")))
		if err != nil {
			return nil, err
		}
		return &ast.ArrayType{Elt: e}, nil
	}
	if i := strings.IndexByte(s, '<'); i >= 0 && strings.HasSuffix(s, ">") {
		name := strings.TrimSpace(s[:i])
		args := splitJava(s[i+1 : len(s)-1])
		switch name {
		case "List", "ArrayList", "Optional":
			if len(args) != 1 {
				return nil, fmt.Errorf("%s needs one type", name)
			}
			e, err := javaType(args[0])
			if err != nil {
				return nil, err
			}
			if name == "Optional" {
				return &ast.StarExpr{X: e}, nil
			}
			return &ast.ArrayType{Elt: e}, nil
		case "Map", "HashMap":
			if len(args) != 2 {
				return nil, fmt.Errorf("Map needs two types")
			}
			k, err := javaType(args[0])
			if err != nil {
				return nil, err
			}
			v, err := javaType(args[1])
			if err != nil {
				return nil, err
			}
			return &ast.MapType{Key: k, Value: v}, nil
		}
		return nil, fmt.Errorf("unsupported Java generic %s", name)
	}
	if s == "Instant" {
		return &ast.SelectorExpr{X: ast.NewIdent("time"), Sel: ast.NewIdent("Time")}, nil
	}
	primitives := map[string]string{"String": "string", "boolean": "bool", "Boolean": "bool", "byte": "int8", "Byte": "int8", "short": "int16", "Short": "int16", "int": "int32", "Integer": "int32", "long": "int64", "Long": "int64", "float": "float32", "Float": "float32", "double": "float64", "Double": "float64"}
	if name, ok := primitives[s]; ok {
		return ast.NewIdent(name), nil
	}
	if !token.IsIdentifier(s) {
		return nil, fmt.Errorf("unsupported Java type %q", s)
	}
	return ast.NewIdent(s), nil
}
func splitJava(s string) []string {
	depth, start := 0, 0
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
