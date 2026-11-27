package cpp

import (
	"fmt"
	"go/ast"
	"go/token"
	"os"
	"regexp"
	"strconv"
	"strings"
	"text/scanner"

	"github.com/banditmoscow1337/benc/cmd/generator/common"
)

// Parse accepts the deliberately small C++20 declaration subset emitted by
// this package as well as ordinary structs using the same standard types.
// Types are translated directly into Go AST; method bodies are not a model.
func Parse(ctx *common.Context) error {
	data, err := os.ReadFile(ctx.InputFile)
	if err != nil {
		return err
	}
	ctx.FileSet = token.NewFileSet()
	f := ctx.FileSet.AddFile(ctx.InputFile, -1, len(data))
	var lex scanner.Scanner
	lex.Init(strings.NewReader(string(data)))
	lex.Mode = scanner.ScanIdents | scanner.ScanInts | scanner.ScanStrings | scanner.ScanComments
	var lexErr error
	lex.Error = func(_ *scanner.Scanner, msg string) { lexErr = fmt.Errorf("%s: %s", ctx.InputFile, msg) }
	var tokens []cppToken
	for tok := lex.Scan(); tok != scanner.EOF; tok = lex.Scan() {
		if tok != scanner.Comment {
			tokens = append(tokens, cppToken{lex.TokenText(), f.Pos(lex.Pos().Offset), lex.Pos().Line})
		}
	}
	if lexErr != nil {
		return lexErr
	}
	ctx.PkgName = regexp.MustCompile(`[^a-zA-Z0-9_]`).ReplaceAllString(ctx.BaseName, "_")
	if ctx.PkgName == "" || ctx.PkgName[0] >= '0' && ctx.PkgName[0] <= '9' {
		ctx.PkgName = "cpp_" + ctx.PkgName
	}
	for i := 0; i < len(tokens); i++ {
		if tokens[i].text == "namespace" && i+2 < len(tokens) && tokens[i+2].text == "{" {
			ctx.PkgName = tokens[i+1].text
		}
		if tokens[i].text != "struct" || i+2 >= len(tokens) {
			continue
		}
		name := tokens[i+1]
		if !token.IsIdentifier(name.text) {
			return fmt.Errorf("%s:%d: expected struct name", ctx.InputFile, name.line)
		}
		j := i + 2
		var base []cppToken
		if tokens[j].text == ":" {
			j++
			if j < len(tokens) && tokens[j].text == "public" {
				j++
			}
			start := j
			for j < len(tokens) && tokens[j].text != "{" && tokens[j].text != ";" {
				j++
			}
			base = tokens[start:j]
		}
		if j >= len(tokens) || tokens[j].text == ";" {
			continue // forward declaration
		}
		if tokens[j].text != "{" {
			return fmt.Errorf("%s:%d: malformed struct %s", ctx.InputFile, name.line, name.text)
		}
		var fields []*ast.Field
		depth, start := 1, j+1
		for j++; j < len(tokens) && depth > 0; j++ {
			switch tokens[j].text {
			case "{":
				depth++
				if depth == 2 {
					start = j + 1
				} // method body
			case "}":
				depth--
				if depth == 1 {
					start = j + 1
				}
			case ";":
				if depth != 1 {
					continue
				}
				part := tokens[start:j]
				start = j + 1
				if len(part) == 0 || cppHas(part, "(") || cppHas(part, "=") || part[0].text == "using" {
					continue // generated methods and type aliases, never data fields
				}
				if len(part) == 2 && part[1].text == ":" {
					continue
				}
				if len(part) < 2 || !token.IsIdentifier(part[len(part)-1].text) {
					return fmt.Errorf("%s:%d: unsupported C++ field", ctx.InputFile, part[0].line)
				}
				e, eErr := cppType(part[:len(part)-1])
				if eErr != nil {
					return fmt.Errorf("%s:%d: %s.%s: %w", ctx.InputFile, part[0].line, name.text, part[len(part)-1].text, eErr)
				}
				id := ast.NewIdent(part[len(part)-1].text)
				id.NamePos = part[len(part)-1].pos
				fields = append(fields, &ast.Field{Names: []*ast.Ident{id}, Type: e})
			}
		}
		if depth != 0 {
			return fmt.Errorf("%s:%d: unterminated struct %s", ctx.InputFile, name.line, name.text)
		}
		id := ast.NewIdent(name.text)
		id.NamePos = name.pos
		ts := &ast.TypeSpec{Name: id}
		if len(base) != 0 {
			e, eErr := cppType(base)
			if eErr != nil {
				return fmt.Errorf("%s:%d: %s: %w", ctx.InputFile, name.line, name.text, eErr)
			}
			if _, ok := e.(*ast.MapType); !ok {
				return fmt.Errorf("%s:%d: only std::map may be a named base", ctx.InputFile, name.line)
			}
			ts.Type = e
		} else {
			ts.Type = &ast.StructType{Fields: &ast.FieldList{List: fields}}
		}
		ctx.Types = append(ctx.Types, ts)
		i = j
	}
	return nil
}

type cppToken struct {
	text string
	pos  token.Pos
	line int
}

func cppHas(tokens []cppToken, value string) bool {
	for _, t := range tokens {
		if t.text == value {
			return true
		}
	}
	return false
}

func cppType(tokens []cppToken) (ast.Expr, error) {
	p := &cppTypes{tokens: tokens}
	e, err := p.parse()
	if err != nil {
		return nil, err
	}
	if p.i != len(tokens) {
		return nil, fmt.Errorf("unexpected token %q in type", tokens[p.i].text)
	}
	return e, nil
}

type cppTypes struct {
	tokens []cppToken
	i      int
}

func (p *cppTypes) take(value string) bool {
	if p.i < len(p.tokens) && p.tokens[p.i].text == value {
		p.i++
		return true
	}
	return false
}
func (p *cppTypes) parse() (ast.Expr, error) {
	if p.i >= len(p.tokens) {
		return nil, fmt.Errorf("missing C++ type")
	}
	name := p.tokens[p.i].text
	p.i++
	for p.take(":") {
		if !p.take(":") || p.i >= len(p.tokens) {
			return nil, fmt.Errorf("invalid qualified C++ type")
		}
		name += "::" + p.tokens[p.i].text
		p.i++
	}
	if p.take("<") {
		first, err := p.parse()
		if err != nil {
			return nil, err
		}
		switch name {
		case "std::vector":
			if !p.take(">") {
				return nil, fmt.Errorf("expected > in std::vector")
			}
			return &ast.ArrayType{Elt: first}, nil
		case "std::optional":
			if !p.take(">") {
				return nil, fmt.Errorf("expected > in std::optional")
			}
			return &ast.StarExpr{X: first}, nil
		case "std::map":
			if !p.take(",") {
				return nil, fmt.Errorf("expected comma in std::map")
			}
			second, e := p.parse()
			if e != nil {
				return nil, e
			}
			if !p.take(">") {
				return nil, fmt.Errorf("expected > in std::map")
			}
			return &ast.MapType{Key: first, Value: second}, nil
		case "std::array":
			if !p.take(",") || p.i >= len(p.tokens) {
				return nil, fmt.Errorf("expected length in std::array")
			}
			length := p.tokens[p.i].text
			if _, e := strconv.ParseUint(length, 0, 31); e != nil {
				return nil, fmt.Errorf("invalid std::array length %q", length)
			}
			p.i++
			if !p.take(">") {
				return nil, fmt.Errorf("expected > in std::array")
			}
			return &ast.ArrayType{Len: &ast.BasicLit{Kind: token.INT, Value: length}, Elt: first}, nil
		default:
			return nil, fmt.Errorf("unsupported C++ template %s", name)
		}
	}
	primitives := map[string]string{
		"bool": "bool", "std::string": "string", "std::byte": "uint8", "bstd::time_point": "time.Time",
		"intptr_t": "int", "uintptr_t": "uint", "int8_t": "int8", "uint8_t": "uint8",
		"int16_t": "int16", "uint16_t": "uint16", "int32_t": "int32", "uint32_t": "uint32",
		"int64_t": "int64", "uint64_t": "uint64", "float": "float32", "double": "float64",
	}
	if goName, ok := primitives[name]; ok {
		if goName == "time.Time" {
			return &ast.SelectorExpr{X: ast.NewIdent("time"), Sel: ast.NewIdent("Time")}, nil
		}
		return ast.NewIdent(goName), nil
	}
	if strings.Contains(name, "::") {
		return nil, fmt.Errorf("unsupported C++ type %s", name)
	}
	return ast.NewIdent(name), nil
}
