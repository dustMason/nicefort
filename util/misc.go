package util

func ClampedInt(i, min, max int) int {
	if i < min {
		return min
	}
	if i > max {
		return max
	}
	return i
}
