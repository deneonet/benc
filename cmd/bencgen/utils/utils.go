package utils

import (
	"go.kine.bz/benc/cmd/bencgen/parser"
)

func FormatType(t *parser.Type) string {
	if t.IsArray {
		return "[]" + FormatType(t.Type)
	}

	if t.CtrName != "" {
		return t.CtrName
	}

	return t.UT.Golang()
}
