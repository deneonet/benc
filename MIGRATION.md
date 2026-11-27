# Migrating to Benc 2.0 / wire v1

This release locks the byte format in [`spec/wire-v1.md`](spec/wire-v1.md).
Parser input remains ordinary language declarations, normalized to Go AST.
There is no schema migration file or additional type representation.

1. Pin the generator and all runtimes to the same release. Run `benc -version`
   and check the package versions in [`spec/compatibility.md`](spec/compatibility.md).
2. Keep the original field declarations and their order. Generate both sides
   again; check errors for unsupported or ambiguous fields. A dropped field
   shifts every following field's bytes.
3. Exercise actual producer → consumer traffic, including empty and nonzero
   values, byte strings containing NUL, fixed-width integers, nested
   collections, pointers and nanosecond timestamps. Use the wire-v1 corpus
   and native smoke commands before deploying.
4. If persisted records were written by an older generator, read them with
   that generator/runtime in a separate migration step and write new wire-v1
   records with regenerated code. Mark the old/new format outside the Benc
   payload (for example in your storage version or application protocol).

Earlier C/C++ generator paths sometimes encoded integer fields as fixed-width
values where Go encoded varints, and other ports had inconsistent bounds,
terminators and timestamp handling. A wire-v1 decoder cannot reliably infer
which encoding an old C/C++ payload used from the bytes alone. Retain the
exact old declaration, producer version and runtime for that migration; do
not enable an automatic, guess-based fallback. If the old producer version
is unknown, recover it from deployment or storage metadata before decoding.

Wire v1 rejects malformed booleans, collection trailers, invalid UTF-8 and
lengths beyond decoder limits. Timestamp convenience date types in JS,
Python, Swift and C# may lose nanoseconds; use the exact nanosecond API when
round-trip precision matters. `int`/`uint` have architecture-dependent width;
new cross-platform declarations should use explicit `int32`, `uint64`, etc.
Map entry order is not canonical, so compare decoded entries for multi-entry
maps rather than raw byte strings.

The supported release architectures are linux/amd64 and linux/arm64. Treat
32-bit deployment as a separate compatibility effort.
