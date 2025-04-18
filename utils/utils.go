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

func namePattern(name string) string {
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
