package parser

import (
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"

	"github.com/deneonet/benc/cmd/bencgen/lexer"
)

type Parser struct {
	lexer *lexer.Lexer
	token lexer.Token
	lit   string
	pos   lexer.Position
}

func NewParser(reader io.Reader, fileContent string) *Parser {
	return &Parser{
		lexer: lexer.NewLexer(reader, fileContent),
	}
}

func (p *Parser) nextToken() {
	pos, token, lit := p.lexer.Lex()
	p.pos = pos
	p.token = token
	p.lit = lit
}

func (p *Parser) match(expected lexer.Token) bool {
	return p.token == expected
}

func (p *Parser) matchAny(expected ...lexer.Token) bool {
	for _, token := range expected {
		if p.token == token {
			return true
		}
	}
	return false
}

func (p *Parser) expect(expected lexer.Token) {
	if !p.match(expected) {
		p.error(fmt.Sprintf("Unexpected token: `%s`. Expected: `%s`", p.token, expected))
	}
	p.nextToken()
}

func (p *Parser) Parse() []Node {
	var nodes []Node
	p.nextToken()
	for !p.match(lexer.EOF) {
		nodes = append(nodes, p.parseStatement())
	}
	return nodes
}

func (p *Parser) parseStatement() Node {
	switch {
	case p.match(lexer.CTR):
		return p.parseContainerStmt()
	case p.match(lexer.ENUM):
		return p.parseEnumStmt()
	case p.match(lexer.DEFINE):
		return p.parseDefineStmt()
	case p.match(lexer.VAR):
		return p.parseVarStmt()
	case p.match(lexer.USE):
		return p.parseUseStmt()
	default:
		p.error(fmt.Sprintf("Unexpected token: `%s`. Expected: `Container, Enum, Define, Use or Var`", p.token))
		return nil
	}
}

func (p *Parser) parseContainerStmt() Node {
	p.expect(lexer.CTR)
	containerName := p.lit
	p.expect(lexer.IDENT)
	p.errorIfContainsDot(containerName, "Container names")

	p.expect(lexer.OPEN_BRACE)

	reservedIDs := p.parseReservedIDs()
	fields := p.parseFields()

	p.expect(lexer.CLOSE_BRACE)
	return &ContainerStmt{Name: containerName, ReservedIDs: reservedIDs, Fields: fields}
}

func (p *Parser) parseEnumStmt() Node {
	p.expect(lexer.ENUM)
	enumName := p.lit
	p.expect(lexer.IDENT)
	p.errorIfContainsDot(enumName, "Enum names")

	p.expect(lexer.OPEN_BRACE)

	values := p.parseEnumValues()

	p.expect(lexer.CLOSE_BRACE)
	return &EnumStmt{Name: enumName, Values: values}
}

func (p *Parser) parseDefineStmt() Node {
	p.expect(lexer.DEFINE)
	definePackage := p.lit
	p.expect(lexer.IDENT)
	p.expect(lexer.SEMICOLON)
	return &DefineStmt{Package: definePackage}
}

func (p *Parser) parseVarStmt() Node {
	p.expect(lexer.VAR)
	name := p.lit
	p.expect(lexer.IDENT)
	p.errorIfContainsDot(name, "Var names")

	p.expect(lexer.EQUALS)
	value := p.lit
	p.expect(lexer.STR_VALUE)
	p.expect(lexer.SEMICOLON)
	return &VarStmt{Name: name, Value: value}
}

func (p *Parser) parseUseStmt() Node {
	p.expect(lexer.USE)
	path := p.lit
	p.expect(lexer.STR_VALUE)
	p.expect(lexer.SEMICOLON)
	return &UseStmt{Path: path}
}

func (p *Parser) parseEnumValues() []string {
	var values []string
	for !p.match(lexer.CLOSE_BRACE) {
		values = append(values, p.parseEnumValue())
	}
	return values
}

func (p *Parser) parseEnumValue() string {
	value := p.lit
	p.expect(lexer.IDENT)
	p.errorIfContainsDot(value, "Enum value names")

	if !p.match(lexer.CLOSE_BRACE) {
		p.expect(lexer.COMMA)
	}
	return value
}

func (p *Parser) parseReservedIDs() []uint16 {
	if p.match(lexer.RESERVED) {
		p.nextToken()
		return p.parseIDList()
	}
	return nil
}

func (p *Parser) parseIDList() []uint16 {
	var ids []uint16
	for {
		id, err := strconv.ParseUint(p.lit, 10, 16)
		p.expect(lexer.NUMBER)
		if err != nil {
			p.error("Error parsing reserved ID: " + err.Error())
		}
		ids = append(ids, uint16(id))

		if !p.match(lexer.COMMA) {
			break
		}
		p.nextToken()
	}
	p.expect(lexer.SEMICOLON)
	return ids
}

func (p *Parser) parseFields() []Field {
	var fields []Field
	for !p.match(lexer.CLOSE_BRACE) {
		fields = append(fields, p.parseField())
	}
	return fields
}

func (p *Parser) parseField() Field {
	fieldType := p.expectType()

	fieldName := p.lit
	p.expect(lexer.IDENT)
	p.errorIfContainsDot(fieldName, "Container field names")

	p.expect(lexer.EQUALS)

	id, err := strconv.ParseUint(p.lit, 10, 16)
	p.expect(lexer.NUMBER)
	if err != nil {
		p.error("Error parsing field ID: " + err.Error())
	}

	p.expect(lexer.SEMICOLON)
	return Field{ID: uint16(id), Name: fieldName, Type: fieldType}
}

func (p *Parser) expectType() *Type {
	switch {
	case p.match(lexer.OPEN_BRACKET):
		p.nextToken()
		p.expect(lexer.CLOSE_BRACKET)
		return &Type{IsArray: true, ChildType: p.expectType()}

	case p.match(lexer.OPEN_ARROW):
		p.nextToken()
		keyType := p.expectType()
		p.expect(lexer.COMMA)
		valueType := p.expectType()
		p.expect(lexer.CLOSE_ARROW)
		return &Type{IsMap: true, MapKeyType: keyType, ChildType: valueType}

	case p.match(lexer.IDENT):
		ctrName := p.lit
		p.nextToken()
		return &Type{ExternalStructure: ctrName}

	case p.match(lexer.UNSAFE):
		p.nextToken()
		tokenType := p.token
		p.nextToken()
		if tokenType != lexer.STRING {
			p.error("`unsafe` can only be applied to `string` types")
		}
		return &Type{IsUnsafe: true, TokenType: tokenType}

	case p.match(lexer.RCOPY):
		p.nextToken()
		tokenType := p.token
		p.nextToken()
		if tokenType != lexer.BYTES {
			p.error("`rcopy` can only be applied to `bytes` types")
		}
		return &Type{IsReturnCopy: true, TokenType: tokenType}

	default:
		if p.matchAny(lexer.STRING, lexer.BYTES, lexer.INT, lexer.INT16, lexer.INT32, lexer.INT64, lexer.UINT, lexer.UINT16, lexer.UINT32, lexer.UINT64, lexer.FLOAT32, lexer.FLOAT64, lexer.BYTE, lexer.BOOL) {
			tokenType := p.token
			p.nextToken()
			return &Type{TokenType: tokenType}
		}
		p.error("Unexpected token, expected a type")
		return nil
	}
}

func (p *Parser) error(message string) {
	errorMessage := "\n\033[1;31m[bencgen] Error:\033[0m\n"
	errorMessage += fmt.Sprintf("    \033[1;37m%d:%d\033[0m %s\n", p.pos.Line, p.pos.Column, highlightError(p.lexer.Content, p.pos.Line, p.pos.Column))
	errorMessage += fmt.Sprintf("    \033[1;37mMessage:\033[0m %s\n", message)
	fmt.Println(errorMessage)
	os.Exit(-1)
}

func highlightError(text string, lineNumber, columnNumber int) string {
	lines := strings.Split(text, "\n")
	if lineNumber <= 0 || lineNumber > len(lines) {
		return "Invalid line number <- report"
	}

	line := lines[lineNumber-1]
	if columnNumber <= 0 || columnNumber > len(line) {
		return "Invalid column number <- report"
	}

	highlightedLine := fmt.Sprintf("%s\033[1;31m%c\033[0m%s", line[:columnNumber], line[columnNumber], line[columnNumber+1:])
	arrow := strings.Repeat(" ", columnNumber-1+6+len(fmt.Sprintf("%d:%d", lineNumber, columnNumber))) + "\033[1;31m^\033[0m"
	return highlightedLine + "\n" + arrow
}

func (p *Parser) errorIfContainsDot(ident string, context string) {
	if strings.Contains(ident, ".") {
		p.error(context + " may not contain a dot '.'")
	}
}

type Node any

type (
	Type struct {
		TokenType         lexer.Token
		MapKeyType        *Type
		ChildType         *Type
		ExternalStructure string `json:"ctrName"`
		IsUnsafe          bool
		IsReturnCopy      bool
		IsArray           bool
		IsMap             bool
	}
	ContainerStmt struct {
		Name        string
		Fields      []Field
		ReservedIDs []uint16
	}
	EnumStmt struct {
		Name   string
		Values []string
	}
	DefineStmt struct {
		Package string
	}
	VarStmt struct {
		Name  string
		Value string
	}
	UseStmt struct {
		Path string
	}
	Field struct {
		ID   uint16 `json:"id"`
		Name string
		Type *Type
	}
)

func (t *Type) IsAnExternalStructure() bool {
	return t.ExternalStructure != ""
}

func (t *Type) AppendUnsafeIfPresent() string {
	if t.IsUnsafe {
		return "Unsafe"
	}
	return ""
}

func (t *Type) AppendReturnCopyIfPresent() string {
	if t.IsReturnCopy {
		return "Copied"
	}
	if t.TokenType == lexer.BYTES {
		return "Cropped"
	}
	return ""
}
