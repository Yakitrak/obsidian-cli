// Package scoring provides APIs that calculates similarity scores between two strings.
package scoring

// Calculate calculates a similarity score between s1 and s2.
// The length of s1 must be greater or equal than the length of s2.
func Calculate(s1, s2 string) (int, [2]int) {
	if len(s1) < len(s2) {
		panic("len(s1) must be greater than or equal to len(s2)")
	}

	return smithWaterman([]rune(s1), []rune(s2))
}

// max returns the biggest number from passed args.
// If the number of args is 0, it always returns 0.
func max(n ...int32) (min int32) {
	if len(n) == 0 {
		return 0
	}
	min = n[0]
	for _, a := range n[1:] {
		if a > min {
			min = a
		}
	}
	return
}
