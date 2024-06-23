//go:generate bencgen --in schemas/complex.benc --out generated/complex --lang go

package tests

import (
	"fmt"
	"reflect"
	"testing"

	"go.kine.bz/benc/testing/generated/complex"
)

func TestComplex(t *testing.T) {
	data := complex.A{
		Id:   123456789,
		Name: "Example Structure",
		SubItems: []complex.BB{
			{
				IsActive: true,
				Details: [][]complex.C{
					{
						{
							Value: 12.34,
							Measurements: []complex.D{
								{
									Timestamp: 1616161616,
									Note:      "Measurement 1",
									Events: []complex.E{
										{EID: 1, Description: "Event 1"},
										{EID: 2, Description: "Event 2"},
									},
								},
								{
									Timestamp: 1616161617,
									Note:      "Measurement 2",
									Events: []complex.E{
										{EID: 3, Description: "Event 3"},
									},
								},
							},
						},
						{
							Value: 56.78,
							Measurements: []complex.D{
								{
									Timestamp: 1616161618,
									Note:      "Measurement 3",
									Events: []complex.E{
										{EID: 4, Description: "Event 4"},
									},
								},
							},
						},
					},
				},
			},
			{
				IsActive: false,
				Details: [][]complex.C{
					{
						{
							Value: 90.12,
							Measurements: []complex.D{
								{
									Timestamp: 1616161619,
									Note:      "Measurement 4",
									Events: []complex.E{
										{EID: 5, Description: "Event 5"},
									},
								},
							},
						},
					},
				},
			},
		},
		ComplexData: [][][]complex.C{
			{
				{
					{
						Value: 101.23,
						Measurements: []complex.D{
							{
								Timestamp: 1616161620,
								Note:      "Complex Measurement 1",
								Events: []complex.E{
									{EID: 6, Description: "Complex Event 1"},
								},
							},
						},
					},
					{
						Value: 202.34,
						Measurements: []complex.D{
							{
								Timestamp: 1616161621,
								Note:      "Complex Measurement 2",
								Events: []complex.E{
									{EID: 7, Description: "Complex Event 2"},
									{EID: 8, Description: "Complex Event 3"},
								},
							},
						},
					},
				},
				{
					{
						Value: 303.45,
						Measurements: []complex.D{
							{
								Timestamp: 1616161622,
								Note:      "Complex Measurement 3",
								Events: []complex.E{
									{EID: 9, Description: "Complex Event 4"},
								},
							},
						},
					},
				},
			},
		},
	}

	s := data.Size()
	b := make([]byte, s)
	data.Marshal(b)
	fmt.Println(b)

	var retData complex.A
	if err := retData.Unmarshal(b); err != nil {
		t.Fatal(err)
	}

	if !reflect.DeepEqual(data, retData) {
		t.Fatalf("no match\norg: %v\ndec: %v\n", data, retData)
	}
}

func BenchmarkComplex(b *testing.B) {
	data := complex.A{
		Id:   123456789,
		Name: "Example Structure",
		SubItems: []complex.BB{
			{
				IsActive: true,
				Details: [][]complex.C{
					{
						{
							Value: 12.34,
							Measurements: []complex.D{
								{
									Timestamp: 1616161616,
									Note:      "Measurement 1",
									Events: []complex.E{
										{EID: 1, Description: "Event 1"},
										{EID: 2, Description: "Event 2"},
									},
								},
								{
									Timestamp: 1616161617,
									Note:      "Measurement 2",
									Events: []complex.E{
										{EID: 3, Description: "Event 3"},
									},
								},
							},
						},
						{
							Value: 56.78,
							Measurements: []complex.D{
								{
									Timestamp: 1616161618,
									Note:      "Measurement 3",
									Events: []complex.E{
										{EID: 4, Description: "Event 4"},
									},
								},
							},
						},
					},
				},
			},
			{
				IsActive: false,
				Details: [][]complex.C{
					{
						{
							Value: 90.12,
							Measurements: []complex.D{
								{
									Timestamp: 1616161619,
									Note:      "Measurement 4",
									Events: []complex.E{
										{EID: 5, Description: "Event 5"},
									},
								},
							},
						},
					},
				},
			},
		},
		ComplexData: [][][]complex.C{
			{
				{
					{
						Value: 101.23,
						Measurements: []complex.D{
							{
								Timestamp: 1616161620,
								Note:      "Complex Measurement 1",
								Events: []complex.E{
									{EID: 6, Description: "Complex Event 1"},
								},
							},
						},
					},
					{
						Value: 202.34,
						Measurements: []complex.D{
							{
								Timestamp: 1616161621,
								Note:      "Complex Measurement 2",
								Events: []complex.E{
									{EID: 7, Description: "Complex Event 2"},
									{EID: 8, Description: "Complex Event 3"},
								},
							},
						},
					},
				},
				{
					{
						Value: 303.45,
						Measurements: []complex.D{
							{
								Timestamp: 1616161622,
								Note:      "Complex Measurement 3",
								Events: []complex.E{
									{EID: 9, Description: "Complex Event 4"},
								},
							},
						},
					},
				},
			},
		},
	}

	for i := 0; i < b.N; i++ {
		s := data.Size()
		buf := make([]byte, s)
		data.Marshal(buf)

		var retData complex.A
		if err := retData.Unmarshal(buf); err != nil {
			b.Fatal(err)
		}
	}
}
