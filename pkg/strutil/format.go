package strutil

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

var (
	numberRegex = regexp.MustCompile("[0-9]+")
	upperRegex  = regexp.MustCompile("[A-Z]")
	titleCaser  = cases.Title(language.English)
)

// ToSnake converts a string to snake_case
func ToSnake(s string) string {
	var result strings.Builder
	for i, r := range s {
		if i > 0 && upperRegex.MatchString(string(r)) {
			result.WriteRune('_')
		}
		result.WriteRune(r)
	}
	return strings.ToLower(result.String())
}

// ToCamel converts a string to camelCase
func ToCamel(s string) string {
	s = strings.TrimSpace(s)
	if s == "" {
		return s
	}

	words := strings.FieldsFunc(s, func(r rune) bool {
		return r == '_' || r == ' ' || r == '-'
	})

	for i := 1; i < len(words); i++ {
		words[i] = titleCaser.String(words[i])
	}

	return strings.ToLower(words[0]) + strings.Join(words[1:], "")
}

// ToPascal converts a string to PascalCase
func ToPascal(s string) string {
	s = ToCamel(s)
	if s == "" {
		return s
	}
	return strings.ToUpper(s[:1]) + s[1:]
}

// ToKebab converts a string to kebab-case
func ToKebab(s string) string {
	return strings.ReplaceAll(ToSnake(s), "_", "-")
}

// FormatNumber formats a number with thousand separators and decimal places
func FormatNumber(n float64, decimals int) string {
	format := fmt.Sprintf("%%.%df", decimals)
	str := fmt.Sprintf(format, n)
	parts := strings.Split(str, ".")

	intPart := parts[0]
	var result strings.Builder

	for i, r := range intPart {
		if i > 0 && (len(intPart)-i)%3 == 0 {
			result.WriteRune(',')
		}
		result.WriteRune(r)
	}

	if len(parts) > 1 {
		result.WriteRune('.')
		result.WriteString(parts[1])
	}

	return result.String()
}

// FormatBytes formats bytes to human readable format (KB, MB, etc)
func FormatBytes(bytes int64) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}

	div, exp := int64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}

	return fmt.Sprintf("%.1f %cB", float64(bytes)/float64(div), "KMGTPE"[exp])
}

// Truncate truncates a string to the specified length and adds ellipsis if needed
func Truncate(s string, length int) string {
	if length <= 0 {
		return ""
	}

	if len(s) <= length {
		return s
	}

	return s[:length-3] + "..."
}

// RemoveSpecialChars removes all special characters from a string
func RemoveSpecialChars(s string) string {
	reg := regexp.MustCompile("[^a-zA-Z0-9]+")
	return reg.ReplaceAllString(s, "")
}

// ExtractNumbers extracts all numbers from a string and returns them as a slice of integers
func ExtractNumbers(s string) []int {
	matches := numberRegex.FindAllString(s, -1)
	numbers := make([]int, 0, len(matches))

	for _, match := range matches {
		if n, err := strconv.Atoi(match); err == nil {
			numbers = append(numbers, n)
		}
	}

	return numbers
}
