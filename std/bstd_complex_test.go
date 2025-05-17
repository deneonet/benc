package bstd

import (
	"reflect"
	"testing"
)

type ComplexData struct {
	Id                int64
	Title             string
	Items             []SubItem
	Metadata          map[string]int32
	Sub_data          SubComplexData
	Large_binary_data [][]byte
	Huge_list         []int64
}

func (complexData *ComplexData) SizePlain() (s int) {
	s += SizeInt64()
	s += SizeString(complexData.Title)
	s += SizeSlice(complexData.Items, func(s SubItem) int { return s.SizePlain() })
	s += SizeMap(complexData.Metadata, SizeString, SizeInt32)
	s += complexData.Sub_data.SizePlain()
	s += SizeSlice(complexData.Large_binary_data, SizeBytes)
	s += SizeFixedSlice(complexData.Huge_list, SizeInt64())
	return
}

func (complexData *ComplexData) MarshalPlain(tn int, b []byte) (n int) {
	n = tn
	n = MarshalInt64(n, b, complexData.Id)
	n = MarshalString(n, b, complexData.Title)
	n = MarshalSlice(n, b, complexData.Items, func(n int, b []byte, s SubItem) int { return s.MarshalPlain(n, b) })
	n = MarshalMap(n, b, complexData.Metadata, MarshalString, MarshalInt32)
	n = complexData.Sub_data.MarshalPlain(n, b)
	n = MarshalSlice(n, b, complexData.Large_binary_data, MarshalBytes)
	n = MarshalSlice(n, b, complexData.Huge_list, MarshalInt64)
	return n
}

func (complexData *ComplexData) UnmarshalPlain(tn int, b []byte) (n int, err error) {
	n = tn
	if n, complexData.Id, err = UnmarshalInt64(n, b); err != nil {
		return
	}
	if n, complexData.Title, err = UnmarshalString(n, b); err != nil {
		return
	}
	if n, complexData.Items, err = UnmarshalSlice[SubItem](n, b, func(n int, b []byte, s *SubItem) (int, error) { return s.UnmarshalPlain(n, b) }); err != nil {
		return
	}
	if n, complexData.Metadata, err = UnmarshalMap[string, int32](n, b, UnmarshalString, UnmarshalInt32); err != nil {
		return
	}
	if n, err = complexData.Sub_data.UnmarshalPlain(n, b); err != nil {
		return
	}
	if n, complexData.Large_binary_data, err = UnmarshalSlice[[]byte](n, b, UnmarshalBytesCropped); err != nil {
		return
	}
	if n, complexData.Huge_list, err = UnmarshalSlice[int64](n, b, UnmarshalInt64); err != nil {
		return
	}
	return
}

type SubItem struct {
	Sub_id      int32
	Description string
	Sub_items   []SubSubItem
}

func (subItem *SubItem) SizePlain() (s int) {
	s += SizeInt32()
	s += SizeString(subItem.Description)
	s += SizeSlice(subItem.Sub_items, func(s SubSubItem) int { return s.SizePlain() })
	return
}

func (subItem *SubItem) MarshalPlain(tn int, b []byte) (n int) {
	n = tn
	n = MarshalInt32(n, b, subItem.Sub_id)
	n = MarshalString(n, b, subItem.Description)
	n = MarshalSlice(n, b, subItem.Sub_items, func(n int, b []byte, s SubSubItem) int { return s.MarshalPlain(n, b) })
	return n
}

func (subItem *SubItem) UnmarshalPlain(tn int, b []byte) (n int, err error) {
	n = tn
	if n, subItem.Sub_id, err = UnmarshalInt32(n, b); err != nil {
		return
	}
	if n, subItem.Description, err = UnmarshalString(n, b); err != nil {
		return
	}
	if n, subItem.Sub_items, err = UnmarshalSlice[SubSubItem](n, b, func(n int, b []byte, s *SubSubItem) (int, error) { return s.UnmarshalPlain(n, b) }); err != nil {
		return
	}
	return
}

type SubSubItem struct {
	Sub_sub_id   string
	Sub_sub_data []byte
}

func (subSubItem *SubSubItem) SizePlain() (s int) {
	s += SizeString(subSubItem.Sub_sub_id)
	s += SizeBytes(subSubItem.Sub_sub_data)
	return
}

func (subSubItem *SubSubItem) MarshalPlain(tn int, b []byte) (n int) {
	n = tn
	n = MarshalString(n, b, subSubItem.Sub_sub_id)
	n = MarshalBytes(n, b, subSubItem.Sub_sub_data)
	return n
}

func (subSubItem *SubSubItem) UnmarshalPlain(tn int, b []byte) (n int, err error) {
	n = tn
	if n, subSubItem.Sub_sub_id, err = UnmarshalString(n, b); err != nil {
		return
	}
	if n, subSubItem.Sub_sub_data, err = UnmarshalBytesCropped(n, b); err != nil {
		return
	}
	return
}

type SubComplexData struct {
	Sub_id          int32
	Sub_title       string
	Sub_binary_data [][]byte
	Sub_items       []SubItem
	Sub_metadata    map[string]string
}

func (subComplexData *SubComplexData) SizePlain() (s int) {
	s += SizeInt32()
	s += SizeString(subComplexData.Sub_title)
	s += SizeSlice(subComplexData.Sub_binary_data, SizeBytes)
	s += SizeSlice(subComplexData.Sub_items, func(s SubItem) int { return s.SizePlain() })
	s += SizeMap(subComplexData.Sub_metadata, SizeString, SizeString)
	return
}

func (subComplexData *SubComplexData) MarshalPlain(tn int, b []byte) (n int) {
	n = tn
	n = MarshalInt32(n, b, subComplexData.Sub_id)
	n = MarshalString(n, b, subComplexData.Sub_title)
	n = MarshalSlice(n, b, subComplexData.Sub_binary_data, MarshalBytes)
	n = MarshalSlice(n, b, subComplexData.Sub_items, func(n int, b []byte, s SubItem) int { return s.MarshalPlain(n, b) })
	n = MarshalMap(n, b, subComplexData.Sub_metadata, MarshalString, MarshalString)
	return n
}

func (subComplexData *SubComplexData) UnmarshalPlain(tn int, b []byte) (n int, err error) {
	n = tn
	if n, subComplexData.Sub_id, err = UnmarshalInt32(n, b); err != nil {
		return
	}
	if n, subComplexData.Sub_title, err = UnmarshalString(n, b); err != nil {
		return
	}
	if n, subComplexData.Sub_binary_data, err = UnmarshalSlice[[]byte](n, b, UnmarshalBytesCropped); err != nil {
		return
	}
	if n, subComplexData.Sub_items, err = UnmarshalSlice[SubItem](n, b, func(n int, b []byte, s *SubItem) (int, error) { return s.UnmarshalPlain(n, b) }); err != nil {
		return
	}
	if n, subComplexData.Sub_metadata, err = UnmarshalMap[string, string](n, b, UnmarshalString, UnmarshalString); err != nil {
		return
	}
	return
}

func TestComplex(t *testing.T) {
	data := ComplexData{
		Id:    12345,
		Title: "Example Complex Data",
		Items: []SubItem{
			{
				Sub_id:      1,
				Description: "SubItem 1",
				Sub_items: []SubSubItem{
					{
						Sub_sub_id:   "subsub1",
						Sub_sub_data: []byte{0x01, 0x02, 0x03},
					},
				},
			},
		},
		Metadata: map[string]int32{
			"key1": 10,
			"key2": 20,
		},
		Sub_data: SubComplexData{
			Sub_id:    999,
			Sub_title: "Sub Complex Data",
			Sub_binary_data: [][]byte{
				{0x11, 0x22, 0x33},
				{0x44, 0x55, 0x66},
			},
			Sub_items: []SubItem{
				{
					Sub_id:      2,
					Description: "SubItem 2",
					Sub_items: []SubSubItem{
						{
							Sub_sub_id:   "subsub2",
							Sub_sub_data: []byte{0xAA, 0xBB, 0xCC},
						},
					},
				},
			},
			Sub_metadata: map[string]string{
				"meta1": "value1",
				"meta2": "value2",
			},
		},
		Large_binary_data: [][]byte{
			{0xFF, 0xEE, 0xDD},
		},
		Huge_list: []int64{1000000, 2000000, 3000000},
	}

	s := data.SizePlain()
	b := make([]byte, s)
	data.MarshalPlain(0, b)

	var retData ComplexData
	if _, err := retData.UnmarshalPlain(0, b); err != nil {
		t.Fatal(err)
	}

	if !reflect.DeepEqual(data, retData) {
		t.Fatalf("no match\norg: %v\ndec: %v\n", data, retData)
	}
}
