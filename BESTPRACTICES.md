# Best practices when using BENC

- Use fixed int types, like `int16`, `int32`, `int64`, for less serialized bytes performance.
- Prefer `uint` over `int` if negative values aren't required. [difference?](https://www.quora.com/Whats-the-difference-between-uint-and-int-in-golang)
- Do pre allocations:
  - Call `bpre.Marshal(s)` once, before doing all marshals, s represents the max length of the byte slice, make sure s
    is always bigger than the calculated size when doing marshal
  - The same is for `bpre.UnmarshalMF(s)`, just for message framing unmarshal
  - Use `btag.UMarshal(...)` over `btag.SMarshal(...)`, the first function uses a uint as tag, therefore less serialized
    bytes
- For string encoding and pre-allocation/buffer reuse example:
  - You can use `bunsafe` to perform zero allocations string conversions ([]byte -> string, string -> []byte)
  - You can use `bpre` to reuse buffers and not have to allocate a new one every `bstd.Marshal(...)`
  - Example Benchmark:
    ```go
    // You pre allocation to really get that zero allocations and squeeze the performance out of BENC
    func BenchmarkUnsafeStringConversionPreAllocation(b *testing.B) {
      bpre.Marshal(2000)
      b.ReportAllocs()
      for i := 0; i < b.N; i++ {
          data := "Lorem ipsum dolor sit amet... (1,368 bytes)"
          s := SizeString(data)
          n, buf := Marshal(s)
          n = bunsafe.MarshalString(n, buf, data)

          _, data, _ = bunsafe.UnmarshalString(0, buf)
      }
    }

    func BenchmarkUnsafeStringConversion(b *testing.B) {
      b.ReportAllocs()
      for i := 0; i < b.N; i++ {
          data := "Lorem ipsum dolor sit amet... (1,368 bytes)"
          s := SizeString(data)
          n, buf := Marshal(s)
          n = bunsafe.MarshalString(n, buf, data)

          _, data, _ = bunsafe.UnmarshalString(0, buf)
      }
    }

    func BenchmarkStringConversion(b *testing.B) {
      b.ReportAllocs()
      for i := 0; i < b.N; i++ {
          data := "Lorem ipsum dolor sit amet... (1,368 bytes)"
          s := SizeString(data)
          n, buf := Marshal(s)
          n = MarshalString(n, buf, data)

          _, data, _ = UnmarshalString(0, buf)
      }
    }```
  - Results:
    ```bash
    goos: windows
    goarch: amd64
    pkg: github.com/deneonet/benc
    cpu: 11th Gen Intel(R) Core(TM) i5-11300H @ 3.10GHz
    BenchmarkUnsafeStringConversionPreAllocation-8    64019119        18.17 ns/op     0 B/op        0 allocs/op
    BenchmarkUnsafeStringConversion-8                 4851391         225.2 ns/op     1408 B/op     1 allocs/op
    BenchmarkStringConversion-8                       2734479         437.2 ns/op     2816 B/op     2 allocs/op
    PASS
    ```
- For byte slice encoding:
  - Consider optimizing byte slice encoding methods for better performance.
  - Example Benchmark:
    ```go
   	func BenchmarkBS(b *testing.B) {
		for i := 0; i < b.N; i++ {
			data := []byte("Lorem ipsum dolor sit amet... (1,368 bytes)")
			s := bstd.SizeSlice[byte](data, bstd.SizeByte)
			n, buf := bstd.Marshal(s)
			n = bstd.MarshalSlice[byte](n, buf, data, bstd.MarshalByte)

			_, data, _ = bstd.UnmarshalSlice(0, buf, bstd.UnmarshalByte)
		}
	}

	func BenchmarkFasterBS(b *testing.B) {
		for i := 0; i < b.N; i++ {
			data := []byte("Lorem ipsum dolor sit amet... (1,368 bytes)")
			s := bstd.SizeByteSlice(data)
			n, buf := bstd.Marshal(s)
			n = bstd.MarshalByteSlice(n, buf, data)

			_, data, _ = bstd.UnmarshalByteSlice(0, buf)
		}
	}
    ```
  - Results:
    ```bash
    goos: windows
    goarch: amd64
    pkg: github.com/deneonet/benc
    cpu: 11th Gen Intel(R) Core(TM) i5-11300H @ 3.10GHz
    BenchmarkBS-8             141571              7760 ns/op
    BenchmarkFasterBS-8      5825556               244.6 ns/op
    PASS
    ```
    Optimized byte slice encoding methods significantly outperform the normal byte slice encoding for the given data (
    1,368 bytes).
    It's recommended to use the faster byte slice encoding methods for improved performance.