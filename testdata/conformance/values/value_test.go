package models

import (
	"fmt"
	"testing"
	"time"
)

func TestWireValue(t *testing.T) {
	v := Root{S: "A\x00B", I: -7, U: 300, Count: 201, Fixed: [3]byte{1, 2, 3}, Numbers: [2]int32{-2, 5}, Raw: []byte{0, 255}, Children: []Child{{ID: -9, Text: "x", Scores: []uint8{4, 5}}}, Groups: [][]byte{{0, 1}}, P: &Child{ID: 3, Text: "p", Scores: []uint8{6}}, Dict: map[string]Child{"d": {ID: 4, Text: "z", Scores: []uint8{7}}}, Named: Labels{"n": -5}, When: time.Unix(123456789, 123)}
	b := make([]byte, v.Size())
	if n := v.Marshal(0, b); n != len(b) {
		t.Fatal(n, len(b))
	}
	fmt.Printf("WIRE=%x\n", b)
}
