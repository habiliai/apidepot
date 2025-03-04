package util

import "strings"

func SplitStringToPair(s string, sep string) (string, string) {
	parts := strings.SplitN(s, sep, 2)
	if len(parts) == 1 {
		return parts[0], ""
	}
	return parts[0], parts[1]
}

func DefaultIfEmpty(s string, defaultValue string) string {
	if s == "" {
		return defaultValue
	}
	return s
}
