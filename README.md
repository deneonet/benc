# BENC

![go workflow](https://github.com/deneonet/benc/actions/workflows/go.yml/badge.svg)
[![go report card](https://goreportcard.com/badge/github.com/deneonet/benc)](https://goreportcard.com/report/github.com/deneonet/benc)
[![go reference](https://pkg.go.dev/badge/github.com/deneonet/benc.svg)](https://pkg.go.dev/github.com/deneonet/benc)

The fastest serializer in pure Golang.

## Features

- [Fastest serialization](https://github.com/alecthomas/go_serialization_benchmarks#readme)
- [Buffer Reuse](#buffer-reuse)
- [Ability to add custom marshal and unmarshal functions](#custom-marshal-and-unmarshal-1)
- [Struct support](#struct-serialization)
- [Message framing support](#message-framing)
- [DataType validation](#datatype-validation)
- [Slices and Maps support](#slices-and-maps-serialization)
- [Out of Order Deserialization](#out-of-order-deserialization)

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

## [Custom Marshal And Unmarshal](#custom-marshal-and-unmarshal-1)

## Struct serialization

[With DataType validation in the unmarshal process](#datatype-validation)

```go
package main

import (
	"github.com/deneonet/benc"
	"github.com/deneonet/benc/bstd"
)

type TestData struct {
	str  string
	id   uint64
	cool bool
}

func MarshalTestData(t *TestData) ([]byte, error) {
	// - Calculate the size of the struct -

	s, err := bstd.SizeString(t.str)
	if err != nil {
		return nil, err
	}

	s += bstd.SizeUInt64()
	s += bstd.SizeBool()

	// - Serialize the struct into a byte slice -

	n, buf := benc.Marshal(s)
	if n, err = bstd.MarshalString(n, buf, t.str); err != nil {
		return nil, err
	}

	n = bstd.MarshalUInt64(n, buf, t.id)
	n = bstd.MarshalBool(n, buf, t.cool)

	// - Lastly verify the marshal process -

	err = benc.VerifyMarshal(n, buf)

	return buf, err
}

func UnmarshalTestData(buf []byte, t *TestData) (err error) {
	var n int

	// - Deserialize the byte slice into the struct -

	n, t.str, err = bstd.UnmarshalString(0, buf)
	if err != nil {
		return
	}

	n, t.id, err = bstd.UnmarshalUInt64(n, buf)
	if err != nil {
		return
	}

	_, t.cool, err = bstd.UnmarshalBool(n, buf)
	if err != nil {
		return
	}

	
	// - Lastly verify the unmarshal process -
	return benc.VerifyUnmarshal(n, buf)
}

func main() {
	// - Create a TestData -

	t := &TestData{
		str:  "I am a Test",
		id:   10,
		cool: true,
	}

	// - Serialize the TestData -

	bytes, err := MarshalTestData(t)
	if err != nil {
		panic(err.Error())
	}
	// You can now use the byte slice `bytes`

	// - Deserialize the TestData -

	var t2 TestData
	if err = UnmarshalTestData(bytes, &t2); err != nil {
		panic(err.Error())
	}

	// "I am a Test"
	println(t2.str)
}
```

## Slices And Maps serialization

```go
package main

import (
	"reflect"

	"github.com/deneonet/benc"
	"github.com/deneonet/benc/bstd"
)

func main() {
	// - Example data -
	sliceData := []string{"DATA_1", "DATA_2"}

	mapData := make(map[string]float64)
	mapData["DATA_1"] = 13531.523400123
	mapData["DATA_2"] = 2561.1512312313

	// - Sizing -

	s, err := bstd.SizeSlice(sliceData, bstd.SizeString)
	if err != nil {
		panic(err.Error())
	}

	ts, err := bstd.SizeMap(mapData, bstd.SizeString, bstd.SizeFloat64)
	if err != nil {
		panic(err.Error())
	}

	s += ts

	// - Serialization -

	n, buf := benc.Marshal(s)
	if n, err = bstd.MarshalSlice(n, buf, sliceData, bstd.MarshalString); err != nil {
		panic(err.Error())
	}

	if n, err = bstd.MarshalMap(n, buf, mapData, bstd.MarshalString, bstd.MarshalFloat64); err != nil {
		panic(err.Error())
	}

	if err := benc.VerifyMarshal(n, buf); err != nil {
		panic(err.Error())
	}

	// - Deserialization -

	n, resSliceData, err := bstd.UnmarshalSlice(0, buf, bstd.UnmarshalString)
	if err != nil {
		panic(err.Error())
	}

	n, resMapData, err := bstd.UnmarshalMap(n, buf, bstd.UnmarshalString, bstd.UnmarshalFloat64)
	if err != nil {
		panic(err.Error())
	}

	if err := benc.VerifyUnmarshal(n, buf); err != nil {
		panic(err.Error())
	}

	// - Verification -

	if !reflect.DeepEqual(sliceData, resSliceData) {
		panic("slice doesn't match")
	}

	if !reflect.DeepEqual(mapData, resMapData) {
		panic("map doesn't match")
	}
}

```

## Buffer Reuse

#### Without Concurrency:
Just create once a `buf` variable, like: `buf := make([]byte, 1024)` and use it instead of `benc.Marshal`, as `n`, returned by `benc.Marshal`, use 0

#### With Concurrency:

Either use the above explained and combine it with mutexs or use `benc.BufPool`, as shown here (Read the comments for concurrency-safety):

```go
package main

import (
	"github.com/deneonet/benc"
	"github.com/deneonet/benc/bstd"
)

func main() {
	// Allocates a byte slice of size 1024, default is 1024 (without `benc.WithBufferSize(...)`)
	bufPool := benc.NewBufPool(benc.WithBufferSize(512))

	s, err := bstd.SizeString("Hello World!")
	if err != nil {
		panic(err.Error())
	}

	s += bstd.SizeFloat64()

	// Doesn't allocate any buffer now, because it gets the needed buffer, from the buffer pool
	buf, err := bufPool.Marshal(s, func(b []byte) (n int) {
		n, err := bstd.MarshalString(n, b, "Hello World!")
		if err != nil {
			panic(err.Error())
		}

		n = bstd.MarshalFloat64(n, b, 1231.5131)

		if err := benc.VerifyMarshal(n, b); err != nil {
			panic(err.Error())
		}
		return
	})

	// You are now able to write `buf` to disk or transmit it over the network,
	// but you cannot read & write to it, only in the function that was specified as argument, an example is in `tests/benchs_test.go`

	if err != nil {
		panic(err.Error())
	}

	_ = buf
}

```

Obviously, is the `without concurrency` the fastest. For low concurrency, mutexs are faster and pooling (`benc.BufPool`) is faster for higher concurrency.

## Out-Of-Order Deserialization

When using out-of-order deserialization, you don't have to follow the order that the data was serialiazed, as shown here:

```go
package main

import (
	"github.com/deneonet/benc"
	"github.com/deneonet/benc/bstd"
)

func main() {
	s, err := bstd.SizeString("Hello World!")
	if err != nil {
		panic(err.Error())
	}

	s += bstd.SizeFloat64()
	s += bstd.SizeFloat32()
	n, buf := benc.Marshal(s)

	// Marshal Order:
	// Hello World! : String
	// 1231.5131 : Float64
	// 1231.5132 : Float32

	n, err = bstd.MarshalString(n, buf, "Hello World!")
	if err != nil {
		panic(err.Error())
	}

	n = bstd.MarshalFloat64(n, buf, 1231.5131)
	n = bstd.MarshalFloat32(n, buf, 1231.5132)

	if err := benc.VerifyMarshal(n, buf); err != nil {
		panic(err.Error())
	}

	// Unmarshal Order:
	// 1231.5131 : Float64
	// Hello World! : String
	// 1231.5132 : Float32

	n, err = bstd.SkipString(0, buf)
	if err != nil {
		panic(err.Error())
	}

	n, randomFloat64, err := bstd.UnmarshalFloat64(n, buf)
	if err != nil {
		panic(err.Error())
	}

	if randomFloat64 != 1231.5131 {
		panic("randomFloat64: doesn't match")
	}

	_, helloWorld, err := bstd.UnmarshalString(0, buf)
	if err != nil {
		panic(err.Error())
	}

	if helloWorld != "Hello World!" {
		panic("helloWorld: doesn't match")
	}

	n, randomFloat32, err := bstd.UnmarshalFloat32(n, buf)
	if err != nil {
		panic(err.Error())
	}

	if randomFloat32 != 1231.5132 {
		panic("randomFloat32: doesn't match")
	}

	if err := benc.VerifyUnmarshal(n, buf); err != nil {
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
	"fmt"

	"github.com/deneonet/benc"
	"github.com/deneonet/benc/bstd"
)

func main() {
	var buf bytes.Buffer

	testStr := "Hello World!"
	s, err := bstd.SizeString(testStr)
	if err != nil {
		panic(err.Error())
	}

	n, b := benc.MarshalMF(s)
	_, err = bstd.MarshalString(n, b, testStr)
	if err != nil {
		panic(err.Error())
	}

	// concatenated bytes of serialized "Hello World!" in the benc format
	buf.Write(b)
	buf.Write(b)

	unconcatenatedBytes, err := benc.UnmarshalMF(buf.Bytes())
	if err != nil {
		panic(err.Error())
	}

	for i, bytes := range unconcatenatedBytes {
		_, str, err := bstd.UnmarshalString(0, bytes)
		if err != nil {
			panic(err.Error())
		}

		if str != testStr {
			fmt.Printf("data %d: decoded str: %s\n", i, str)
		}
	}
}
```

## DataType validation
DataType validation appends in every marshal the data type serialized, and then checks in the unmarshal if the data type matches with the one expected. 

### Infos
- Type mismatch errors, should be treated as uncontinueable errors.
- When you serialize a slice or map with datatype validation, you can either use, as marshaller, the standard marshals `bstd` or data type validation marshals `bmd`, your choice, same with the standard (`bstd`) slices and maps.

### Snippet, that fails:

```go
package main

import (
	"fmt"

	"github.com/deneonet/benc"
	"github.com/deneonet/benc/bmd"
)

func main() {
	n, b := benc.Marshal(2) // 1 is the size of a byte + 1 for the data type
	bmd.MarshalByte(n, b, 128)

	// we serialized a byte, but now try to deserialize a bool -> type mismatch
	_, _, err := bmd.UnmarshalBool(n, b)

	// "type mismatch: expected Bool, got Byte"
	fmt.Println(err.Error())
}
```

### Full example, working:

```go
package main

import (
	"github.com/deneonet/benc"
	"github.com/deneonet/benc/bmd"
)

type TestData struct {
	str  string
	id   uint64
	cool bool
}

func MarshalTestData(t *TestData) ([]byte, error) {
	// - Calculate the size of the struct -

	s, err := bmd.SizeString(t.str)
	if err != nil {
		return nil, err
	}

	s += bmd.SizeUInt64()
	s += bmd.SizeBool()

	// - Serialize the struct into a byte slice -

	n, buf := benc.Marshal(s)
	if n, err = bmd.MarshalString(n, buf, t.str); err != nil {
		return nil, err
	}

	n = bmd.MarshalUInt64(n, buf, t.id)
	n = bmd.MarshalBool(n, buf, t.cool)

	// - Lastly verify the marshal process -

	err = benc.VerifyMarshal(n, buf)

	return buf, err
}

func UnmarshalTestData(buf []byte, t *TestData) (err error) {
	var n int

	// - Deserialize the byte slice into the struct -

	n, t.str, err = bmd.UnmarshalString(0, buf)
	if err != nil {
		return
	}

	n, t.id, err = bmd.UnmarshalUInt64(n, buf)
	if err != nil {
		return
	}

	_, t.cool, err = bmd.UnmarshalBool(n, buf)
	if err != nil {
		return
	}

	// - Lastly verify the unmarshal process -
	return benc.VerifyUnmarshal(n, buf)
}

func main() {
	// - Create a TestData -

	t := &TestData{
		str:  "I am a Test",
		id:   10,
		cool: true,
	}

	// - Serialize the TestData -

	bytes, err := MarshalTestData(t)
	if err != nil {
		panic(err.Error())
	}
	// You can now use the byte slice `bytes`

	// - Deserialize the TestData -

	var t2 TestData
	if err = UnmarshalTestData(bytes, &t2); err != nil {
		panic(err.Error())
	}

	// "I am a Test"
	println(t2.str)
}

```

## Custom Marshal and Unmarshal

Using custom marshal and unmarshal functions you can serialize, for example, a struct or an custom type into a slice or map.

To do that, we need actually 3 functions: `Size`, `Marshal` and `Unmarshal`  
An example:

```go
package main

import (
	"encoding/binary"
	"time"
)

func SizeTime() int {
	return 8 // Size of uint64 (size of what we write to the buffer)
}

func MarshalTime(n int, b []byte, time time.Time) (int, error) {
	binary.LittleEndian.PutUint64(b[n:], uint64(time.Unix()))

	// The size of data that was written to `b` has to be added to `n`
	return n + 8, nil
}

func UnmarshalTime(n int, b []byte) (int, time.Time, error) {
	unix := binary.LittleEndian.Uint64(b[n:])
	
	// The size of data that was read from `b` has to be added to `n`
	return n + 8, time.Unix(int64(unix), 0), nil
}
```

#### Lets expand it, by adding data type validation:

```go
package main

import (
	"encoding/binary"
	"fmt"
	"time"

	"github.com/deneonet/benc/bmd"
)

const (
	// Data type ID, make sure there are no duplicated IDS
	// Also make sure to always use `bmd.AllowedDataTypeStartIndex` as starting ID, to avoid conflicts with built-in types in the future
	Time byte = bmd.AllowedDataTypeStartID
)

func getDataTypeName(dt byte) string {
	switch dt {
	case Time:
		return "Time"
	default:
		return bmd.GetDataTypeName(dt)
	}
}

func SizeTime() int {
	return 9 // Size of uint64 + 1 byte for the data type (what we write to the buffer)
}

func MarshalTime(n int, b []byte, time time.Time) (int, error) {
	// First byte serialized should be the datatype ID
	b[n] = Time
	n++ // 1 here added

	binary.LittleEndian.PutUint64(b[n:], uint64(time.Unix()))

	// The size of data that was written to `b` has to be added to n
	// 1 was already added to `n`, so we only add 8 to `n`
	return n + 8, nil
}

func UnmarshalTime(n int, b []byte) (int, time.Time, error) {
	// First byte deserialized should be the datatype ID
	dt := b[n]
	if dt != Time {
		return n, time.Time{}, fmt.Errorf("type mismatch: expected Time, got %s", getDataTypeName(dt))
	}
	n++ // 1 here added

	unix := binary.LittleEndian.Uint64(b[n:])

	// The size of data that was read from `b` has to be added to n
	// 1 was already added to `n`, so we only add 8 to `n`
	return n + 8, time.Unix(int64(unix), 0), nil
}

```

## License

[MIT](https://choosealicense.com/licenses/mit/)
