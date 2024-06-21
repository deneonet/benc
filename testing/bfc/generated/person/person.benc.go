// Code generated by bencgen golang. DO NOT EDIT.
// source: schemas/person.benc

package person

import (
    "go.kine.bz/benc/std"
    "go.kine.bz/benc/impl/gen"
)

// Struct - Person
type Person struct {
    rIds []uint16

    Age byte
    Name string
    Parents Parents
    Child Child
}

// Reserved Ids - Person
var personRIds = []uint16{}

// Size - Person
func (person *Person) Size(id uint16) (s int, err error) {
    var ts int
    s += bstd.SizeByte() + 2
    if ts, err = bstd.SizeString(person.Name); err != nil {
        return
    }
    s += ts + 2
    if ts, err = person.Parents.Size(3); err != nil {
        return
    }
    s += ts
    if ts, err = person.Child.Size(4); err != nil {
        return
    }
    s += ts

    _ = ts
    if id > 255 {
        s += 5
        return
    }
    s += 4
    return
}

// Marshal - Person
func (person *Person) Marshal(tn int, b []byte, id uint16) (n int, err error) {
    n = bgenimpl.MarshalTag(tn, b, bgenimpl.Container, id)
    n = bgenimpl.MarshalTag(n, b, bgenimpl.Fixed8, 1)
    n = bstd.MarshalByte(n, b, person.Age)
    n = bgenimpl.MarshalTag(n, b, bgenimpl.Bytes, 2)
    if n, err = bstd.MarshalString(n, b, person.Name); err != nil {
        return
    }
    if n, err = person.Parents.Marshal(n, b, 3); err != nil {
        return
    }
    if n, err = person.Child.Marshal(n, b, 4); err != nil {
        return
    }

    n += 2
    b[n-2] = 1
    b[n-1] = 1
    return
}

// Nested Unmarshal - Person
func (person *Person) unmarshal(n int, b []byte, r []uint16, id uint16) (int, error) {
    person.rIds = r
    return person.Unmarshal(n, b, id)
}

// Unmarshal - Person
func (person *Person) Unmarshal(tn int, b []byte, id uint16) (n int, err error) {
    var ok bool
    if n, ok, err = bgenimpl.HandleCompatibility(tn, b, person.rIds, id); !ok {
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
    if n, err = person.Parents.unmarshal(n, b, personRIds, 3); err != nil {
        return
    }
    if n, err = person.Child.unmarshal(n, b, personRIds, 4); err != nil {
        return
    }
    n += 2
    return
}

// Struct - Child
type Child struct {
    rIds []uint16

    Age byte
    Name string
    Parents Parents
}

// Reserved Ids - Child
var childRIds = []uint16{}

// Size - Child
func (child *Child) Size(id uint16) (s int, err error) {
    var ts int
    s += bstd.SizeByte() + 2
    if ts, err = bstd.SizeString(child.Name); err != nil {
        return
    }
    s += ts + 2
    if ts, err = child.Parents.Size(3); err != nil {
        return
    }
    s += ts

    _ = ts
    if id > 255 {
        s += 5
        return
    }
    s += 4
    return
}

// Marshal - Child
func (child *Child) Marshal(tn int, b []byte, id uint16) (n int, err error) {
    n = bgenimpl.MarshalTag(tn, b, bgenimpl.Container, id)
    n = bgenimpl.MarshalTag(n, b, bgenimpl.Fixed8, 1)
    n = bstd.MarshalByte(n, b, child.Age)
    n = bgenimpl.MarshalTag(n, b, bgenimpl.Bytes, 2)
    if n, err = bstd.MarshalString(n, b, child.Name); err != nil {
        return
    }
    if n, err = child.Parents.Marshal(n, b, 3); err != nil {
        return
    }

    n += 2
    b[n-2] = 1
    b[n-1] = 1
    return
}

// Nested Unmarshal - Child
func (child *Child) unmarshal(n int, b []byte, r []uint16, id uint16) (int, error) {
    child.rIds = r
    return child.Unmarshal(n, b, id)
}

// Unmarshal - Child
func (child *Child) Unmarshal(tn int, b []byte, id uint16) (n int, err error) {
    var ok bool
    if n, ok, err = bgenimpl.HandleCompatibility(tn, b, child.rIds, id); !ok {
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

// Struct - Parents
type Parents struct {
    rIds []uint16

    Mother string
    Father string
}

// Reserved Ids - Parents
var parentsRIds = []uint16{}

// Size - Parents
func (parents *Parents) Size(id uint16) (s int, err error) {
    var ts int
    if ts, err = bstd.SizeString(parents.Mother); err != nil {
        return
    }
    s += ts + 2
    if ts, err = bstd.SizeString(parents.Father); err != nil {
        return
    }
    s += ts + 2

    _ = ts
    if id > 255 {
        s += 5
        return
    }
    s += 4
    return
}

// Marshal - Parents
func (parents *Parents) Marshal(tn int, b []byte, id uint16) (n int, err error) {
    n = bgenimpl.MarshalTag(tn, b, bgenimpl.Container, id)
    n = bgenimpl.MarshalTag(n, b, bgenimpl.Bytes, 1)
    if n, err = bstd.MarshalString(n, b, parents.Mother); err != nil {
        return
    }
    n = bgenimpl.MarshalTag(n, b, bgenimpl.Bytes, 2)
    if n, err = bstd.MarshalString(n, b, parents.Father); err != nil {
        return
    }

    n += 2
    b[n-2] = 1
    b[n-1] = 1
    return
}

// Nested Unmarshal - Parents
func (parents *Parents) unmarshal(n int, b []byte, r []uint16, id uint16) (int, error) {
    parents.rIds = r
    return parents.Unmarshal(n, b, id)
}

// Unmarshal - Parents
func (parents *Parents) Unmarshal(tn int, b []byte, id uint16) (n int, err error) {
    var ok bool
    if n, ok, err = bgenimpl.HandleCompatibility(tn, b, parents.rIds, id); !ok {
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
