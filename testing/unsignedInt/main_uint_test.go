//go:generate bencgen --in ../schemas/uint.benc --out ./ --file uint_test --lang go

package unsignedInt

import (
	"math/rand"
	"reflect"
	"testing"
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

	data := UintTest{
		Ui64: ui64,
		Ui32: ui32,
		Ui16: ui16,
		Ui:   ui,

		Ui64Arr: ui64Arr,
		Ui64Map: ui64Map,
	}

	buf := make([]byte, data.Size())
	data.Marshal(buf)

	var deserData UintTest
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
