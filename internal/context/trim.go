package contextbuilder

func trimTo(value string, max int) string {
	if max <= 0 || len([]rune(value)) <= max {
		return value
	}
	runes := []rune(value)
	return string(runes[len(runes)-max:])
}
