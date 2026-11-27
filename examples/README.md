# Executable generator examples

From the repository root, run `bash examples/quickstart.sh js`, `py`, `c`,
`cpp` or `java`. Each run starts with a Go type declaration, passes it through
Go AST to the selected target, compiles or interprets the generated consumer,
and runs its generated test in a fresh temporary directory. Install the Go
toolchain and the selected language toolchain first.

For Rust, Kotlin, C#, Swift, Scala and TypeScript, run
`bash testdata/conformance/generated-native.sh rust` (or `kotlin`, `csharp`,
`swift`, `scala`, `typescript`). The optional second argument selects another
source parser, for example `bash testdata/conformance/generated-native.sh
rust py` for Python declarations → Go AST → Rust output. That fixture
compares nonzero wire bytes with the independent Go value.
