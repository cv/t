package codes

// IsValidIATA checks if a string is a valid IATA code.
// Valid codes are 3 alphanumeric ASCII characters, starting with a letter.
func IsValidIATA(s string) bool {
	if len(s) != 3 {
		return false
	}
	// First character must be a letter
	if !isASCIILetter(s[0]) {
		return false
	}
	// Remaining characters must be alphanumeric
	for i := 1; i < len(s); i++ {
		if !isASCIIAlphanumeric(s[i]) {
			return false
		}
	}
	return true
}

func isASCIILetter(c byte) bool {
	return (c >= 'A' && c <= 'Z') || (c >= 'a' && c <= 'z')
}

func isASCIIAlphanumeric(c byte) bool {
	return isASCIILetter(c) || (c >= '0' && c <= '9')
}
