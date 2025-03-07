package secutil

import (
	"crypto/hmac"
	"crypto/md5"
	"crypto/sha256"
	"crypto/sha512"
	"encoding/hex"
	"fmt"
	"hash"

	"golang.org/x/crypto/bcrypt"
)

const (
	// DefaultCost is the default bcrypt cost factor
	DefaultCost = 12
)

// HashPassword creates a bcrypt hash of a password
func HashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), DefaultCost)
	if err != nil {
		return "", fmt.Errorf("failed to hash password: %w", err)
	}
	return string(bytes), nil
}

// VerifyPassword checks if a password matches its hash
func VerifyPassword(hashedPassword, password string) error {
	return bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password))
}

// HashString creates a hash of a string using the specified algorithm
func HashString(s, algorithm string) (string, error) {
	var h hash.Hash
	switch algorithm {
	case "md5":
		h = md5.New()
	case "sha256":
		h = sha256.New()
	case "sha512":
		h = sha512.New()
	default:
		return "", fmt.Errorf("unsupported hash algorithm: %s", algorithm)
	}

	h.Write([]byte(s))
	return hex.EncodeToString(h.Sum(nil)), nil
}

// CreateHMAC creates an HMAC of a message using the specified key and hash algorithm
func CreateHMAC(message, key string, algorithm string) (string, error) {
	var h func() hash.Hash
	switch algorithm {
	case "sha256":
		h = sha256.New
	case "sha512":
		h = sha512.New
	default:
		return "", fmt.Errorf("unsupported HMAC algorithm: %s", algorithm)
	}

	mac := hmac.New(h, []byte(key))
	mac.Write([]byte(message))
	return hex.EncodeToString(mac.Sum(nil)), nil
}

// VerifyHMAC verifies an HMAC signature
func VerifyHMAC(message, key, signature, algorithm string) (bool, error) {
	expectedMAC, err := CreateHMAC(message, key, algorithm)
	if err != nil {
		return false, err
	}
	return hmac.Equal([]byte(signature), []byte(expectedMAC)), nil
}

// CompareHashes compares two hashes in constant time to prevent timing attacks
func CompareHashes(hash1, hash2 string) bool {
	return hmac.Equal([]byte(hash1), []byte(hash2))
}
