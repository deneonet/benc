# benc std

The Benc standard provides a suite of methods for raw sizing, skipping, marshalling, and unmarshalling of Go types. When I refer to "raw", it means that only the essential elements are serialized, for example, serialized data is not prefixed with their corresponding type information. 

## Installation
```bash
go get github.com/deneonet/benc/std
```

## Tests
Code coverage of `bstd.go` is approximately 95%

## Usage

Benc Standard provides four primary functions, for all of these types (`string`, `unsafe string`, `slice`, `map`, `bool`, `byte`, `bytes` (slice of type byte), `float32`, `float64`, `int` (var int), `int16`, `int32`, `int64`, `uint` (var uint), `uint16`, `uint32`, `uint64`):

- **Skip**: Skips the requested type.
- **Size**: Calculate the needed size for the requested type (and data).
- **Marshal**: Marshals the requested type (and data) into the buffer at a given offset `n`.
- **Unmarshal**: Unmarshals the requested type.

Append the type (listed above) in CamelCase to the end of each function to skip/size/marshal or unmarshal the requested type.  
**Exception**: `int` and `uint`, the skip function for both of them is: `bstd.SkipVarint`

## Basic Type Example

Marshaling and Unmarshalling a string:

```go
package main

import (
	"fmt"
	"github.com/deneonet/benc/std"
)

func main() {
	mystr := "My string"

	// Calculate the size needed
	s := bstd.SizeString(mystr)

	// Create buffer
	buf := make([]byte, s)

	// Marshal the string into buffer
	_ = bstd.MarshalString(0, buf, mystr)

	// Unmarshal string
	_, deserMyStr, err := bstd.UnmarshalString(0, buf)
	if err != nil {
		panic(err)
	}
	if mystr != deserMyStr {
		panic("no match")
	}

	// Success
	fmt.Println("Marshaling and unmarshaling successful:", deserMyStr)
}
```

## Complex Data Type Example

Complex data types, like slices and maps:

```go
package main

import (
	"fmt"
	"reflect"

	bstd "github.com/deneonet/benc/std"
)

func main() {
	myslice := []string{"Str Element 1", "Str Element 2", "Str Element 3"}
	mymap := make(map[string]string)
	mymap["Str Key 1"] = "Str Val 1"
	mymap["Str Key 2"] = "Str Val 2"
	mymap["Str Key 3"] = "Str Val 3"

	// Calculate the size needed
	s := bstd.SizeSlice(myslice, bstd.SizeString)
	s += bstd.SizeMap(mymap, bstd.SizeString, bstd.SizeString)

	// Create buffer
	buf := make([]byte, s)

	// Marshal the slice and map into buffer
	n := bstd.MarshalSlice(0, buf, myslice, bstd.MarshalString)
	_ = bstd.MarshalMap(n, buf, mymap, bstd.MarshalString, bstd.MarshalString)

	// Unmarshal slice
	n, deserMySlice, err := bstd.UnmarshalSlice[string](0, buf, bstd.UnmarshalString)
	if err != nil {
		panic(err)
	}
	if !reflect.DeepEqual(myslice, deserMySlice) {
		panic("slice: no match")
	}

	// Unmarshal map
	_, deserMyMap, err := bstd.UnmarshalMap[string, string](n, buf, bstd.UnmarshalString, bstd.UnmarshalString)
	if err != nil {
		panic(err)
	}
	if !reflect.DeepEqual(mymap, deserMyMap) {
		panic("map: no match")
	}

	// Success
	fmt.Println("Slice marshaling and unmarshaling successful:")
	for _, str := range deserMySlice {
		fmt.Println(str)
	}

	fmt.Println("\nMap marshaling and unmarshaling successful:")
	for key, val := range deserMyMap {
		fmt.Println(key + ": " + val)
	}
}
```

Note: Maps and slice are able to nest, for example:

```go
// Size
bstd.SizeSlice(myslice, func (slice []string) int {
	return bstd.SizeSlice(slice, bstd.SizeString)
})

// Marshal
bstd.MarshalSlice(0, buf, mySliceSlice, func(n int, buf []byte, slice []string) int {
	return bstd.MarshalSlice(n, buf, slice, bstd.MarshalString)
})

// Unmarshal
bstd.UnmarshalSlice[[]string](0, buf, func (n int, buf []byte) (int, []string, error) {
	return bstd.UnmarshalSlice[string](n, buf, bstd.UnmarshalString)
})
```