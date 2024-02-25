# BENC

[![ko-fi](https://ko-fi.com/img/githubbutton_sm.svg)](https://ko-fi.com/U7U4T5BU3)

![go workflow](https://github.com/deneonet/benc/actions/workflows/go.yml/badge.svg)
[![go report card](https://goreportcard.com/badge/github.com/deneonet/benc)](https://goreportcard.com/report/github.com/deneonet/benc)
[![go reference](https://pkg.go.dev/badge/github.com/deneonet/benc.svg)](https://pkg.go.dev/github.com/deneonet/benc)

The fastest serializer in pure Golang.

## Features

- Fastest serialization (fastest serializer out there)
- [Slices and Maps support (manual)](#slices-and-maps-serialization)
- [Struct support (manual)](#struct-serialization)
- [Message framing support](#message-framing)
- [Tagging support](#tagging)
- [Pre-Allocation/Buffer reuse](#pre-allocation)
- [Out of Order Deserialization](#out-of-order-deserialization)
- [DataType validation](#datatype-validation)

## Changelog

- [See changelog here](CHANGELOG.md)

## Benchmarks

- [See benchmarks here](https://github.com/alecthomas/go_serialization_benchmarks)

## Best Practices
- [See best practices here](BESTPRACTICES.md)

## Installation

Install BENC in any Golang Project

```bash
go get github.com/deneonet/benc
```

## Struct serialization

[With DataType validation in the unmarshal process](#datatype-validation)

```go
package main

import (
	bstd "github.com/deneonet/benc"
)

type TestData struct {
	str  string
	id   uint64
	cool bool
}

func MarshalTestData(t *TestData) (buf []byte, err error) {
	// Calculate the size of the struct
	s := bstd.SizeString(t.str)
	s += bstd.SizeUInt64()
	s += bstd.SizeBool()

	// Serialize the struct into a byte slice
	n, buf := bstd.Marshal(s)
	n = bstd.MarshalString(n, buf, t.str)
	n = bstd.MarshalUInt64(n, buf, t.id)
	n = bstd.MarshalBool(n, buf, t.cool)

	err = bstd.VerifyMarshal(n, buf)
	return
}

func UnMarshalTestData(b []byte, t *TestData) (err error) {
	var n int

	// Deserialize the byte slice into the struct
	n, t.str, err = bstd.UnmarshalString(0, b)
	if err != nil {
		return
	}

	n, t.id, err = bstd.UnmarshalUInt64(n, b)
	if err != nil {
		return
	}

	n, t.cool, err = bstd.UnmarshalBool(n, b)
	if err != nil {
		return
	}
	return
}

func main() {
	// Create a TestData
	t := &TestData{
		str:  "I am a Test",
		id:   10,
		cool: true,
	}

	// Serialize the TestData
	bytes, err := MarshalTestData(t)
	if err != nil {
		panic(err.Error())
	}

	// You can now share the byte slice `bytes`

	var t2 TestData

	// Deserialize the TestData
	if err = UnMarshalTestData(bytes, &t2); err != nil {
		panic(err.Error())
	}

	// "I am a Test"
	println(t2.str)
}
```

## Tagging

`btag.SMarshal` (string tag) or `btag.UMarshal` (uint tag) just replace `bstd.Marshal`

```go
package main

import (
	bstd "github.com/deneonet/benc"
	"github.com/deneonet/benc/btag"
)

func main() {
	// for a uint tag: btag.UMarshal(0, UINT16)
	n, b := btag.SMarshal(0, "v1")
	if err := bstd.VerifyMarshal(n, b); err != nil {
		panic(err.Error())
	}

	// a string/uint tag is the first thing that has to be deserialized
	n, tag, err := btag.SUnmarshal(0, b) // or for a uint tag: btag.UUnmarshal(0, b)
	if err != nil {
		panic(err.Error())
	}
	if tag != "v1" {
		panic("tag doesn't match")
	}

	if err := bstd.VerifyUnmarshal(n, b); err != nil {
		panic(err.Error())
	}
}
```

Little benchmark, `btag.SMarshal` (string tag) vs `btag.UMarshal` (uint tag):
You can find all benchmarks in `benchs_test.go`.

```bash
goos: windows
goarch: amd64
pkg: github.com/deneonet/benc
cpu: 11th Gen Intel(R) Core(TM) i5-11300H @ 3.10GHz
BenchmarkStringTag-8    100000000  10.80 ns/op   3 B/op  1 allocs/op
BenchmarkUIntTag-8      147857434  7.944 ns/op  2 B/op  1 allocs/op
```

Not a big difference but still faster, here is the code:

```go
package main

import (
	bstd "github.com/deneonet/benc"
	"github.com/deneonet/benc/btag"
	"testing"
)

func BenchmarkStringTag(b *testing.B) {
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		// for a uint tag: btag.UMarshal(0, UINT16)
		n, b := btag.SMarshal(0, "v1")
		if err := bstd.VerifyMarshal(n, b); err != nil {
			panic(err.Error())
		}

		// a string/uint tag is the first thing that has to be deserialized
		n, tag, err := btag.SUnmarshal(0, b) // or for a uint tag: btag.UUnmarshal(0, b)
		if err != nil {
			panic(err.Error())
		}
		if tag != "v1" {
			panic("tag doesn't match")
		}

		if err := bstd.VerifyUnmarshal(n, b); err != nil {
			panic(err.Error())
		}
	}
}

const (
	v1 uint16 = iota
)

func BenchmarkUIntTag(b *testing.B) {
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		n, b := btag.UMarshal(0, v1)
		if err := bstd.VerifyMarshal(n, b); err != nil {
			panic(err.Error())
		}

		n, tag, err := btag.UUnmarshal(0, b)
		if err != nil {
			panic(err.Error())
		}
		if tag != v1 {
			panic("tag doesn't match")
		}

		if err := bstd.VerifyUnmarshal(n, b); err != nil {
			panic(err.Error())
		}
	}
}
```

## Slices And Maps serialization

```go
package main

import (
	bstd "github.com/deneonet/benc"
)

func main() {
	sliceData := []string{"DATA_1", "DATA_2"}

	mapData := make(map[string]float64)
	mapData["DATA_1"] = 13531.523400123
	mapData["DATA_2"] = 2561.1512312313

	s := bstd.SizeSlice(sliceData, bstd.SizeString)
	s += bstd.SizeMap(mapData, bstd.SizeString, bstd.SizeFloat64)

	n, buf := bstd.Marshal(s)
	n = bstd.MarshalSlice(n, buf, sliceData, bstd.MarshalString)
	n = bstd.MarshalMap(n, buf, mapData, bstd.MarshalString, bstd.MarshalFloat64)

	if err := bstd.VerifyMarshal(n, buf); err != nil {
		panic(err.Error())
	}

	var err error

	n, sliceData, err = bstd.UnmarshalSlice(0, buf, bstd.UnmarshalString)
	if err != nil {
		panic(err.Error())
	}

	n, mapData, err = bstd.UnmarshalMap(n, buf, bstd.UnmarshalString, bstd.UnmarshalFloat64)
	if err != nil {
		panic(err.Error())
	}

	if err := bstd.VerifyUnmarshal(n, buf); err != nil {
		panic(err.Error())
	}

	if sliceData[0] != "DATA_1" || sliceData[1] != "DATA_2" {
		panic("slice doesn't match")
	}

	if mapData["DATA_1"] != 13531.523400123 || mapData["DATA_2"] != 2561.1512312313 {
		panic("map doesn't match")
	}
}
```

## Pre-Allocation

Using pre-allocation, the buffer is reused instead of allocating one each time serialiazing

```go
package main

import (
	bstd "github.com/deneonet/benc"
	"github.com/deneonet/benc/bpre"
	"github.com/deneonet/benc/bunsafe"
)

func main() {
	// pre-allocates a byte slice of size 1000
	bpre.Marshal(1000)

	s := bstd.SizeString("Hello World!")
	s += bstd.SizeFloat64()

	// doesn't allocate any memory now, because it takes the needed bytes, from the pre-allocated byte slice
	n, buf := bstd.Marshal(s)
	n = bunsafe.MarshalString(n, buf, "Hello World!")
	n = bstd.MarshalFloat64(n, buf, 1231.5131)

	if err := bstd.VerifyMarshal(n, buf); err != nil {
		panic(err.Error())
	}

	// for simplicity, we just skip the string and float64
	n, err := bstd.SkipString(0, buf)
	if err != nil {
		panic(err.Error())
	}

	n, err = bstd.SkipFloat64(n, buf)
	if err != nil {
		panic(err.Error())
	}

	if err := bstd.VerifyUnmarshal(n, buf); err != nil {
		panic(err.Error())
	}

	// resets the buffer that is reused, so it's not going to be reused again
	bpre.Reset()
}
```

Little benchmark, with pre-allocation/buffer reuse and without, you can find these benchmarks in `benchs_test.go`. A similar benchmark can be found [here](BESTPRACTICES.md):

```bash
goos: windows
goarch: amd64
pkg: github.com/deneonet/benc
cpu: 11th Gen Intel(R) Core(TM) i5-11300H @ 3.10GHz
BenchmarkPreAllocations-8       320253640   3.612 ns/op   0 B/op   0 allocs/op
BenchmarkNoPreAllocations-8     54564552   22.89 ns/op   24 B/op   1 allocs/op
```

## Out-Of-Order Deserialization

Using out-of-order deserialization, you don't have to follow the order that the data was serialiazed, though it's not
recommended

```go
package main

import (
	bstd "github.com/deneonet/benc"
	"github.com/deneonet/benc/bunsafe"
)

func main() {
	s := bstd.SizeString("Hello World!")
	s += bstd.SizeFloat64()
	s += bstd.SizeFloat32()
	n, buf := bstd.Marshal(s)

	// Marshal - Order:
	// Hello World! : bunsafe.UnmarshalString(...)
	// 1231.5131 : UnmarshalFloat64(...)
	// 1231.5132 : UnmarshalFloat32(...)

	n = bunsafe.MarshalString(n, buf, "Hello World!")
	n = bstd.MarshalFloat64(n, buf, 1231.5131)
	n = bstd.MarshalFloat32(n, buf, 1231.5132)

	if err := bstd.VerifyMarshal(n, buf); err != nil {
		panic(err.Error())
	}

	// Unmarshal - Order:
	// 1231.5131 : UnmarshalFloat64(...)
	// Hello World! : bunsafe.UnmarshalString(...)
	// 1231.5132 : UnmarshalFloat32(...)

	n, err := bstd.SkipString(0, buf)
	if err != nil {
		panic(err.Error())
	}

	var randomFloat64 float64
	n, randomFloat64, err = bstd.UnmarshalFloat64(n, buf)
	if err != nil {
		panic(err.Error())
	}

	if randomFloat64 != 1231.5131 {
		panic("randomFloat64: float64 doesn't match")
	}

	var helloWorld string
	_, helloWorld, err = bunsafe.UnmarshalString(0, buf)
	if err != nil {
		panic(err.Error())
	}

	if helloWorld != "Hello World!" {
		panic("helloWorld: string doesn't match")
	}

	var randomFloat32 float32
	n, randomFloat32, err = bstd.UnmarshalFloat32(n, buf)
	if err != nil {
		panic(err.Error())
	}

	if randomFloat32 != 1231.5132 {
		panic("randomFloat32: float32 doesn't match")
	}

	if err := bstd.VerifyUnmarshal(n, buf); err != nil {
		panic(err.Error())
	}
}
```

## Message framing

Message framing prefixes the serialized byte slice with the size of the
data, [useful for TCP/IP](https://blog.stephencleary.com/2009/04/message-framing.html)

```go
package main

import (
	"bytes"
	bstd "github.com/deneonet/benc"
)

func main() {
	var buffer bytes.Buffer

	s := bstd.SizeString("Hello World!")
	s += bstd.SizeFloat64()

	n, buf := bstd.MarshalMF(s)
	n = bstd.MarshalString(n, buf, "Hello World!")
	n = bstd.MarshalFloat64(n, buf, 1231.5131)
	if err := bstd.VerifyMarshal(n, buf); err != nil {
		panic(err.Error())
	}

	// Write the byte slice containing the encoded data twice into buffer
	// = two concatenated BENC encoded byte slices
	buffer.Write(buf)
	buffer.Write(buf)

	// Extracts the two concatenated byte slices, into a slice of byte slices
	data, err := bstd.UnmarshalMF(buffer.Bytes())
	if err != nil {
		panic(err.Error())
	}

	for _, bs := range data {
		var helloWorld string
		n, helloWorld, err = bstd.UnmarshalString(0, bs)
		if err != nil {
			panic(err.Error())
		}
		if helloWorld != "Hello World!" {
			panic("helloWorld: string doesn't match")
		}

		var randomFloat64 float64
		n, randomFloat64, err = bstd.UnmarshalFloat64(n, bs)
		if err != nil {
			panic(err.Error())
		}
		if randomFloat64 != 1231.5131 {
			panic("randomFloat64: float64 doesn't match")
		}
	}

	if err := bstd.VerifyUnmarshalMF(n, buf); err != nil {
		panic(err.Error())
	}
}
```

## DataType validation

```go
package main

import (
	bstd "github.com/deneonet/benc"
	"github.com/deneonet/benc/bmd"
)

type TestData struct {
	str  string
	id   uint64
	cool bool
}

func MarshalTestData(t *TestData) (buf []byte, err error) {
	// Calculate the size of the struct
	s := bmd.SizeString(t.str)
	s += bmd.SizeUInt64()
	s += bmd.SizeByte() // same as bmd.SizeBool()

	// Serialize the struct into a byte slice
	n, buf := bstd.Marshal(s)
	n = bmd.MarshalString(n, buf, t.str)
	n = bmd.MarshalUInt64(n, buf, t.id)

	// Let's marshal a byte instead of a bool
	n = bmd.MarshalByte(n, buf, 1)

	// Verify the marshal process
	err = bstd.VerifyMarshal(n, buf)
	return
}

func UnmarshalTestData(b []byte, t *TestData) (err error) {
	var n int

	// Deserialize the byte slice into the struct
	n, t.str, err = bmd.UnmarshalString(0, b)
	if err != nil {
		return
	}

	n, t.id, err = bmd.UnmarshalUInt64(n, b)
	if err != nil {
		return
	}

	// Here we unmarshal a bool, but in the marshal process, we used a byte, so this will return an error
	n, t.cool, err = bmd.UnmarshalBool(n, b)
	if err != nil {
		return
	}

	// Verify the unmarshal process
	err = bstd.VerifyUnmarshal(n, b)
	return
}

func main() {
	// Create a new TestData struct
	t := &TestData{
		str:  "I am a Test",
		id:   10,
		cool: true,
	}

	// Serialize the TestData
	bytes, err := MarshalTestData(t)
	if err != nil {
		// do error handling here
		panic(err.Error())
	}

	// You can now share the byte slice `bytes`

	var t2 TestData

	// Decode the TestData
	if err := UnmarshalTestData(bytes, &t2); err != nil {
		// do error handling here
		// note that UnmarshalTestData will always return an error: see line 48
	}
}
```

## License

[MIT](https://choosealicense.com/licenses/mit/)
