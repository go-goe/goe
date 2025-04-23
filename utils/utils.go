package utils

import (
	"strings"
	"unicode"
)

// TableNamePattern is the default name patterning for mapping struct to table
func TableNamePattern(name string) string {
	name = namePattern(name)
	if name[len(name)-1] != 's' {
		name += "s"
	}
	return name
}

// ColumnNamePattern is the default name patterning for mapping struct fields to table columns
func ColumnNamePattern(name string) string {
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
	result.WriteRune(unicode.ToLower(rune(name[0])))
	for _, r := range name[1:] {
		if unicode.IsUpper(r) {
			result.WriteRune('_')
			result.WriteRune(unicode.ToLower(r))
			continue
		}
		result.WriteRune(r)
	}
	return result.String()
}
