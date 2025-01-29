package bcd

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/deneonet/benc/cmd/bencgen/parser"
	"github.com/deneonet/benc/cmd/bencgen/utils"
	"golang.org/x/exp/slices"
)

type Msgs struct {
	Msgs map[string]Msg `json:"msgs"`
}

type Msg struct {
	ReservedIDs []uint16                `json:"rIds"`
	Fields      map[uint16]parser.Field `json:"fields"`
}

type Bcd struct {
	File  string
	Nodes []parser.Node
}

func NewBcd(nodes []parser.Node, file string) *Bcd {
	return &Bcd{
		File:  file,
		Nodes: nodes,
	}
}

func (b *Bcd) handleError(message string) {
	fmt.Printf("\n\033[1;31m[bencgen] Error:\033[0m\n    \033[1;37mFile:\033[0m %s\n    \033[1;37mMessage:\033[0m %s\n", b.File, message)
	os.Exit(-1)
}

func (b *Bcd) Analyze(force bool) {
	contentBytes, err := os.ReadFile(b.File)
	if err != nil {
		panic(err)
	}

	content := string(contentBytes)
	start, end := strings.Index(content, "# [meta_s]"), strings.Index(content, "[meta_e]")

	var existingMsgs Msgs
	if start != -1 && end != -1 {
		base64Content := strings.TrimSpace(content[start+10 : end])
		if decodedContent, err := base64.StdEncoding.DecodeString(base64Content); err == nil {
			if err := json.Unmarshal(decodedContent, &existingMsgs); err != nil {
				panic(err)
			}
		} else {
			panic(err)
		}
	}

	newMsgs := b.buildMsgs(existingMsgs, force)
	b.updateContent(content, newMsgs, start, end)
}

func (b *Bcd) buildMsgs(existingMsgs Msgs, force bool) Msgs {
	newMsgs := Msgs{Msgs: make(map[string]Msg)}

	for _, node := range b.Nodes {
		if stmt, ok := node.(*parser.ContainerStmt); ok {
			fields := make(map[uint16]parser.Field)
			for _, field := range stmt.Fields {
				fields[field.ID] = field
			}

			if existingMsg, exists := existingMsgs.Msgs[stmt.Name]; !force && exists {
				b.checkForConflicts(existingMsg, stmt, fields)
			}

			newMsgs.Msgs[stmt.Name] = Msg{
				Fields:      fields,
				ReservedIDs: stmt.ReservedIDs,
			}
		}
	}

	b.mergeUnchangedFields(existingMsgs, &newMsgs)
	return newMsgs
}

func (b *Bcd) checkForConflicts(existingMsg Msg, stmt *parser.ContainerStmt, fields map[uint16]parser.Field) {
	for _, existingField := range existingMsg.Fields {
		currentField, exists := fields[existingField.ID]
		if !exists && !slices.Contains(stmt.ReservedIDs, existingField.ID) {
			b.handleError(fmt.Sprintf("Field '%s' (id '%d') on msg '%s' was removed, but '%d' is not marked as reserved.", existingField.Name, existingField.ID, stmt.Name, existingField.ID))
		}

		if exists {
			if slices.Contains(stmt.ReservedIDs, currentField.ID) {
				b.handleError(fmt.Sprintf("Field '%s' (id '%d') on msg '%s' may not be marked as reserved.", currentField.Name, currentField.ID, stmt.Name))
			}

			if existingField.ID == currentField.ID && !utils.CompareTypes(existingField.Type, currentField.Type) {
				b.handleError(fmt.Sprintf("Field '%s' (id '%d') on msg '%s' changed type from '%s' to '%s'.", currentField.Name, currentField.ID, stmt.Name, utils.FormatType(existingField.Type), utils.FormatType(currentField.Type)))
			}
		}
	}
}

func (b *Bcd) mergeUnchangedFields(existingMsgs Msgs, newMsgs *Msgs) {
	for name, msg := range existingMsgs.Msgs {
		if updatedMsg, exists := newMsgs.Msgs[name]; exists {
			for id, field := range msg.Fields {
				if _, exists := updatedMsg.Fields[id]; !exists {
					updatedMsg.Fields[id] = field
				}
			}
		}
	}
}

func (b *Bcd) updateContent(content string, newMsgs Msgs, start, end int) {
	marshaledMsgs, err := json.Marshal(&newMsgs)
	if err != nil {
		panic(err)
	}

	updatedMeta := base64.StdEncoding.EncodeToString(marshaledMsgs)
	var newContent strings.Builder

	if start == -1 || end == -1 {
		newContent.WriteString(content)
		newContent.WriteString("\n\n# DO NOT EDIT.\n# [meta_s] ")
		newContent.WriteString(updatedMeta)
		newContent.WriteString(" [meta_e]")
	} else {
		newContent.WriteString(content[:start+11])
		newContent.WriteString(updatedMeta)
		newContent.WriteString(content[end-1:])
	}

	if err = os.WriteFile(b.File, []byte(newContent.String()), os.ModePerm); err != nil {
		panic(err)
	}
}
