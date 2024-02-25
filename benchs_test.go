package bstd

import (
	"github.com/deneonet/benc/bpre"
	"github.com/deneonet/benc/btag"
	"github.com/deneonet/benc/bunsafe"
	"testing"
)

func BenchmarkStringTag(b *testing.B) {
	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_, buf := btag.SMarshal(0, "1")
		_, t, _ := btag.SUnmarshal(0, buf)

		if t != "1" {
			b.Fatal("tag don't match")
		}
	}
}

func BenchmarkUIntTag(b *testing.B) {
	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_, buf := btag.UMarshal(0, 1)
		_, t, _ := btag.UUnmarshal(0, buf)

		if t != 1 {
			b.Fatal("tag don't match")
		}
	}
}

func BenchmarkPreAllocations(b *testing.B) {
	b.ReportAllocs()
	b.ResetTimer()

	// pre-allocates a byte slice of size 1000
	bpre.Marshal(1000)

	for i := 0; i < b.N; i++ {
		s := SizeString("Hello World!")
		s += SizeFloat64()

		// doesn't allocate any memory now, because it takes the needed bytes, from the pre-allocated byte slice
		n, buf := Marshal(s)
		n = bunsafe.MarshalString(n, buf, "Hello World!")
		n = MarshalFloat64(n, buf, 1231.5131)

		if err := VerifyMarshal(n, buf); err != nil {
			b.Fatal(err.Error())
		}

		// for simplicity, we just skip the string and float64
		n, err := SkipString(0, buf)
		if err != nil {
			b.Fatal(err.Error())
		}

		n, err = SkipFloat64(n, buf)
		if err != nil {
			b.Fatal(err.Error())
		}

		if err := VerifyUnmarshal(n, buf); err != nil {
			b.Fatal(err.Error())
		}
	}

	// resets the buffer that is reused, so it's not going to be reused again
	bpre.Reset()
}

func BenchmarkNoPreAllocations(b *testing.B) {
	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		s := SizeString("Hello World!")
		s += SizeFloat64()

		// allocates memory each op of the size needed
		n, buf := Marshal(s)
		n = bunsafe.MarshalString(n, buf, "Hello World!")
		n = MarshalFloat64(n, buf, 1231.5131)

		if err := VerifyMarshal(n, buf); err != nil {
			b.Fatal(err.Error())
		}

		// for simplicity, we just skip the string and float64
		n, err := SkipString(0, buf)
		if err != nil {
			b.Fatal(err.Error())
		}

		n, err = SkipFloat64(n, buf)
		if err != nil {
			b.Fatal(err.Error())
		}

		if err := VerifyUnmarshal(n, buf); err != nil {
			b.Fatal(err.Error())
		}
	}
}
