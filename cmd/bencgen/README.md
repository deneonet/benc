# bencgen

The code generator for benc, handling both forward and backward compatibility.

## Table Of Contents
- [Requirements](#requirements)
- [Installation](#installation)
- [Usage](#usage)
- [Generating Example](#generating-example)
- [Go Usage Example](#go-usage-example)
- [Breaking Changes Detector](#breaking-changes-detector-bcd)
- [Maintaining](#maintaining)
- [Examples and Tests](#examples-and-tests)
- [Schema Grammar](#header)
- [Languages](#languages)
- [License](#license)

## Requirements
- Go for installing and executing `bencgen`.
- [benc standard](../../std/README.md)
- [bencgen impl](../../impl/gen/README.md)

## Installation

1. Install `bencgen`:
```bash
go install go.kine.bz/benc/cmd/bencgen
```

## Usage

Arguments:

- `--in`: The .benc input file (required)
- `--out`: The output directory (optional)
- `--lang`: The [language](#languages) to compile into (required)
- `--file`: The output file name (optional)
- `--force`: Disables the breaking changes detector (optional, not recommended for production)

## Generating Example

1. Create a .benc file (e.g. `person.benc`).
2. Write a schema, for example:

    ```plaintext
    header person;

    ctr Person {
        byte age = 1;
        string name = 2;
        Parents parents = 3;
        Child child = 4;
    }

    ctr Child {
        byte age = 1;
        string name = 2;
        Parents parents = 3;
    }

    ctr Parents {
        string mother = 1;
        string father = 2;
    }
    ```

3. Generate Go code using `bencgen --in person.benc --lang go`.
4. Find instructions for using the generated code in your selected language:
    - [Go](#go-usage-example)

## Go Usage Example

After generating, a file called `out/person.benc.go` will be created. To marshal and unmarshal the person:

```go
package main

import (
	"go.kine.bz/benc"
	person ".../out"
)

func main() {
	data := person.Person{
		Age:  24,
		Name: "Johnny",
		Parents: person.Parents{
			Mother: "Johna",
			Father: "John",
		},
		Child: person.Child{
			Name: "Johnny Jr.",
			Age:  3,
			Parents: person.Parents{
				Mother: "Johna Jr.",
				Father: "Johnny",
			},
		},
	}

	b, err := benc.MarshalCtr(&data)
	if err != nil {
		panic(err)
	}

	var retData person.Person
	if err = benc.UnmarshalCtr(&retData); err != nil {
		panic(err)
	}
}
```

## Breaking Changes Detector (BCD)

BCD detects breaking changes, such as:
- A field exists but is marked as reserved.
- A field never existed but is marked as reserved.
- A field was removed but is not marked as reserved.
- The type of a field changed, but its ID stayed the same.

## Maintaining

To maintain your benc schema, follow these rules:
- Mark a field as reserved when removed.
- New fields must be appended at the bottom (fields must be ordered by their IDs in ascending order).
- If the type of a field changes, it requires a new ID, and the old ID must be marked as reserved.
- Use [BCD](#breaking-changes-detector-bcd) (on by default); it catches and reports compatibility issues.

### Reserving IDs

Using the `person` schema from [earlier](#generating-example), if we remove the `parents` field, which had ID `3`, ID `3` must be marked as reserved:

`person2.benc`:
```plaintext
header person2;

ctr Person {
    reserved 3; # reserved the parents field ID

    byte age = 1;
    string name = 2;
    Child child = 4;
}

ctr Child {
    byte age = 1;
    string name = 2;
    Parents parents = 3;
}

ctr Parents {
    string mother = 1;
    string father = 2;
}
```

## Examples and Tests

See all tests [here](../../testing).  
See tests specifically about forward and backward compatibility [here](../../testing/bfc/person_test.go).

## Header

A header consists of: `header` IDENTIFIER `;`

| **Benc** | **Golang** |
|:--------:|:----------:|
| `header ...` | `package ...` |

## Fields

A field consists of: [TYPE](#types) IDENTIFIER `=` ID `;`

- The ID may not be larger than `65535`.
- A field with type `string` or `bytes` may have type attributes.

### Examples

Field:
`string name = 1;`

Type Attributes:
`string @unsafe name = 1;`

### Type Attributes

- **unsafe** (only in Go): Uses the `unsafe` package, allowing faster unmarshal operations.
- **bytes2**: Sets the max size of the field to 2 bytes (`65535` characters/bytes).
- **bytes4**: Like `bytes2` but with 4 bytes (`4294967295` characters/bytes).
- **bytes8**: Like `bytes2` but with 8 bytes (`9223372036854775807` characters/bytes).

Multiple `unsafe` type attributes are ignored.  
Multiple `bytes...` type attributes are ignored; the last one specified wins.

## Types

| **Benc** | **Golang** |
|:--------:|:----------:|
| `byte` | `byte` |
| `bytes` | `[]byte` |
| `int16` | `int16` |
| `int32` | `int32` |
| `int64` | `int64` |
| `uint16` | `uint16` |
| `uint32` | `uint32` |
| `uint64` | `uint64` |
| `float32` | `float32` |
| `float64` | `float64` |
| `bool` | `bool` |
| `string` | `string` |
| `[]T` | `[]T` |
| `map[K]V` | `map[K]V` |

The name of another container is also a type (`Container`).

## Languages

Valid values for `--lang` are:
- `go`

## License

MIT