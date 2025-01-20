package lexer

import (
	"bufio"
	"io"
	"strings"
	"unicode"
)

type Token int

const (
	EOF = iota
	ILLEGAL
	NUMBER
	IDENT

	HEADER   // header ...
	RESERVED // reserved ...

	// types
	INT64
	INT32
	INT16
	INT

	UINT64
	UINT32
	UINT16
	UINT

	BYTES
	STRING

	FLOAT64
	FLOAT32

	BOOL
	BYTE
	// types

	UNSAFE // @unsafe
	// type attributes

	OPEN_BRACKET  // [
	CLOSE_BRACKET // ]
	OPEN_BRACE    // {
	CLOSE_BRACE   // }

	OPEN_ARROW  // <
	CLOSE_ARROW // >

	CTR       // container
	COMMA     // ,
	EQUALS    // =
	SEMICOLON // ;
)

var tokens = []string{
	EOF:      "EOF",
	ILLEGAL:  "Illegal",
	NUMBER:   "Number",
	HEADER:   "Header",
	RESERVED: "Reserved",
	IDENT:    "Identifier",
	CTR:      "Container",

	INT64: "Int64",
	INT32: "Int32",
	INT16: "Int16",
	INT:   "Int",

	UINT64: "Uint64",
	UINT32: "Uint32",
	UINT16: "Uint16",
	UINT:   "Uint",

	FLOAT32: "Float32",
	FLOAT64: "Float64",

	BYTE: "Byte",
	BOOL: "Bool",

	BYTES:  "Bytes",
	STRING: "String",

	UNSAFE: "Unsafe",

	OPEN_BRACKET:  "[",
	CLOSE_BRACKET: "]",
	OPEN_BRACE:    "{",
	CLOSE_BRACE:   "}",

	OPEN_ARROW:  "<",
	CLOSE_ARROW: ">",

	COMMA:     ",",
	EQUALS:    "=",
	SEMICOLON: ";",
}

var keywords = map[string]Token{
	"reserved": RESERVED,
	"header":   HEADER,
	"ctr":      CTR,

	"int64": INT64,
	"int32": INT32,
	"int16": INT16,
	"int":   INT,

	"uint64": UINT64,
	"uint32": UINT32,
	"uint16": UINT16,
	"uint":   UINT,

	"float32": FLOAT32,
	"float64": FLOAT64,

	"bool": BOOL,
	"byte": BYTE,

	"bytes":  BYTES,
	"string": STRING,

	"@unsafe": UNSAFE,
}

func (t Token) String() string {
	return tokens[t]
}

func (t Token) Golang() string {
	switch t {
	case INT64:
		return "int64"
	case INT32:
		return "int32"
	case INT16:
		return "int16"
	case INT:
		return "int"
	case UINT64:
		return "uint64"
	case UINT32:
		return "uint32"
	case UINT16:
		return "uint16"
	case UINT:
		return "uint"
	case FLOAT32:
		return "float32"
	case FLOAT64:
		return "float64"
	case BYTE:
		return "byte"
	case BOOL:
		return "bool"
	case BYTES:
		return "[]byte"
	case STRING:
		return "string"
	}
	return "invalid type"
}

type Position struct {
	Line   int
	Column int
}

type Lexer struct {
	pos     Position
	reader  *bufio.Reader
	Content string
}

func NewLexer(reader io.Reader, content string) *Lexer {
	return &Lexer{
		Content: content,
		pos:     Position{Line: 1, Column: 0},
		reader:  bufio.NewReader(reader),
	}
}

func (l *Lexer) Lex() (Position, Token, string) {
	comment := false

	for {
		r, _, err := l.reader.ReadRune()
		if err != nil {
			if err == io.EOF {
				return l.pos, EOF, ""
			}

			panic(err)
		}

		l.pos.Column++

		if r == '\n' {
			comment = false
		}

		if comment {
			continue
		}

		switch r {
		case '\n':
			l.resetPosition()
		case '#':
			comment = true
			continue
		case '[':
			return l.pos, OPEN_BRACKET, "["
		case ']':
			return l.pos, CLOSE_BRACKET, "]"
		case '{':
			return l.pos, OPEN_BRACE, "{"
		case '}':
			return l.pos, CLOSE_BRACE, "}"
		case '<':
			return l.pos, OPEN_ARROW, "<"
		case '>':
			return l.pos, CLOSE_ARROW, ">"
		case ',':
			return l.pos, COMMA, ","
		case '=':
			return l.pos, EQUALS, "="
		case ';':
			return l.pos, SEMICOLON, ";"
		default:
			if unicode.IsSpace(r) {
				continue
			}

			if unicode.IsDigit(r) {
				startPos := l.pos
				l.backup()
				lit := l.lexNumber()
				return startPos, NUMBER, lit
			}

			if unicode.IsLetter(r) || r == '_' {
				startPos := l.pos
				l.backup()
				lit := l.lexIdent()
				if token, ok := keywords[lit]; ok {
					return startPos, token, lit
				}
				return startPos, IDENT, lit
			}

			if r == '@' {
				startPos := l.pos
				lit := l.lexAtPrefixedIdent()
				if token, ok := keywords[lit]; ok {
					return startPos, token, lit
				}
				return startPos, ILLEGAL, lit
			}

			return l.pos, ILLEGAL, string(r)
		}
	}
}

func (l *Lexer) resetPosition() {
	l.pos.Line++
	l.pos.Column = 0
}

func (l *Lexer) backup() {
	if err := l.reader.UnreadRune(); err != nil {
		panic(err)
	}
	l.pos.Column--
}

func (l *Lexer) lexNumber() string {
	var sb strings.Builder
	for {
		r, _, err := l.reader.ReadRune()
		if err != nil || !unicode.IsDigit(r) {
			l.backup()
			break
		}
		l.pos.Column++
		sb.WriteRune(r)
	}
	return sb.String()
}

func (l *Lexer) lexIdent() string {
	var sb strings.Builder
	for {
		r, _, err := l.reader.ReadRune()
		if err != nil || !(unicode.IsLetter(r) || unicode.IsDigit(r) || r == '_') {
			l.backup()
			break
		}
		l.pos.Column++
		sb.WriteRune(r)
	}
	return sb.String()
}

func (l *Lexer) lexAtPrefixedIdent() string {
	var sb strings.Builder
	sb.WriteRune('@')
	for {
		r, _, err := l.reader.ReadRune()
		if err != nil || !(unicode.IsLetter(r) || unicode.IsDigit(r) || r == '_') {
			l.backup()
			break
		}
		l.pos.Column++
		sb.WriteRune(r)
	}
	return sb.String()
}
