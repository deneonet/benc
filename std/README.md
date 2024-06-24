# benc std

The Benc standard provides a suite of methods for raw sizing, skipping, marshalling, and unmarshalling of Go types. When I refer to "raw", it means that only the essential elements are serialized, for example, serialized data is not prefixed with their corresponding type information. 

## Installation
```bash
go get go.kine.bz/benc/std
```

## Tests
Code coverage of `bstd.go` is approximately 95%