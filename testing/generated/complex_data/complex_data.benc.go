// Code generated by bencgen golang. DO NOT EDIT.
// source: schemas/complex_data.benc

package complex_data

import (
    "go.kine.bz/benc/std"
    "go.kine.bz/benc/impl/gen"
)

// Struct - ComplexData
type ComplexData struct {
    Id int64
    Title string
    Items []SubItem
    Metadata map[string]int32
    Sub_data SubComplexData
    Large_binary_data [][]byte
    Huge_list []int64
}

// Reserved Ids - ComplexData
var complexDataRIds = []uint16{}

// Size - ComplexData
func (complexData *ComplexData) Size() int {
    return complexData.size(0)
}

// Nested Size - ComplexData
func (complexData *ComplexData) size(id uint16) (s int) {
    s += bstd.SizeInt64() + 2
    s += bstd.SizeString(complexData.Title) + 2
    s += bstd.SizeSlice(complexData.Items, func (s SubItem) int { return s.SizePlain() }) + 2
    s += bstd.SizeMap(complexData.Metadata, bstd.SizeString, bstd.SizeInt32) + 2
    s += complexData.Sub_data.size(5)
    s += bstd.SizeSlice(complexData.Large_binary_data, bstd.SizeBytes) + 2
    s += bstd.SizeSlice(complexData.Huge_list, bstd.SizeInt64) + 2

    if id > 255 {
        s += 5
        return
    }
    s += 4
    return
}

// SizePlain - ComplexData
func (complexData *ComplexData) SizePlain() (s int) {
    s += bstd.SizeInt64()
    s += bstd.SizeString(complexData.Title)
    s += bstd.SizeSlice(complexData.Items, func (s SubItem) int { return s.SizePlain() })
    s += bstd.SizeMap(complexData.Metadata, bstd.SizeString, bstd.SizeInt32)
    s += complexData.Sub_data.SizePlain()
    s += bstd.SizeSlice(complexData.Large_binary_data, bstd.SizeBytes)
    s += bstd.SizeSlice(complexData.Huge_list, bstd.SizeInt64)
    return
}

// Marshal - ComplexData
func (complexData *ComplexData) Marshal(b []byte) {
    complexData.marshal(0, b, 0)
}

// Nested Marshal - ComplexData
func (complexData *ComplexData) marshal(tn int, b []byte, id uint16) (n int) {
    n = bgenimpl.MarshalTag(tn, b, bgenimpl.Container, id)
    n = bgenimpl.MarshalTag(n, b, bgenimpl.Fixed64, 1)
    n = bstd.MarshalInt64(n, b, complexData.Id)
    n = bgenimpl.MarshalTag(n, b, bgenimpl.Bytes, 2)
    n = bstd.MarshalString(n, b, complexData.Title)
    n = bgenimpl.MarshalTag(n, b, bgenimpl.Array, 3)
    n = bstd.MarshalSlice(n, b, complexData.Items, func (n int, b []byte, s SubItem) int { return s.MarshalPlain(n, b) })
    n = bgenimpl.MarshalTag(n, b, bgenimpl.Array, 4)
    n = bstd.MarshalMap(n, b, complexData.Metadata, bstd.MarshalString, bstd.MarshalInt32)
    n = complexData.Sub_data.marshal(n, b, 5)
    n = bgenimpl.MarshalTag(n, b, bgenimpl.Array, 6)
    n = bstd.MarshalSlice(n, b, complexData.Large_binary_data, bstd.MarshalBytes)
    n = bgenimpl.MarshalTag(n, b, bgenimpl.Array, 7)
    n = bstd.MarshalSlice(n, b, complexData.Huge_list, bstd.MarshalInt64)

    n += 2
    b[n-2] = 1
    b[n-1] = 1
    return
}

// MarshalPlain - ComplexData
func (complexData *ComplexData) MarshalPlain(tn int, b []byte) (n int) {
    n = tn
    n = bstd.MarshalInt64(n, b, complexData.Id)
    n = bstd.MarshalString(n, b, complexData.Title)
    n = bstd.MarshalSlice(n, b, complexData.Items, func (n int, b []byte, s SubItem) int { return s.MarshalPlain(n, b) })
    n = bstd.MarshalMap(n, b, complexData.Metadata, bstd.MarshalString, bstd.MarshalInt32)
    n = complexData.Sub_data.MarshalPlain(n, b)
    n = bstd.MarshalSlice(n, b, complexData.Large_binary_data, bstd.MarshalBytes)
    n = bstd.MarshalSlice(n, b, complexData.Huge_list, bstd.MarshalInt64)
    return n
}

// Unmarshal - ComplexData
func (complexData *ComplexData) Unmarshal(b []byte) (err error) {
    _, err = complexData.unmarshal(0, b, []uint16{}, 0)
    return
}

// Nested Unmarshal - ComplexData
func (complexData *ComplexData) unmarshal(tn int, b []byte, r []uint16, id uint16) (n int, err error) {
    var ok bool
    if n, ok, err = bgenimpl.HandleCompatibility(tn, b, r, id); !ok {
        if err == bgenimpl.ErrEof {
            return n, nil
        }
        return
    }
    if n, ok, err = bgenimpl.HandleCompatibility(n, b, complexDataRIds, 1); err != nil {
        if err == bgenimpl.ErrEof {
            return n, nil
        }
        return
    }
    if ok {
        if n, complexData.Id, err = bstd.UnmarshalInt64(n, b); err != nil {
            return
        }
    }
    if n, ok, err = bgenimpl.HandleCompatibility(n, b, complexDataRIds, 2); err != nil {
        if err == bgenimpl.ErrEof {
            return n, nil
        }
        return
    }
    if ok {
        if n, complexData.Title, err = bstd.UnmarshalString(n, b); err != nil {
            return
        }
    }
    if n, ok, err = bgenimpl.HandleCompatibility(n, b, complexDataRIds, 3); err != nil {
        if err == bgenimpl.ErrEof {
            return n, nil
        }
        return
    }
    if ok {
        if n, complexData.Items, err = bstd.UnmarshalSlice[SubItem](n, b, func (n int, b []byte, s *SubItem) (int, error) { return s.UnmarshalPlain(n, b) }); err != nil {
            return
        }
    }
    if n, ok, err = bgenimpl.HandleCompatibility(n, b, complexDataRIds, 4); err != nil {
        if err == bgenimpl.ErrEof {
            return n, nil
        }
        return
    }
    if ok {
        if n, complexData.Metadata, err = bstd.UnmarshalMap[string, int32](n, b, bstd.UnmarshalString, bstd.UnmarshalInt32); err != nil {
            return
        }
    }
    if n, ok, err = bgenimpl.HandleCompatibility(n, b, complexDataRIds, 5); err != nil {
        if err == bgenimpl.ErrEof {
            return n, nil
        }
        return
    }
    if ok {
        if n, err = complexData.Sub_data.unmarshal(n, b, complexDataRIds, 5); err != nil {
            return
        }
    }
    if n, ok, err = bgenimpl.HandleCompatibility(n, b, complexDataRIds, 6); err != nil {
        if err == bgenimpl.ErrEof {
            return n, nil
        }
        return
    }
    if ok {
        if n, complexData.Large_binary_data, err = bstd.UnmarshalSlice[[]byte](n, b, bstd.UnmarshalBytes); err != nil {
            return
        }
    }
    if n, ok, err = bgenimpl.HandleCompatibility(n, b, complexDataRIds, 7); err != nil {
        if err == bgenimpl.ErrEof {
            return n, nil
        }
        return
    }
    if ok {
        if n, complexData.Huge_list, err = bstd.UnmarshalSlice[int64](n, b, bstd.UnmarshalInt64); err != nil {
            return
        }
    }
    n += 2
    return
}

// UnmarshalPlain - ComplexData
func (complexData *ComplexData) UnmarshalPlain(tn int, b []byte) (n int, err error) {
    n = tn
    if n, complexData.Id, err = bstd.UnmarshalInt64(n, b); err != nil {
        return
    }
    if n, complexData.Title, err = bstd.UnmarshalString(n, b); err != nil {
        return
    }
    if n, complexData.Items, err = bstd.UnmarshalSlice[SubItem](n, b, func (n int, b []byte, s *SubItem) (int, error) { return s.UnmarshalPlain(n, b) }); err != nil {
        return
    }
    if n, complexData.Metadata, err = bstd.UnmarshalMap[string, int32](n, b, bstd.UnmarshalString, bstd.UnmarshalInt32); err != nil {
        return
    }
    if n, err = complexData.Sub_data.UnmarshalPlain(n, b); err != nil {
        return
    }
    if n, complexData.Large_binary_data, err = bstd.UnmarshalSlice[[]byte](n, b, bstd.UnmarshalBytes); err != nil {
        return
    }
    if n, complexData.Huge_list, err = bstd.UnmarshalSlice[int64](n, b, bstd.UnmarshalInt64); err != nil {
        return
    }
    return
}

// Struct - SubItem
type SubItem struct {
    Sub_id int32
    Description string
    Sub_items []SubSubItem
}

// Reserved Ids - SubItem
var subItemRIds = []uint16{}

// Size - SubItem
func (subItem *SubItem) Size() int {
    return subItem.size(0)
}

// Nested Size - SubItem
func (subItem *SubItem) size(id uint16) (s int) {
    s += bstd.SizeInt32() + 2
    s += bstd.SizeString(subItem.Description) + 2
    s += bstd.SizeSlice(subItem.Sub_items, func (s SubSubItem) int { return s.SizePlain() }) + 2

    if id > 255 {
        s += 5
        return
    }
    s += 4
    return
}

// SizePlain - SubItem
func (subItem *SubItem) SizePlain() (s int) {
    s += bstd.SizeInt32()
    s += bstd.SizeString(subItem.Description)
    s += bstd.SizeSlice(subItem.Sub_items, func (s SubSubItem) int { return s.SizePlain() })
    return
}

// Marshal - SubItem
func (subItem *SubItem) Marshal(b []byte) {
    subItem.marshal(0, b, 0)
}

// Nested Marshal - SubItem
func (subItem *SubItem) marshal(tn int, b []byte, id uint16) (n int) {
    n = bgenimpl.MarshalTag(tn, b, bgenimpl.Container, id)
    n = bgenimpl.MarshalTag(n, b, bgenimpl.Fixed32, 1)
    n = bstd.MarshalInt32(n, b, subItem.Sub_id)
    n = bgenimpl.MarshalTag(n, b, bgenimpl.Bytes, 2)
    n = bstd.MarshalString(n, b, subItem.Description)
    n = bgenimpl.MarshalTag(n, b, bgenimpl.Array, 3)
    n = bstd.MarshalSlice(n, b, subItem.Sub_items, func (n int, b []byte, s SubSubItem) int { return s.MarshalPlain(n, b) })

    n += 2
    b[n-2] = 1
    b[n-1] = 1
    return
}

// MarshalPlain - SubItem
func (subItem *SubItem) MarshalPlain(tn int, b []byte) (n int) {
    n = tn
    n = bstd.MarshalInt32(n, b, subItem.Sub_id)
    n = bstd.MarshalString(n, b, subItem.Description)
    n = bstd.MarshalSlice(n, b, subItem.Sub_items, func (n int, b []byte, s SubSubItem) int { return s.MarshalPlain(n, b) })
    return n
}

// Unmarshal - SubItem
func (subItem *SubItem) Unmarshal(b []byte) (err error) {
    _, err = subItem.unmarshal(0, b, []uint16{}, 0)
    return
}

// Nested Unmarshal - SubItem
func (subItem *SubItem) unmarshal(tn int, b []byte, r []uint16, id uint16) (n int, err error) {
    var ok bool
    if n, ok, err = bgenimpl.HandleCompatibility(tn, b, r, id); !ok {
        if err == bgenimpl.ErrEof {
            return n, nil
        }
        return
    }
    if n, ok, err = bgenimpl.HandleCompatibility(n, b, subItemRIds, 1); err != nil {
        if err == bgenimpl.ErrEof {
            return n, nil
        }
        return
    }
    if ok {
        if n, subItem.Sub_id, err = bstd.UnmarshalInt32(n, b); err != nil {
            return
        }
    }
    if n, ok, err = bgenimpl.HandleCompatibility(n, b, subItemRIds, 2); err != nil {
        if err == bgenimpl.ErrEof {
            return n, nil
        }
        return
    }
    if ok {
        if n, subItem.Description, err = bstd.UnmarshalString(n, b); err != nil {
            return
        }
    }
    if n, ok, err = bgenimpl.HandleCompatibility(n, b, subItemRIds, 3); err != nil {
        if err == bgenimpl.ErrEof {
            return n, nil
        }
        return
    }
    if ok {
        if n, subItem.Sub_items, err = bstd.UnmarshalSlice[SubSubItem](n, b, func (n int, b []byte, s *SubSubItem) (int, error) { return s.UnmarshalPlain(n, b) }); err != nil {
            return
        }
    }
    n += 2
    return
}

// UnmarshalPlain - SubItem
func (subItem *SubItem) UnmarshalPlain(tn int, b []byte) (n int, err error) {
    n = tn
    if n, subItem.Sub_id, err = bstd.UnmarshalInt32(n, b); err != nil {
        return
    }
    if n, subItem.Description, err = bstd.UnmarshalString(n, b); err != nil {
        return
    }
    if n, subItem.Sub_items, err = bstd.UnmarshalSlice[SubSubItem](n, b, func (n int, b []byte, s *SubSubItem) (int, error) { return s.UnmarshalPlain(n, b) }); err != nil {
        return
    }
    return
}

// Struct - SubSubItem
type SubSubItem struct {
    Sub_sub_id string
    Sub_sub_data []byte
}

// Reserved Ids - SubSubItem
var subSubItemRIds = []uint16{}

// Size - SubSubItem
func (subSubItem *SubSubItem) Size() int {
    return subSubItem.size(0)
}

// Nested Size - SubSubItem
func (subSubItem *SubSubItem) size(id uint16) (s int) {
    s += bstd.SizeString(subSubItem.Sub_sub_id) + 2
    s += bstd.SizeBytes(subSubItem.Sub_sub_data) + 2

    if id > 255 {
        s += 5
        return
    }
    s += 4
    return
}

// SizePlain - SubSubItem
func (subSubItem *SubSubItem) SizePlain() (s int) {
    s += bstd.SizeString(subSubItem.Sub_sub_id)
    s += bstd.SizeBytes(subSubItem.Sub_sub_data)
    return
}

// Marshal - SubSubItem
func (subSubItem *SubSubItem) Marshal(b []byte) {
    subSubItem.marshal(0, b, 0)
}

// Nested Marshal - SubSubItem
func (subSubItem *SubSubItem) marshal(tn int, b []byte, id uint16) (n int) {
    n = bgenimpl.MarshalTag(tn, b, bgenimpl.Container, id)
    n = bgenimpl.MarshalTag(n, b, bgenimpl.Bytes, 1)
    n = bstd.MarshalString(n, b, subSubItem.Sub_sub_id)
    n = bgenimpl.MarshalTag(n, b, bgenimpl.Bytes, 2)
    n = bstd.MarshalBytes(n, b, subSubItem.Sub_sub_data)

    n += 2
    b[n-2] = 1
    b[n-1] = 1
    return
}

// MarshalPlain - SubSubItem
func (subSubItem *SubSubItem) MarshalPlain(tn int, b []byte) (n int) {
    n = tn
    n = bstd.MarshalString(n, b, subSubItem.Sub_sub_id)
    n = bstd.MarshalBytes(n, b, subSubItem.Sub_sub_data)
    return n
}

// Unmarshal - SubSubItem
func (subSubItem *SubSubItem) Unmarshal(b []byte) (err error) {
    _, err = subSubItem.unmarshal(0, b, []uint16{}, 0)
    return
}

// Nested Unmarshal - SubSubItem
func (subSubItem *SubSubItem) unmarshal(tn int, b []byte, r []uint16, id uint16) (n int, err error) {
    var ok bool
    if n, ok, err = bgenimpl.HandleCompatibility(tn, b, r, id); !ok {
        if err == bgenimpl.ErrEof {
            return n, nil
        }
        return
    }
    if n, ok, err = bgenimpl.HandleCompatibility(n, b, subSubItemRIds, 1); err != nil {
        if err == bgenimpl.ErrEof {
            return n, nil
        }
        return
    }
    if ok {
        if n, subSubItem.Sub_sub_id, err = bstd.UnmarshalString(n, b); err != nil {
            return
        }
    }
    if n, ok, err = bgenimpl.HandleCompatibility(n, b, subSubItemRIds, 2); err != nil {
        if err == bgenimpl.ErrEof {
            return n, nil
        }
        return
    }
    if ok {
        if n, subSubItem.Sub_sub_data, err = bstd.UnmarshalBytes(n, b); err != nil {
            return
        }
    }
    n += 2
    return
}

// UnmarshalPlain - SubSubItem
func (subSubItem *SubSubItem) UnmarshalPlain(tn int, b []byte) (n int, err error) {
    n = tn
    if n, subSubItem.Sub_sub_id, err = bstd.UnmarshalString(n, b); err != nil {
        return
    }
    if n, subSubItem.Sub_sub_data, err = bstd.UnmarshalBytes(n, b); err != nil {
        return
    }
    return
}

// Struct - SubComplexData
type SubComplexData struct {
    Sub_id int32
    Sub_title string
    Sub_binary_data [][]byte
    Sub_items []SubItem
    Sub_metadata map[string]string
}

// Reserved Ids - SubComplexData
var subComplexDataRIds = []uint16{}

// Size - SubComplexData
func (subComplexData *SubComplexData) Size() int {
    return subComplexData.size(0)
}

// Nested Size - SubComplexData
func (subComplexData *SubComplexData) size(id uint16) (s int) {
    s += bstd.SizeInt32() + 2
    s += bstd.SizeString(subComplexData.Sub_title) + 2
    s += bstd.SizeSlice(subComplexData.Sub_binary_data, bstd.SizeBytes) + 2
    s += bstd.SizeSlice(subComplexData.Sub_items, func (s SubItem) int { return s.SizePlain() }) + 2
    s += bstd.SizeMap(subComplexData.Sub_metadata, bstd.SizeString, bstd.SizeString) + 2

    if id > 255 {
        s += 5
        return
    }
    s += 4
    return
}

// SizePlain - SubComplexData
func (subComplexData *SubComplexData) SizePlain() (s int) {
    s += bstd.SizeInt32()
    s += bstd.SizeString(subComplexData.Sub_title)
    s += bstd.SizeSlice(subComplexData.Sub_binary_data, bstd.SizeBytes)
    s += bstd.SizeSlice(subComplexData.Sub_items, func (s SubItem) int { return s.SizePlain() })
    s += bstd.SizeMap(subComplexData.Sub_metadata, bstd.SizeString, bstd.SizeString)
    return
}

// Marshal - SubComplexData
func (subComplexData *SubComplexData) Marshal(b []byte) {
    subComplexData.marshal(0, b, 0)
}

// Nested Marshal - SubComplexData
func (subComplexData *SubComplexData) marshal(tn int, b []byte, id uint16) (n int) {
    n = bgenimpl.MarshalTag(tn, b, bgenimpl.Container, id)
    n = bgenimpl.MarshalTag(n, b, bgenimpl.Fixed32, 1)
    n = bstd.MarshalInt32(n, b, subComplexData.Sub_id)
    n = bgenimpl.MarshalTag(n, b, bgenimpl.Bytes, 2)
    n = bstd.MarshalString(n, b, subComplexData.Sub_title)
    n = bgenimpl.MarshalTag(n, b, bgenimpl.Array, 3)
    n = bstd.MarshalSlice(n, b, subComplexData.Sub_binary_data, bstd.MarshalBytes)
    n = bgenimpl.MarshalTag(n, b, bgenimpl.Array, 4)
    n = bstd.MarshalSlice(n, b, subComplexData.Sub_items, func (n int, b []byte, s SubItem) int { return s.MarshalPlain(n, b) })
    n = bgenimpl.MarshalTag(n, b, bgenimpl.Array, 5)
    n = bstd.MarshalMap(n, b, subComplexData.Sub_metadata, bstd.MarshalString, bstd.MarshalString)

    n += 2
    b[n-2] = 1
    b[n-1] = 1
    return
}

// MarshalPlain - SubComplexData
func (subComplexData *SubComplexData) MarshalPlain(tn int, b []byte) (n int) {
    n = tn
    n = bstd.MarshalInt32(n, b, subComplexData.Sub_id)
    n = bstd.MarshalString(n, b, subComplexData.Sub_title)
    n = bstd.MarshalSlice(n, b, subComplexData.Sub_binary_data, bstd.MarshalBytes)
    n = bstd.MarshalSlice(n, b, subComplexData.Sub_items, func (n int, b []byte, s SubItem) int { return s.MarshalPlain(n, b) })
    n = bstd.MarshalMap(n, b, subComplexData.Sub_metadata, bstd.MarshalString, bstd.MarshalString)
    return n
}

// Unmarshal - SubComplexData
func (subComplexData *SubComplexData) Unmarshal(b []byte) (err error) {
    _, err = subComplexData.unmarshal(0, b, []uint16{}, 0)
    return
}

// Nested Unmarshal - SubComplexData
func (subComplexData *SubComplexData) unmarshal(tn int, b []byte, r []uint16, id uint16) (n int, err error) {
    var ok bool
    if n, ok, err = bgenimpl.HandleCompatibility(tn, b, r, id); !ok {
        if err == bgenimpl.ErrEof {
            return n, nil
        }
        return
    }
    if n, ok, err = bgenimpl.HandleCompatibility(n, b, subComplexDataRIds, 1); err != nil {
        if err == bgenimpl.ErrEof {
            return n, nil
        }
        return
    }
    if ok {
        if n, subComplexData.Sub_id, err = bstd.UnmarshalInt32(n, b); err != nil {
            return
        }
    }
    if n, ok, err = bgenimpl.HandleCompatibility(n, b, subComplexDataRIds, 2); err != nil {
        if err == bgenimpl.ErrEof {
            return n, nil
        }
        return
    }
    if ok {
        if n, subComplexData.Sub_title, err = bstd.UnmarshalString(n, b); err != nil {
            return
        }
    }
    if n, ok, err = bgenimpl.HandleCompatibility(n, b, subComplexDataRIds, 3); err != nil {
        if err == bgenimpl.ErrEof {
            return n, nil
        }
        return
    }
    if ok {
        if n, subComplexData.Sub_binary_data, err = bstd.UnmarshalSlice[[]byte](n, b, bstd.UnmarshalBytes); err != nil {
            return
        }
    }
    if n, ok, err = bgenimpl.HandleCompatibility(n, b, subComplexDataRIds, 4); err != nil {
        if err == bgenimpl.ErrEof {
            return n, nil
        }
        return
    }
    if ok {
        if n, subComplexData.Sub_items, err = bstd.UnmarshalSlice[SubItem](n, b, func (n int, b []byte, s *SubItem) (int, error) { return s.UnmarshalPlain(n, b) }); err != nil {
            return
        }
    }
    if n, ok, err = bgenimpl.HandleCompatibility(n, b, subComplexDataRIds, 5); err != nil {
        if err == bgenimpl.ErrEof {
            return n, nil
        }
        return
    }
    if ok {
        if n, subComplexData.Sub_metadata, err = bstd.UnmarshalMap[string, string](n, b, bstd.UnmarshalString, bstd.UnmarshalString); err != nil {
            return
        }
    }
    n += 2
    return
}

// UnmarshalPlain - SubComplexData
func (subComplexData *SubComplexData) UnmarshalPlain(tn int, b []byte) (n int, err error) {
    n = tn
    if n, subComplexData.Sub_id, err = bstd.UnmarshalInt32(n, b); err != nil {
        return
    }
    if n, subComplexData.Sub_title, err = bstd.UnmarshalString(n, b); err != nil {
        return
    }
    if n, subComplexData.Sub_binary_data, err = bstd.UnmarshalSlice[[]byte](n, b, bstd.UnmarshalBytes); err != nil {
        return
    }
    if n, subComplexData.Sub_items, err = bstd.UnmarshalSlice[SubItem](n, b, func (n int, b []byte, s *SubItem) (int, error) { return s.UnmarshalPlain(n, b) }); err != nil {
        return
    }
    if n, subComplexData.Sub_metadata, err = bstd.UnmarshalMap[string, string](n, b, bstd.UnmarshalString, bstd.UnmarshalString); err != nil {
        return
    }
    return
}

