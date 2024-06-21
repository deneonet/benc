package codegens

import (
	"fmt"
	"slices"

	"go.kine.bz/benc/cmd/bencgen/lexer"
	"go.kine.bz/benc/cmd/bencgen/parser"
)

type GoGenerator struct {
	Generator

	file      string
	generated string
}

func (gg GoGenerator) File() string {
	return gg.file
}

func (gg GoGenerator) Lang() GeneratorLanguage {
	return GenGolang
}

func (gen GoGenerator) GenHeader(containsMaxSize bool, stmt *parser.HeaderStmt) string {
	gen.generated += "package " + stmt.Name + "\n\nimport (\n    "
	if containsMaxSize {
		gen.generated += "\"go.kine.bz/benc\"\n    "
	}
	return gen.generated + "\"go.kine.bz/benc/std\"\n    \"go.kine.bz/benc/impl/gen\"\n)\n\n"
}

func (gen GoGenerator) GenReservedIds(stmt *parser.CtrStmt) string {
	rl := ""
	li := len(stmt.ReservedIds) - 1

	for i, id := range stmt.ReservedIds {
		if i == li {
			rl += fmt.Sprintf("%d", id)
			continue
		}

		rl += fmt.Sprintf("%d, ", id)
	}

	return gen.generated + fmt.Sprintf("// Reserved Ids - %s\nvar %sRIds = []uint16{%s}\n\n", stmt.Name, toLower(stmt.Name), rl)
}

func (gen GoGenerator) GenStruct(ctrDeclarations []string, stmt *parser.CtrStmt) string {
	stmtName := toUpper(stmt.Name)
	gen.generated += fmt.Sprintf("// Struct - %s\ntype %s struct {\n    rIds []uint16\n\n", stmtName, stmtName)
	for _, field := range stmt.Fields {
		fieldName := toUpper(field.Name)

		array := ""
		if field.IsArray {
			array = "[]"
		}

		if field.Type == lexer.CTR {
			if !slices.Contains(ctrDeclarations, field.CtrName) {
				createError(gen, fmt.Sprintf("Container \"%s\" is not declared.", field.CtrName))
			}

			gen.generated += fmt.Sprintf("    %s %s%s\n", fieldName, array, toUpper(field.CtrName))
			continue
		}

		gen.generated += fmt.Sprintf("    %s %s%s\n", fieldName, array, field.Type.Golang())
	}
	return gen.generated + "}\n\n"
}

func (gen GoGenerator) GenSize(stmt *parser.CtrStmt) string {
	privStmtName := toLower(stmt.Name)
	publStmtName := toUpper(stmt.Name)

	gen.generated += fmt.Sprintf("// Size - %s\nfunc (%s *%s) Size(id uint16) (s int, err error) {\n    var ts int\n", stmt.Name, privStmtName, publStmtName)
	for _, field := range stmt.Fields {
		fieldName := toUpper(field.Name)
		typeString := field.Type.String()

		tagSize := 2
		if field.Id > 255 {
			tagSize = 3
		}

		switch field.Type {
		case lexer.CTR:
			gen.generated += fmt.Sprintf("    if ts, err = %s.%s.Size(%d); err != nil {\n        return\n    }\n    s += ts\n", privStmtName, fieldName, field.Id)
		case lexer.STRING, lexer.BYTES:
			gen.generated += fmt.Sprintf("    if ts, err = bstd.Size%s(%s.%s); err != nil {\n        return\n    }\n    s += ts + %d\n", typeString, privStmtName, fieldName, tagSize)
		default:
			gen.generated += fmt.Sprintf("    s += bstd.Size%s() + %d\n", typeString, tagSize)
		}
	}
	return gen.generated + "\n    _ = ts\n    if id > 255 {\n        s += 5\n        return\n    }\n    s += 4\n    return\n}\n\n"
}

func (gen GoGenerator) GenMarshal(stmt *parser.CtrStmt) string {
	privStmtName := toLower(stmt.Name)
	publStmtName := toUpper(stmt.Name)

	gen.generated += fmt.Sprintf("// Marshal - %s\nfunc (%s *%s) Marshal(tn int, b []byte, id uint16) (n int, err error) {\n    n = bgenimpl.MarshalTag(tn, b, bgenimpl.Container, id)\n", stmt.Name, privStmtName, publStmtName)

	for _, field := range stmt.Fields {
		fieldName := toUpper(field.Name)
		typeString := field.Type.String()

		if field.Type == lexer.CTR {
			gen.generated += fmt.Sprintf("    if n, err = %s.%s.Marshal(n, b, %d); err != nil {\n        return\n    }\n", privStmtName, fieldName, field.Id)
			continue
		}

		bgenimplType := "Fixed64"
		switch field.Type {
		case lexer.STRING, lexer.BYTES:
			bgenimplType = "Bytes"
			gen.generated += fmt.Sprintf("    n = bgenimpl.MarshalTag(n, b, bgenimpl.%s, %d)\n    if n, err = bstd.Marshal%s%s(n, b, %s.%s%s); err != nil {\n        return\n    }\n", bgenimplType, field.Id, field.GetUnsafeStr(), typeString, privStmtName, fieldName, field.GetMaxSizeStr())
			continue
		case lexer.BOOL, lexer.BYTE:
			bgenimplType = "Fixed8"
		case lexer.INT16, lexer.UINT16:
			bgenimplType = "Fixed16"
		case lexer.INT32, lexer.UINT32, lexer.FLOAT32:
			bgenimplType = "Fixed32"
		}

		gen.generated += fmt.Sprintf("    n = bgenimpl.MarshalTag(n, b, bgenimpl.%s, %d)\n    n = bstd.Marshal%s(n, b, %s.%s)\n", bgenimplType, field.Id, typeString, privStmtName, fieldName)

	}
	return gen.generated + "\n    n += 2\n    b[n-2] = 1\n    b[n-1] = 1\n    return\n}\n\n"
}

func (gen GoGenerator) GenNestedUnmarshal(stmt *parser.CtrStmt) string {
	privStmtName := toLower(stmt.Name)
	publStmtName := toUpper(stmt.Name)
	return fmt.Sprintf("// Nested Unmarshal - %s\nfunc (%s *%s) unmarshal(n int, b []byte, r []uint16, id uint16) (int, error) {\n    %s.rIds = r\n    return %s.Unmarshal(n, b, id)\n}\n\n", stmt.Name, privStmtName, publStmtName, privStmtName, privStmtName)
}

func (gen GoGenerator) GenUnmarshal(stmt *parser.CtrStmt) string {
	privStmtName := toLower(stmt.Name)
	publStmtName := toUpper(stmt.Name)

	gen.generated += fmt.Sprintf("// Unmarshal - %s\nfunc (%s *%s) Unmarshal(tn int, b []byte, id uint16) (n int, err error) {\n    var ok bool\n    if n, ok, err = bgenimpl.HandleCompatibility(tn, b, %s.rIds, id); !ok {\n        if err == bgenimpl.ErrEof {\n            return n, nil\n        }\n        return\n    }\n", stmt.Name, privStmtName, publStmtName, privStmtName)

	for _, field := range stmt.Fields {
		fieldName := toUpper(field.Name)
		typeString := field.Type.String()

		if field.Type == lexer.CTR {
			gen.generated += fmt.Sprintf("    if n, err = %s.%s.unmarshal(n, b, %sRIds, %d); err != nil {\n        return\n    }\n", privStmtName, fieldName, privStmtName, field.Id)
			continue
		}

		gen.generated += fmt.Sprintf("    if n, ok, err = bgenimpl.HandleCompatibility(n, b, %sRIds, %d); err != nil {\n        if err == bgenimpl.ErrEof {\n            return n, nil\n        }\n        return\n    }\n", privStmtName, field.Id)
		gen.generated += fmt.Sprintf("    if ok {\n        if n, %s.%s, err = bstd.Unmarshal%s%s(n, b); err != nil {\n            return\n        }\n    }\n", privStmtName, fieldName, field.GetUnsafeStr(), typeString)
	}
	return gen.generated + "    n += 2\n    return\n}\n\n"
}
