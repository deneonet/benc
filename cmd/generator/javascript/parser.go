package javascript

import (
	"bufio"
	"bytes"
	"fmt"
	"github.com/banditmoscow1337/benc/cmd/generator/common"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"regexp"
	"strings"
)

var jsClass = regexp.MustCompile(`^\s*class\s+([A-Za-z_$][\w$]*)(?:\s+extends\s+Map)?\s*\{`)
var jsField = regexp.MustCompile(`^\s*this\.([A-Za-z_$][\w$]*)\s*=\s*(.+?);\s*(?://\s*(.*))?$`)
var jsInt = regexp.MustCompile(`^-?\d+n?$`)
var jsFloat = regexp.MustCompile(`^-?\d+\.\d+$`)
var jsNew = regexp.MustCompile(`^new\s+([A-Za-z_$][\w$]*)\s*\(`)

// Parse reads constructor declarations and explicit Go type hints into Go AST.
// An ambiguous initializer is an error: it cannot silently become any.
func Parse(ctx *common.Context) error {
	data, err := os.ReadFile(ctx.InputFile)
	if err != nil {
		return err
	}
	ctx.FileSet = token.NewFileSet()
	astFile := ctx.FileSet.AddFile(ctx.InputFile, -1, len(data))
	lineStarts := []int{0}
	for i, b := range data {
		if b == '\n' && i+1 < len(data) {
			astFile.AddLine(i + 1)
			lineStarts = append(lineStarts, i+1)
		}
	}
	ctx.PkgName = strings.ToLower(ctx.BaseName)
	ctx.PkgName = regexp.MustCompile(`[^a-z0-9_]`).ReplaceAllString(ctx.PkgName, "_")
	if ctx.PkgName == "" || (ctx.PkgName[0] >= '0' && ctx.PkgName[0] <= '9') {
		ctx.PkgName = "js_" + ctx.PkgName
	}
	scan := bufio.NewScanner(bytes.NewReader(data))
	var current *ast.TypeSpec
	var fields []*ast.Field
	inCtor, ctorSeen, alias := false, false, false
	depth, line := 0, 0
	var hint string
	finish := func() error {
		if current == nil {
			return nil
		}
		if alias {
			e, err := parser.ParseExpr(hint)
			if err != nil {
				return fmt.Errorf("%s:%d: invalid map type hint: %w", ctx.InputFile, line, err)
			}
			if _, ok := e.(*ast.MapType); !ok {
				return fmt.Errorf("%s:%d: named map requires //benc:type map[K]V", ctx.InputFile, line)
			}
			current.Type = e
		} else {
			if !ctorSeen {
				return fmt.Errorf("%s:%d: %s has no constructor", ctx.InputFile, line, current.Name.Name)
			}
			current.Type = &ast.StructType{Fields: &ast.FieldList{List: fields}}
		}
		ctx.Types = append(ctx.Types, current)
		current = nil
		fields = nil
		hint = ""
		alias = false
		ctorSeen = false
		return nil
	}
	for scan.Scan() {
		line++
		raw := scan.Text()
		trimmed := strings.TrimSpace(raw)
		if current == nil {
			if strings.HasPrefix(trimmed, "//benc:type ") {
				hint = strings.TrimSpace(strings.TrimPrefix(trimmed, "//benc:type "))
				continue
			}
			if m := jsClass.FindStringSubmatch(raw); m != nil {
				name := ast.NewIdent(m[1])
				name.NamePos = astFile.Pos(lineStarts[line-1])
				current = &ast.TypeSpec{Name: name}
				alias = strings.Contains(raw, "extends Map") || strings.HasPrefix(hint, "map[")
				depth = 1
			} else if strings.HasPrefix(trimmed, "class ") {
				return fmt.Errorf("%s:%d: unsupported class declaration", ctx.InputFile, line)
			}
			continue
		}
		if alias {
			depth += strings.Count(raw, "{") - strings.Count(raw, "}")
			if depth == 0 {
				if err := finish(); err != nil {
					return err
				}
			}
			continue
		}
		if strings.HasPrefix(trimmed, "constructor(") && depth == 1 {
			inCtor = true
			ctorSeen = true
		}
		if inCtor && strings.HasPrefix(trimmed, "this.") {
			m := jsField.FindStringSubmatch(raw)
			if m == nil {
				return fmt.Errorf("%s:%d: unsupported constructor assignment", ctx.InputFile, line)
			}
			if m[3] == "benc:ignore" {
				continue
			}
			typeHint := strings.TrimSpace(m[3])
			var e ast.Expr
			if typeHint != "" {
				var err error
				e, err = parser.ParseExpr(typeHint)
				if err != nil {
					return fmt.Errorf("%s:%d: %s: invalid Go type hint %q: %w", ctx.InputFile, line, m[1], typeHint, err)
				}
			} else {
				var err error
				e, err = inferJS(m[2])
				if err != nil {
					return fmt.Errorf("%s:%d: %s: %w (add // GoType)", ctx.InputFile, line, m[1], err)
				}
			}
			name := ast.NewIdent(m[1])
			name.NamePos = astFile.Pos(lineStarts[line-1])
			fields = append(fields, &ast.Field{Names: []*ast.Ident{name}, Type: e})
		}
		depth += strings.Count(raw, "{") - strings.Count(raw, "}")
		if inCtor && depth == 1 {
			inCtor = false
		}
		if depth == 0 {
			if err := finish(); err != nil {
				return err
			}
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

func inferJS(value string) (ast.Expr, error) {
	value = strings.TrimSpace(value)
	switch {
	case value == "true" || value == "false":
		return ast.NewIdent("bool"), nil
	case jsInt.MatchString(value):
		if strings.HasSuffix(value, "n") {
			return ast.NewIdent("int64"), nil
		}
		return ast.NewIdent("int"), nil
	case jsFloat.MatchString(value):
		return ast.NewIdent("float64"), nil
	case strings.HasPrefix(value, `"`) || strings.HasPrefix(value, "'"):
		return ast.NewIdent("string"), nil
	case strings.HasPrefix(value, "new Date("):
		return parser.ParseExpr("time.Time")
	case strings.HasPrefix(value, "new Uint8Array("):
		return parser.ParseExpr("[]byte")
	}
	if m := jsNew.FindStringSubmatch(value); m != nil && m[1] != "Map" && m[1] != "Array" {
		return ast.NewIdent(m[1]), nil
	}
	return nil, fmt.Errorf("ambiguous initializer %q", value)
}
