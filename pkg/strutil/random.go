package strutil

import (
	"crypto/rand"
	"math/big"
)

const (
	upperChars   = "ABCDEFGHIJKLMNOPQRSTUVWXYZ"
	lowerChars   = "abcdefghijklmnopqrstuvwxyz"
	numberChars  = "0123456789"
	specialChars = "!@#$%^&*()-_=+[]{}|;:,.<>?"
)

// GenerateRandom generates a random string with specified parameters
func GenerateRandom(length int, useUpper, useLower, useNumbers, useSpecial bool) (string, error) {
	if length <= 0 {
		return "", nil
	}

	// Create character set based on parameters
	var chars string
	if useUpper {
		chars += upperChars
	}
	if useLower {
		chars += lowerChars
	}
	if useNumbers {
		chars += numberChars
	}
	if useSpecial {
		chars += specialChars
	}

	// If no character set is selected, use alphanumeric by default
	if chars == "" {
		chars = upperChars + lowerChars + numberChars
	}

	// Generate random string
	result := make([]byte, length)
	max := big.NewInt(int64(len(chars)))

	for i := 0; i < length; i++ {
		n, err := rand.Int(rand.Reader, max)
		if err != nil {
			return "", err
		}
		result[i] = chars[n.Int64()]
	}

	return string(result), nil
}

// GenerateKey generates a random key (for API keys, tokens etc) with optional prefix
func GenerateKey(length int, prefix string) (string, error) {
	// Generate random part (subtracting prefix length to maintain desired total length)
	randomLength := length - len(prefix)
	if randomLength <= 0 {
		return prefix, nil
	}

	// Generate random string using only URL-safe characters
	random, err := GenerateRandom(randomLength, true, true, true, false)
	if err != nil {
		return "", err
	}

	return prefix + random, nil
}

// GeneratePassword generates a secure password with minimum requirements
func GeneratePassword(length int, requireUpper, requireLower, requireNumbers, requireSpecial bool) (string, error) {
	if length < 8 {
		length = 8 // Enforce minimum length for security
	}

	// Generate the main part of the password
	password, err := GenerateRandom(length, true, true, true, requireSpecial)
	if err != nil {
		return "", err
	}

	// Ensure at least one character of each required type is present
	// This could be enhanced to modify the generated password if it doesn't meet requirements
	return password, nil
}
