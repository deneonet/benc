package lexer

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strings"
	"unicode"
)

type Token int

const (
	EOF = iota
	ILLEGAL
	NUMBER
	IDENT

	USE      // use ...
	VAR      // var ...
	CTR      // container ...
	ENUM     // enum ...
	DEFINE   // define ...
	RESERVED // reserved ...

	STR_VALUE // "..."

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

	UNSAFE // unsafe
	RCOPY  // return copy
	// type attributes

	OPEN_BRACKET  // [
	CLOSE_BRACKET // ]
	OPEN_BRACE    // {
	CLOSE_BRACE   // }

	OPEN_ARROW  // <
	CLOSE_ARROW // >

	COMMA     // ,
	EQUALS    // =
	SEMICOLON // ;
)

var tokens = []string{
	EOF:       "EOF",
	ILLEGAL:   "Illegal",
	NUMBER:    "Number",
	DEFINE:    "Define",
	RESERVED:  "Reserved",
	VAR:       "Var",
	USE:       "Use",
	IDENT:     "Identifier",
	STR_VALUE: "String Value",
	CTR:       "Container",
	ENUM:      "Enum",

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

	RCOPY:  "ReturnCopy",
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
	"define":   DEFINE,
	"var":      VAR,
	"enum":     ENUM,
	"use":      USE,
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

	"unsafe": UNSAFE,
	"rcopy":  RCOPY,
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

func (l *Lexer) error(message string) {
	errorMessage := "\n\033[1;31m[bencgen] Error:\033[0m\n"
	errorMessage += fmt.Sprintf("    \033[1;37m%d:%d\033[0m %s\n", l.pos.Line, l.pos.Column, highlightError(l.Content, l.pos.Line, l.pos.Column))
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
		case '"':
			var sb strings.Builder
			for {
				r, _, err := l.reader.ReadRune()
				if r == '\n' {
					l.error("String isn't valid (no end, expected: \"...\").")
				}
				if err != nil || r == '"' {
					break
				}
				l.pos.Column++
				sb.WriteRune(r)
			}
			return l.pos, STR_VALUE, sb.String()
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
		if err != nil || !(unicode.IsLetter(r) || unicode.IsDigit(r) || r == '.' || r == '_') {
			l.backup()
			break
		}
		l.pos.Column++
		sb.WriteRune(r)
	}
	return sb.String()
}
