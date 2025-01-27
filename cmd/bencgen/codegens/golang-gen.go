package codegens

import (
	"fmt"
	"slices"
	"strings"

	"github.com/deneonet/benc/cmd/bencgen/lexer"
	"github.com/deneonet/benc/cmd/bencgen/parser"
	"github.com/deneonet/benc/cmd/bencgen/utils"
)

type GoCtrStmt struct {
	PublicName  string
	PrivateName string

	DefaultName string

	Fields      []parser.Field
	ReservedIds []uint16
}

type GoEnumStmt struct {
	PublicName  string
	PrivateName string

	DefaultName string

	Fields []string
}

type GoField struct {
	Id uint16

	PublicName  string
	PrivateName string
	DefaultName string

	Type *parser.Type
}

func (f *GoField) AppendUnsafeIfPresent() string {
	return f.Type.AppendUnsafeIfPresent()
}

type GoGen struct {
	file string

	ctrDecls  []string
	enumDecls []string

	plainGen   bool
	headerStmt *parser.HeaderStmt

	// currently generated...

	field    GoField
	ctrStmt  GoCtrStmt
	enumStmt GoEnumStmt
}

func NewGoGen(file string) *GoGen {
	return &GoGen{file: file}
}

func (g *GoGen) File() string {
	return g.file
}

func (*GoGen) Lang() GenLang {
	return GoGenLang
}

func (g *GoGen) IsCtr(extStructureName string) bool {
	return slices.Contains(g.ctrDecls, extStructureName)
}

func (g *GoGen) IsEnum(extStructureName string) bool {
	return slices.Contains(g.enumDecls, extStructureName)
}

func (g *GoGen) ForEachCtrFields(f func(i int)) {
	for i, field := range g.ctrStmt.Fields {
		g.field = GoField{
			Id: field.Id,

			PublicName:  utils.ToUpper(field.Name),
			PrivateName: utils.ToLower(field.Name),
			DefaultName: field.Name,

			Type: field.Type,
		}
		f(i)
	}
}

func (g *GoGen) ForEachEnumFields(f func(i int, field string)) {
	for i, field := range g.enumStmt.Fields {
		f(i, utils.ToUpper(field))
	}
}

func (g *GoGen) HasHeader() bool {
	return g.headerStmt != nil
}

func (g *GoGen) SetCtrStatement(stmt *parser.CtrStmt) {
	g.ctrStmt = GoCtrStmt{
		PublicName:  utils.ToUpper(stmt.Name),
		PrivateName: utils.ToLower(stmt.Name),

		DefaultName: stmt.Name,
		Fields:      stmt.Fields,
		ReservedIds: stmt.ReservedIds,
	}
}

func (g *GoGen) SetEnumStatement(stmt *parser.EnumStmt) {
	g.enumStmt = GoEnumStmt{
		PublicName:  utils.ToUpper(stmt.Name),
		PrivateName: utils.ToLower(stmt.Name),

		DefaultName: stmt.Name,
		Fields:      stmt.Fields,
	}
}

func (g *GoGen) SetHeaderStatement(stmt *parser.HeaderStmt) {
	g.headerStmt = stmt
}

func (g *GoGen) SetCtrDecls(ctrDecls []string) {
	g.ctrDecls = ctrDecls
}

func (g *GoGen) SetEnumDecls(enumDecls []string) {
	g.enumDecls = enumDecls
}

func (g *GoGen) GenHeader() string {
	return fmt.Sprintf(
		`package %s

import (
    "github.com/deneonet/benc/std"
    "github.com/deneonet/benc/impl/gen"
)

`, g.headerStmt.Name)
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

func (g *GoGen) GenReservedIds() string {
	ctr := g.ctrStmt
	return fmt.Sprintf("// Reserved Ids - %s\nvar %sRIds = []uint16{%s}\n\n",
		ctr.DefaultName, ctr.PrivateName, joinUint16(ctr.ReservedIds))
}

func (g *GoGen) GenStruct() string {
	var sb strings.Builder
	ctr := g.ctrStmt

	sb.WriteString(fmt.Sprintf("// Struct - %s\ntype %s struct {\n",
		ctr.DefaultName, ctr.PublicName))

	g.ForEachCtrFields(func(i int) {
		field := g.field
		sb.WriteString(fmt.Sprintf("    %s %s\n",
			field.PublicName, utils.BencTypeToGolang(field.Type)))
	})

	sb.WriteString("}\n\n")
	return sb.String()
}

func (g *GoGen) GenEnum() string {
	var sb strings.Builder
	enum := g.enumStmt

	sb.WriteString(fmt.Sprintf("// Enum - %s\ntype %s int\nconst (\n",
		enum.DefaultName, enum.PublicName))

	g.ForEachEnumFields(func(i int, field string) {
		if i == 0 {
			sb.WriteString(fmt.Sprintf("    %s%s %s = iota\n",
				enum.PublicName, field, enum.PublicName))
			return
		}
		sb.WriteString(fmt.Sprintf("    %s%s\n",
			enum.PublicName, field))
	})

	sb.WriteString(")\n\n")
	return sb.String()
}

func (g *GoGen) getSizeFunc() string {
	ctr := g.ctrStmt
	field := g.field

	switch {
	case field.Type.IsArray:
		return fmt.Sprintf("bstd.SizeSlice(%s.%s, %s)",
			ctr.PrivateName, field.PublicName, g.getElemSizeFunc(field.Type.ChildType))
	case field.Type.IsMap:
		return fmt.Sprintf("bstd.SizeMap(%s.%s, %s, %s)",
			ctr.PrivateName, field.PublicName, g.getElemSizeFunc(field.Type.MapKeyType), g.getElemSizeFunc(field.Type.ChildType))
	case field.Type.ExtStructure != "":
		if g.IsEnum(field.Type.ExtStructure) {
			return fmt.Sprintf("bgenimpl.SizeEnum(%s.%s)",
				ctr.PrivateName, field.PublicName)
		}

		if g.plainGen {
			return fmt.Sprintf("%s.%s.SizePlain()",
				ctr.PrivateName, field.PublicName)
		}
		return fmt.Sprintf("%s.%s.size(%d)",
			ctr.PrivateName, field.PublicName, field.Id)
	default:
		switch field.Type.TokenType {
		case lexer.STRING, lexer.BYTES, lexer.INT, lexer.UINT:
			return fmt.Sprintf("bstd.Size%s(%s.%s)",
				field.Type.TokenType.String(), ctr.PrivateName, field.PublicName)
		}

		return fmt.Sprintf("bstd.Size%s()",
			field.Type.TokenType.String())
	}
}

func (g *GoGen) getElemSizeFunc(t *parser.Type) string {
	switch {
	case t.IsArray:
		return fmt.Sprintf("func (s %s) int { return bstd.SizeSlice(s, %s) }",
			utils.BencTypeToGolang(t), g.getElemSizeFunc(t.ChildType))
	case t.IsMap:
		return fmt.Sprintf("func (s %s) int { return bstd.SizeMap(s, %s, %s) }",
			utils.BencTypeToGolang(t), g.getElemSizeFunc(t.MapKeyType), g.getElemSizeFunc(t.ChildType))
	case t.ExtStructure != "":
		if g.IsEnum(t.ExtStructure) {
			return "bgenimpl.SizeEnum"
		}

		return fmt.Sprintf("func (s %s) int { return s.SizePlain() }",
			utils.ToUpper(t.ExtStructure))
	default:
		return "bstd.Size" + t.TokenType.String()
	}
}

func (g *GoGen) GenSize() string {
	var sb strings.Builder
	ctr := g.ctrStmt

	sb.WriteString(fmt.Sprintf("// Size - %s\nfunc (%s *%s) Size() int {\n    return %s.size(0)\n}\n\n",
		ctr.DefaultName, ctr.PrivateName, ctr.PublicName, ctr.PrivateName))

	sb.WriteString(fmt.Sprintf("// Nested Size - %s\nfunc (%s *%s) size(id uint16) (s int) {\n",
		ctr.DefaultName, ctr.PrivateName, ctr.PublicName))

	g.ForEachCtrFields(func(_ int) {
		field := g.field

		tagSize := 2
		if g.field.Id > 255 {
			tagSize = 3
		}

		sb.WriteString(fmt.Sprintf("    s += %s", g.getSizeFunc()))

		if !g.IsCtr(field.Type.ExtStructure) {
			sb.WriteString(fmt.Sprintf(" + %d\n", tagSize))
		} else {
			sb.WriteString("\n")
		}
	})

	sb.WriteString("\n    if id > 255 {\n        s += 5\n        return\n    }\n    s += 4\n    return\n}\n\n")
	return sb.String()
}

func (g *GoGen) GenSizePlain() string {
	var sb strings.Builder
	ctr := g.ctrStmt

	g.plainGen = true
	defer func() { g.plainGen = false }()

	sb.WriteString(fmt.Sprintf("// SizePlain - %s\nfunc (%s *%s) SizePlain() (s int) {\n",
		ctr.DefaultName, ctr.PrivateName, ctr.PublicName))

	g.ForEachCtrFields(func(_ int) {
		sb.WriteString(fmt.Sprintf("    s += %s\n", g.getSizeFunc()))
	})

	sb.WriteString("    return\n}\n\n")
	return sb.String()
}

func (g *GoGen) getMarshalFunc() string {
	ctr := g.ctrStmt
	field := g.field

	switch {
	case field.Type.IsArray:
		return fmt.Sprintf("bstd.MarshalSlice(n, b, %s.%s, %s)",
			ctr.PrivateName, field.PublicName, g.getElemMarshalFunc(field.Type.ChildType))
	case field.Type.IsMap:
		return fmt.Sprintf("bstd.MarshalMap(n, b, %s.%s, %s, %s)",
			ctr.PrivateName, field.PublicName, g.getElemMarshalFunc(field.Type.MapKeyType), g.getElemMarshalFunc(field.Type.ChildType))
	case field.Type.ExtStructure != "":
		if g.IsEnum(field.Type.ExtStructure) {
			return fmt.Sprintf("bgenimpl.MarshalEnum(n, b, %s.%s)",
				ctr.PrivateName, field.PublicName)
		}

		if g.plainGen {
			return fmt.Sprintf("%s.%s.MarshalPlain(n, b)",
				ctr.PrivateName, field.PublicName)
		}
		return fmt.Sprintf("%s.%s.marshal(n, b, %d)",
			ctr.PrivateName, field.PublicName, field.Id)
	default:
		return fmt.Sprintf("bstd.Marshal%s%s(n, b, %s.%s)",
			field.AppendUnsafeIfPresent(), field.Type.TokenType.String(), ctr.PrivateName, field.PublicName)
	}
}

func (g *GoGen) getElemMarshalFunc(t *parser.Type) string {
	switch {
	case t.IsArray:
		return fmt.Sprintf("func (n int, b []byte, s %s) int { return bstd.MarshalSlice(n, b, s, %s) }",
			utils.BencTypeToGolang(t), g.getElemMarshalFunc(t.ChildType))
	case t.IsMap:
		return fmt.Sprintf("func (n int, b []byte, s %s) int { return bstd.MarshalMap(n, b, s, %s, %s) }",
			utils.BencTypeToGolang(t), g.getElemMarshalFunc(t.MapKeyType), g.getElemMarshalFunc(t.ChildType))
	case t.ExtStructure != "":
		if g.IsEnum(t.ExtStructure) {
			return "bgenimpl.MarshalEnum"
		}

		return fmt.Sprintf("func (n int, b []byte, s %s) int { return s.MarshalPlain(n, b) }",
			utils.ToUpper(t.ExtStructure))
	default:
		return "bstd.Marshal" + t.AppendUnsafeIfPresent() + t.TokenType.String()
	}
}

func (g *GoGen) GenMarshal() string {
	var sb strings.Builder
	ctr := g.ctrStmt

	sb.WriteString(fmt.Sprintf("// Marshal - %s\nfunc (%s *%s) Marshal(b []byte) {\n    %s.marshal(0, b, 0)\n}\n\n",
		ctr.DefaultName, ctr.PrivateName, ctr.PublicName, ctr.PrivateName))

	sb.WriteString(fmt.Sprintf("// Nested Marshal - %s\nfunc (%s *%s) marshal(tn int, b []byte, id uint16) (n int) {\n    n = bgenimpl.MarshalTag(tn, b, bgenimpl.Container, id)\n",
		ctr.DefaultName, ctr.PrivateName, ctr.PublicName))

	g.ForEachCtrFields(func(_ int) {
		field := g.field

		if !g.IsCtr(field.Type.ExtStructure) {
			sb.WriteString(fmt.Sprintf("    n = bgenimpl.MarshalTag(n, b, bgenimpl.%s, %d)\n",
				g.mapTokenTypeToBgenimplType(field.Type.TokenType), field.Id))
		}
		sb.WriteString(fmt.Sprintf("    n = %s\n", g.getMarshalFunc()))
	})

	sb.WriteString("\n    n += 2\n    b[n-2] = 1\n    b[n-1] = 1\n    return\n}\n\n")
	return sb.String()
}

func (g *GoGen) GenMarshalPlain() string {
	var sb strings.Builder
	ctr := g.ctrStmt

	g.plainGen = true
	defer func() { g.plainGen = false }()

	sb.WriteString(fmt.Sprintf("// MarshalPlain - %s\nfunc (%s *%s) MarshalPlain(tn int, b []byte) (n int) {\n    n = tn\n",
		ctr.DefaultName, ctr.PrivateName, ctr.PublicName))

	g.ForEachCtrFields(func(_ int) {
		sb.WriteString(fmt.Sprintf("    n = %s\n", g.getMarshalFunc()))
	})

	sb.WriteString("    return n\n}\n\n")
	return sb.String()
}

func (g *GoGen) mapTokenTypeToBgenimplType(t lexer.Token) string {
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

func (g *GoGen) getUnmarshalFunc() string {
	ctr := g.ctrStmt
	field := g.field

	switch {
	case field.Type.IsArray:
		return fmt.Sprintf("bstd.UnmarshalSlice[%s](n, b, %s)",
			utils.BencTypeToGolang(field.Type.ChildType), g.getElemUnmarshalFunc(field.Type.ChildType))
	case field.Type.IsMap:
		return fmt.Sprintf("bstd.UnmarshalMap[%s, %s](n, b, %s, %s)",
			utils.BencTypeToGolang(field.Type.MapKeyType), utils.BencTypeToGolang(field.Type.ChildType), g.getElemUnmarshalFunc(field.Type.MapKeyType), g.getElemUnmarshalFunc(field.Type.ChildType))
	case field.Type.ExtStructure != "":
		if g.IsEnum(field.Type.ExtStructure) {
			return fmt.Sprintf("bgenimpl.UnmarshalEnum[%s](n, b)", field.Type.ExtStructure)
		}
		if g.plainGen {
			return fmt.Sprintf("%s.%s.UnmarshalPlain(n, b)", ctr.PrivateName, field.PublicName)
		}
		return fmt.Sprintf("bstd.Unmarshal%s%s(n, b)", field.AppendUnsafeIfPresent(), field.Type.TokenType.String())
	default:
		return fmt.Sprintf("bstd.Unmarshal%s%s(n, b)", field.AppendUnsafeIfPresent(), field.Type.TokenType.String())
	}
}

func (g *GoGen) getElemUnmarshalFunc(t *parser.Type) string {
	switch {
	case t.IsArray:
		return fmt.Sprintf("func (n int, b []byte) (int, %s, error) { return bstd.UnmarshalSlice[%s](n, b, %s) }",
			utils.BencTypeToGolang(t), utils.BencTypeToGolang(t.ChildType), g.getElemUnmarshalFunc(t.ChildType))
	case t.IsMap:
		return fmt.Sprintf("func (n int, b []byte) (int, %s, error) { return bstd.UnmarshalMap[%s, %s](n, b, %s, %s) }",
			utils.BencTypeToGolang(t), utils.BencTypeToGolang(t.MapKeyType), utils.BencTypeToGolang(t.ChildType), g.getElemUnmarshalFunc(t.MapKeyType), g.getElemUnmarshalFunc(t.ChildType))
	case t.ExtStructure != "":
		if g.IsEnum(t.ExtStructure) {
			return "bgenimpl.UnmarshalEnum"
		}
		return fmt.Sprintf("func (n int, b []byte, s *%s) (int, error) { return s.UnmarshalPlain(n, b) }",
			utils.ToUpper(t.ExtStructure))
	default:
		return "bstd.Unmarshal" + t.AppendUnsafeIfPresent() + t.TokenType.String()
	}
}

func (g *GoGen) GenUnmarshal() string {
	var sb strings.Builder
	ctr := g.ctrStmt

	sb.WriteString(fmt.Sprintf("// Unmarshal - %s\nfunc (%s *%s) Unmarshal(b []byte) (err error) {\n    _, err = %s.unmarshal(0, b, []uint16{}, 0)\n    return\n}\n\n",
		ctr.DefaultName, ctr.PrivateName, ctr.PublicName, ctr.PrivateName))

	sb.WriteString(fmt.Sprintf("// Nested Unmarshal - %s\nfunc (%s *%s) unmarshal(tn int, b []byte, r []uint16, id uint16) (n int, err error) {\n    var ok bool\n    if n, ok, err = bgenimpl.HandleCompatibility(tn, b, r, id); !ok {\n        if err == bgenimpl.ErrEof {\n            return n, nil\n        }\n        return\n    }\n",
		ctr.DefaultName, ctr.PrivateName, ctr.PublicName))

	g.ForEachCtrFields(func(_ int) {
		field := g.field

		if g.IsCtr(field.Type.ExtStructure) {
			sb.WriteString(fmt.Sprintf("    if n, err = %s.%s.unmarshal(n, b, %sRIds, %d); err != nil {\n        return\n    }\n",
				ctr.PrivateName, field.PublicName, ctr.PrivateName, field.Id))
			return
		}

		sb.WriteString(fmt.Sprintf("    if n, ok, err = bgenimpl.HandleCompatibility(n, b, %sRIds, %d); err != nil {\n        if err == bgenimpl.ErrEof {\n            return n, nil\n        }\n        return\n    }\n",
			ctr.PrivateName, field.Id))

		sb.WriteString(fmt.Sprintf("    if ok {\n        if n, %s.%s, err = %s; err != nil {\n            return\n        }\n    }\n",
			ctr.PrivateName, field.PublicName, g.getUnmarshalFunc()))
	})

	sb.WriteString("    n += 2\n    return\n}\n\n")
	return sb.String()
}

func (g *GoGen) GenUnmarshalPlain() string {
	var sb strings.Builder
	ctr := g.ctrStmt

	g.plainGen = true
	defer func() { g.plainGen = false }()

	sb.WriteString(fmt.Sprintf("// UnmarshalPlain - %s\nfunc (%s *%s) UnmarshalPlain(tn int, b []byte) (n int, err error) {\n    n = tn\n",
		ctr.DefaultName, ctr.PrivateName, ctr.PublicName))

	g.ForEachCtrFields(func(_ int) {
		field := g.field

		if g.IsCtr(field.Type.ExtStructure) {
			sb.WriteString(fmt.Sprintf("    if n, err = %s; err != nil {\n        return\n    }\n",
				g.getUnmarshalFunc()))
			return
		}

		sb.WriteString(fmt.Sprintf("    if n, %s.%s, err = %s; err != nil {\n        return\n    }\n",
			ctr.PrivateName, field.PublicName, g.getUnmarshalFunc()))
	})

	sb.WriteString("    return\n}\n\n")
	return sb.String()
}
