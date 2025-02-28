//go:generate bencgen --in ../schemas/others.benc --out ./ --file ... --lang go --import-dir ../schemas

package others

import (
	"math/rand"
	"reflect"
	"testing"

	"github.com/deneonet/benc/testing/person"
)

func TestUint(t *testing.T) {
	ui64 := rand.Uint64()
	ui32 := rand.Uint32()
	ui16 := uint16(65000)
	ui := uint(rand.Uint64())

	ui64Arr := []uint64{rand.Uint64(), rand.Uint64(), rand.Uint64()}

	ui64Map := make(map[uint64]uint32)
	ui64Map[rand.Uint64()] = rand.Uint32()
	ui64Map[rand.Uint64()] = rand.Uint32()
	ui64Map[rand.Uint64()] = rand.Uint32()

	bankMap := make(map[Bank]Citizen)
	bankMap[Bank{Name: "VR Bank"}] = Citizen{Name: "John"}
	bankMap[Bank{Name: "Sparkasse"}] = Citizen{Name: "John Jr."}

	data := OthersTest{
		Ui64: ui64,
		Ui32: ui32,
		Ui16: ui16,
		Ui:   ui,

		Ui64Arr: ui64Arr,
		Ui64Map: ui64Map,

		ExampleEnum:  ExampleEnumOne,
		ExampleEnum2: ExampleEnum2Six,

		Person2: [][][]person.Person2{{{{Age: 10}}}},

		BankMap: bankMap,
	}

	buf := make([]byte, data.Size())
	data.Marshal(buf)

	var deserData OthersTest
	err := deserData.Unmarshal(buf)
	if err != nil {
		t.Fatal(err)
	}

	if !reflect.DeepEqual(deserData, data) {
		t.Logf("%v", deserData)
		t.Logf("%v", data)
		t.Errorf("Deserialized- and original data don't match!")
	}
}
