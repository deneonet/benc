package bcd

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"os"
	"slices"
	"strings"

	"go.kine.bz/benc/cmd/bencgen/parser"
	"go.kine.bz/benc/cmd/bencgen/utils"
)

type Msgs struct {
	Msgs map[string]Msg `json:"msgs"`
}

type Msg struct {
	RIds   []uint16                `json:"rIds"`
	Fields map[uint16]parser.Field `json:"fields"`
}

type Bcd struct {
	file  string
	nodes []parser.Node
}

func NewBcd(nodes []parser.Node, file string) *Bcd {
	return &Bcd{
		file:  file,
		nodes: nodes,
	}
}

func (b *Bcd) Analyze(force bool) {
	contentBytes, _ := os.ReadFile(b.file)
	content := string(contentBytes)

	start := strings.Index(content, "# [meta_s]")
	end := strings.Index(content, "[meta_e]")

	msgs := Msgs{}

	if start != -1 && end != -1 {
		base64Content := strings.TrimSpace(content[start+10 : end])

		b, err := base64.StdEncoding.DecodeString(base64Content)
		if err != nil {
			panic(err)
		}

		err = json.Unmarshal(b, &msgs)
		if err != nil {
			panic(err)
		}
	}

	newMsgs := Msgs{
		Msgs: make(map[string]Msg),
	}

	for _, node := range b.nodes {
		switch stmt := node.(type) {
		case *parser.CtrStmt:
			fields := make(map[uint16]parser.Field)
			for _, field := range stmt.Fields {
				fields[field.Id] = field
			}

			if m, ok := msgs.Msgs[stmt.Name]; ok {
				for _, id := range stmt.ReservedIds {
					if _, ok := m.Fields[id]; !force && !ok {
						b.error(fmt.Sprintf("ID \"%d\" on msg \"%s\" is marked as reserved, but never existed.", id, stmt.Name))
					}
				}

				for _, field := range m.Fields {
					f, ok := fields[field.Id]
					if !force && !ok && !slices.Contains(stmt.ReservedIds, field.Id) {
						b.error(fmt.Sprintf("Field \"%s\" (id \"%d\") on msg \"%s\" was removed, but \"%d\" is not marked as reserved.", field.Name, field.Id, stmt.Name, field.Id))
					}

					if !force && ok {
						if slices.Contains(stmt.ReservedIds, f.Id) {
							b.error(fmt.Sprintf("Field \"%s\" (id \"%d\") on msg \"%s\" may not be marked as reserved.", f.Name, f.Id, stmt.Name))
						}

						if field.Id == f.Id && (field.Type != f.Type || field.IsArray != f.IsArray) {
							b.error(fmt.Sprintf("Field \"%s\" (id \"%d\") on msg \"%s\" changed type from \"%s\" to \"%s\".", f.Name, f.Id, stmt.Name, utils.FormatType(field), utils.FormatType(f)))
						}
					}
				}
			}

			newMsgs.Msgs[stmt.Name] = Msg{
				Fields: fields,
				RIds:   stmt.ReservedIds,
			}
		}
	}

	for n, m := range msgs.Msgs {
		if msg, ok := newMsgs.Msgs[n]; ok {
			for id, f := range m.Fields {
				if _, ok := msg.Fields[id]; !ok {
					msg.Fields[id] = f
				}
			}
		}
	}

	mbs, err := json.Marshal(&newMsgs)
	if err != nil {
		panic(err)
	}

	newContent := ""
	updatedMeta := base64.StdEncoding.EncodeToString(mbs)

	if start == -1 || end == -1 {
		newContent = content + "\n\n" + "# DO NOT EDIT." + "\n" + "# [meta_s] " + updatedMeta + " [meta_e]"
	} else {
		newContent = content[:start+11] + updatedMeta + content[end-1:]
	}

	if err = os.WriteFile(b.file, []byte(newContent), os.ModePerm); err != nil {
		panic(err)
	}
}
