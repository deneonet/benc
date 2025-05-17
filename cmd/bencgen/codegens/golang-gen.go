package codegens

import (
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strings"

	"github.com/deneonet/benc/cmd/bencgen/lexer"
	"github.com/deneonet/benc/cmd/bencgen/parser"
	"github.com/deneonet/benc/cmd/bencgen/utils"
	"golang.org/x/exp/maps"
)

type GoContainerStmt struct {
	PublicName  string
	PrivateName string

	DefaultName string

	Fields      []parser.Field
	ReservedIDs []uint16
}

type GoEnumStmt struct {
	PublicName  string
	PrivateName string

	DefaultName string

	Values []string
}

type GoField struct {
	ID uint16

	PublicName  string
	PrivateName string
	DefaultName string

	Type *parser.Type
}

func (f *GoField) AppendUnsafeIfPresent() string {
	return f.Type.AppendUnsafeIfPresent()
}

func (f *GoField) AppendReturnCopyIfPresent() string {
	return f.Type.AppendReturnCopyIfPresent()
}

type GoGen struct {
	file string

	enumDecls      []string
	containerDecls []string

	varMap map[string]string

	importedPackages          []string
	importedEnumsOrContainers map[string]string

	plainGen   bool
	defineStmt *parser.DefineStmt

	// currently generated...

	field         GoField
	enumStmt      GoEnumStmt
	containerStmt GoContainerStmt
}

func NewGoGen(file string) *GoGen {
	return &GoGen{file: file, importedEnumsOrContainers: make(map[string]string)}
}

func (g *GoGen) File() string {
	return g.file
}

func (*GoGen) Lang() GenLang {
	return GoGenLang
}

func (g *GoGen) IsEnum(externalStructure string) bool {
	return slices.Contains(g.enumDecls, externalStructure)
}

func (g *GoGen) IsContainer(externalStructure string) bool {
	return slices.Contains(g.containerDecls, externalStructure)
}

func (g *GoGen) ForEachCtrFields(f func(i int)) {
	for i, field := range g.containerStmt.Fields {
		g.field = GoField{
			ID: field.ID,

			PublicName:  utils.ToUpper(field.Name),
			PrivateName: utils.ToLower(field.Name),
			DefaultName: field.Name,

			Type: field.Type,
		}
		f(i)
	}
}

func (g *GoGen) ForEachEnumValues(f func(i int, value string)) {
	for i, value := range g.enumStmt.Values {
		f(i, utils.ToUpper(value))
	}
}

func (g *GoGen) HasPackageDefined() bool {
	return g.defineStmt != nil
}

func (g *GoGen) SetVarMap(varMap map[string]string) {
	g.varMap = varMap
}

func (g *GoGen) SetDefineStatement(stmt *parser.DefineStmt) {
	g.defineStmt = stmt
}

func (g *GoGen) SetEnumStatement(stmt *parser.EnumStmt) {
	g.enumStmt = GoEnumStmt{
		PublicName:  utils.ToUpper(stmt.Name),
		PrivateName: utils.ToLower(stmt.Name),

		DefaultName: stmt.Name,
		Values:      stmt.Values,
	}
}

func (g *GoGen) adjustExternalStructureToImports(t *parser.Type) {
	if t.IsAnExternalStructure() {
		if replacement, ok := g.importedEnumsOrContainers[t.ExternalStructure]; ok {
			t.ExternalStructure = replacement
		}
		return
	}

	if t.IsArray {
		g.adjustExternalStructureToImports(t.ChildType)
		return
	}

	if t.IsMap {
		g.adjustExternalStructureToImports(t.MapKeyType)
		g.adjustExternalStructureToImports(t.ChildType)
		return
	}
}

func (g *GoGen) SetContainerStatement(stmt *parser.ContainerStmt) {
	fields := stmt.Fields
	for _, field := range fields {
		g.adjustExternalStructureToImports(field.Type)
	}

	g.containerStmt = GoContainerStmt{
		PublicName:  utils.ToUpper(stmt.Name),
		PrivateName: utils.ToLower(stmt.Name),

		Fields:      fields,
		DefaultName: stmt.Name,
		ReservedIDs: stmt.ReservedIDs,
	}
}

func (g *GoGen) AddEnumDecls(enumDecls []string) {
	g.enumDecls = append(g.enumDecls, enumDecls...)
}

func (g *GoGen) AddContainerDecls(containerDecls []string) {
	g.containerDecls = append(g.containerDecls, containerDecls...)
}

func (g *GoGen) ProcessImport(stmt *parser.UseStmt, importDirs []string) ([]string, []string) {
	var content []byte
	var err error

	importDirs = append(importDirs, "./")

	for _, importDir := range importDirs {
		trimmedDir := strings.TrimSuffix(importDir, "/")
		fullPath := filepath.Join(trimmedDir, stmt.Path)

		if _, err = os.Stat(fullPath); os.IsNotExist(err) {
			continue
		}

		content, err = os.ReadFile(fullPath)
		if err == nil {
			break
		}
	}

	if err != nil {
		LogErrorAndExit(g, fmt.Sprintf("Failed to read file: %s. Check again whether your import dirs are correct.", stmt.Path))
	}

	importParser := parser.NewParser(strings.NewReader(string(content)), string(content))
	importNodes := importParser.Parse()

	var goPackage string
	var definePackage string

	var enumDecls = []string{}
	var containerDecls = []string{}

	for _, node := range importNodes {
		switch n := node.(type) {
		case *parser.VarStmt:
			if n.Name == "go_package" {
				goPackage = n.Value
			}
		case *parser.DefineStmt:
			definePackage = n.Package
		case *parser.EnumStmt:
			enumDecls = append(enumDecls, n.Name)
		case *parser.ContainerStmt:
			containerDecls = append(containerDecls, n.Name)
		}
	}

	if goPackage == "" {
		LogErrorAndExit(g, fmt.Sprintf("No 'go_package' variable has been set in imported file '%s'.", stmt.Path))
	}

	splitPackage := strings.Split(goPackage, "/")
	if len(splitPackage) <= 1 {
		LogErrorAndExit(g, fmt.Sprintf("Invalid 'go_package' variable has been set in imported file '%s'.", stmt.Path))
	}

	packageAlias := splitPackage[len(splitPackage)-1]

	importedEnums := make(map[string]string)
	importedContainers := make(map[string]string)

	for _, enum := range enumDecls {
		importedEnums[definePackage+"."+enum] = packageAlias + "." + enum
	}
	for _, container := range containerDecls {
		importedContainers[definePackage+"."+container] = packageAlias + "." + container
	}

	g.enumDecls = append(g.enumDecls, maps.Values(importedEnums)...)
	g.containerDecls = append(g.containerDecls, maps.Values(importedContainers)...)

	if !slices.Contains(g.importedPackages, goPackage) {
		g.importedPackages = append(g.importedPackages, goPackage)
	}

	maps.Copy(g.importedEnumsOrContainers, importedEnums)
	maps.Copy(g.importedEnumsOrContainers, importedContainers)

	return maps.Keys(importedEnums), maps.Keys(importedContainers)
}

func (g *GoGen) joinImportedPackages() string {
	var sb strings.Builder
	iEnd := len(g.importedPackages) - 1
	for i, pkg := range g.importedPackages {
		if i == iEnd {
			sb.WriteString(fmt.Sprintf("	\"%s\"", pkg))
			break
		}
		sb.WriteString(fmt.Sprintf("	\"%s\"\n", pkg))
	}
	return sb.String()
}

func (g *GoGen) GenDefine() string {
	goPackage, found := g.varMap["go_package"]
	if !found {
		LogErrorAndExit(g, "No 'go_package' variable has been set.")
	}

	splitPackage := strings.Split(goPackage, "/")
	if len(splitPackage) <= 1 {
		LogErrorAndExit(g, "Invalid 'go_package' variable has been set.")
	}

	packageAlias := splitPackage[len(splitPackage)-1]

	return fmt.Sprintf(
		`package %s

import (
    "github.com/deneonet/benc/std"
    "github.com/deneonet/benc/impl/gen"

%s
)

`, packageAlias, g.joinImportedPackages())
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
	ctr := g.containerStmt
	return fmt.Sprintf("// Reserved Ids - %s\nvar %sRIds = []uint16{%s}\n\n",
		ctr.DefaultName, ctr.PrivateName, joinUint16(ctr.ReservedIDs))
}

func (g *GoGen) GenStruct() string {
	var sb strings.Builder
	ctr := g.containerStmt

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

	g.ForEachEnumValues(func(i int, value string) {
		if i == 0 {
			sb.WriteString(fmt.Sprintf("    %s%s %s = iota\n",
				enum.PublicName, value, enum.PublicName))
			return
		}
		sb.WriteString(fmt.Sprintf("    %s%s\n",
			enum.PublicName, value))
	})

	sb.WriteString(")\n\n")
	return sb.String()
}

func (g *GoGen) getSizeFunc() string {
	ctr := g.containerStmt
	field := g.field

	switch {
	case field.Type.IsArray:
		if field.Type.ChildType.TokenType == lexer.STRING || field.Type.ChildType.TokenType == lexer.BYTES || field.Type.ChildType.IsAnExternalStructure() || field.Type.ChildType.IsMap || field.Type.ChildType.IsArray {
			return fmt.Sprintf("bstd.SizeSlice(%s.%s, %s)",
				ctr.PrivateName, field.PublicName, g.getElemSizeFunc(field.Type.ChildType))
		}

		return fmt.Sprintf("bstd.SizeFixedSlice(%s.%s, %s())",
			ctr.PrivateName, field.PublicName, g.getElemSizeFunc(field.Type.ChildType))
	case field.Type.IsMap:
		return fmt.Sprintf("bstd.SizeMap(%s.%s, %s, %s)",
			ctr.PrivateName, field.PublicName, g.getElemSizeFunc(field.Type.MapKeyType), g.getElemSizeFunc(field.Type.ChildType))
	case field.Type.IsAnExternalStructure():
		if g.IsEnum(field.Type.ExternalStructure) {
			return fmt.Sprintf("bgenimpl.SizeEnum(%s.%s)",
				ctr.PrivateName, field.PublicName)
		}

		if g.plainGen {
			return fmt.Sprintf("%s.%s.SizePlain()",
				ctr.PrivateName, field.PublicName)
		}
		return fmt.Sprintf("%s.%s.NestedSize(%d)",
			ctr.PrivateName, field.PublicName, field.ID)
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

func makeExternalStructureUpperOrNot(externalStructure string) string {
	if strings.Contains(externalStructure, ".") {
		return externalStructure
	}
	return utils.ToUpper(externalStructure)
}

func (g *GoGen) getElemSizeFunc(t *parser.Type) string {
	switch {
	case t.IsArray:
		if t.ChildType.TokenType == lexer.STRING || t.ChildType.TokenType == lexer.BYTES || t.ChildType.IsAnExternalStructure() || t.ChildType.IsMap || t.ChildType.IsArray {
			return fmt.Sprintf("func (s %s) int { return bstd.SizeSlice(s, %s) }",
				utils.BencTypeToGolang(t), g.getElemSizeFunc(t.ChildType))
		}

		return fmt.Sprintf("func (s %s) int { return bstd.SizeFixedSlice(s, %s()) }",
			utils.BencTypeToGolang(t), g.getElemSizeFunc(t.ChildType))
	case t.IsMap:
		return fmt.Sprintf("func (s %s) int { return bstd.SizeMap(s, %s, %s) }",
			utils.BencTypeToGolang(t), g.getElemSizeFunc(t.MapKeyType), g.getElemSizeFunc(t.ChildType))
	case t.IsAnExternalStructure():
		if g.IsEnum(t.ExternalStructure) {
			return "bgenimpl.SizeEnum"
		}

		return fmt.Sprintf("func (s %s) int { return s.SizePlain() }",
			makeExternalStructureUpperOrNot(t.ExternalStructure))
	default:
		return "bstd.Size" + t.TokenType.String()
	}
}

func (g *GoGen) GenSize() string {
	var sb strings.Builder
	ctr := g.containerStmt

	sb.WriteString(fmt.Sprintf("// Size - %s\nfunc (%s *%s) Size() int {\n    return %s.NestedSize(0)\n}\n\n",
		ctr.DefaultName, ctr.PrivateName, ctr.PublicName, ctr.PrivateName))

	sb.WriteString(fmt.Sprintf("// Nested Size - %s\nfunc (%s *%s) NestedSize(id uint16) (s int) {\n",
		ctr.DefaultName, ctr.PrivateName, ctr.PublicName))

	g.ForEachCtrFields(func(_ int) {
		field := g.field

		tagSize := 2
		if g.field.ID > 255 {
			tagSize = 3
		}

		sb.WriteString(fmt.Sprintf("    s += %s", g.getSizeFunc()))

		if !g.IsContainer(field.Type.ExternalStructure) {
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
	ctr := g.containerStmt

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
	ctr := g.containerStmt
	field := g.field

	switch {
	case field.Type.IsArray:
		return fmt.Sprintf("bstd.MarshalSlice(n, b, %s.%s, %s)",
			ctr.PrivateName, field.PublicName, g.getElemMarshalFunc(field.Type.ChildType))
	case field.Type.IsMap:
		return fmt.Sprintf("bstd.MarshalMap(n, b, %s.%s, %s, %s)",
			ctr.PrivateName, field.PublicName, g.getElemMarshalFunc(field.Type.MapKeyType), g.getElemMarshalFunc(field.Type.ChildType))
	case field.Type.IsAnExternalStructure():
		if g.IsEnum(field.Type.ExternalStructure) {
			return fmt.Sprintf("bgenimpl.MarshalEnum(n, b, %s.%s)",
				ctr.PrivateName, field.PublicName)
		}

		if g.plainGen {
			return fmt.Sprintf("%s.%s.MarshalPlain(n, b)",
				ctr.PrivateName, field.PublicName)
		}
		return fmt.Sprintf("%s.%s.NestedMarshal(n, b, %d)",
			ctr.PrivateName, field.PublicName, field.ID)
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
	case t.IsAnExternalStructure():
		if g.IsEnum(t.ExternalStructure) {
			return "bgenimpl.MarshalEnum"
		}

		return fmt.Sprintf("func (n int, b []byte, s %s) int { return s.MarshalPlain(n, b) }",
			makeExternalStructureUpperOrNot(t.ExternalStructure))
	default:
		return "bstd.Marshal" + t.AppendUnsafeIfPresent() + t.TokenType.String()
	}
}

func (g *GoGen) GenMarshal() string {
	var sb strings.Builder
	ctr := g.containerStmt

	sb.WriteString(fmt.Sprintf("// Marshal - %s\nfunc (%s *%s) Marshal(b []byte) {\n    %s.NestedMarshal(0, b, 0)\n}\n\n",
		ctr.DefaultName, ctr.PrivateName, ctr.PublicName, ctr.PrivateName))

	sb.WriteString(fmt.Sprintf("// Nested Marshal - %s\nfunc (%s *%s) NestedMarshal(tn int, b []byte, id uint16) (n int) {\n    n = bgenimpl.MarshalTag(tn, b, bgenimpl.Container, id)\n",
		ctr.DefaultName, ctr.PrivateName, ctr.PublicName))

	g.ForEachCtrFields(func(_ int) {
		field := g.field

		if !g.IsContainer(field.Type.ExternalStructure) {
			sb.WriteString(fmt.Sprintf("    n = bgenimpl.MarshalTag(n, b, bgenimpl.%s, %d)\n",
				g.mapTokenTypeToBgenimplType(field.Type.TokenType), field.ID))
		}
		sb.WriteString(fmt.Sprintf("    n = %s\n", g.getMarshalFunc()))
	})

	sb.WriteString("\n    n += 2\n    b[n-2] = 1\n    b[n-1] = 1\n    return\n}\n\n")
	return sb.String()
}

func (g *GoGen) GenMarshalPlain() string {
	var sb strings.Builder
	ctr := g.containerStmt

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
	ctr := g.containerStmt
	field := g.field

	switch {
	case field.Type.IsArray:
		return fmt.Sprintf("bstd.UnmarshalSlice[%s](n, b, %s)",
			utils.BencTypeToGolang(field.Type.ChildType), g.getElemUnmarshalFunc(field.Type.ChildType))
	case field.Type.IsMap:
		return fmt.Sprintf("bstd.UnmarshalMap[%s, %s](n, b, %s, %s)",
			utils.BencTypeToGolang(field.Type.MapKeyType), utils.BencTypeToGolang(field.Type.ChildType), g.getElemUnmarshalFunc(field.Type.MapKeyType), g.getElemUnmarshalFunc(field.Type.ChildType))
	case field.Type.IsAnExternalStructure():
		if g.IsEnum(field.Type.ExternalStructure) {
			return fmt.Sprintf("bgenimpl.UnmarshalEnum[%s](n, b)", field.Type.ExternalStructure)
		}
		if g.plainGen {
			return fmt.Sprintf("%s.%s.UnmarshalPlain(n, b)", ctr.PrivateName, field.PublicName)
		}
		return fmt.Sprintf("bstd.Unmarshal%s%s%s(n, b)", field.AppendUnsafeIfPresent(), field.Type.TokenType.String(), field.AppendReturnCopyIfPresent())
	default:
		return fmt.Sprintf("bstd.Unmarshal%s%s%s(n, b)", field.AppendUnsafeIfPresent(), field.Type.TokenType.String(), field.AppendReturnCopyIfPresent())
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
	case t.IsAnExternalStructure():
		if g.IsEnum(t.ExternalStructure) {
			return "bgenimpl.UnmarshalEnum"
		}
		return fmt.Sprintf("func (n int, b []byte, s *%s) (int, error) { return s.UnmarshalPlain(n, b) }",
			makeExternalStructureUpperOrNot(t.ExternalStructure))
	default:
		return "bstd.Unmarshal" + t.AppendUnsafeIfPresent() + t.TokenType.String() + t.AppendReturnCopyIfPresent()
	}
}

func (g *GoGen) GenUnmarshal() string {
	var sb strings.Builder
	ctr := g.containerStmt

	sb.WriteString(fmt.Sprintf("// Unmarshal - %s\nfunc (%s *%s) Unmarshal(b []byte) (err error) {\n    _, err = %s.NestedUnmarshal(0, b, []uint16{}, 0)\n    return\n}\n\n",
		ctr.DefaultName, ctr.PrivateName, ctr.PublicName, ctr.PrivateName))

	sb.WriteString(fmt.Sprintf("// Nested Unmarshal - %s\nfunc (%s *%s) NestedUnmarshal(tn int, b []byte, r []uint16, id uint16) (n int, err error) {\n    var ok bool\n    if n, ok, err = bgenimpl.HandleCompatibility(tn, b, r, id); !ok {\n        if err == bgenimpl.ErrEof {\n            return n, nil\n        }\n        return\n    }\n",
		ctr.DefaultName, ctr.PrivateName, ctr.PublicName))

	g.ForEachCtrFields(func(_ int) {
		field := g.field

		if g.IsContainer(field.Type.ExternalStructure) {
			sb.WriteString(fmt.Sprintf("    if n, err = %s.%s.NestedUnmarshal(n, b, %sRIds, %d); err != nil {\n        return\n    }\n",
				ctr.PrivateName, field.PublicName, ctr.PrivateName, field.ID))
			return
		}

		sb.WriteString(fmt.Sprintf("    if n, ok, err = bgenimpl.HandleCompatibility(n, b, %sRIds, %d); err != nil {\n        if err == bgenimpl.ErrEof {\n            return n, nil\n        }\n        return\n    }\n",
			ctr.PrivateName, field.ID))

		sb.WriteString(fmt.Sprintf("    if ok {\n        if n, %s.%s, err = %s; err != nil {\n            return\n        }\n    }\n",
			ctr.PrivateName, field.PublicName, g.getUnmarshalFunc()))
	})

	sb.WriteString("    n += 2\n    return\n}\n\n")
	return sb.String()
}

func (g *GoGen) GenUnmarshalPlain() string {
	var sb strings.Builder
	ctr := g.containerStmt

	g.plainGen = true
	defer func() { g.plainGen = false }()

	sb.WriteString(fmt.Sprintf("// UnmarshalPlain - %s\nfunc (%s *%s) UnmarshalPlain(tn int, b []byte) (n int, err error) {\n    n = tn\n",
		ctr.DefaultName, ctr.PrivateName, ctr.PublicName))

	g.ForEachCtrFields(func(_ int) {
		field := g.field

		if g.IsContainer(field.Type.ExternalStructure) {
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
