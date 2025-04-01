package utils

import "strings"

func TableNamePattern(name string) string {
	if name[len(name)-1] != 's' {
		name += "s"
	}
	return strings.ToLower(name)
}

func ColumnNamePattern(name string) string {
	return strings.ToLower(name)
}
