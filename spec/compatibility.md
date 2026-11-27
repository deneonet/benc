# Release compatibility

Release: `2.0.0` · wire: `v1` · generator input/output: Go AST.

The source declaration determines field order and type; the wire stream has
no type identifiers. Regenerate consumers and producers from matching source
declarations. There is no independent schema or IDL. Generated files must be
used with a runtime from the same release. Go libraries and the Swift package
take their release version from the repository Git tag `v2.0.0`; other package
manifests carry `2.0.0` directly. This document describes a release candidate
until the tag is published.

| Component | Source or installation | Version |
| --- | --- | --- |
| CLI and Go runtime | `go build ./cmd`, `github.com/banditmoscow1337/benc/std/golang` | Git tag `v2.0.0` |
| C / C++ | `std/c/benc.h`, `std/c++/std.hpp` | repository tag `v2.0.0` |
| JavaScript / TypeScript | `npm install ./std/javascript` | `2.0.0`, one package |
| Python | `python -m pip install ./std/python` | `2.0.0` |
| Rust | `std/rust/Cargo.toml` | `2.0.0` |
| Java | `std/java/pom.xml` | `2.0.0` |
| Kotlin | `std/kotlin/build.gradle.kts` | `2.0.0` |
| Scala | `std/scala/build.sbt` | `2.0.0` |
| C# | `std/c#/benc/benc.csproj` | `2.0.0` |
| Swift | `std/swift/Package.swift` | repository tag `v2.0.0` |

Supported CI architectures are **linux/amd64 and linux/arm64**. Native 32-bit
targets are not part of this release's test matrix. `int`, `uint`, `intptr_t`
and `uintptr_t` reflect native word width; cross-architecture interchange
must use fixed-width fields. C, C++ and other runtimes must agree on the
same declarations and the wire-v1 contract in [`wire-v1.md`](wire-v1.md).

On every PR and main push, CI runs Go AST parser → generator tests for all
12 sources × 12 targets, compiles/runs Go, C, C++, JS, Python and Java for
each source, compiles the remaining targets from all 12 source parsers, and
checks their nonzero wire against Go. It also runs each runtime's unit tests, package installation
or build, repeat-generation checks, malformed-wire tests, bounded Go fuzzing,
C/C++ sanitizers, and Go tests on both supported architectures. The weekly and manually
triggered workflows repeat the full source × target matrix before a release. A compiler
missing from a CI runner is a failure, not a successful skip.

The single nonzero cross-language fixture checks exact wire bytes and each
runtime's own decode. These checks cover a real exchange for that fixture:
every producer's bytes equal the same independent value that every tested
consumer decodes. The shared JSON corpus runs fully in Go, JS/TS and Python;
other ports have targeted malformed-wire tests. It does **not** establish
arbitrary combinations of nested values or untested platform support.

See [`../MIGRATION.md`](../MIGRATION.md) when upgrading data produced by
earlier versions.
