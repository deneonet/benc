// Code generated by bencgen go. DO NOT EDIT.
// source: ../schemas/others.benc

package others

import (
    "github.com/deneonet/benc/std"
    "github.com/deneonet/benc/impl/gen"
)

// Enum - ExampleEnum
type ExampleEnum int
const (
    ExampleEnumOne ExampleEnum = iota
    ExampleEnumTwo
    ExampleEnumThree
    ExampleEnumFour
)

// Enum - ExampleEnum2
type ExampleEnum2 int
const (
    ExampleEnum2Five ExampleEnum2 = iota
    ExampleEnum2Six
)

// Struct - OthersTest
type OthersTest struct {
    Ui64 uint64
    Ui64Arr []uint64
    Ui64Map map[uint64]uint32
    Ui32 uint32
    Ui16 uint16
    Ui uint
    ExampleEnum ExampleEnum
    ExampleEnum2 ExampleEnum2
}

// Reserved Ids - OthersTest
var othersTestRIds = []uint16{}

// Size - OthersTest
func (othersTest *OthersTest) Size() int {
    return othersTest.size(0)
}

// Nested Size - OthersTest
func (othersTest *OthersTest) size(id uint16) (s int) {
    s += bstd.SizeUint64() + 2
    s += bstd.SizeSlice(othersTest.Ui64Arr, bstd.SizeUint64) + 2
    s += bstd.SizeMap(othersTest.Ui64Map, bstd.SizeUint64, bstd.SizeUint32) + 2
    s += bstd.SizeUint32() + 2
    s += bstd.SizeUint16() + 2
    s += bstd.SizeUint(othersTest.Ui) + 2
    s += bgenimpl.SizeEnum(othersTest.ExampleEnum) + 2
    s += bgenimpl.SizeEnum(othersTest.ExampleEnum2) + 2

    if id > 255 {
        s += 5
        return
    }
    s += 4
    return
}

// SizePlain - OthersTest
func (othersTest *OthersTest) SizePlain() (s int) {
    s += bstd.SizeUint64()
    s += bstd.SizeSlice(othersTest.Ui64Arr, bstd.SizeUint64)
    s += bstd.SizeMap(othersTest.Ui64Map, bstd.SizeUint64, bstd.SizeUint32)
    s += bstd.SizeUint32()
    s += bstd.SizeUint16()
    s += bstd.SizeUint(othersTest.Ui)
    s += bgenimpl.SizeEnum(othersTest.ExampleEnum)
    s += bgenimpl.SizeEnum(othersTest.ExampleEnum2)
    return
}

// Marshal - OthersTest
func (othersTest *OthersTest) Marshal(b []byte) {
    othersTest.marshal(0, b, 0)
}

// Nested Marshal - OthersTest
func (othersTest *OthersTest) marshal(tn int, b []byte, id uint16) (n int) {
    n = bgenimpl.MarshalTag(tn, b, bgenimpl.Container, id)
    n = bgenimpl.MarshalTag(n, b, bgenimpl.Fixed64, 1)
    n = bstd.MarshalUint64(n, b, othersTest.Ui64)
    n = bgenimpl.MarshalTag(n, b, bgenimpl.ArrayMap, 2)
    n = bstd.MarshalSlice(n, b, othersTest.Ui64Arr, bstd.MarshalUint64)
    n = bgenimpl.MarshalTag(n, b, bgenimpl.ArrayMap, 3)
    n = bstd.MarshalMap(n, b, othersTest.Ui64Map, bstd.MarshalUint64, bstd.MarshalUint32)
    n = bgenimpl.MarshalTag(n, b, bgenimpl.Fixed32, 4)
    n = bstd.MarshalUint32(n, b, othersTest.Ui32)
    n = bgenimpl.MarshalTag(n, b, bgenimpl.Fixed16, 5)
    n = bstd.MarshalUint16(n, b, othersTest.Ui16)
    n = bgenimpl.MarshalTag(n, b, bgenimpl.Varint, 7)
    n = bstd.MarshalUint(n, b, othersTest.Ui)
    n = bgenimpl.MarshalTag(n, b, bgenimpl.ArrayMap, 8)
    n = bgenimpl.MarshalEnum(n, b, othersTest.ExampleEnum)
    n = bgenimpl.MarshalTag(n, b, bgenimpl.ArrayMap, 9)
    n = bgenimpl.MarshalEnum(n, b, othersTest.ExampleEnum2)

    n += 2
    b[n-2] = 1
    b[n-1] = 1
    return
}

// MarshalPlain - OthersTest
func (othersTest *OthersTest) MarshalPlain(tn int, b []byte) (n int) {
    n = tn
    n = bstd.MarshalUint64(n, b, othersTest.Ui64)
    n = bstd.MarshalSlice(n, b, othersTest.Ui64Arr, bstd.MarshalUint64)
    n = bstd.MarshalMap(n, b, othersTest.Ui64Map, bstd.MarshalUint64, bstd.MarshalUint32)
    n = bstd.MarshalUint32(n, b, othersTest.Ui32)
    n = bstd.MarshalUint16(n, b, othersTest.Ui16)
    n = bstd.MarshalUint(n, b, othersTest.Ui)
    n = bgenimpl.MarshalEnum(n, b, othersTest.ExampleEnum)
    n = bgenimpl.MarshalEnum(n, b, othersTest.ExampleEnum2)
    return n
}

// Unmarshal - OthersTest
func (othersTest *OthersTest) Unmarshal(b []byte) (err error) {
    _, err = othersTest.unmarshal(0, b, []uint16{}, 0)
    return
}

// Nested Unmarshal - OthersTest
func (othersTest *OthersTest) unmarshal(tn int, b []byte, r []uint16, id uint16) (n int, err error) {
    var ok bool
    if n, ok, err = bgenimpl.HandleCompatibility(tn, b, r, id); !ok {
        if err == bgenimpl.ErrEof {
            return n, nil
        }
        return
    }
    if n, ok, err = bgenimpl.HandleCompatibility(n, b, othersTestRIds, 1); err != nil {
        if err == bgenimpl.ErrEof {
            return n, nil
        }
        return
    }
    if ok {
        if n, othersTest.Ui64, err = bstd.UnmarshalUint64(n, b); err != nil {
            return
        }
    }
    if n, ok, err = bgenimpl.HandleCompatibility(n, b, othersTestRIds, 2); err != nil {
        if err == bgenimpl.ErrEof {
            return n, nil
        }
        return
    }
    if ok {
        if n, othersTest.Ui64Arr, err = bstd.UnmarshalSlice[uint64](n, b, bstd.UnmarshalUint64); err != nil {
            return
        }
    }
    if n, ok, err = bgenimpl.HandleCompatibility(n, b, othersTestRIds, 3); err != nil {
        if err == bgenimpl.ErrEof {
            return n, nil
        }
        return
    }
    if ok {
        if n, othersTest.Ui64Map, err = bstd.UnmarshalMap[uint64, uint32](n, b, bstd.UnmarshalUint64, bstd.UnmarshalUint32); err != nil {
            return
        }
    }
    if n, ok, err = bgenimpl.HandleCompatibility(n, b, othersTestRIds, 4); err != nil {
        if err == bgenimpl.ErrEof {
            return n, nil
        }
        return
    }
    if ok {
        if n, othersTest.Ui32, err = bstd.UnmarshalUint32(n, b); err != nil {
            return
        }
    }
    if n, ok, err = bgenimpl.HandleCompatibility(n, b, othersTestRIds, 5); err != nil {
        if err == bgenimpl.ErrEof {
            return n, nil
        }
        return
    }
    if ok {
        if n, othersTest.Ui16, err = bstd.UnmarshalUint16(n, b); err != nil {
            return
        }
    }
    if n, ok, err = bgenimpl.HandleCompatibility(n, b, othersTestRIds, 7); err != nil {
        if err == bgenimpl.ErrEof {
            return n, nil
        }
        return
    }
    if ok {
        if n, othersTest.Ui, err = bstd.UnmarshalUint(n, b); err != nil {
            return
        }
    }
    if n, ok, err = bgenimpl.HandleCompatibility(n, b, othersTestRIds, 8); err != nil {
        if err == bgenimpl.ErrEof {
            return n, nil
        }
        return
    }
    if ok {
        if n, othersTest.ExampleEnum, err = bgenimpl.UnmarshalEnum[ExampleEnum](n, b); err != nil {
            return
        }
    }
    if n, ok, err = bgenimpl.HandleCompatibility(n, b, othersTestRIds, 9); err != nil {
        if err == bgenimpl.ErrEof {
            return n, nil
        }
        return
    }
    if ok {
        if n, othersTest.ExampleEnum2, err = bgenimpl.UnmarshalEnum[ExampleEnum2](n, b); err != nil {
            return
        }
    }
    n += 2
    return
}

// UnmarshalPlain - OthersTest
func (othersTest *OthersTest) UnmarshalPlain(tn int, b []byte) (n int, err error) {
    n = tn
    if n, othersTest.Ui64, err = bstd.UnmarshalUint64(n, b); err != nil {
        return
    }
    if n, othersTest.Ui64Arr, err = bstd.UnmarshalSlice[uint64](n, b, bstd.UnmarshalUint64); err != nil {
        return
    }
    if n, othersTest.Ui64Map, err = bstd.UnmarshalMap[uint64, uint32](n, b, bstd.UnmarshalUint64, bstd.UnmarshalUint32); err != nil {
        return
    }
    if n, othersTest.Ui32, err = bstd.UnmarshalUint32(n, b); err != nil {
        return
    }
    if n, othersTest.Ui16, err = bstd.UnmarshalUint16(n, b); err != nil {
        return
    }
    if n, othersTest.Ui, err = bstd.UnmarshalUint(n, b); err != nil {
        return
    }
    if n, othersTest.ExampleEnum, err = bgenimpl.UnmarshalEnum[ExampleEnum](n, b); err != nil {
        return
    }
    if n, othersTest.ExampleEnum2, err = bgenimpl.UnmarshalEnum[ExampleEnum2](n, b); err != nil {
        return
    }
    return
}

