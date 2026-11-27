# Benc

**Benc** is a binary serialization format and a collection of language runtimes.

This project is a fork of [deneonet/benc](https://github.com/deneonet/benc). Its generator parses language declarations into **Go AST** (`go/ast`) and generates source for a requested target from that AST. There is no separate schema or IDL.

The CLI has parsers and generators for Go, C, C++, JavaScript, TypeScript, Python, Rust, Java, Kotlin, C#, Swift and Scala. Its pipeline tests compare Go AST types for all 12 source languages, generate all 12 targets, and compile/run Go, C, C++, JavaScript, Python and Java outputs. [Native smoke CI](.github/workflows/generated-native.yml) compiles/runs Rust, Kotlin, C#, Swift and Scala and typechecks TypeScript for every source language. See the [supported AST subset](spec/ast-mapping.md), [wire v1](spec/wire-v1.md), [compatibility and release gates](spec/compatibility.md), and [migration guide](MIGRATION.md).

## Getting Started

### 1. Installation

Build the CLI from the repository:

```sh
git clone https://github.com/banditmoscow1337/benc.git
cd benc
go build -o benc ./cmd
./benc -version
```

### 2. Define source types

Benc reads supported declarations from each listed language. These declarations are parsed into Go AST. For example:
```
package models

// The parser converts this declaration to Go AST for target generators.
type MyData struct {
    ID      uint64
    Name    string
    Tags    []string
    Payload []byte
}
```

### 3. Generate code

The CLI accepts one positional source file and `-lang` with one or more implemented targets. Generated files are placed next to the input. Use `-list-languages` for the current inventory.

```
benc -lang=cpp ./models/types.go
benc -lang=go,js,ts,py,java,c ./models/types.go
```

### 4. Integration
Import the generated code and the Benc runtime (`std`) for your target. The output is named `<source>_benc.*` with companion `<source>_benc_test.*` files. TypeScript produces executable JavaScript and a `.d.ts` declaration. Install the JS/TS runtime from `std/javascript` (one npm package), or the Python runtime with `python -m pip install ./std/python`. Generated JS currently loads a colocated `std.js` from that installed package. The other install/build commands are in the [compatibility table](spec/compatibility.md). The supported declaration subset and inline Go type annotations are described in [the AST mapping guide](spec/ast-mapping.md).

Run the [executable examples](examples/README.md) to generate and execute a consumer in a clean directory.

##  Project Structure

-   **`cmd/`**: CLI and parsers/generators feeding one Go AST.
-   **`std/`**: Language runtime implementations; their conformance is tracked separately from parser and generator availability.
-   **`testdata/conformance/`**: Independent wire vectors and equivalent source declarations for parser comparison.
-   **`spec/`**: Wire v1 and Go AST mapping rules.
    

## Contributing

We welcome contributions, especially for adding new language ports or optimizing existing generators.

1.  Check the [CONTRIBUTING.md](CONTRIBUTING.md) file.
2.  Ensure you run tests for the specific language `std` lib you are modifying.
3.  If modifying the protocol, update `spec/wire-v1.md` and the independent vectors before porting the change across languages.
    

## License

[MIT](LICENSE)
