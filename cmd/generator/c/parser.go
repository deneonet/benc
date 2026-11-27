package c

import (
	"fmt"
	"github.com/banditmoscow1337/benc/cmd/generator/common"
	"go/ast"
	"go/token"
	"os"
	"strconv"
	"strings"
	"text/scanner"
)

// Parse translates C typedef structs into Go AST without inferring ambiguous pointers.
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
	var s scanner.Scanner
	s.Init(strings.NewReader(string(data)))
	s.Mode = scanner.ScanIdents | scanner.ScanInts | scanner.ScanComments
	var scanErr error
	s.Error = func(_ *scanner.Scanner, msg string) {
		if scanErr == nil {
			scanErr = fmt.Errorf("%s: %s", ctx.InputFile, msg)
		}
	}
	var toks []cToken
	for tok := s.Scan(); tok != scanner.EOF; tok = s.Scan() {
		if tok != scanner.Comment {
			toks = append(toks, cToken{s.TokenText(), s.Pos().Line})
		}
	}
	if scanErr != nil {
		return scanErr
	}
	ctx.PkgName = "c_out"
	var declarations []*ast.TypeSpec
	for i := 0; i < len(toks); i++ {
		if toks[i].text != "typedef" || i+1 >= len(toks) || toks[i+1].text != "struct" {
			continue
		}
		i += 2
		if i < len(toks) && toks[i].text != "{" {
			i++
		}
		if i >= len(toks) || toks[i].text != "{" {
			return fmt.Errorf("%s: expected struct body", ctx.InputFile)
		}
		var fields [][]cToken
		for i++; i < len(toks) && toks[i].text != "}"; {
			start := i
			for i < len(toks) && toks[i].text != ";" && toks[i].text != "}" {
				i++
			}
			if i >= len(toks) || toks[i].text != ";" {
				return fmt.Errorf("%s:%d: expected field terminator", ctx.InputFile, toks[start].line)
			}
			fields = append(fields, toks[start:i])
			i++
		}
		if i+2 >= len(toks) || toks[i].text != "}" || toks[i+2].text != ";" {
			return fmt.Errorf("%s: unterminated typedef struct", ctx.InputFile)
		}
		name := toks[i+1].text
		nameLine := toks[i+1].line
		i += 2
		raw := make([]cDeclaration, 0, len(fields))
		for _, f := range fields {
			d, err := parseCDeclaration(f)
			if err != nil {
				return fmt.Errorf("%s:%d: %w", ctx.InputFile, f[0].line, err)
			}
			raw = append(raw, d)
		}
		converted, err := cFields(raw)
		if err != nil {
			return fmt.Errorf("%s: %s: %w", ctx.InputFile, name, err)
		}
		for _, f := range converted {
			for _, d := range raw {
				if d.name == f.Names[0].Name || d.name == f.Names[0].Name+"_keys" {
					f.Names[0].NamePos = astFile.Pos(lineStarts[d.line-1])
					break
				}
			}
		}
		ident := ast.NewIdent(name)
		ident.NamePos = astFile.Pos(lineStarts[nameLine-1])
		declarations = append(declarations, &ast.TypeSpec{Name: ident, Type: &ast.StructType{Fields: &ast.FieldList{List: converted}}})
	}
	helpers := map[string]ast.Expr{}
	for _, ts := range declarations {
		if !strings.HasPrefix(ts.Name.Name, "benc_generated_t") {
			continue
		}
		fields := ts.Type.(*ast.StructType).Fields.List
		if len(fields) != 1 {
			return fmt.Errorf("%s: malformed generated helper %s", ctx.InputFile, ts.Name.Name)
		}
		field := fields[0]
		if len(field.Names) != 1 {
			return fmt.Errorf("%s: malformed generated helper %s", ctx.InputFile, ts.Name.Name)
		}
		if field.Names[0].Name == "unix_nano" {
			helpers[ts.Name.Name] = &ast.SelectorExpr{X: ast.NewIdent("time"), Sel: ast.NewIdent("Time")}
		} else {
			helpers[ts.Name.Name] = field.Type
		}
	}
	var resolve func(ast.Expr, map[string]bool) (ast.Expr, error)
	resolve = func(e ast.Expr, seen map[string]bool) (ast.Expr, error) {
		switch t := e.(type) {
		case *ast.Ident:
			if h, ok := helpers[t.Name]; ok {
				if seen[t.Name] {
					return nil, fmt.Errorf("recursive C helper %s", t.Name)
				}
				seen[t.Name] = true
				defer delete(seen, t.Name)
				return resolve(h, seen)
			}
		case *ast.StarExpr:
			x, err := resolve(t.X, seen)
			if err != nil {
				return nil, err
			}
			return &ast.StarExpr{X: x}, nil
		case *ast.ArrayType:
			x, err := resolve(t.Elt, seen)
			if err != nil {
				return nil, err
			}
			return &ast.ArrayType{Len: t.Len, Elt: x}, nil
		case *ast.MapType:
			k, err := resolve(t.Key, seen)
			if err != nil {
				return nil, err
			}
			v, err := resolve(t.Value, seen)
			if err != nil {
				return nil, err
			}
			return &ast.MapType{Key: k, Value: v}, nil
		}
		return e, nil
	}
	for _, ts := range declarations {
		if strings.HasPrefix(ts.Name.Name, "benc_generated_t") {
			continue
		}
		st := ts.Type.(*ast.StructType)
		for _, f := range st.Fields.List {
			var err error
			f.Type, err = resolve(f.Type, map[string]bool{})
			if err != nil {
				return fmt.Errorf("%s: %s.%s: %w", ctx.InputFile, ts.Name.Name, f.Names[0].Name, err)
			}
		}
		if len(st.Fields.List) == 1 && st.Fields.List[0].Names[0].Name == "benc_value" {
			if _, ok := st.Fields.List[0].Type.(*ast.MapType); !ok {
				return fmt.Errorf("%s: malformed named map %s", ctx.InputFile, ts.Name.Name)
			}
			ts.Type = st.Fields.List[0].Type
		}
		ctx.Types = append(ctx.Types, ts)
	}
	return nil
}

type cToken struct {
	text string
	line int
}
type cDeclaration struct {
	typ, name string
	pointers  int
	length    string
	line      int
}

func parseCDeclaration(f []cToken) (cDeclaration, error) {
	d := cDeclaration{line: f[0].line}
	i := 0
	for i < len(f) && (f[i].text == "const" || f[i].text == "volatile") {
		i++
	}
	if i < len(f) && f[i].text == "struct" {
		i++
	}
	if i >= len(f) {
		return d, fmt.Errorf("missing C type")
	}
	d.typ = f[i].text
	i++
	if d.typ == "unsigned" || d.typ == "signed" || d.typ == "long" || d.typ == "short" {
		if i < len(f) && f[i].text != "*" {
			d.typ += " " + f[i].text
			i++
		}
	}
	for i < len(f) && f[i].text == "*" {
		d.pointers++
		i++
	}
	if i >= len(f) {
		return d, fmt.Errorf("missing field name")
	}
	d.name = f[i].text
	i++
	if i < len(f) && f[i].text == "[" {
		if i+2 >= len(f) || f[i+2].text != "]" {
			return d, fmt.Errorf("%s: only literal fixed arrays are supported", d.name)
		}
		d.length = f[i+1].text
		i += 3
	}
	if i != len(f) {
		return d, fmt.Errorf("%s: unsupported C declarator", d.name)
	}
	return d, nil
}

func cFields(raw []cDeclaration) ([]*ast.Field, error) {
	var out []*ast.Field
	for i := 0; i < len(raw); i++ {
		d := raw[i]
		if strings.HasSuffix(d.name, "_count") || strings.HasSuffix(d.name, "_values") || strings.HasSuffix(d.name, "_len") {
			return nil, fmt.Errorf("line %d: orphaned companion field %s", d.line, d.name)
		}
		var expr ast.Expr
		name := d.name
		if strings.HasSuffix(name, "_keys") && i+2 < len(raw) && raw[i+1].name == strings.TrimSuffix(name, "_keys")+"_values" && raw[i+2].name == strings.TrimSuffix(name, "_keys")+"_count" {
			v, n := raw[i+1], raw[i+2]
			if d.pointers != 1 || v.pointers != 1 || n.typ != "size_t" || n.pointers != 0 {
				return nil, fmt.Errorf("line %d: malformed map %s", d.line, name)
			}
			k, e := cType(d.typ, 0)
			if e != nil {
				return nil, e
			}
			val, e := cType(v.typ, 0)
			if e != nil {
				return nil, e
			}
			expr = &ast.MapType{Key: k, Value: val}
			name = strings.TrimSuffix(name, "_keys")
			i += 2
		} else if i+1 < len(raw) && raw[i+1].name == name+"_count" {
			count := raw[i+1]
			if d.pointers != 1 || count.typ != "size_t" || count.pointers != 0 {
				return nil, fmt.Errorf("line %d: malformed slice %s", d.line, name)
			}
			elt, e := cType(d.typ, 0)
			if e != nil {
				return nil, e
			}
			expr = &ast.ArrayType{Elt: elt}
			i++
		} else if i+1 < len(raw) && raw[i+1].name == name+"_len" {
			length := raw[i+1]
			if d.typ != "char" || d.pointers != 1 || length.typ != "size_t" || length.pointers != 0 {
				return nil, fmt.Errorf("line %d: malformed string %s", d.line, name)
			}
			expr = ast.NewIdent("string")
			i++
		} else if d.length != "" {
			n, e := strconv.ParseUint(d.length, 0, 31)
			if e != nil || n == 0 || d.pointers != 0 {
				return nil, fmt.Errorf("line %d: invalid fixed array %s", d.line, name)
			}
			elt, e := cType(d.typ, 0)
			if e != nil {
				return nil, e
			}
			expr = &ast.ArrayType{Len: &ast.BasicLit{Kind: token.INT, Value: strconv.FormatUint(n, 10)}, Elt: elt}
		} else {
			if strings.HasSuffix(name, "_keys") {
				return nil, fmt.Errorf("line %d: incomplete map %s", d.line, name)
			}
			var e error
			expr, e = cType(d.typ, d.pointers)
			if e != nil {
				return nil, fmt.Errorf("line %d: %w", d.line, e)
			}
		}
		out = append(out, &ast.Field{Names: []*ast.Ident{ast.NewIdent(name)}, Type: expr})
	}
	return out, nil
}

func cType(typ string, pointers int) (ast.Expr, error) {
	if typ == "char" && pointers == 1 {
		return nil, fmt.Errorf("char* requires matching size_t <field>_len for binary-safe strings")
	}
	if pointers > 1 {
		return nil, fmt.Errorf("unsupported pointer depth for %s", typ)
	}
	goName := map[string]string{"bool": "bool", "_Bool": "bool", "int8_t": "int8", "uint8_t": "uint8", "char": "int8", "int16_t": "int16", "short": "int16", "uint16_t": "uint16", "int32_t": "int32", "int": "int32", "uint32_t": "uint32", "unsigned int": "uint32", "int64_t": "int64", "long": "int64", "long long": "int64", "uint64_t": "uint64", "unsigned long": "uint64", "size_t": "uint64", "intptr_t": "int", "uintptr_t": "uint", "float": "float32", "double": "float64"}[typ]
	if goName == "" {
		goName = typ
	}
	var e ast.Expr = ast.NewIdent(goName)
	if pointers == 1 {
		e = &ast.StarExpr{X: e}
	}
	return e, nil
}
