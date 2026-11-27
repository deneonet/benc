package models

import "time"

type Root struct {
	S        string
	I        int
	U        uint
	Count    uint8
	Fixed    [3]byte
	Numbers  [2]int32
	Raw      []byte
	Children []Child
	Groups   [][]byte
	P        *Child
	Dict     map[string]Child
	Named    Labels
	When     time.Time
}

type Labels map[string]int32

type Child struct {
	ID     int64
	Text   string
	Scores []uint8
}
