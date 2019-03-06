package utils

//SubString safely shorten the string from beginning until <size> characters
func SubString(s string, size int) string {
	if size >= 0 && len(s) > 0 && len(s) >= size {
		s = s[:size]
	}
	return s
}

// PosString returns the first index of element in slice.
// If slice does not contain element, returns -1.
func PosString(slice []string, element string) int {
	for index, elem := range slice {
		if elem == element {
			return index
		}
	}
	return -1
}

// ContainsString returns true if slice contains element
func ContainsString(slice []string, element string) bool {
	return !(PosString(slice, element) == -1)
}
