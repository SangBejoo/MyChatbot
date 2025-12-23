package http

import (
	"regexp"
	"strings"
	"unicode/utf8"
)

// Input validation constants
const (
	MaxSlugLength      = 64
	MaxTitleLength     = 256
	MaxPayloadLength   = 10000
	MaxConfigKeyLength = 64
	MaxConfigValLength = 50000 // For AI prompts
	MaxTableNameLength = 128
)

// ValidSlug checks if a slug is safe (alphanumeric + underscore + hyphen)
func ValidSlug(s string) bool {
	if s == "" || len(s) > MaxSlugLength {
		return false
	}
	matched, _ := regexp.MatchString(`^[a-zA-Z0-9_-]+$`, s)
	return matched
}

// ValidTableName checks if a table name is safe
func ValidTableName(s string) bool {
	if s == "" || len(s) > MaxTableNameLength {
		return false
	}
	// Must start with "dt_" prefix (our dynamic tables)
	if !strings.HasPrefix(s, "dt_") {
		return false
	}
	matched, _ := regexp.MatchString(`^[a-zA-Z0-9_]+$`, s)
	return matched
}

// ValidConfigKey checks if a config key is safe
func ValidConfigKey(s string) bool {
	if s == "" || len(s) > MaxConfigKeyLength {
		return false
	}
	matched, _ := regexp.MatchString(`^[a-zA-Z0-9_]+$`, s)
	return matched
}

// SanitizeString removes null bytes and control characters
func SanitizeString(s string) string {
	// Remove null bytes
	s = strings.ReplaceAll(s, "\x00", "")
	
	// Keep only valid UTF-8
	if !utf8.ValidString(s) {
		v := make([]rune, 0, len(s))
		for _, r := range s {
			if r != utf8.RuneError {
				v = append(v, r)
			}
		}
		s = string(v)
	}
	return s
}

// TruncateString safely truncates a string to max length
func TruncateString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen]
}

// ValidateLength checks if string is within bounds
func ValidateLength(s string, min, max int) bool {
	l := len(s)
	return l >= min && l <= max
}

// EscapeHTML escapes HTML special characters to prevent XSS
func EscapeHTML(s string) string {
	s = strings.ReplaceAll(s, "&", "&amp;")
	s = strings.ReplaceAll(s, "<", "&lt;")
	s = strings.ReplaceAll(s, ">", "&gt;")
	s = strings.ReplaceAll(s, "\"", "&quot;")
	s = strings.ReplaceAll(s, "'", "&#x27;")
	return s
}
