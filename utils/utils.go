package utils

import (
	"strings"
	"unicode"
)

// TableNamePattern is the default name patterning for mapping struct to table
func TableNamePattern(name string) string {
	if len(name) == 0 {
		return name
	}
	name = namePattern(name)
	if name[len(name)-1] != 's' {
		name += "s"
	}
	return name
}

// ColumnNamePattern is the default name patterning for mapping struct fields to table columns
func ColumnNamePattern(name string) string {
	if len(name) == 0 {
		return name
	}
	return namePattern(name)
}

func isUpper(name string) bool {
	for _, r := range name[1:] {
		if unicode.IsLower(r) {
			return false
		}
	}
	return true
}

func namePattern(name string) string {
	if isUpper(name) {
		return strings.ToLower(name)
	}

	result := strings.Builder{}
	result.WriteRune(rune(name[0]))
	for i := 1; i < len(name); i++ {
		currentLetter, prevLetter := rune(name[i]), rune(name[i-1])
		if unicode.IsUpper(currentLetter) && unicode.IsLower(prevLetter) {
			result.WriteRune('_')
		} else if c := i + 1; c < len(name) && (unicode.IsUpper(currentLetter) && unicode.IsLower(rune(name[c]))) {
			result.WriteRune('_')
		}
		result.WriteRune(currentLetter)
	}
	return strings.ToLower(result.String())
}
