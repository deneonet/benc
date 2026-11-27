# Source declarations and Go AST

Benc has one internal type model: `go/ast`. Language-specific parsers
produce `*ast.TypeSpec` nodes. Each target generator reads those nodes.
There is no Benc schema, IDL or additional intermediate representation.
Validation happens on the AST before any generated file is written.

## Common subset

Named structs with ordered, exported fields and defined named maps are
supported. Field types can be bool, string, signed/unsigned integers of
8/16/32/64 bits, float32/64, `int`, `uint`, `byte`, `rune`, `time.Time`,
other named structs/maps, pointers, slices, fixed arrays with a positive
integer literal length, and maps. Map keys must be scalar primitives other
than floating-point values; pointers, collections, time and named structs
cannot be map keys. `byte`/`uint8` and `rune`/`int32` are equivalent
aliases for cross-language AST comparison. `[]byte` uses the bytes wire
encoding; `[N]byte` uses exactly N bytes without a prefix.

Empty structs, unknown/imported types (except `time.Time`), type aliases declared with
`=`, named non-map scalar/collection types, embedded fields, recursive
types, interfaces, functions and channels produce an error with a type/field
name and source position. `//benc:ignore` on a Go field explicitly excludes
it; JavaScript's `// benc:ignore` excludes a constructor assignment.
Omission is never an automatic fallback for an unsupported field.

`int` and `uint` use architecture-width varints. C represents them as
`intptr_t`/`uintptr_t`, C++ likewise, and JavaScript uses BigInt. Use
fixed-width integers for cross-architecture interchange. Cycles are
rejected, including pointer cycles, because all targets must represent
the same AST without incomplete by-value declarations.

## Source syntax

| Source | Declaration form | Explicit types needed for |
| --- | --- | --- |
| Go | Top-level `type Name struct { ... }` or `type Name map[K]V` | Standard Go syntax is authoritative. |
| C17 | `typedef struct { ... } Name;` | Binary strings: `char* Name; size_t Name_len;`. Slices: `T* Name; size_t Name_count;`. Maps: adjacent `K* Name_keys; V* Name_values; size_t Name_count;`. Fixed arrays: `T Name[N];`. A plain `char*` is ambiguous and rejected. |
| JavaScript | `class Name { constructor() { this.Field = value; // GoType } }` | Empty arrays, maps, null, and other ambiguous initializers require a Go type expression after `//`. A named map can be declared with `//benc:type map[K]V` immediately before the class. |
| C++20 | `struct Name { T Field; };` or a named `std::map` wrapper | Supported `std::vector`, `std::array`, `std::map` and `std::optional` declarations. |
| TypeScript | `class` / `interface` fields and `type Name = Map<K,V>` | Number/BigInt and container types use inline Go type comments. The generated `.d.ts` carries them. |
| Python | Annotated class / dataclass fields or named dictionary class | Fixed widths, arrays and byte collections use inline Go type comments where Python typing is ambiguous. |
| Rust | `struct Name { pub Field: T, }` / named `HashMap` tuple struct | Generated code annotates fields with their Go type. |
| Java | Public class fields / named map wrapper | Inline Go type comments preserve unsigned widths and array lengths. |
| Kotlin | `class Name { var Field: T = default // GoType }` / named map wrapper | Inline Go type comments preserve `int`, `uint`, byte arrays and fixed lengths. |
| C# | `class Name { public T Field = default; // GoType }` / named map wrapper | Inline Go type comments preserve `time.Time` and fixed lengths. |
| Swift | `struct Name { var Field: T = default // GoType }` / named map wrapper | Inline Go type comments preserve `time.Time` and fixed lengths. |
| Scala 3 | Multiline `case class Name(Field: T = default, // GoType)` / named map case class | Each property needs an inline Go type comment. |

The C generator uses private `benc_generated_t*` typedefs for nested collections,
binary-safe strings, and nanosecond timestamps. Its parser resolves these
declarations back into Go AST expressions and does not expose them as user
types. Named maps are represented by a generated C typedef containing
`benc_value`; the C parser reconstructs the named map type.

## Generated output

`benc -lang=go,c,cpp,js,ts,py,rs,java,kt,cs,swift,scala path/to/model.go` writes
`model_benc.go`, `model_benc.h`, `model_benc.c`,
`model_benc.hpp`, `model_benc.js` and matching
`model_benc_test.*` files next to the input. For C or JS input, Go
generation also writes the declarations recovered from that source.
The Go input already contains declarations, so Go generation writes
methods only. TypeScript generation delegates executable encoding to its
JavaScript output and writes `model_benc.d.ts` for static declarations.

Compile C with `std/c/benc.h` and define `BSTD_IMPLEMENTATION` in exactly
one translation unit. Compile C++20 with `std/c++/std.hpp`. Place
`std/javascript/std.js` beside generated JS. C `*_unmarshal` assigns
an owned decoded value to its output, which must be zero-initialized or
previously freed; release decoded strings/collections with `*_free`.
JavaScript `bstd.Time` retains exact Unix nanoseconds. A native JS
`Date` is accepted for encoding at millisecond precision, and
`bstd.Time.toDate()` provides a Date view.
Swift and C# generated fields use signed 64-bit Unix nanoseconds for
`time.Time` so round trips retain sub-microsecond precision. Scala unsigned
fields use signed JVM storage with unsigned bit patterns.

| Source | Parser → Go AST | Generator from Go AST | Runtime |
| --- | --- | --- | --- |
| Go | yes | yes | yes |
| C | yes (declared subset) | yes | yes |
| C++ | yes | yes | yes |
| JavaScript | yes (declared subset) | yes | yes |
| TypeScript | yes | yes (JS runtime and `.d.ts`) | yes (JS runtime) |
| Python | yes | yes | yes |
| Java | yes | yes | yes |
| Rust | yes | yes, native CI | yes |
| Kotlin | yes | yes, native CI | yes |
| C# | yes | yes, native CI | yes |
| Swift | yes | yes, native CI | yes |
| Scala | yes | yes, native CI | yes |

The pipeline matrix in `cmd/generator/common/pipeline_test.go` parses all
listed source languages, compares their Go AST field types and order, then
generates every target. It compiles and runs generated Go/C/C++/JS/Python/Java
outputs. A nonzero cross-language fixture compares those six encoders' wire
bytes for NUL-containing strings, varints, fixed arrays, bytes, nested
collections, named maps and nanosecond timestamps. The native smoke workflow
also compiles and runs Rust, Kotlin, C#, Swift and Scala against the same
nonzero Go wire value, and typechecks TypeScript declarations with `tsc`.
CI feeds each of the 12 source languages through all six native targets on
PR, main, weekly, and manual runs, checking their generated code and wire. Repeated
generation must reproduce byte-identical files.
Invalid source types produce positioned errors rather than silently omitting
fields from the AST.
