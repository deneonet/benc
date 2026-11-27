package common

import (
	"bytes"
	"fmt"
	"go/ast"
	"go/format"
	"go/token"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

type Generator interface {
	Generate() error
	Tests() error
}

// Context holds the shared state of the generation process (AST info).
type Context struct {
	InputFile, PkgName, BaseName, OutputDir string
	TypeSpecs                               map[string]*ast.TypeSpec
	Types                                   []*ast.TypeSpec
	FileSet                                 *token.FileSet
}

// NewContext creates a new shared context.
func NewContext(inputFile string) (ctx *Context) {
	ctx = &Context{
		InputFile: inputFile,
		TypeSpecs: make(map[string]*ast.TypeSpec),
		BaseName:  strings.TrimSuffix(filepath.Base(inputFile), filepath.Ext(inputFile)),
		OutputDir: filepath.Dir(inputFile),
	}

	return
}

// ExprToString converts an AST expression to its string representation.
func (c *Context) ExprToString(expr ast.Expr) string {
	var b bytes.Buffer
	if err := format.Node(&b, token.NewFileSet(), expr); err != nil {
		return ""
	}
	return b.String()
}

// ShouldIgnoreField checks if the field has a //benc:ignore comment.
func (c *Context) ShouldIgnoreField(field *ast.Field) bool {
	checkGroup := func(cg *ast.CommentGroup) bool {
		if cg == nil {
			return false
		}
		for _, cm := range cg.List {
			if strings.Contains(cm.Text, "//benc:ignore") {
				return true
			}
		}
		return false
	}
	return checkGroup(field.Doc) || checkGroup(field.Comment)
}

// IsUnsupportedType is retained for generated test helpers. Validation, rather than
// filtering, is responsible for rejecting every non-ignored unsupported field.
func (c *Context) IsUnsupportedType(expr ast.Expr) bool {
	return c.validateExpr(expr, nil, false) != nil
}

// GetSupportedFields filters fields of a struct based on ignore tags and unsupported types.
func (c *Context) GetSupportedFields(ts *ast.TypeSpec) []*ast.Field {
	var supportedFields []*ast.Field
	for _, field := range ts.Type.(*ast.StructType).Fields.List {
		if c.ShouldIgnoreField(field) {
			continue
		}
		supportedFields = append(supportedFields, field)
	}
	return supportedFields
}

func (ctx *Context) WriteFile(content *bytes.Buffer, prefix, lang string) error {
	path := filepath.Join(ctx.OutputDir, ctx.BaseName+"_"+strings.TrimPrefix(prefix, "_")+"."+strings.TrimPrefix(lang, "."))
	if err := os.WriteFile(path, content.Bytes(), 0644); err != nil {
		return err
	}

	content.Reset()

	return nil
}

// Validate checks the Go AST used by every generator before creating any output.
// A named type is resolved from the AST itself; there is no second schema.
func (c *Context) Validate() error {
	if len(c.Types) == 0 {
		return fmt.Errorf("%s: no supported declarations", c.InputFile)
	}
	c.TypeSpecs = make(map[string]*ast.TypeSpec, len(c.Types))
	for _, ts := range c.Types {
		if !token.IsIdentifier(ts.Name.Name) || !ts.Name.IsExported() {
			return c.problem(ts.Pos(), "type %s must be exported", ts.Name.Name)
		}
		if _, exists := c.TypeSpecs[ts.Name.Name]; exists {
			return c.problem(ts.Pos(), "duplicate type %s", ts.Name.Name)
		}
		c.TypeSpecs[ts.Name.Name] = ts
	}
	for _, ts := range c.Types {
		if ts.Assign.IsValid() {
			return c.problem(ts.Pos(), "%s: type aliases are not supported; use a defined type", ts.Name.Name)
		}
		switch t := ts.Type.(type) {
		case *ast.StructType:
			if t.Fields == nil || len(t.Fields.List) == 0 {
				return c.problem(ts.Pos(), "%s: empty structs are not supported", ts.Name.Name)
			}
			seen := map[string]bool{}
			included := 0
			for _, f := range t.Fields.List {
				if c.ShouldIgnoreField(f) {
					continue
				}
				included++
				if len(f.Names) == 0 {
					return c.problem(f.Pos(), "%s: embedded fields require //benc:ignore", ts.Name.Name)
				}
				for _, name := range f.Names {
					if seen[name.Name] {
						return c.problem(name.Pos(), "%s.%s: duplicate field", ts.Name.Name, name.Name)
					}
					seen[name.Name] = true
					if !token.IsIdentifier(name.Name) || !name.IsExported() {
						return c.problem(name.Pos(), "%s.%s: field must be exported", ts.Name.Name, name.Name)
					}
					if err := c.validateExpr(f.Type, map[string]bool{ts.Name.Name: true}, false); err != nil {
						return c.problem(f.Pos(), "%s.%s: %v", ts.Name.Name, name.Name, err)
					}
				}
			}
			if included == 0 {
				return c.problem(ts.Pos(), "%s: all fields are excluded", ts.Name.Name)
			}
		case *ast.MapType:
			if err := c.validateExpr(t, map[string]bool{ts.Name.Name: true}, false); err != nil {
				return c.problem(ts.Pos(), "%s: %v", ts.Name.Name, err)
			}
		default:
			return c.problem(ts.Pos(), "%s: only structs and named maps are supported", ts.Name.Name)
		}
	}
	return nil
}

var primitives = map[string]bool{"bool": true, "byte": true, "rune": true, "string": true, "int": true, "uint": true, "int8": true, "uint8": true, "int16": true, "uint16": true, "int32": true, "uint32": true, "int64": true, "uint64": true, "float32": true, "float64": true}

func (c *Context) validateExpr(e ast.Expr, seen map[string]bool, mapKey bool) error {
	switch t := e.(type) {
	case *ast.Ident:
		if primitives[t.Name] {
			if mapKey && (t.Name == "float32" || t.Name == "float64") {
				return fmt.Errorf("floating map keys are not supported across all targets")
			}
			return nil
		}
		if mapKey {
			return fmt.Errorf("named map keys are not supported across all targets")
		}
		ts, ok := c.TypeSpecs[t.Name]
		if !ok {
			return fmt.Errorf("unknown or unsupported type %s", t.Name)
		}
		if seen[t.Name] {
			return fmt.Errorf("recursive type %s is not supported across all targets", t.Name)
		}
		if seen == nil {
			seen = map[string]bool{}
		}
		next := make(map[string]bool, len(seen)+1)
		for n := range seen {
			next[n] = true
		}
		next[t.Name] = true
		switch d := ts.Type.(type) {
		case *ast.MapType:
			return c.validateExpr(d, next, false)
		case *ast.StructType:
			for _, f := range d.Fields.List {
				if !c.ShouldIgnoreField(f) {
					if err := c.validateExpr(f.Type, next, false); err != nil {
						return err
					}
				}
			}
			return nil
		}
		return fmt.Errorf("unsupported named type %s", t.Name)
	case *ast.SelectorExpr:
		if mapKey {
			return fmt.Errorf("time map keys are not supported across all targets")
		}
		if c.ExprToString(t) == "time.Time" {
			return nil
		}
		return fmt.Errorf("unsupported imported type %s", c.ExprToString(t))
	case *ast.StarExpr:
		if mapKey {
			return fmt.Errorf("pointer map keys are not supported")
		}
		return c.validateExpr(t.X, seen, false)
	case *ast.ArrayType:
		if mapKey {
			return fmt.Errorf("collection map keys are not supported")
		}
		if t.Len != nil {
			lit, ok := t.Len.(*ast.BasicLit)
			if !ok || lit.Kind != token.INT {
				return fmt.Errorf("array length must be an integer literal")
			}
			n, err := strconv.ParseUint(lit.Value, 0, 31)
			if err != nil || n == 0 {
				return fmt.Errorf("array length must be a positive integer literal")
			}
		}
		return c.validateExpr(t.Elt, seen, false)
	case *ast.MapType:
		if mapKey {
			return fmt.Errorf("map key cannot be a map")
		}
		if err := c.validateExpr(t.Key, seen, true); err != nil {
			return err
		}
		return c.validateExpr(t.Value, seen, false)
	default:
		return fmt.Errorf("unsupported type %s", c.ExprToString(e))
	}
}

func (c *Context) problem(pos token.Pos, pattern string, args ...any) error {
	loc := c.InputFile
	if c.FileSet != nil && pos.IsValid() {
		loc = c.FileSet.Position(pos).String()
	}
	return fmt.Errorf("%s: %s", loc, fmt.Sprintf(pattern, args...))
}

// OrderedTypes preserves field order while defining dependencies before users
// in C/C++, whose by-value fields require complete declarations.
func (c *Context) OrderedTypes() []*ast.TypeSpec {
	var ordered []*ast.TypeSpec
	visited := map[string]bool{}
	var visit func(*ast.TypeSpec)
	visit = func(ts *ast.TypeSpec) {
		if visited[ts.Name.Name] {
			return
		}
		visited[ts.Name.Name] = true
		ast.Inspect(ts.Type, func(n ast.Node) bool {
			id, ok := n.(*ast.Ident)
			if ok {
				if dep, yes := c.TypeSpecs[id.Name]; yes {
					visit(dep)
				}
			}
			return true
		})
		ordered = append(ordered, ts)
	}
	for _, ts := range c.Types {
		visit(ts)
	}
	return ordered
}
