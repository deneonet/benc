# Best practices when using Benc

- [Use buffer reuse](README.md#buffer-reuse)
- Use `bstd.MarshalByteSlice` and `bstd.UnmarshalByteSlice`, than using `bstd.MarshalSlice` with byte as typ, for better performance