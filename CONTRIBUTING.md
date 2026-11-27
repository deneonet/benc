# Contributing to Benc

Changes to a parser must preserve the supported Go AST field types and order;
changes to a generator must work with that same AST. Do not add a second
schema or silently omit unsupported source fields. When the wire changes,
first update `spec/wire-v1.md` and the independent corpus.

Before sending a PR, run `go test ./...`, the relevant runtime command in
`testdata/conformance/runtime-tests.sh`, and
`python3 testdata/conformance/check-release.py` if packaging changed. CI runs
the cross-language matrix, native compilers, sanitizers and package installation.
See `spec/compatibility.md` for the release gates and supported architectures.

Bug fixes, new ports and performance work are welcome. Include a regression
case for observable wire or AST behavior, and benchmarks for performance
changes.
