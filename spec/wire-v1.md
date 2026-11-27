# Benc wire v1

This describes the byte format implemented by the Go runtime. It is a protocol
contract, not a schema or an intermediate representation. The source declaration
is parsed into Go `go/ast`; a target generator consumes that AST directly.

| Go AST field type | Encoding |
| --- | --- |
| `bool` | One byte: `00` for false, `01` for true. Other values are invalid in strict v1. |
| `byte`, `int8`, `uint8` | One byte. |
| `int16/32/64`, `uint16/32/64` | Two/four/eight bytes, little endian, two's complement for signed integers. |
| `int`, `uint`, `uintptr` | Unsigned base-128 varint; signed `int` first uses ZigZag. Width is the producer's word size; these types are **not portable across 32/64-bit targets**. Use fixed-width types for exchange. |
| `float32/64` | IEEE-754 bit pattern, little endian. |
| `string`, `[]byte` | Length as unsigned varint, then exactly that many bytes. Strings must be valid UTF-8 (without normalization); NUL is ordinary data. `[]byte` accepts arbitrary bytes. |
| `[N]T` | Exactly N consecutive encodings of T, without a length prefix or terminator. |
| `[]T` (other than `[]byte`) | Element count as unsigned varint, elements in order, then `01 01 01 01`. |
| `map[K]V` | Pair count as unsigned varint, then key/value pairs, then `01 01 01 01`. Duplicate decoded keys are invalid. Go map iteration order is unspecified: compare decoded entries, not serialized order, unless canonical ordering is introduced in a future version. |
| `*T` | `00` for nil, or `01` followed by T for present. |
| `struct` | Supported exported fields in declaration order, with no field names or delimiters. The exact type/field order is therefore essential to decoding. |
| `time.Time` | Signed 64-bit little-endian Unix nanoseconds. Time zone and monotonic clock are not encoded. |

Decode rejects duplicate keys where the runtime materializes a keyed map.
Generic `SkipMap` callbacks only traverse the encoding and cannot compare
decoded keys. C's generic map API exposes key/value arrays, so callers must
check key uniqueness using their field's decoded key type.

Base-128 varints use seven low bits per byte and continuation bit `80`; at most
five bytes for `uint32` and ten for `uint64`. At the last byte only four bits
are legal for 32-bit values and one bit for 64-bit values. Truncated, overlong
and overflowing values are invalid. Nonminimal in-range varints are accepted by
the existing Go/C decoders for compatibility; encoders emit the shortest form.

## Decoder policy

Strict v1 validates **boolean bytes, collection terminators, UTF-8 strings** and bounds:
missing/incorrect terminators, truncated payloads, out-of-range cursors,
overflowing lengths/counts and excessive allocation requests produce errors.
These are invalid wire values, not alternate encodings. An invalid callback
argument is a programmer error (Go may panic, C returns `BSTD_ERR_INVALID_ARGUMENT`).
Go decoding methods return offset zero on error; C restores the supplied cursor
to its initial value on composite error and leaves output parameters untouched.
Default per-field budgets in Go and C are 64 MiB of payload/allocation and
1,048,576 collection elements. C also enforces 128 nested composite calls per
thread; callers can set `BSTD_MAX_DECODE_BYTES`, `BSTD_MAX_DECODE_ELEMENTS`
and `BSTD_MAX_DECODE_DEPTH` before including `benc.h`. The historical C
`_alloc` entrypoints without a freer are suitable only for plain elements
without nested ownership when errors are possible; use their `_alloc_ex`
versions with cleanup callbacks for owned strings, pointers, slices and maps.
For Go recursive declarations, create `NewDecodeDepth(128)` and use the
`...WithDepth` entrypoints, forwarding the same budget through each nested
callback. The historical Go callbacks cannot pass this state automatically;
generated decoders must adopt the state-aware entrypoints before their
recursive inputs have an enforceable depth limit. This is an API limitation,
not a wire-format change.

The same per-field byte and element limits are used by the other runtime
decoders. JavaScript and TypeScript return `bigint` for decoded `int` and
`uint` regardless of magnitude; unsafe `number` arguments are rejected.
Python, C# and Swift expose signed 64-bit Unix nanoseconds for exact
timestamp round trips; their convenience date types have lower precision.
Encoding a date outside the signed-64 nanosecond window is an error. Errors
do not return a successful cursor; runtimes with mutable read cursors commit
composite reads only after successful validation where supported.

The v1 corpus is in [`testdata/conformance/wire-v1.json`](../testdata/conformance/wire-v1.json).
Integers in JSON are decimal strings so a JavaScript test runner cannot round
64-bit values. Each hex string is an independent expected byte sequence, not
an output copied from the implementation under test.
