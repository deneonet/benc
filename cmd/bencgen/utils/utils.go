package utils

import (
	"go.kine.bz/benc/cmd/bencgen/parser"
)

func FormatTypeGolang(t *parser.Type) string {
	if t.IsArray {
		return "[]" + FormatTypeGolang(t.Type)
	}
	if t.IsMap {
		return "map[" + FormatTypeGolang(t.Key) + "]" + FormatTypeGolang(t.Type)
	}

	if t.CtrName != "" {
		return t.CtrName
	}

	return t.UT.Golang()
}
