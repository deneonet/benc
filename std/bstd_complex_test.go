package bstd

import (
	"reflect"
	"testing"
)

type E struct {
	EID         int32
	Description string
}

func (e *E) Size() (s int) {
	s += SizeInt32()
	s += SizeString(e.Description)
	return
}

func (e *E) Marshal(tn int, b []byte) (n int) {
	n = tn
	n = MarshalInt32(n, b, e.EID)
	n = MarshalString(n, b, e.Description)
	return n
}

func (e *E) Unmarshal(tn int, b []byte) (n int, err error) {
	n = tn
	if n, e.EID, err = UnmarshalInt32(n, b); err != nil {
		return
	}
	if n, e.Description, err = UnmarshalString(n, b); err != nil {
		return
	}
	return
}

type D struct {
	Timestamp uint32
	Note      string
	Events    []E
}

func (d *D) Size() (s int) {
	s += SizeUInt32()
	s += SizeString(d.Note)
	s += SizeSlice(d.Events, func(s E) int { return s.Size() })
	return
}

func (d *D) Marshal(tn int, b []byte) (n int) {
	n = tn
	n = MarshalUInt32(n, b, d.Timestamp)
	n = MarshalString(n, b, d.Note)
	n = MarshalSlice(n, b, d.Events, func(n int, b []byte, s E) int { return s.Marshal(n, b) })
	return n
}

func (d *D) Unmarshal(tn int, b []byte) (n int, err error) {
	n = tn
	if n, d.Timestamp, err = UnmarshalUInt32(n, b); err != nil {
		return
	}
	if n, d.Note, err = UnmarshalString(n, b); err != nil {
		return
	}
	if n, d.Events, err = UnmarshalSlice[E](n, b, func(n int, b []byte, s *E) (int, error) { return s.Unmarshal(n, b) }); err != nil {
		return
	}
	return
}

type C struct {
	Value        float64
	Measurements []D
}

func (c *C) Size() (s int) {
	s += SizeFloat64()
	s += SizeSlice(c.Measurements, func(s D) int { return s.Size() })
	return
}

func (c *C) Marshal(tn int, b []byte) (n int) {
	n = tn
	n = MarshalFloat64(n, b, c.Value)
	n = MarshalSlice(n, b, c.Measurements, func(n int, b []byte, s D) int { return s.Marshal(n, b) })
	return n
}

func (c *C) Unmarshal(tn int, b []byte) (n int, err error) {
	n = tn
	if n, c.Value, err = UnmarshalFloat64(n, b); err != nil {
		return
	}
	if n, c.Measurements, err = UnmarshalSlice[D](n, b, func(n int, b []byte, s *D) (int, error) { return s.Unmarshal(n, b) }); err != nil {
		return
	}
	return
}

type BB struct {
	IsActive bool
	Details  [][]C
}

func (bB *BB) Size() (s int) {
	s += SizeBool()
	s += SizeSlice(bB.Details, func(s []C) int { return SizeSlice(s, func(s C) int { return s.Size() }) })
	return
}

func (bB *BB) Marshal(tn int, b []byte) (n int) {
	n = tn
	n = MarshalBool(n, b, bB.IsActive)
	n = MarshalSlice(n, b, bB.Details, func(n int, b []byte, s []C) int {
		return MarshalSlice(n, b, s, func(n int, b []byte, s C) int { return s.Marshal(n, b) })
	})
	return n
}

func (bB *BB) Unmarshal(tn int, b []byte) (n int, err error) {
	n = tn
	if n, bB.IsActive, err = UnmarshalBool(n, b); err != nil {
		return
	}
	if n, bB.Details, err = UnmarshalSlice[[]C](n, b, func(n int, b []byte) (int, []C, error) {
		return UnmarshalSlice[C](n, b, func(n int, b []byte, s *C) (int, error) { return s.Unmarshal(n, b) })
	}); err != nil {
		return
	}
	return
}

type A struct {
	Id          int64
	Name        string
	SubItems    []BB
	ComplexData [][][]C
}

func (a *A) Size() (s int) {
	s += SizeInt64()
	s += SizeString(a.Name)
	s += SizeSlice(a.SubItems, func(s BB) int { return s.Size() })
	s += SizeSlice(a.ComplexData, func(s [][]C) int {
		return SizeSlice(s, func(s []C) int { return SizeSlice(s, func(s C) int { return s.Size() }) })
	})
	return
}

func (a *A) Marshal(tn int, b []byte) (n int) {
	n = tn
	n = MarshalInt64(n, b, a.Id)
	n = MarshalString(n, b, a.Name)
	n = MarshalSlice(n, b, a.SubItems, func(n int, b []byte, s BB) int { return s.Marshal(n, b) })
	n = MarshalSlice(n, b, a.ComplexData, func(n int, b []byte, s [][]C) int {
		return MarshalSlice(n, b, s, func(n int, b []byte, s []C) int {
			return MarshalSlice(n, b, s, func(n int, b []byte, s C) int { return s.Marshal(n, b) })
		})
	})
	return n
}

func (a *A) Unmarshal(tn int, b []byte) (n int, err error) {
	n = tn
	if n, a.Id, err = UnmarshalInt64(n, b); err != nil {
		return
	}
	if n, a.Name, err = UnmarshalString(n, b); err != nil {
		return
	}
	if n, a.SubItems, err = UnmarshalSlice[BB](n, b, func(n int, b []byte, s *BB) (int, error) { return s.Unmarshal(n, b) }); err != nil {
		return
	}
	if n, a.ComplexData, err = UnmarshalSlice[[][]C](n, b, func(n int, b []byte) (int, [][]C, error) {
		return UnmarshalSlice[[]C](n, b, func(n int, b []byte) (int, []C, error) {
			return UnmarshalSlice[C](n, b, func(n int, b []byte, s *C) (int, error) { return s.Unmarshal(n, b) })
		})
	}); err != nil {
		return
	}
	return
}

func TestComplex(t *testing.T) {
	data := A{
		Id:   123456789,
		Name: "Example Structure",
		SubItems: []BB{
			{
				IsActive: true,
				Details: [][]C{
					{
						{
							Value: 12.34,
							Measurements: []D{
								{
									Timestamp: 1616161616,
									Note:      "Measurement 1",
									Events: []E{
										{EID: 1, Description: "Event 1"},
										{EID: 2, Description: "Event 2"},
									},
								},
								{
									Timestamp: 1616161617,
									Note:      "Measurement 2",
									Events: []E{
										{EID: 3, Description: "Event 3"},
									},
								},
							},
						},
						{
							Value: 56.78,
							Measurements: []D{
								{
									Timestamp: 1616161618,
									Note:      "Measurement 3",
									Events: []E{
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
				Details: [][]C{
					{
						{
							Value: 90.12,
							Measurements: []D{
								{
									Timestamp: 1616161619,
									Note:      "Measurement 4",
									Events: []E{
										{EID: 5, Description: "Event 5"},
									},
								},
							},
						},
					},
				},
			},
		},
		ComplexData: [][][]C{
			{
				{
					{
						Value: 101.23,
						Measurements: []D{
							{
								Timestamp: 1616161620,
								Note:      "Complex Measurement 1",
								Events: []E{
									{EID: 6, Description: "Complex Event 1"},
								},
							},
						},
					},
					{
						Value: 202.34,
						Measurements: []D{
							{
								Timestamp: 1616161621,
								Note:      "Complex Measurement 2",
								Events: []E{
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
						Measurements: []D{
							{
								Timestamp: 1616161622,
								Note:      "Complex Measurement 3",
								Events: []E{
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
	data.Marshal(0, b)

	var retData A
	if _, err := retData.Unmarshal(0, b); err != nil {
		t.Fatal(err)
	}

	if !reflect.DeepEqual(data, retData) {
		t.Fatalf("no match\norg: %v\ndec: %v\n", data, retData)
	}
}
