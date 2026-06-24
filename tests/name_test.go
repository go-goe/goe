package tests

import (
	"testing"

	"github.com/go-goe/goe/utils"
)

func TestToSnakeCase(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "single lowercase letter",
			input:    "a",
			expected: "a",
		},
		{
			name:     "single uppercase letter",
			input:    "A",
			expected: "a",
		},
		{
			name:     "two letters camelCase",
			input:    "aB",
			expected: "a_b",
		},
		{
			name:     "two letters uppercase",
			input:    "AB",
			expected: "ab",
		},
		{
			name:     "single word lowercase",
			input:    "user",
			expected: "user",
		},
		{
			name:     "single word uppercase",
			input:    "USER",
			expected: "user",
		},
		{
			name:     "all uppercase",
			input:    "URL",
			expected: "url",
		},
		{
			name:     "all uppercase longer",
			input:    "HTTP",
			expected: "http",
		},
		{
			name:     "simple camelCase",
			input:    "userName",
			expected: "user_name",
		},
		{
			name:     "camelCase with multiple words",
			input:    "getUserName",
			expected: "get_user_name",
		},
		{
			name:     "camelCase ending with uppercase",
			input:    "userID",
			expected: "user_id",
		},
		{
			name:     "PascalCase",
			input:    "UserName",
			expected: "user_name",
		},
		{
			name:     "camelCase with consecutive uppercase",
			input:    "parseXMLData",
			expected: "parse_xml_data",
		},
		{
			name:     "PascalCase with consecutive uppercase",
			input:    "XMLParser",
			expected: "xml_parser",
		},
		{
			name:     "mixed case with numbers",
			input:    "user123",
			expected: "user123",
		},
		{
			name:     "already snake_case",
			input:    "user_name",
			expected: "user_name",
		},
		// {
		// 	name:     "complex camelCase",
		// 	input:    "getUserXMLHTTPRequest",
		// 	expected: "get_user_xml_http_request",
		// },
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := utils.ToSnakeCase(tt.input)
			if result != tt.expected {
				t.Errorf("ToSnakeCase(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}
