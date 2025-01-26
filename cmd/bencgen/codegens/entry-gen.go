package codegens

import (
	"fmt"
	"os"
	"slices"

	"github.com/deneonet/benc/cmd/bencgen/parser"
	"github.com/deneonet/benc/cmd/bencgen/utils"
)

type GeneratorLanguage string

const (
	GenGolang GeneratorLanguage = "go"
)

func (lang GeneratorLanguage) String() string {
	switch lang {
	case GenGolang:
		return "golang"
	default:
		return "invalid"
	}
}

type Generator interface {
	File() string
	Lang() GeneratorLanguage
	GenHeader(stmt *parser.HeaderStmt) string
	GenEnum(stmt *parser.EnumStmt) string
	GenStruct(stmt *parser.CtrStmt) string
	GenReservedIds(stmt *parser.CtrStmt) string
	GenSize(stmt *parser.CtrStmt, enumDeclarations []string) string
	GenMarshal(stmt *parser.CtrStmt, enumDeclarations []string) string
	GenUnmarshal(stmt *parser.CtrStmt, enumDeclarations []string) string
	GenSizePlain(stmt *parser.CtrStmt, enumDeclarations []string) string
	GenMarshalPlain(stmt *parser.CtrStmt, enumDeclarations []string) string
	GenUnmarshalPlain(stmt *parser.CtrStmt, enumDeclarations []string) string
}

func NewGenerator(lang GeneratorLanguage, file string) Generator {
	switch lang {
	case GenGolang:
		return NewGoGenerator(file)
	default:
		return nil
	}
}

func logErrorAndExit(g Generator, msg string) {
	fmt.Printf("\n\033[1;31m[bencgen] Error:\033[0m\n"+
		"    \033[1;37mFile:\033[0m %s\n"+
		"    \033[1;37mMessage:\033[0m %s\n", g.File(), msg)
	os.Exit(1)
}

var unallowedNames = []string{"b", "n", "id", "r"}

func Generate(gen Generator, nodes []parser.Node) string {
	containsHeader := false

	ctrDeclarations := []string{}
	enumDeclarations := []string{}

	res := fmt.Sprintf("// Code generated by bencgen %s. DO NOT EDIT.\n// source: %s\n\n", gen.Lang().String(), gen.File())
	for _, node := range nodes {
		switch stmt := node.(type) {
		case *parser.CtrStmt:
			validateCtrStmt(gen, stmt, ctrDeclarations, enumDeclarations)
			ctrDeclarations = append(ctrDeclarations, stmt.Name)
		case *parser.EnumStmt:
			validateEnumStmt(gen, stmt, enumDeclarations, ctrDeclarations)
			enumDeclarations = append(enumDeclarations, stmt.Name)
		}
	}

	for _, node := range nodes {
		switch stmt := node.(type) {
		case *parser.HeaderStmt:
			if containsHeader {
				logErrorAndExit(gen, "Multiple `header` declarations.")
			}
			res += gen.GenHeader(stmt)
			containsHeader = true
		case *parser.CtrStmt:
			if !containsHeader {
				logErrorAndExit(gen, "A `header` was not declared.")
			}
			validateContainerFields(gen, stmt, ctrDeclarations, enumDeclarations)
			res += generateContainerCode(gen, stmt, enumDeclarations)
		case *parser.EnumStmt:
			if !containsHeader {
				logErrorAndExit(gen, "A `header` was not declared.")
			}
			validateEnumFields(gen, stmt)
			res += gen.GenEnum(stmt)
		}
	}

	return res
}

func validateCtrStmt(gen Generator, stmt *parser.CtrStmt, ctrDeclarations []string, enumDeclarations []string) {
	if slices.Contains(ctrDeclarations, stmt.Name) {
		logErrorAndExit(gen, fmt.Sprintf("Multiple containers with the same name \"%s\".", stmt.Name))
	}
	if slices.Contains(enumDeclarations, stmt.Name) {
		logErrorAndExit(gen, fmt.Sprintf("A enum with the same name \"%s\" is already declared.", stmt.Name))
	}

	if len(stmt.Fields) == 0 {
		logErrorAndExit(gen, fmt.Sprintf("Empty container \"%s\".", stmt.Name))
	}
}

func validateEnumStmt(gen Generator, stmt *parser.EnumStmt, enumDeclarations []string, ctrDeclarations []string) {
	if slices.Contains(enumDeclarations, stmt.Name) {
		logErrorAndExit(gen, fmt.Sprintf("Multiple enums with the same name \"%s\".", stmt.Name))
	}
	if slices.Contains(ctrDeclarations, stmt.Name) {
		logErrorAndExit(gen, fmt.Sprintf("A container with the same name \"%s\" is already declared.", stmt.Name))
	}

	if len(stmt.Fields) == 0 {
		logErrorAndExit(gen, fmt.Sprintf("Empty enum \"%s\".", stmt.Name))
	}
}

func validateContainerFields(gen Generator, stmt *parser.CtrStmt, ctrDeclarations []string, enumDeclarations []string) {
	var ids []uint16
	var lastID uint16
	var fieldNames []string

	for _, field := range stmt.Fields {
		_, enumNotFound := utils.FindUndeclaredContainersOrEnums(enumDeclarations, field.Type)
		if ctrEnum, notFound := utils.FindUndeclaredContainersOrEnums(ctrDeclarations, field.Type); notFound && enumNotFound {
			logErrorAndExit(gen, fmt.Sprintf("Container/Enum \"%s\" not declared on \"%s\" (\"%s\").", ctrEnum, stmt.Name, field.Name))
		}

		if field.Id == 0 {
			logErrorAndExit(gen, fmt.Sprintf("Field \"%s\" has an ID of \"0\" on \"%s\".", field.Name, stmt.Name))
		}

		if slices.Contains(ids, field.Id) {
			logErrorAndExit(gen, fmt.Sprintf("Multiple fields with the same ID \"%d\" on \"%s\".", field.Id, stmt.Name))
		}

		if lastID > field.Id {
			logErrorAndExit(gen, fmt.Sprintf("Fields must be ordered by their IDs in ascending order on \"%s\".", stmt.Name))
		}

		if slices.Contains(fieldNames, field.Name) {
			logErrorAndExit(gen, fmt.Sprintf("Multiple fields with the same name \"%s\" on \"%s\".", field.Name, stmt.Name))
		}

		if slices.Contains(unallowedNames, utils.ToLower(stmt.Name)) {
			logErrorAndExit(gen, fmt.Sprintf("Unallowed container name \"%s\".", stmt.Name))
		}

		lastID = field.Id
		ids = append(ids, field.Id)
		fieldNames = append(fieldNames, field.Name)
	}
}

func validateEnumFields(gen Generator, stmt *parser.EnumStmt) {
	var fieldNames []string

	for _, field := range stmt.Fields {
		if slices.Contains(fieldNames, field) {
			logErrorAndExit(gen, fmt.Sprintf("Multiple fields with the same name \"%s\" on \"%s\".", field, stmt.Name))
		}
		fieldNames = append(fieldNames, field)
	}
}

func generateContainerCode(gen Generator, stmt *parser.CtrStmt, enumDeclarations []string) string {
	return gen.GenStruct(stmt) +
		gen.GenReservedIds(stmt) +
		gen.GenSize(stmt, enumDeclarations) +
		gen.GenSizePlain(stmt, enumDeclarations) +
		gen.GenMarshal(stmt, enumDeclarations) +
		gen.GenMarshalPlain(stmt, enumDeclarations) +
		gen.GenUnmarshal(stmt, enumDeclarations) +
		gen.GenUnmarshalPlain(stmt, enumDeclarations)
}
