package utils

//SubString safely shorten the string from beginning until <size> characters
func SubString(s string, size int) string {
	if size >= 0 && len(s) > 0 && len(s) >= size {
		s = s[:size]
	}
	return s
}
