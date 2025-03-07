package strutil

import (
	"net/mail"
	"net/url"
	"regexp"
	"unicode"
)

var (
	phoneRegex    = regexp.MustCompile(`^\+?[0-9]{10,15}$`)
	usernameRegex = regexp.MustCompile(`^[a-zA-Z0-9_-]{3,30}$`)
	uuidRegex     = regexp.MustCompile(`^[0-9a-fA-F]{8}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{12}$`)
)

// IsEmail validates if a string is a valid email address
func IsEmail(email string) bool {
	_, err := mail.ParseAddress(email)
	return err == nil
}

// IsPhone validates if a string is a valid phone number
// This is a basic implementation; consider using a phone number library for production
func IsPhone(phone string) bool {
	return phoneRegex.MatchString(phone)
}

// IsURL validates if a string is a valid URL
func IsURL(s string) bool {
	u, err := url.Parse(s)
	return err == nil && u.Scheme != "" && u.Host != ""
}

// IsUsername validates if a string is a valid username
func IsUsername(username string) bool {
	return usernameRegex.MatchString(username)
}

// IsUUID validates if a string is a valid UUID
func IsUUID(uuid string) bool {
	return uuidRegex.MatchString(uuid)
}

// IsStrongPassword checks if a password meets security requirements
func IsStrongPassword(password string) bool {
	if len(password) < 8 {
		return false
	}

	var (
		hasUpper   bool
		hasLower   bool
		hasNumber  bool
		hasSpecial bool
	)

	for _, char := range password {
		switch {
		case unicode.IsUpper(char):
			hasUpper = true
		case unicode.IsLower(char):
			hasLower = true
		case unicode.IsNumber(char):
			hasNumber = true
		case unicode.IsPunct(char) || unicode.IsSymbol(char):
			hasSpecial = true
		}
	}

	return hasUpper && hasLower && hasNumber && hasSpecial
}

// HasMinLength checks if a string meets the minimum length requirement
func HasMinLength(s string, min int) bool {
	return len(s) >= min
}

// HasMaxLength checks if a string meets the maximum length requirement
func HasMaxLength(s string, max int) bool {
	return len(s) <= max
}

// IsAlphanumeric checks if a string contains only letters and numbers
func IsAlphanumeric(s string) bool {
	for _, r := range s {
		if !unicode.IsLetter(r) && !unicode.IsNumber(r) {
			return false
		}
	}
	return true
}

// ContainsOnly checks if a string contains only characters from the allowed set
func ContainsOnly(s string, allowed string) bool {
	allowedMap := make(map[rune]bool)
	for _, r := range allowed {
		allowedMap[r] = true
	}

	for _, r := range s {
		if !allowedMap[r] {
			return false
		}
	}
	return true
}
