package util

// Ordered is a type constraint that matches any ordered type.
// An ordered type is one that supports the <, <=, >, and >= operators.
type Ordered interface {
	int | int8 | int16 | int32 | int64 |
		uint | uint8 | uint16 | uint32 | uint64 | uintptr |
		float32 | float64 |
		string
}

func Min[T Ordered](x, y T) T {
	if x < y {
		return x
	}
	return y
}

func Max[T Ordered](x, y T) T {
	if x < y {
		return y
	}
	return x
}
