[Back to README](README.md)

# Best practices when using BENC
- Try to use fixed int types, like: int16, int32, int64
- Try to use uint instead of int, if you don't need negative values [difference?](https://www.quora.com/Whats-the-difference-between-uint-and-int-in-golang)
- Avoid using maps, they are really slow
- Use for encoding a byte slice, ByteSlice (MarshalByteSlice, UnmarshalByteSlice)

#### Benchmark: Normal string encoding vs byte slice string encoding:
Byte slice string encoding:
```go
func BenchmarkBS(b *testing.B) {
	data := []byte("Hello")

	for i := 0; i < b.N; i++ {
		s := bstd.SizeSlice[byte](data, bstd.SizeByte)
		n, buf := bstd.Marshal(s)
		n = bstd.MarshalSlice[byte](n, buf, data, bstd.MarshalByte)

		_, data, _ = bstd.UnmarshalSlice[byte](0, buf, bstd.UnmarshalByte)
	}
}
```
Normal string encoding:
```go
func BenchmarkString(b *testing.B) {
	str := "Hello"

	for i := 0; i < b.N; i++ {
		s := bstd.SizeString(str)
		n, buf := bstd.Marshal(s)
		n = bstd.MarshalString(n, buf, str)

		_, _, _ = bstd.UnmarshalString(0, buf)
	}
}
```
Results:
Command:
```bash
perflock -governor 70% go test -bench='.*' ./ -count=1 > "results.txt"
```
```
goos: linux
goarch: amd64
pkg: github.com/deneonet/benc/test
cpu: 11th Gen Intel(R) Core(TM) i5-11300H @ 3.10GHz
BenchmarkString-8            	41685588	        27.21 ns/op
BenchmarkBS-8   	            44105101	        24.73 ns/op
```
As you can see byte slice string encoding is slightly faster.


