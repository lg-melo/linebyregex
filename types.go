package linebyregex

func isSpecial(c byte) bool {
	switch c {
	case '(', ')', '[', ']', '{', ',', '}', '?', '+', '*', '\\', '|', '.':
		return true
	}

	return false
}

func isRepetition(c byte) bool {
	return c == '?' || c == '+' || c == '*'
}

func isClass(c byte) bool {
	switch c {
	case 'w', 'W', 'd', 'D', 's', 'S':
		return true
	}

	return false
}

func isAlphaNum(c byte) bool {
	return isDigit(c) || 'a' <= c && c <= 'z' || 'A' <= c && c <= 'Z' || c == '_';
}

func isDigit(c byte) bool {
	return ('0' <= c && c <= '9')
}

func isSpace(c byte) bool {
	return (c == ' ' || c == '\t' || c == '\r' || c == '\n' || c == '\f')
}