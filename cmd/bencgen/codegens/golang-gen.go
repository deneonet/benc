package codegens

import (
	"fmt"

	"go.kine.bz/benc/cmd/bencgen/lexer"
	"go.kine.bz/benc/cmd/bencgen/parser"
	"go.kine.bz/benc/cmd/bencgen/utils"
)

type GoGenerator struct {
	Generator

	file      string
	generated string
}

func getSizeFunc(name string, field parser.Field, plain bool) string {
	fieldName := toUpper(field.Name)

	if field.Type.IsArray {
		return fmt.Sprintf("bstd.SizeSlice(%s.%s, %s)", name, fieldName, getElemSizeFunc(field.Type.Type))
	}
	if field.Type.IsMap {
		return fmt.Sprintf("bstd.SizeMap(%s.%s, %s, %s)", name, fieldName, getElemSizeFunc(field.Type.Key), getElemSizeFunc(field.Type.Type))
	}

	if field.Type.CtrName != "" {
		if plain {
			return fmt.Sprintf("%s.%s.SizePlain()", name, fieldName)
		}
		return fmt.Sprintf("%s.%s.size(%d)", name, fieldName, field.Id)
	}

	typeString := field.Type.UT.String()
	switch field.Type.UT {
	case lexer.STRING, lexer.BYTES:
		return fmt.Sprintf("bstd.Size%s(%s.%s)", typeString, name, fieldName)
	default:
		return fmt.Sprintf("bstd.Size%s()", typeString)
	}
}

func getElemSizeFunc(t *parser.Type) string {
	if t.IsArray {
		return fmt.Sprintf("func (s %s) int { return bstd.SizeSlice(s, %s) }", utils.FormatTypeGolang(t), getElemSizeFunc(t.Type))
	}
	if t.IsMap {
		return fmt.Sprintf("func (s %s) int { return bstd.SizeMap(s, %s, %s) }", utils.FormatTypeGolang(t), getElemSizeFunc(t.Key), getElemSizeFunc(t.Type))
	}

	if t.CtrName != "" {
		return fmt.Sprintf("func (s %s) int { return s.SizePlain() }", toUpper(t.CtrName))
	}
	return "bstd.Size" + t.UT.String()
}

func getMarshalFunc(name string, field parser.Field, plain bool) string {
	fieldName := toUpper(field.Name)

	if field.Type.IsArray {
		return fmt.Sprintf("bstd.MarshalSlice(n, b, %s.%s, %s)", name, fieldName, getElemMarshalFunc(field.Type.Type))
	}
	if field.Type.IsMap {
		return fmt.Sprintf("bstd.MarshalMap(n, b, %s.%s, %s, %s)", name, fieldName, getElemMarshalFunc(field.Type.Key), getElemMarshalFunc(field.Type.Type))
	}

	if field.Type.CtrName != "" {
		if plain {
			return fmt.Sprintf("%s.%s.MarshalPlain(n, b)", name, fieldName)
		}
		return fmt.Sprintf("%s.%s.marshal(n, b, %d)", name, fieldName, field.Id)
	}

	typeString := field.Type.UT.String()
	return fmt.Sprintf("bstd.Marshal%s(n, b, %s.%s)", typeString, name, fieldName)
}

func getElemMarshalFunc(t *parser.Type) string {
	if t.IsArray {
		return fmt.Sprintf("func (n int, b []byte, s %s) int { return bstd.MarshalSlice(n, b, s, %s) }", utils.FormatTypeGolang(t), getElemMarshalFunc(t.Type))
	}
	if t.IsMap {
		return fmt.Sprintf("func (n int, b []byte, s %s) int { return bstd.MarshalMap(n, b, s, %s, %s) }", utils.FormatTypeGolang(t), getElemMarshalFunc(t.Key), getElemMarshalFunc(t.Type))
	}

	if t.CtrName != "" {
		return fmt.Sprintf("func (n int, b []byte, s %s) int { return s.MarshalPlain(n, b) }", toUpper(t.CtrName))
	}
	return "bstd.Marshal" + t.UT.String()
}

func getUnmarshalFunc(name string, field parser.Field, plain bool) string {
	fieldName := toUpper(field.Name)

	if field.Type.IsArray {
		return fmt.Sprintf("bstd.UnmarshalSlice[%s](n, b, %s)", utils.FormatTypeGolang(field.Type.Type), getElemUnmarshalFunc(field.Type.Type))
	}
	if field.Type.IsMap {
		return fmt.Sprintf("bstd.UnmarshalMap[%s, %s](n, b, %s, %s)", utils.FormatTypeGolang(field.Type.Key), utils.FormatTypeGolang(field.Type.Type), getElemUnmarshalFunc(field.Type.Key), getElemUnmarshalFunc(field.Type.Type))
	}

	if field.Type.CtrName != "" && plain {
		return fmt.Sprintf("%s.%s.UnmarshalPlain(n, b)", name, fieldName)
	}

	typeString := field.Type.UT.String()
	return fmt.Sprintf("bstd.Unmarshal%s(n, b)", typeString)
}

func getElemUnmarshalFunc(t *parser.Type) string {
	if t.IsArray {
		return fmt.Sprintf("func (n int, b []byte) (int, %s, error) { return bstd.UnmarshalSlice[%s](n, b, %s) }", utils.FormatTypeGolang(t), utils.FormatTypeGolang(t.Type), getElemUnmarshalFunc(t.Type))
	}
	if t.IsMap {
		return fmt.Sprintf("func (n int, b []byte) (int, %s, error) { return bstd.UnmarshalMap[%s, %s](n, b, %s, %s) }", utils.FormatTypeGolang(t), utils.FormatTypeGolang(t.Key), utils.FormatTypeGolang(t.Type), getElemUnmarshalFunc(t.Key), getElemUnmarshalFunc(t.Type))
	}

	if t.CtrName != "" {
		return fmt.Sprintf("func (n int, b []byte, s *%s) (int, error) { return s.UnmarshalPlain(n, b) }", toUpper(t.CtrName))
	}
	return "bstd.Unmarshal" + t.UT.String()
}

func (gg GoGenerator) File() string {
	return gg.file
}

func (gg GoGenerator) Lang() GeneratorLanguage {
	return GenGolang
}

func (gen GoGenerator) GenHeader(stmt *parser.HeaderStmt) string {
	return "package " + stmt.Name + "\n\nimport (\n    \"go.kine.bz/benc/std\"\n    \"go.kine.bz/benc/impl/gen\"\n)\n\n"
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
	gen.generated += fmt.Sprintf("// Struct - %s\ntype %s struct {\n", stmtName, stmtName)
	for _, field := range stmt.Fields {
		fieldName := toUpper(field.Name)

		/*
			TODO: do ctr check

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
			}*/

		gen.generated += fmt.Sprintf("    %s %s\n", fieldName, utils.FormatTypeGolang(field.Type))
	}
	return gen.generated + "}\n\n"
}

func (gen GoGenerator) GenSize(stmt *parser.CtrStmt) string {
	privStmtName := toLower(stmt.Name)
	publStmtName := toUpper(stmt.Name)

	gen.generated += fmt.Sprintf("// Size - %s\nfunc (%s *%s) Size() int {\n    return %s.size(0)\n}\n\n", stmt.Name, privStmtName, publStmtName, privStmtName)
	gen.generated += fmt.Sprintf("// Nested Size - %s\nfunc (%s *%s) size(id uint16) (s int) {\n", stmt.Name, privStmtName, publStmtName)

	for _, field := range stmt.Fields {
		tagSize := 2
		if field.Id > 255 {
			tagSize = 3
		}

		t := getSizeFunc(privStmtName, field, false)
		gen.generated += fmt.Sprintf("    s += %s", t)
		if field.Type.CtrName == "" {
			gen.generated += fmt.Sprintf(" + %d\n", tagSize)
			continue
		}
		gen.generated += "\n"
	}
	return gen.generated + "\n    if id > 255 {\n        s += 5\n        return\n    }\n    s += 4\n    return\n}\n\n"
}

func (gen GoGenerator) GenSizePlain(stmt *parser.CtrStmt) string {
	privStmtName := toLower(stmt.Name)
	publStmtName := toUpper(stmt.Name)

	gen.generated += fmt.Sprintf("// SizePlain - %s\nfunc (%s *%s) SizePlain() (s int) {\n", stmt.Name, privStmtName, publStmtName)
	for _, field := range stmt.Fields {
		t := getSizeFunc(privStmtName, field, true)
		gen.generated += fmt.Sprintf("    s += %s\n", t)
	}
	return gen.generated + "    return\n}\n\n"
}

func (gen GoGenerator) GenMarshal(stmt *parser.CtrStmt) string {
	privStmtName := toLower(stmt.Name)
	publStmtName := toUpper(stmt.Name)

	gen.generated += fmt.Sprintf("// Marshal - %s\nfunc (%s *%s) Marshal(b []byte) {\n    %s.marshal(0, b, 0)\n}\n\n", stmt.Name, privStmtName, publStmtName, privStmtName)
	gen.generated += fmt.Sprintf("// Nested Marshal - %s\nfunc (%s *%s) marshal(tn int, b []byte, id uint16) (n int) {\n    n = bgenimpl.MarshalTag(tn, b, bgenimpl.Container, id)\n", stmt.Name, privStmtName, publStmtName)

	for _, field := range stmt.Fields {
		bgenimplType := "Array"
		switch field.Type.UT {
		case lexer.STRING, lexer.BYTES:
			bgenimplType = "Bytes"
		case lexer.BOOL, lexer.BYTE:
			bgenimplType = "Fixed8"
		case lexer.INT16, lexer.UINT16:
			bgenimplType = "Fixed16"
		case lexer.INT32, lexer.UINT32, lexer.FLOAT32:
			bgenimplType = "Fixed32"
		case lexer.INT64, lexer.UINT64, lexer.FLOAT64:
			bgenimplType = "Fixed64"
		}

		if field.Type.CtrName == "" {
			gen.generated += fmt.Sprintf("    n = bgenimpl.MarshalTag(n, b, bgenimpl.%s, %d)\n", bgenimplType, field.Id)
		}
		gen.generated += fmt.Sprintf("    n = %s\n", getMarshalFunc(privStmtName, field, false))

	}
	return gen.generated + "\n    n += 2\n    b[n-2] = 1\n    b[n-1] = 1\n    return\n}\n\n"
}

func (gen GoGenerator) GenMarshalPlain(stmt *parser.CtrStmt) string {
	privStmtName := toLower(stmt.Name)
	publStmtName := toUpper(stmt.Name)

	gen.generated += fmt.Sprintf("// MarshalPlain - %s\nfunc (%s *%s) MarshalPlain(tn int, b []byte) (n int) {\n    n = tn\n", stmt.Name, privStmtName, publStmtName)

	for _, field := range stmt.Fields {
		gen.generated += fmt.Sprintf("    n = %s\n", getMarshalFunc(privStmtName, field, true))

	}
	return gen.generated + "    return n\n}\n\n"
}

func (gen GoGenerator) GenUnmarshal(stmt *parser.CtrStmt) string {
	privStmtName := toLower(stmt.Name)
	publStmtName := toUpper(stmt.Name)
	gen.generated += fmt.Sprintf("// Unmarshal - %s\nfunc (%s *%s) Unmarshal(b []byte) (err error) {\n    _, err = %s.unmarshal(0, b, []uint16{}, 0)\n    return\n}\n\n", stmt.Name, privStmtName, publStmtName, privStmtName)
	gen.generated += fmt.Sprintf("// Nested Unmarshal - %s\nfunc (%s *%s) unmarshal(tn int, b []byte, r []uint16, id uint16) (n int, err error) {\n    var ok bool\n    if n, ok, err = bgenimpl.HandleCompatibility(tn, b, r, id); !ok {\n        if err == bgenimpl.ErrEof {\n            return n, nil\n        }\n        return\n    }\n", stmt.Name, privStmtName, publStmtName)

	for _, field := range stmt.Fields {
		fieldName := toUpper(field.Name)
		gen.generated += fmt.Sprintf("    if n, ok, err = bgenimpl.HandleCompatibility(n, b, %sRIds, %d); err != nil {\n        if err == bgenimpl.ErrEof {\n            return n, nil\n        }\n        return\n    }\n", privStmtName, field.Id)
		if field.Type.CtrName != "" {
			gen.generated += fmt.Sprintf("    if ok {\n        if n, err = %s.%s.unmarshal(n, b, %sRIds, %d); err != nil {\n            return\n        }\n    }\n", privStmtName, fieldName, privStmtName, field.Id)
			continue
		}
		t := getUnmarshalFunc(privStmtName, field, false)
		gen.generated += fmt.Sprintf("    if ok {\n        if n, %s.%s, err = %s; err != nil {\n            return\n        }\n    }\n", privStmtName, fieldName, t)
	}
	return gen.generated + "    n += 2\n    return\n}\n\n"
}

func (gen GoGenerator) GenUnmarshalPlain(stmt *parser.CtrStmt) string {
	privStmtName := toLower(stmt.Name)
	publStmtName := toUpper(stmt.Name)

	gen.generated += fmt.Sprintf("// UnmarshalPlain - %s\nfunc (%s *%s) UnmarshalPlain(tn int, b []byte) (n int, err error) {\n    n = tn\n", stmt.Name, privStmtName, publStmtName)

	for _, field := range stmt.Fields {
		fieldName := toUpper(field.Name)
		t := getUnmarshalFunc(privStmtName, field, true)
		if field.Type.CtrName != "" {
			gen.generated += fmt.Sprintf("    if n, err = %s; err != nil {\n        return\n    }\n", t)
			continue
		}
		gen.generated += fmt.Sprintf("    if n, %s.%s, err = %s; err != nil {\n        return\n    }\n", privStmtName, fieldName, t)
	}
	return gen.generated + "    return\n}\n\n"
}
