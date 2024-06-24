//go:generate bencgen --in ../schemas/complex_data.benc --out ./ --file complex_data_test --lang go

package complex_data

import (
	"fmt"
	"reflect"
	"testing"
)

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

	s := data.Size()
	b := make([]byte, s)
	data.Marshal(b)
	fmt.Println(b)

	var retData ComplexData
	if err := retData.Unmarshal(b); err != nil {
		t.Fatal(err)
	}

	if !reflect.DeepEqual(data, retData) {
		t.Fatalf("no match\norg: %v\ndec: %v\n", data, retData)
	}
}

func BenchmarkComplex(b *testing.B) {
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
	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		s := data.Size()
		buf := make([]byte, s)
		data.Marshal(buf)

		var retData ComplexData
		if err := retData.Unmarshal(buf); err != nil {
			b.Fatal(err)
		}
	}
}
