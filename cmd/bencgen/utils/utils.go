package utils

import (
	"strings"

	"go.kine.bz/benc/cmd/bencgen/lexer"
	"go.kine.bz/benc/cmd/bencgen/parser"
)

func FormatType(f parser.Field) string {
	array := ""
	if f.IsArray {
		array = "[]"
	}

	if f.Type == lexer.CTR {
		return array + f.CtrName
	}

	return array + strings.ToLower(f.Type.String())
}
