package utils

import (
	"slices"
	"strings"

	"github.com/deneonet/benc/cmd/bencgen/parser"
)

func ToUpper(s string) string {
	return strings.ToUpper(s[:1]) + s[1:]
}

func ToLower(s string) string {
	return strings.ToLower(s[:1]) + s[1:]
}

func FormatType(t *parser.Type) string {
	return formatTypeHelper(t, false)
}

func BencTypeToGolang(t *parser.Type) string {
	return formatTypeHelper(t, true)
}

func formatTypeHelper(t *parser.Type, useGoFormat bool) string {
	if t.IsArray {
		return "[]" + formatTypeHelper(t.ChildType, useGoFormat)
	}
	if t.IsMap {
		keyFormat := formatTypeHelper(t.MapKeyType, useGoFormat)
		valueFormat := formatTypeHelper(t.ChildType, useGoFormat)
		if useGoFormat {
			return "map[" + keyFormat + "]" + valueFormat
		}
		return "<" + keyFormat + ", " + valueFormat + ">"
	}

	if t.CtrName != "" {
		return t.CtrName
	}

	if useGoFormat {
		return t.TokenType.Golang()
	}
	return t.TokenType.String()
}

func CompareTypes(t1 *parser.Type, t2 *parser.Type) bool {
	return compareTypes(t1, t2)
}

func FindUndeclaredContainers(ctrDeclarations []string, t *parser.Type) (string, bool) {
	if t.CtrName != "" && !slices.Contains(ctrDeclarations, t.CtrName) {
		return t.CtrName, true
	}

	if t.ChildType != nil {
		if ctr, notFound := FindUndeclaredContainers(ctrDeclarations, t.ChildType); notFound {
			return ctr, true
		}
	}

	if t.MapKeyType != nil {
		if ctr, notFound := FindUndeclaredContainers(ctrDeclarations, t.MapKeyType); notFound {
			return ctr, true
		}
	}

	return "", false
}

func compareTypes(t1 *parser.Type, t2 *parser.Type) bool {
	if t1 == nil || t2 == nil {
		return t1 == t2
	}

	return t1.IsArray == t2.IsArray &&
		t1.IsMap == t2.IsMap &&
		t1.TokenType == t2.TokenType &&
		t1.CtrName == t2.CtrName &&
		compareTypes(t1.MapKeyType, t2.MapKeyType) &&
		compareTypes(t1.ChildType, t2.ChildType)
}
