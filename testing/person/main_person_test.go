//go:generate bencgen --in ../schemas/person.benc --out ./ --file person_test --lang go
//go:generate bencgen --in ../schemas/person2.benc --out ./ --file person2_test --lang go

package person

import "testing"

// Forward compatibility
func TestPersonToPerson2(t *testing.T) {
	originalPerson := Person{
		Age:  30,
		Name: "John Doe",
		Parents: Parents{
			Mother: "Jane Doe",
			Father: "John Smith",
		},
		Child: Child{
			Age:  10,
			Name: "Junior Doe",
			Parents: Parents{
				Mother: "Jane Doe",
				Father: "John Doe",
			},
		},
	}

	expectedPerson2 := Person2{
		Age:  30,
		Name: "John Doe",
		Child: Child2{
			Age:  10,
			Name: "Junior Doe",
			Parents: Parents2{
				Mother: "Jane Doe",
				Father: "John Doe",
			},
		},
	}

	buf := make([]byte, originalPerson.Size())
	originalPerson.Marshal(buf)

	var deserPerson2 Person2
	err := deserPerson2.Unmarshal(buf)
	if err != nil {
		t.Fatal(err)
	}

	if deserPerson2.Age != expectedPerson2.Age {
		t.Errorf("Expected Age %d, got %d", expectedPerson2.Age, deserPerson2.Age)
	}
	if deserPerson2.Name != expectedPerson2.Name {
		t.Errorf("Expected Name %s, got %s", expectedPerson2.Name, deserPerson2.Name)
	}
	if deserPerson2.Child.Age != expectedPerson2.Child.Age {
		t.Errorf("Expected Child Age %d, got %d", expectedPerson2.Child.Age, deserPerson2.Child.Age)
	}
	if deserPerson2.Child.Name != expectedPerson2.Child.Name {
		t.Errorf("Expected Child Name %s, got %s", expectedPerson2.Child.Name, deserPerson2.Child.Name)
	}
	if deserPerson2.Child.Parents.Mother != expectedPerson2.Child.Parents.Mother {
		t.Errorf("Expected Child's Mother %s, got %s", expectedPerson2.Child.Parents.Mother, deserPerson2.Child.Parents.Mother)
	}
	if deserPerson2.Child.Parents.Father != expectedPerson2.Child.Parents.Father {
		t.Errorf("Expected Child's Father %s, got %s", expectedPerson2.Child.Parents.Father, deserPerson2.Child.Parents.Father)
	}
}

// Backward compatibility
func TestPerson2ToPerson(t *testing.T) {
	originalPerson2 := Person2{
		Age:  30,
		Name: "John Doe",
		Child: Child2{
			Age:  10,
			Name: "Junior Doe",
			Parents: Parents2{
				Mother: "Jane Doe",
				Father: "John Doe",
			},
		},
	}

	expectedPerson := Person{
		Age:  30,
		Name: "John Doe",
		Child: Child{
			Age:  10,
			Name: "Junior Doe",
			Parents: Parents{
				Mother: "Jane Doe",
				Father: "John Doe",
			},
		},
	}

	buf := make([]byte, originalPerson2.Size())
	originalPerson2.Marshal(buf)

	var deserPerson Person
	err := deserPerson.Unmarshal(buf)
	if err != nil {
		t.Fatal(err)
	}

	if deserPerson.Age != expectedPerson.Age {
		t.Errorf("Expected Age %d, got %d", expectedPerson.Age, deserPerson.Age)
	}
	if deserPerson.Name != expectedPerson.Name {
		t.Errorf("Expected Name %s, got %s", expectedPerson.Name, deserPerson.Name)
	}
	if deserPerson.Child.Age != expectedPerson.Child.Age {
		t.Errorf("Expected Child Age %d, got %d", expectedPerson.Child.Age, deserPerson.Child.Age)
	}
	if deserPerson.Child.Name != expectedPerson.Child.Name {
		t.Errorf("Expected Child Name %s, got %s", expectedPerson.Child.Name, deserPerson.Child.Name)
	}
	if deserPerson.Child.Parents.Mother != expectedPerson.Child.Parents.Mother {
		t.Errorf("Expected Child's Mother %s, got %s", expectedPerson.Child.Parents.Mother, deserPerson.Child.Parents.Mother)
	}
	if deserPerson.Child.Parents.Father != expectedPerson.Child.Parents.Father {
		t.Errorf("Expected Child's Father %s, got %s", expectedPerson.Child.Parents.Father, deserPerson.Child.Parents.Father)
	}
}
