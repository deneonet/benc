package codegens

import (
	"fmt"
	"strings"

	"github.com/deneonet/benc/cmd/bencgen/lexer"
	"github.com/deneonet/benc/cmd/bencgen/parser"
	"github.com/deneonet/benc/cmd/bencgen/utils"
)

type GoGenerator struct {
	file string
}

func NewGoGenerator(file string) *GoGenerator {
	return &GoGenerator{file: file}
}

func (gg GoGenerator) File() string {
	return gg.file
}

func (gg GoGenerator) Lang() GeneratorLanguage {
	return GenGolang
}

func getSizeFunc(name string, field parser.Field, plain bool) string {
	fieldName := utils.ToUpper(field.Name)
	fieldType := field.Type
	switch {
	case field.Type.IsArray:
		return fmt.Sprintf("bstd.SizeSlice(%s.%s, %s)", name, fieldName, getElemSizeFunc(fieldType.ChildType))
	case field.Type.IsMap:
		return fmt.Sprintf("bstd.SizeMap(%s.%s, %s, %s)", name, fieldName, getElemSizeFunc(fieldType.MapKeyType), getElemSizeFunc(fieldType.ChildType))
	case field.Type.CtrName != "":
		if plain {
			return fmt.Sprintf("%s.%s.SizePlain()", name, fieldName)
		}
		return fmt.Sprintf("%s.%s.size(%d)", name, fieldName, field.Id)
	default:
		switch fieldType.TokenType {
		case lexer.STRING, lexer.BYTES, lexer.INT, lexer.UINT:
			return fmt.Sprintf("bstd.Size%s(%s.%s)", fieldType.TokenType.String(), name, fieldName)
		}
		return fmt.Sprintf("bstd.Size%s()", fieldType.TokenType.String())
	}
}

func getElemSizeFunc(t *parser.Type) string {
	switch {
	case t.IsArray:
		return fmt.Sprintf("func (s %s) int { return bstd.SizeSlice(s, %s) }", utils.BencTypeToGolang(t), getElemSizeFunc(t.ChildType))
	case t.IsMap:
		return fmt.Sprintf("func (s %s) int { return bstd.SizeMap(s, %s, %s) }", utils.BencTypeToGolang(t), getElemSizeFunc(t.MapKeyType), getElemSizeFunc(t.ChildType))
	case t.CtrName != "":
		return fmt.Sprintf("func (s %s) int { return s.SizePlain() }", utils.ToUpper(t.CtrName))
	default:
		return "bstd.Size" + t.TokenType.String()
	}
}

func getMarshalFunc(name string, field parser.Field, plain bool) string {
	fieldName := utils.ToUpper(field.Name)
	switch {
	case field.Type.IsArray:
		return fmt.Sprintf("bstd.MarshalSlice(n, b, %s.%s, %s)", name, fieldName, getElemMarshalFunc(field.Type.ChildType))
	case field.Type.IsMap:
		return fmt.Sprintf("bstd.MarshalMap(n, b, %s.%s, %s, %s)", name, fieldName, getElemMarshalFunc(field.Type.MapKeyType), getElemMarshalFunc(field.Type.ChildType))
	case field.Type.CtrName != "":
		if plain {
			return fmt.Sprintf("%s.%s.MarshalPlain(n, b)", name, fieldName)
		}
		return fmt.Sprintf("%s.%s.marshal(n, b, %d)", name, fieldName, field.Id)
	default:
		return fmt.Sprintf("bstd.Marshal%s%s(n, b, %s.%s)", field.GetUnsafeStr(), field.Type.TokenType.String(), name, fieldName)
	}
}

func getElemMarshalFunc(t *parser.Type) string {
	switch {
	case t.IsArray:
		return fmt.Sprintf("func (n int, b []byte, s %s) int { return bstd.MarshalSlice(n, b, s, %s) }", utils.BencTypeToGolang(t), getElemMarshalFunc(t.ChildType))
	case t.IsMap:
		return fmt.Sprintf("func (n int, b []byte, s %s) int { return bstd.MarshalMap(n, b, s, %s, %s) }", utils.BencTypeToGolang(t), getElemMarshalFunc(t.MapKeyType), getElemMarshalFunc(t.ChildType))
	case t.CtrName != "":
		return fmt.Sprintf("func (n int, b []byte, s %s) int { return s.MarshalPlain(n, b) }", utils.ToUpper(t.CtrName))
	default:
		return "bstd.Marshal" + t.GetUnsafeStr() + t.TokenType.String()
	}
}

func getUnmarshalFunc(name string, field parser.Field, plain bool) string {
	fieldName := utils.ToUpper(field.Name)
	switch {
	case field.Type.IsArray:
		return fmt.Sprintf("bstd.UnmarshalSlice[%s](n, b, %s)", utils.BencTypeToGolang(field.Type.ChildType), getElemUnmarshalFunc(field.Type.ChildType))
	case field.Type.IsMap:
		return fmt.Sprintf("bstd.UnmarshalMap[%s, %s](n, b, %s, %s)", utils.BencTypeToGolang(field.Type.MapKeyType), utils.BencTypeToGolang(field.Type.ChildType), getElemUnmarshalFunc(field.Type.MapKeyType), getElemUnmarshalFunc(field.Type.ChildType))
	case field.Type.CtrName != "" && plain:
		return fmt.Sprintf("%s.%s.UnmarshalPlain(n, b)", name, fieldName)
	default:
		return fmt.Sprintf("bstd.Unmarshal%s%s(n, b)", field.GetUnsafeStr(), field.Type.TokenType.String())
	}
}

func getElemUnmarshalFunc(t *parser.Type) string {
	switch {
	case t.IsArray:
		return fmt.Sprintf("func (n int, b []byte) (int, %s, error) { return bstd.UnmarshalSlice[%s](n, b, %s) }", utils.BencTypeToGolang(t), utils.BencTypeToGolang(t.ChildType), getElemUnmarshalFunc(t.ChildType))
	case t.IsMap:
		return fmt.Sprintf("func (n int, b []byte) (int, %s, error) { return bstd.UnmarshalMap[%s, %s](n, b, %s, %s) }", utils.BencTypeToGolang(t), utils.BencTypeToGolang(t.MapKeyType), utils.BencTypeToGolang(t.ChildType), getElemUnmarshalFunc(t.MapKeyType), getElemUnmarshalFunc(t.ChildType))
	case t.CtrName != "":
		return fmt.Sprintf("func (n int, b []byte, s *%s) (int, error) { return s.UnmarshalPlain(n, b) }", utils.ToUpper(t.CtrName))
	default:
		return "bstd.Unmarshal" + t.GetUnsafeStr() + t.TokenType.String()
	}
}

func (gen GoGenerator) GenHeader(stmt *parser.HeaderStmt) string {
	return fmt.Sprintf(
		`package %s

import (
    "github.com/deneonet/benc/std"
    "github.com/deneonet/benc/impl/gen"
)

`, stmt.Name)
}

func (gen GoGenerator) GenReservedIds(stmt *parser.CtrStmt) string {
	return fmt.Sprintf("// Reserved Ids - %s\nvar %sRIds = []uint16{%s}\n\n",
		stmt.Name, utils.ToLower(stmt.Name), joinUint16(stmt.ReservedIds))
}

func joinUint16(ids []uint16) string {
	var sb strings.Builder
	for i, id := range ids {
		if i > 0 {
			sb.WriteString(", ")
		}
		sb.WriteString(fmt.Sprintf("%d", id))
	}
	return sb.String()
}

func (gen GoGenerator) GenStruct(ctrDeclarations []string, stmt *parser.CtrStmt) string {
	var sb strings.Builder
	stmtName := utils.ToUpper(stmt.Name)
	sb.WriteString(fmt.Sprintf("// Struct - %s\ntype %s struct {\n", stmtName, stmtName))
	for _, field := range stmt.Fields {
		sb.WriteString(fmt.Sprintf("    %s %s\n", utils.ToUpper(field.Name), utils.BencTypeToGolang(field.Type)))
	}
	sb.WriteString("}\n\n")
	return sb.String()
}

func (gen GoGenerator) GenSize(stmt *parser.CtrStmt) string {
	privStmtName := utils.ToLower(stmt.Name)
	publStmtName := utils.ToUpper(stmt.Name)

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("// Size - %s\nfunc (%s *%s) Size() int {\n    return %s.size(0)\n}\n\n", stmt.Name, privStmtName, publStmtName, privStmtName))
	sb.WriteString(fmt.Sprintf("// Nested Size - %s\nfunc (%s *%s) size(id uint16) (s int) {\n", stmt.Name, privStmtName, publStmtName))

	for _, field := range stmt.Fields {
		tagSize := 2
		if field.Id > 255 {
			tagSize = 3
		}

		sb.WriteString(fmt.Sprintf("    s += %s", getSizeFunc(privStmtName, field, false)))
		if field.Type.CtrName == "" {
			sb.WriteString(fmt.Sprintf(" + %d\n", tagSize))
		} else {
			sb.WriteString("\n")
		}
	}

	sb.WriteString("\n    if id > 255 {\n        s += 5\n        return\n    }\n    s += 4\n    return\n}\n\n")
	return sb.String()
}

func (gen GoGenerator) GenSizePlain(stmt *parser.CtrStmt) string {
	privStmtName := utils.ToLower(stmt.Name)
	publStmtName := utils.ToUpper(stmt.Name)

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("// SizePlain - %s\nfunc (%s *%s) SizePlain() (s int) {\n", stmt.Name, privStmtName, publStmtName))
	for _, field := range stmt.Fields {
		sb.WriteString(fmt.Sprintf("    s += %s\n", getSizeFunc(privStmtName, field, true)))
	}
	sb.WriteString("    return\n}\n\n")
	return sb.String()
}

func (gen GoGenerator) GenMarshal(stmt *parser.CtrStmt) string {
	privStmtName := utils.ToLower(stmt.Name)
	publStmtName := utils.ToUpper(stmt.Name)

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("// Marshal - %s\nfunc (%s *%s) Marshal(b []byte) {\n    %s.marshal(0, b, 0)\n}\n\n", stmt.Name, privStmtName, publStmtName, privStmtName))
	sb.WriteString(fmt.Sprintf("// Nested Marshal - %s\nfunc (%s *%s) marshal(tn int, b []byte, id uint16) (n int) {\n    n = bgenimpl.MarshalTag(tn, b, bgenimpl.Container, id)\n", stmt.Name, privStmtName, publStmtName))

	for _, field := range stmt.Fields {
		if field.Type.CtrName == "" {
			sb.WriteString(fmt.Sprintf("    n = bgenimpl.MarshalTag(n, b, bgenimpl.%s, %d)\n", mapTokenTypeToBgenimplType(field.Type.TokenType), field.Id))
		}
		sb.WriteString(fmt.Sprintf("    n = %s\n", getMarshalFunc(privStmtName, field, false)))
	}

	sb.WriteString("\n    n += 2\n    b[n-2] = 1\n    b[n-1] = 1\n    return\n}\n\n")
	return sb.String()
}

func mapTokenTypeToBgenimplType(t lexer.Token) string {
	switch t {
	case lexer.INT, lexer.UINT:
		return "Varint"
	case lexer.STRING, lexer.BYTES:
		return "Bytes"
	case lexer.BOOL, lexer.BYTE:
		return "Fixed8"
	case lexer.INT16, lexer.UINT16:
		return "Fixed16"
	case lexer.INT32, lexer.UINT32, lexer.FLOAT32:
		return "Fixed32"
	case lexer.INT64, lexer.UINT64, lexer.FLOAT64:
		return "Fixed64"
	default:
		return "ArrayMap"
	}
}

func (gen GoGenerator) GenMarshalPlain(stmt *parser.CtrStmt) string {
	privStmtName := utils.ToLower(stmt.Name)
	publStmtName := utils.ToUpper(stmt.Name)

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("// MarshalPlain - %s\nfunc (%s *%s) MarshalPlain(tn int, b []byte) (n int) {\n    n = tn\n", stmt.Name, privStmtName, publStmtName))
	for _, field := range stmt.Fields {
		sb.WriteString(fmt.Sprintf("    n = %s\n", getMarshalFunc(privStmtName, field, true)))
	}
	sb.WriteString("    return n\n}\n\n")
	return sb.String()
}

func (gen GoGenerator) GenUnmarshal(stmt *parser.CtrStmt) string {
	privStmtName := utils.ToLower(stmt.Name)
	publStmtName := utils.ToUpper(stmt.Name)

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("// Unmarshal - %s\nfunc (%s *%s) Unmarshal(b []byte) (err error) {\n    _, err = %s.unmarshal(0, b, []uint16{}, 0)\n    return\n}\n\n", stmt.Name, privStmtName, publStmtName, privStmtName))
	sb.WriteString(fmt.Sprintf("// Nested Unmarshal - %s\nfunc (%s *%s) unmarshal(tn int, b []byte, r []uint16, id uint16) (n int, err error) {\n    var ok bool\n    if n, ok, err = bgenimpl.HandleCompatibility(tn, b, r, id); !ok {\n        if err == bgenimpl.ErrEof {\n            return n, nil\n        }\n        return\n    }\n", stmt.Name, privStmtName, publStmtName))

	for _, field := range stmt.Fields {
		fieldName := utils.ToUpper(field.Name)
		if field.Type.CtrName != "" {
			sb.WriteString(fmt.Sprintf("    if n, err = %s.%s.unmarshal(n, b, %sRIds, %d); err != nil {\n        return\n    }\n", privStmtName, fieldName, privStmtName, field.Id))
			continue
		}
		sb.WriteString(fmt.Sprintf("    if n, ok, err = bgenimpl.HandleCompatibility(n, b, %sRIds, %d); err != nil {\n        if err == bgenimpl.ErrEof {\n            return n, nil\n        }\n        return\n    }\n", privStmtName, field.Id))
		sb.WriteString(fmt.Sprintf("    if ok {\n        if n, %s.%s, err = %s; err != nil {\n            return\n        }\n    }\n", privStmtName, fieldName, getUnmarshalFunc(privStmtName, field, false)))
	}

	sb.WriteString("    n += 2\n    return\n}\n\n")
	return sb.String()
}

func (gen GoGenerator) GenUnmarshalPlain(stmt *parser.CtrStmt) string {
	privStmtName := utils.ToLower(stmt.Name)
	publStmtName := utils.ToUpper(stmt.Name)

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("// UnmarshalPlain - %s\nfunc (%s *%s) UnmarshalPlain(tn int, b []byte) (n int, err error) {\n    n = tn\n", stmt.Name, privStmtName, publStmtName))
	for _, field := range stmt.Fields {
		if field.Type.CtrName != "" {
			sb.WriteString(fmt.Sprintf("    if n, err = %s; err != nil {\n        return\n    }\n", getUnmarshalFunc(privStmtName, field, true)))
			continue
		}

		sb.WriteString(fmt.Sprintf("    if n, %s.%s, err = %s; err != nil {\n        return\n    }\n", privStmtName, utils.ToUpper(field.Name), getUnmarshalFunc(privStmtName, field, true)))
	}
	sb.WriteString("    return\n}\n\n")
	return sb.String()
}
