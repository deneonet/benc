package bstd

// MaxDecodeDepth is the recommended nesting budget for the state-aware APIs.
const MaxDecodeDepth = 128

// DecodeDepth is a per-operation recursion budget. Pass the same value through
// every nested state-aware callback; it is not safe to share across goroutines.
// The original callback API remains available for compatibility and cannot
// automatically carry a depth budget across callbacks.
type DecodeDepth struct {
	maxDepth int
	depth    int
}

func NewDecodeDepth(maxDepth int) *DecodeDepth {
	if maxDepth <= 0 {
		panic("benc: decode depth must be positive")
	}
	return &DecodeDepth{maxDepth: maxDepth}
}

func (d *DecodeDepth) enter() error {
	if d == nil || d.depth >= d.maxDepth {
		return ErrDecodeLimit
	}
	d.depth++
	return nil
}

func (d *DecodeDepth) leave() { d.depth-- }

func SkipSliceWithDepth(n int, b []byte, d *DecodeDepth, skip func(int, []byte, *DecodeDepth) (int, error)) (int, error) {
	if err := d.enter(); err != nil {
		return 0, err
	}
	defer d.leave()
	return SkipSlice(n, b, func(n int, b []byte) (int, error) { return skip(n, b, d) })
}

func SkipMapWithDepth(n int, b []byte, d *DecodeDepth, key, value func(int, []byte, *DecodeDepth) (int, error)) (int, error) {
	if err := d.enter(); err != nil {
		return 0, err
	}
	defer d.leave()
	return SkipMap(n, b,
		func(n int, b []byte) (int, error) { return key(n, b, d) },
		func(n int, b []byte) (int, error) { return value(n, b, d) })
}

func SkipPointerWithDepth(n int, b []byte, d *DecodeDepth, skip func(int, []byte, *DecodeDepth) (int, error)) (int, error) {
	if err := d.enter(); err != nil {
		return 0, err
	}
	defer d.leave()
	return SkipPointer(n, b, func(n int, b []byte) (int, error) { return skip(n, b, d) })
}

func UnmarshalSliceWithDepth[T any](n int, b []byte, d *DecodeDepth, decode func(int, []byte, *DecodeDepth) (int, T, error)) (int, []T, error) {
	if err := d.enter(); err != nil {
		return 0, nil, err
	}
	defer d.leave()
	return UnmarshalSlice[T](n, b, func(n int, b []byte) (int, T, error) { return decode(n, b, d) })
}

func UnmarshalMapWithDepth[K comparable, V any](n int, b []byte, d *DecodeDepth,
	key func(int, []byte, *DecodeDepth) (int, K, error), value func(int, []byte, *DecodeDepth) (int, V, error)) (int, map[K]V, error) {
	if err := d.enter(); err != nil {
		return 0, nil, err
	}
	defer d.leave()
	return UnmarshalMap[K, V](n, b,
		func(n int, b []byte) (int, K, error) { return key(n, b, d) },
		func(n int, b []byte) (int, V, error) { return value(n, b, d) })
}

func UnmarshalPointerWithDepth[T any](n int, b []byte, d *DecodeDepth, decode func(int, []byte, *DecodeDepth) (int, T, error)) (int, *T, error) {
	if err := d.enter(); err != nil {
		return 0, nil, err
	}
	defer d.leave()
	return UnmarshalPointer[T](n, b, func(n int, b []byte) (int, T, error) { return decode(n, b, d) })
}

func UnmarshalArrayWithDepth[T any](n int, b []byte, d *DecodeDepth, dst []T, decode func(int, []byte, *DecodeDepth) (int, T, error)) (int, error) {
	if err := d.enter(); err != nil {
		return 0, err
	}
	defer d.leave()
	return UnmarshalArray(n, b, dst, func(n int, b []byte) (int, T, error) { return decode(n, b, d) })
}
