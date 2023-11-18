# BENC

![go workflow](https://github.com/deneonet/benc/actions/workflows/go.yml/badge.svg)
![go report card]()
[![go reference](https://pkg.go.dev/badge/github.com/deneonet/benc.svg)](https://pkg.go.dev/github.com/deneonet/benc)

The fastest binary encoder/decoder in pure Golang.

## Features

- Fast encoding/decoding
- Slices support (manual)
- Map support (manual)
- Struct support (manual)
- Message framing support
- Robust decoding and validation

## Changelog
v1 to v1.0.1
- benc -> bstd
- all Size function requires 1 argument T (going to be removed again in v1.0.2)
- added Time, Byte, Faster String encoding, Faster byte slice encoding, Maps and Slices, UInt16, UInt32 and Int16, aswell as Float32
- added best practices

See you in v1.0.2

## Benchmarks

- [See benchmarks here](https://github.com/deneonet/go_serialization_benchmarks)

## Best Practices
- [See best practices here](BESTPRACTICES.md)

## Installation

Install BENC in any Golang Project

```bash
go get github.com/deneonet/benc
```

## Get Started

Basic encoding and decoding of a string:

```go
package main

import (
	bstd "github.com/deneonet/benc"
)

func main() {
	str := "Hello!"
	// Calculate the size of the struct
	s := bstd.SizeString(str)
	// Encode the string
	n, buf := bstd.Marshal(s)
	n = bstd.MarshalString(n, buf, "Hello!")
	// Verify the marshal process
	err := bstd.VerifyMarshal(n, buf)
	if err != nil {
		panic(err.Error())
	}

	// You can now share the byte slice `buf`

	// Decode the string
	n, hello, err := bstd.UnmarshalString(0, buf)
	if err != nil {
		panic(err.Error())
	}
	// Verify the unmarshal process
	err = bstd.VerifyUnmarshal(n, buf)
	if err != nil {
		panic(err.Error())
	}
	// Prints: Hello
	println(hello)
}
```

Basic encoding and decoding of a struct:

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

func MarshalTestData(t TestData) []byte {
	// Calculate the size of the struct
	s := bstd.SizeString(t.str)
	s += bstd.SizeUInt64()
	s += bstd.SizeBool()

	// Encode the struct into a byte slice
	n, buf := bstd.Marshal(s)
	n = bstd.MarshalString(n, buf, t.str)
	n = bstd.MarshalUInt64(n, buf, t.id)
	n = bstd.MarshalBool(n, buf, t.cool)
	// And return the byte slice
	return buf
}

func UnMarshalTestData(b []byte) (TestData, error) {
	t := TestData{}
	var err error
	var n int
	// Decode the byte slice into the struct
	n, t.str, err = bstd.UnmarshalString(0, b)
	if err != nil {
		return TestData{}, err
	}
	n, t.id, err = bstd.UnmarshalUInt64(n, b)
	if err != nil {
		return TestData{}, err
	}
	n, t.cool, err = bstd.UnmarshalBool(n, b)
	if err != nil {
		return TestData{}, err
	}
	// And return the testdata
	return t, nil
}

func main() {
	// Create a TestData
	t := TestData{
		str:  "I am a Test",
		id:   10,
		cool: true,
	}
	// Encode the TestData
	bytes := MarshalTestData(t)

	// You can now share the byte slice `bytes`

	// Decode the TestData
	t2, err := UnMarshalTestData(bytes)
	if err != nil {
		panic(err.Error())
	}
	// Prints: I am a Test
	println(t2.str)
}
```

#### More examples in the examples folder 

## Todos

- Automatic struct encoding/decoding
- Automatic slice encoding/decoding
- Less encoded byte slice length
- Versioning
## License

[MIT](https://choosealicense.com/licenses/mit/)
