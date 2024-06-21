//go:generate bencgen --in schemas/person.benc --out generated/person --lang go
//go:generate bencgen --in schemas/person2.benc --out generated/person2 --lang go

package bfc

import (
	"reflect"
	"testing"

	"go.kine.bz/benc"
	"go.kine.bz/benc/testing/bfc/generated/person"
	"go.kine.bz/benc/testing/bfc/generated/person2"
)

func CustomDeepEqual(a, b interface{}) bool {
	aVal := reflect.ValueOf(a)
	bVal := reflect.ValueOf(b)

	if aVal.Kind() != reflect.Struct || bVal.Kind() != reflect.Struct {
		return false
	}

	for i := 0; i < aVal.NumField(); i++ {
		fieldA := aVal.Field(i)
		fieldTypeA := aVal.Type().Field(i)

		fieldB := bVal.FieldByName(fieldTypeA.Name)
		fieldTypeB, ok := bVal.Type().FieldByName(fieldTypeA.Name)

		if !ok {
			continue
		}

		if !fieldTypeA.IsExported() || !fieldTypeB.IsExported() {
			continue
		}

		// Recursively check nested structs.
		if fieldA.Kind() == reflect.Struct && fieldB.Kind() == reflect.Struct {
			if !CustomDeepEqual(fieldA, fieldB) {
				return false
			}
			continue
		}

		if !reflect.DeepEqual(fieldA.Interface(), fieldB.Interface()) {
			return false
		}
	}

	return true
}

func TestPersonToPerson2(t *testing.T) {
	data := person.Person{
		Age:  24,
		Name: "Johnny",
		Parents: person.Parents{
			Mother: "Johna",
			Father: "John",
		},
		Child: person.Child{
			Name: "Johnny Jr.",
			Age:  3,
			Parents: person.Parents{
				Mother: "Johna Jr.",
				Father: "Johnny",
			},
		},
	}

	b, err := benc.MarshalCtr(&data)
	if err != nil {
		t.Fatal(err)
	}

	var retData person2.Person
	if err = benc.UnmarshalCtr(b, &retData); err != nil {
		t.Fatal(err)
	}

	if !CustomDeepEqual(data, retData) {
		t.Fatalf("no match\norg: %v\ndec: %v\n", data, retData)
	}
}

func TestPerson2ToPerson(t *testing.T) {
	data := person2.Person{
		Age:  24,
		Name: "Johnny",
		Child: person2.Child{
			Name: "Johnny Jr.",
			Age:  3,
			Parents: person2.Parents{
				Mother: "Johna Jr.",
				Father: "Johnny",
			},
		},
	}

	b, err := benc.MarshalCtr(&data)
	if err != nil {
		t.Fatal(err)
	}

	var retData person.Person
	if err = benc.UnmarshalCtr(b, &retData); err != nil {
		t.Fatal(err)
	}

	if !CustomDeepEqual(data, retData) {
		t.Fatalf("no match\norg: %v\ndec: %v\n", data, retData)
	}
}
