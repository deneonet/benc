package bidv

import (
	"fmt"

	bstd "github.com/deneonet/benc/std"
)

const (
	Int16 uint = iota + 2
	Int32
	Int64
	UInt16
	UInt32
	UInt64
	Float32
	Float64
	Bool
	Byte
	String
	Slice
	Map
	ByteSlice
)

const AllowedStartId = 16

// Returns the nickname for the standard IDs for all data types
func GetDefaultIdNickname(id uint) string {
	switch id {
	case Int16:
		return "Int16"
	case Int32:
		return "Int32"
	case Int64:
		return "Int64"
	case UInt16:
		return "Uint16"
	case UInt32:
		return "Uint32"
	case UInt64:
		return "Uint64"
	case Float32:
		return "Float32"
	case Float64:
		return "Float64"
	case Bool:
		return "Bool"
	case Byte:
		return "Byte"
	case String:
		return "String"
	case Slice:
		return "Slice"
	case Map:
		return "Map"
	case ByteSlice:
		return "Byte slice"
	default:
		return "N/A"
	}
}

var GetIdNickname = GetDefaultIdNickname

type SkipFunc func(n int, b []byte) (int, error)
type MarshalFunc[T any] func(n int, b []byte, t T) int

func Skip(n int, b []byte, id uint, skipper SkipFunc) (int, error) {
	n, dId, err := bstd.UnmarshalUint(n, b)
	if err != nil {
		return 0, err
	}

	if dId != id {
		nn := GetIdNickname(id)
		dNn := GetIdNickname(dId)
		return 0, fmt.Errorf("id mismatch: expected %s (%d), got %s (%d)", nn, id, dNn, dId)
	}

	return skipper(n, b)
}

func Size(id uint, s int) int {
	return bstd.SizeUint(id) + s
}

func Marshal(n int, b []byte, id uint) int {
	return bstd.MarshalUint(n, b, id)
}

func Unmarshal[T any](tn int, b []byte, id uint, unmarshaler any) (n int, t T, err error) {
	n, dId, err := bstd.UnmarshalUint(tn, b)
	if err != nil {
		n = 0
		return
	}

	if dId != id {
		n = 0
		nn := GetIdNickname(id)
		dNn := GetIdNickname(dId)
		err = fmt.Errorf("id mismatch: expected %s (%d), got %s (%d)", nn, id, dNn, dId)
		return
	}

	switch p := unmarshaler.(type) {
	case func(n int, b []byte) (int, T, error):
		n, t, err = p(n, b)
	case func(n int, b []byte, v *T) (int, error):
		n, err = p(n, b, &t)
	default:
		panic("benc: invalid `unmarshaler` provided in `bidv.Unmarshal`")
	}
	return
}
