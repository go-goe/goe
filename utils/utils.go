package utils

import "strings"

// TableNamePattern is the default name patterning for mapping struct to table
func TableNamePattern(name string) string {
	if name[len(name)-1] != 's' {
		name += "s"
	}
	return strings.ToLower(name)
}

// ColumnNamePattern is the default name patterning for mapping struct fields to table columns
func ColumnNamePattern(name string) string {
	return strings.ToLower(name)
}
