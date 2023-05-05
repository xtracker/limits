package util

import "time"

func Max[T int | int64 | float64 | time.Duration](x, y T) T {
	if x < y {
		return y
	}

	return x
}

func Min[T int | int64 | float64 | time.Duration](x, y T) T {
	if x < y {
		return x
	}

	return y
}
