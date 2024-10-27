# benc idv

The Benc ID Validation (Benc IDV) provides a suite of methods for prefixing Benc standard's raw size, marshaling, and unmarshaling with an ID. When referring to "prefixing with ID," it means that the marshaled Go type is prefixed with a provided ID of any size. Upon unmarshaling, this ID is then checked against the deserialized ID.

## Installation
```bash
go get github.com/deneonet/benc/idv
```

## Tests
The code coverage of `bidv.go` is 100% (~97% with uncalled panics).

## Usage

Benc IDV provides four primary functions:

- **Skip**: Skips the ID (+ validates it) and the requested type.
- **Size**: Adds the needed size for `id` and returns `s` plus the calculated ID size.
- **Marshal**: Marshals `id` into the buffer at a given offset `n`.
- **Unmarshal**: Unmarshals and validates the deserialized ID, then unmarshals the requested type.

## Example

Marshaling and Unmarshalling a string with the ID of `1`:  
See [bstd examples](../std/README.md#complex-data-type-example)

```go
package main

import (
	"fmt"
	"github.com/deneonet/benc/idv"
	"github.com/deneonet/benc/std"
)

func main() {
	var id uint64 = 1
	mystr := "My string"

	// Calculate the size needed
	s := bidv.Size(id, bstd.SizeString(mystr))

	// Create buffer
	buf := make([]byte, s)

	// Marshal ID into buffer
	n := bidv.Marshal(0, buf, id)

	// Marshal string into buffer
	_ = bstd.MarshalString(n, buf, mystr)

	// Unmarshal ID and string
	_, deserMyStr, err := bidv.Unmarshal[string](0, buf, id, bstd.UnmarshalString)
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