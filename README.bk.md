# Benc

![go workflow](https://github.com/deneonet/benc/actions/workflows/go.yml/badge.svg)
[![go report card](https://goreportcard.com/badge/github.com/deneonet/benc)](https://goreportcard.com/report/github.com/deneonet/benc)
[![go reference](https://pkg.go.dev/badge/github.com/deneonet/benc.svg)](https://pkg.go.dev/github.com/deneonet/benc)

The fastest serializer in pure Golang, with the option for backward/forward compatibile generated code.

This module is split into four main packages:

- **[cmd/bencgen](cmd/bencgen/README.md)** - the code-generator for benc
- **[impl/gen](impl/gen/README.md)** - the implementation for bencgen, for handling backward and forward compatibility
- **[std](std/README.md)** - the benc standard, raw serialization
- **[idv](idv/README.md)** - the benc ID validation, raw serialization with ID prefixing

### Changelog

- [See changelog here](CHANGELOG.md)

### Benchmarks

- [See benchmarks here](https://github.com/alecthomas/go_serialization_benchmarks)

### Best Practices
- [See best practices here](BESTPRACTICES.md)

## License

[MIT](https://choosealicense.com/licenses/mit/)
