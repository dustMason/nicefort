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

func AbsInt(i int) int {
	if i < 0 {
		return i * -1
	}
	return i
}

var Arrows = map[[2]int]string{
	[2]int{0, 0}:   "•",
	[2]int{1, 1}:   "↘",
	[2]int{1, -1}:  "↗",
	[2]int{-1, 1}:  "↙",
	[2]int{-1, -1}: "↖",
	[2]int{0, -1}:  "↑",
	[2]int{0, 1}:   "↓",
	[2]int{-1, 0}:  "←",
	[2]int{1, 0}:   "→",
}
