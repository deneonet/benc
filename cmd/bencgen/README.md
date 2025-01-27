# bencgen

A code generator for Benc, ensuring both forward and backward compatibility.

## Table of Contents

- [Requirements](#requirements)
- [Installation](#installation)
- [Usage](#usage)
- [Generating Example](#generating-example)
- [Go Usage Example](#go-usage-example)
- [Breaking Changes Detector](#breaking-changes-detector-bcd)
- [Maintaining](#maintaining)
- [Examples and Tests](#examples-and-tests)
- [Enums](#enums)
- [Schema Grammar](#schema-grammar)
- [Languages](#languages)
- [License](#license)

## Requirements

- Go (for installing and running `bencgen`)
- [Benc Standard](../../std/README.md)
- [Bencgen Implementation](../../impl/gen/README.md)

## Installation

1. Install `bencgen` using the following command:

```bash
go install github.com/deneonet/benc/cmd/bencgen
```

## Usage

Arguments:

- `--in`: The input `.benc` file (required)
- `--out`: The output directory (optional)
- `--lang`: The target [language](#languages) to compile into (required)
- `--file`: The name of the output file (optional)
- `--force`: Disable the breaking changes detector (optional, not recommended if the schema is in use, e.g., in software)

## Generating Example

1. Create a `.benc` file (e.g., `person.benc`).
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

3. Generate Go code with the following command:

```bash
bencgen --in person.benc --lang go
```

4. Follow the instructions for using the generated code in the selected language:
   - [Go Usage Example](#go-usage-example)

## Go Usage Example

After generating, a file called `out/person.benc.go` will be created. To marshal and unmarshal the `Person` data:

```go
package main

import (
	"github.com/deneonet/benc"
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

	buf := make([]byte, data.Size())
	data.Marshal(buf)

	var retData person.Person
	if err := retData.Unmarshal(buf); err != nil {
		panic(err)
	}
}
```

## Breaking Changes Detector (BCD)

BCD detects breaking changes, such as:

- A field exists but is marked as reserved.
- A field was removed but is not marked as reserved.
- The type of a field changed, but its ID remained the same.

## Maintaining

To maintain your Benc schema, follow these rules:

- Mark removed fields as reserved.
- New fields must be appended at the bottom (fields should be ordered by their IDs in ascending order).
- If the type of a field changes, assign it a new ID and mark the old ID as reserved.
- Use [BCD](#breaking-changes-detector-bcd) (enabled by default) to catch and report compatibility issues.

### Reserving IDs

For example, using the `person` schema from the [earlier section](#generating-example), if the `parents` field is removed (ID `3`), mark ID `3` as reserved:

`person.benc`:

```plaintext
header person;

ctr Person {
    reserved 3;  # Reserved the 'parents' field ID

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
For tests specifically related to forward and backward compatibility, see [here](../../testing/person/main_person_test.go).

## Enums

Enums example:

```plaintext
header person;

enum JobStatus {
    Employed,
    Unemployed
}

ctr Person {
    byte age = 1;
    string name = 2;
    JobStatus jobStatus = 3;
}
```

Enums are treated as named integers. Forward and backward compatibility is preserved even when fields are added or removed from an enum, as the benc protocol doesn't rely on them.

## Schema Grammar

A schema consists of the following components:

### Header

A header consists of: `header` IDENTIFIER `;`

|   **Benc**   |    **Go**     |
| :----------: | :-----------: |
| `header ...` | `package ...` |

### Fields

A field consists of: "[ATTR](#type-attributes) [TYPE](#types) IDENTIFIER = ID ;" || "[CONTAINER_OR_ENUM_NAME](#containers-or-enums) IDENTIFIER = ID ;"

- The ID must be no larger than `65535`.
- A field of type `string` may have [type attributes](#type-attributes).

Example of a simple field:

```plaintext
string name = 1;
```

Example of a field with type attributes:

```plaintext
@unsafe string name = 1;
```

**Note:** Type attributes **must** precede the type, e.g., for arrays:

```plaintext
[] @unsafe string names = 1;
```

#### Type Attributes

- **unsafe** (Go only): Uses the `unsafe` package, allowing faster unmarshalling operations. Multiple `unsafe` attributes are ignored.

## Types

| **Benc**  |  **Go**   |
| :-------: | :-------: |
|  `byte`   |  `byte`   |
|  `bytes`  | `[]byte`  |
|   `int`   |   `int`   |
|  `int16`  |  `int16`  |
|  `int32`  |  `int32`  |
|  `int64`  |  `int64`  |
|  `uint`   |  `uint`   |
| `uint16`  | `uint16`  |
| `uint32`  | `uint32`  |
| `uint64`  | `uint64`  |
| `float32` | `float32` |
| `float64` | `float64` |
|  `bool`   |  `bool`   |
| `string`  | `string`  |
|   `[]T`   |   `[]T`   |
| `<K, V>`  | `map[K]V` |

### Containers Or Enums

A container or enum name refers to another defined structure.

**Container**:

```
ctr Person {
    byte age = 1;
    string name = 2;
    Child child = 4;
}
```

Reference:

```
ctr Person2 {
    Person person = 1;
}
```

**Enum**:

```
enum JobStatus {
    Employed,
    Unemployed
}
```

Reference:

```
ctr Person {
    JobStatus jobStatus = 1;
}
```

## Languages

Valid values for `--lang` are:

- `go`

## License

MIT
