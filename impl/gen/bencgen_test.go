package bgenimpl

import (
	"reflect"
	"testing"

	bstd "go.kine.bz/benc/std"
)

func TestSlices(t *testing.T) {
	slice := []string{"sliceelement1", "sliceelement2", "sliceelement3", "sliceelement4", "sliceelement5"}
	s := bstd.SizeSlice(slice, bstd.SizeString)
	buf := make([]byte, s)
	bstd.MarshalSlice(0, buf, slice, bstd.MarshalString)

	_, retSlice, err := bstd.UnmarshalSlice[string](0, buf, bstd.UnmarshalString)
	if err != nil {
		t.Fatal(err.Error())
	}

	if !reflect.DeepEqual(retSlice, slice) {
		t.Fatal("no match!")
	}

	t.Logf("org %v\ndec %v", slice, retSlice)
}
