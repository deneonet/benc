# Best practices when using BENC

- Use fixed int types, like `int16`, `int32`, `int64`, for optimal performance.
- Prefer `uint` over `int` if negative values aren't required. [difference?](https://www.quora.com/Whats-the-difference-between-uint-and-int-in-golang)
- Do pre allocations:
  - Call ```bpre.Marshal(s)``` once (only once), before doing marshal, s represents the max length of the byte slice, make sure s is always bigger than the calculated size when doing marshal
  - The same is for ```bpre.MFUnmarshal(s)```, just for message framing unmarshal

- For string encoding:
  - Byte slice string encoding is faster than normal string encoding for larger and slightly faster for smaller data sets.
  <br />

  - Example Benchmark:
      ```go
      func BenchmarkByteSliceStringEncoding(b *testing.B) {
        for i := 0; i < b.N; i++ {
			data := []byte("Lorem ipsum dolor sit amet... (1,368 bytes)")
			s := bstd.SizeByteSlice(data)
			n, buf := bstd.Marshal(s)
			n = bstd.MarshalByteSlice(n, buf, data)

			_, data, _ = bstd.UnmarshalByteSlice(0, buf)
			_ = string(data)
		}
      }

      func BenchmarkNormalStringEncoding(b *testing.B) {
        str := "Lorem ipsum dolor sit amet... (1,368 bytes)"
        for i := 0; i < b.N; i++ {
			s := bstd.SizeString(str)
			n, buf := bstd.Marshal(s)
			n = bstd.MarshalString(n, buf, str)

			_, _, _ = bstd.UnmarshalString(0, buf)
		}
      }
      ```
  - Benchmark Results:
    - Command:
      ```bash
      perflock -governor 70% go test -bench='.*' ./ -count=1 > "results.txt"
      ```
    - Results:
      ```bash
      goos: linux
      goarch: amd64
      pkg: github.com/deneonet/benc/test
      cpu: 11th Gen Intel(R) Core(TM) i5-11300H @ 3.10GHz
      BenchmarkString-8           1923907       634.5 ns/op
      BenchmarkFasterBS-8         2459222       492.4 ns/op
      ```
      Byte slice string encoding demonstrates faster performance with larger data sets (1,368 bytes) compared to normal string encoding.
  
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
  - Benchmark Results:
    - Command:
      ```bash
      perflock -governor 70% go test -bench='.*' ./ -count=1 > "results.txt"
      ```
    - Results:
      ```bash
      goos: linux
      goarch: amd64
      pkg: github.com/deneonet/benc/test
      cpu: 11th Gen Intel(R) Core(TM) i5-11300H @ 3.10GHz
      BenchmarkBS-8          	  155154	      7997 ns/op
      BenchmarkFasterBS-8    	 3759979	       306.2 ns/op
      ```
      Optimized byte slice encoding methods significantly outperform the normal byte slice encoding for the given data (1,368 bytes).
      It's recommended to use the faster byte slice encoding methods for improved performance.
