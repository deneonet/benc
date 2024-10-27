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

func NewParser(reader io.Reader, fc string) *Parser {
	return &Parser{
		lexer: lexer.NewLexer(reader, fc),
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

func (p *Parser) mMatch(expected ...lexer.Token) bool {
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

func (p *Parser) expectType() *Type {
	switch {
	case p.match(lexer.OPEN_BRACKET):
		p.nextToken()
		p.expect(lexer.CLOSE_BRACKET)
		return &Type{IsArray: true, ChildType: p.expectType()}

	case p.match(lexer.OPEN_ARROW):
		p.nextToken()
		key := p.expectType()
		p.expect(lexer.COMMA)
		val := p.expectType()
		p.expect(lexer.CLOSE_ARROW)
		return &Type{IsMap: true, MapKeyType: key, ChildType: val}

	case p.match(lexer.IDENT):
		ctrName := p.lit
		p.nextToken()
		return &Type{CtrName: ctrName}

	case p.match(lexer.UNSAFE):
		p.nextToken()
		typ := p.token
		p.nextToken()
		if typ != lexer.STRING {
			p.error("`@unsafe` can only be applied to `string` types")
		}
		return &Type{IsUnsafe: true, TokenType: typ}

	default:
		if p.mMatch(lexer.STRING, lexer.BYTES, lexer.INT, lexer.INT16, lexer.INT32, lexer.INT64, lexer.UINT, lexer.UINT16, lexer.UINT32, lexer.UINT64, lexer.FLOAT32, lexer.FLOAT64, lexer.BYTE, lexer.BOOL) {
			typ := p.token
			p.nextToken()
			return &Type{TokenType: typ}
		}
		p.error("Unexpected token, expected a type")
		return nil
	}
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

func (p *Parser) error(m string) {
	errorMessage := "\n\033[1;31m[bencgen] Error:\033[0m\n"
	errorMessage += fmt.Sprintf("    \033[1;37m%d:%d\033[0m %s\n", p.pos.Line, p.pos.Column, highlightError(p.lexer.Content, p.pos.Line, p.pos.Column))
	errorMessage += fmt.Sprintf("    \033[1;37mMessage:\033[0m %s\n", m)
	fmt.Println(errorMessage)
	os.Exit(-1)
}

type Node interface{}

type (
	Type struct {
		TokenType  lexer.Token
		MapKeyType *Type
		ChildType  *Type
		CtrName    string
		IsUnsafe   bool
		IsArray    bool
		IsMap      bool
	}
	HeaderStmt struct {
		Name string
	}
	CtrStmt struct {
		Name        string
		Fields      []Field
		ReservedIds []uint16
	}
	Field struct {
		Id   uint16
		Name string
		Type *Type
	}
)

func (f *Field) GetUnsafeStr() string {
	if f.Type.IsUnsafe {
		return "Unsafe"
	}
	return ""
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
	case p.match(lexer.HEADER):
		return p.parseHeaderStmt()
	case p.match(lexer.CTR):
		return p.parseCtrStmt()
	default:
		p.error(fmt.Sprintf("Unexpected token: `%s`. Expected: `Container or Header`", p.token))
		return nil
	}
}

func (p *Parser) parseHeaderStmt() Node {
	p.expect(lexer.HEADER)
	headerName := p.lit
	p.expect(lexer.IDENT)
	p.expect(lexer.SEMICOLON)
	return &HeaderStmt{Name: headerName}
}

func (p *Parser) parseCtrStmt() Node {
	p.expect(lexer.CTR)
	name := p.lit
	p.expect(lexer.IDENT)
	p.expect(lexer.OPEN_BRACE)

	reservedIds := p.parseReservedIds()
	fields := p.parseFields()

	p.expect(lexer.CLOSE_BRACE)
	return &CtrStmt{Name: name, ReservedIds: reservedIds, Fields: fields}
}

func (p *Parser) parseReservedIds() []uint16 {
	if p.match(lexer.RESERVED) {
		p.nextToken()
		return p.parseIdList()
	}
	return nil
}

func (p *Parser) parseIdList() []uint16 {
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
	t := p.expectType()

	n := p.lit
	p.expect(lexer.IDENT)
	p.expect(lexer.EQUALS)

	id, err := strconv.ParseUint(p.lit, 10, 16)
	p.expect(lexer.NUMBER)
	if err != nil {
		p.error("Error parsing field ID: " + err.Error())
	}

	p.expect(lexer.SEMICOLON)
	return Field{Id: uint16(id), Name: n, Type: t}
}
