// Code generated by bencgen golang. DO NOT EDIT.
// source: schemas/person2.benc

package person2

import (
    "go.kine.bz/benc/std"
    "go.kine.bz/benc/impl/gen"
)

// Struct - Person
type Person struct {
    Age byte
    Name string
    Child Child
}

// Reserved Ids - Person
var personRIds = []uint16{3}

// Size - Person
func (person *Person) Size() int {
    return person.size(0)
}

// Nested Size - Person
func (person *Person) size(id uint16) (s int) {
    s += bstd.SizeByte() + 2
    s += bstd.SizeString(person.Name) + 2
    s += person.Child.size(4)

    if id > 255 {
        s += 5
        return
    }
    s += 4
    return
}

// SizePlain - Person
func (person *Person) SizePlain() (s int) {
    s += bstd.SizeByte()
    s += bstd.SizeString(person.Name)
    s += person.Child.SizePlain()
    return
}

// Marshal - Person
func (person *Person) Marshal(b []byte) {
    person.marshal(0, b, 0)
}

// Nested Marshal - Person
func (person *Person) marshal(tn int, b []byte, id uint16) (n int) {
    n = bgenimpl.MarshalTag(tn, b, bgenimpl.Container, id)
    n = bgenimpl.MarshalTag(n, b, bgenimpl.Fixed8, 1)
    n = bstd.MarshalByte(n, b, person.Age)
    n = bgenimpl.MarshalTag(n, b, bgenimpl.Bytes, 2)
    n = bstd.MarshalString(n, b, person.Name)
    n = person.Child.marshal(n, b, 4)

    n += 2
    b[n-2] = 1
    b[n-1] = 1
    return
}

// MarshalPlain - Person
func (person *Person) MarshalPlain(tn int, b []byte) (n int) {
    n = tn
    n = bstd.MarshalByte(n, b, person.Age)
    n = bstd.MarshalString(n, b, person.Name)
    n = person.Child.MarshalPlain(n, b)
    return n
}

// Unmarshal - Person
func (person *Person) Unmarshal(b []byte) (err error) {
    _, err = person.unmarshal(0, b, []uint16{}, 0)
    return
}

// Nested Unmarshal - Person
func (person *Person) unmarshal(tn int, b []byte, r []uint16, id uint16) (n int, err error) {
    var ok bool
    if n, ok, err = bgenimpl.HandleCompatibility(tn, b, r, id); !ok {
        if err == bgenimpl.ErrEof {
            return n, nil
        }
        return
    }
    if n, ok, err = bgenimpl.HandleCompatibility(n, b, personRIds, 1); err != nil {
        if err == bgenimpl.ErrEof {
            return n, nil
        }
        return
    }
    if ok {
        if n, person.Age, err = bstd.UnmarshalByte(n, b); err != nil {
            return
        }
    }
    if n, ok, err = bgenimpl.HandleCompatibility(n, b, personRIds, 2); err != nil {
        if err == bgenimpl.ErrEof {
            return n, nil
        }
        return
    }
    if ok {
        if n, person.Name, err = bstd.UnmarshalString(n, b); err != nil {
            return
        }
    }
    if n, err = person.Child.unmarshal(n, b, personRIds, 4); err != nil {
        return
    }
    n += 2
    return
}

// UnmarshalPlain - Person
func (person *Person) UnmarshalPlain(tn int, b []byte) (n int, err error) {
    n = tn
    if n, person.Age, err = bstd.UnmarshalByte(n, b); err != nil {
        return
    }
    if n, person.Name, err = bstd.UnmarshalString(n, b); err != nil {
        return
    }
    if n, err = person.Child.UnmarshalPlain(n, b); err != nil {
        return
    }
    return
}

// Struct - Child
type Child struct {
    Age byte
    Name string
    Parents Parents
}

// Reserved Ids - Child
var childRIds = []uint16{}

// Size - Child
func (child *Child) Size() int {
    return child.size(0)
}

// Nested Size - Child
func (child *Child) size(id uint16) (s int) {
    s += bstd.SizeByte() + 2
    s += bstd.SizeString(child.Name) + 2
    s += child.Parents.size(3)

    if id > 255 {
        s += 5
        return
    }
    s += 4
    return
}

// SizePlain - Child
func (child *Child) SizePlain() (s int) {
    s += bstd.SizeByte()
    s += bstd.SizeString(child.Name)
    s += child.Parents.SizePlain()
    return
}

// Marshal - Child
func (child *Child) Marshal(b []byte) {
    child.marshal(0, b, 0)
}

// Nested Marshal - Child
func (child *Child) marshal(tn int, b []byte, id uint16) (n int) {
    n = bgenimpl.MarshalTag(tn, b, bgenimpl.Container, id)
    n = bgenimpl.MarshalTag(n, b, bgenimpl.Fixed8, 1)
    n = bstd.MarshalByte(n, b, child.Age)
    n = bgenimpl.MarshalTag(n, b, bgenimpl.Bytes, 2)
    n = bstd.MarshalString(n, b, child.Name)
    n = child.Parents.marshal(n, b, 3)

    n += 2
    b[n-2] = 1
    b[n-1] = 1
    return
}

// MarshalPlain - Child
func (child *Child) MarshalPlain(tn int, b []byte) (n int) {
    n = tn
    n = bstd.MarshalByte(n, b, child.Age)
    n = bstd.MarshalString(n, b, child.Name)
    n = child.Parents.MarshalPlain(n, b)
    return n
}

// Unmarshal - Child
func (child *Child) Unmarshal(b []byte) (err error) {
    _, err = child.unmarshal(0, b, []uint16{}, 0)
    return
}

// Nested Unmarshal - Child
func (child *Child) unmarshal(tn int, b []byte, r []uint16, id uint16) (n int, err error) {
    var ok bool
    if n, ok, err = bgenimpl.HandleCompatibility(tn, b, r, id); !ok {
        if err == bgenimpl.ErrEof {
            return n, nil
        }
        return
    }
    if n, ok, err = bgenimpl.HandleCompatibility(n, b, childRIds, 1); err != nil {
        if err == bgenimpl.ErrEof {
            return n, nil
        }
        return
    }
    if ok {
        if n, child.Age, err = bstd.UnmarshalByte(n, b); err != nil {
            return
        }
    }
    if n, ok, err = bgenimpl.HandleCompatibility(n, b, childRIds, 2); err != nil {
        if err == bgenimpl.ErrEof {
            return n, nil
        }
        return
    }
    if ok {
        if n, child.Name, err = bstd.UnmarshalString(n, b); err != nil {
            return
        }
    }
    if n, err = child.Parents.unmarshal(n, b, childRIds, 3); err != nil {
        return
    }
    n += 2
    return
}

// UnmarshalPlain - Child
func (child *Child) UnmarshalPlain(tn int, b []byte) (n int, err error) {
    n = tn
    if n, child.Age, err = bstd.UnmarshalByte(n, b); err != nil {
        return
    }
    if n, child.Name, err = bstd.UnmarshalString(n, b); err != nil {
        return
    }
    if n, err = child.Parents.UnmarshalPlain(n, b); err != nil {
        return
    }
    return
}

// Struct - Parents
type Parents struct {
    Mother string
    Father string
}

// Reserved Ids - Parents
var parentsRIds = []uint16{}

// Size - Parents
func (parents *Parents) Size() int {
    return parents.size(0)
}

// Nested Size - Parents
func (parents *Parents) size(id uint16) (s int) {
    s += bstd.SizeString(parents.Mother) + 2
    s += bstd.SizeString(parents.Father) + 2

    if id > 255 {
        s += 5
        return
    }
    s += 4
    return
}

// SizePlain - Parents
func (parents *Parents) SizePlain() (s int) {
    s += bstd.SizeString(parents.Mother)
    s += bstd.SizeString(parents.Father)
    return
}

// Marshal - Parents
func (parents *Parents) Marshal(b []byte) {
    parents.marshal(0, b, 0)
}

// Nested Marshal - Parents
func (parents *Parents) marshal(tn int, b []byte, id uint16) (n int) {
    n = bgenimpl.MarshalTag(tn, b, bgenimpl.Container, id)
    n = bgenimpl.MarshalTag(n, b, bgenimpl.Bytes, 1)
    n = bstd.MarshalString(n, b, parents.Mother)
    n = bgenimpl.MarshalTag(n, b, bgenimpl.Bytes, 2)
    n = bstd.MarshalString(n, b, parents.Father)

    n += 2
    b[n-2] = 1
    b[n-1] = 1
    return
}

// MarshalPlain - Parents
func (parents *Parents) MarshalPlain(tn int, b []byte) (n int) {
    n = tn
    n = bstd.MarshalString(n, b, parents.Mother)
    n = bstd.MarshalString(n, b, parents.Father)
    return n
}

// Unmarshal - Parents
func (parents *Parents) Unmarshal(b []byte) (err error) {
    _, err = parents.unmarshal(0, b, []uint16{}, 0)
    return
}

// Nested Unmarshal - Parents
func (parents *Parents) unmarshal(tn int, b []byte, r []uint16, id uint16) (n int, err error) {
    var ok bool
    if n, ok, err = bgenimpl.HandleCompatibility(tn, b, r, id); !ok {
        if err == bgenimpl.ErrEof {
            return n, nil
        }
        return
    }
    if n, ok, err = bgenimpl.HandleCompatibility(n, b, parentsRIds, 1); err != nil {
        if err == bgenimpl.ErrEof {
            return n, nil
        }
        return
    }
    if ok {
        if n, parents.Mother, err = bstd.UnmarshalString(n, b); err != nil {
            return
        }
    }
    if n, ok, err = bgenimpl.HandleCompatibility(n, b, parentsRIds, 2); err != nil {
        if err == bgenimpl.ErrEof {
            return n, nil
        }
        return
    }
    if ok {
        if n, parents.Father, err = bstd.UnmarshalString(n, b); err != nil {
            return
        }
    }
    n += 2
    return
}

// UnmarshalPlain - Parents
func (parents *Parents) UnmarshalPlain(tn int, b []byte) (n int, err error) {
    n = tn
    if n, parents.Mother, err = bstd.UnmarshalString(n, b); err != nil {
        return
    }
    if n, parents.Father, err = bstd.UnmarshalString(n, b); err != nil {
        return
    }
    return
}

