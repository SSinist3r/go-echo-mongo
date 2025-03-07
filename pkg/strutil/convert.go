package strutil

import (
	"encoding/base64"
	"encoding/json"
	"strconv"
	"strings"
	"time"
)

// ToInt converts a string to an integer
func ToInt(s string) (int, error) {
	return strconv.Atoi(s)
}

// ToInt64 converts a string to an int64
func ToInt64(s string) (int64, error) {
	return strconv.ParseInt(s, 10, 64)
}

// ToFloat64 converts a string to a float64
func ToFloat64(s string) (float64, error) {
	return strconv.ParseFloat(s, 64)
}

// ToBool converts a string to a boolean
func ToBool(s string) (bool, error) {
	return strconv.ParseBool(s)
}

// ToTime converts a string to time.Time using the specified layout
func ToTime(s string, layout string) (time.Time, error) {
	return time.Parse(layout, s)
}

// ToBase64 converts a string to base64
func ToBase64(s string) string {
	return base64.StdEncoding.EncodeToString([]byte(s))
}

// FromBase64 converts a base64 string back to a regular string
func FromBase64(s string) (string, error) {
	bytes, err := base64.StdEncoding.DecodeString(s)
	if err != nil {
		return "", err
	}
	return string(bytes), nil
}

// ToJSON converts an interface to a JSON string
func ToJSON(v interface{}) (string, error) {
	bytes, err := json.Marshal(v)
	if err != nil {
		return "", err
	}
	return string(bytes), nil
}

// FromJSON converts a JSON string to an interface
func FromJSON(s string, v interface{}) error {
	return json.Unmarshal([]byte(s), v)
}

// ToSlice splits a string into a slice using the specified separator
func ToSlice(s string, sep string) []string {
	if s == "" {
		return []string{}
	}
	return strings.Split(s, sep)
}

// FromSlice joins a slice of strings using the specified separator
func FromSlice(slice []string, sep string) string {
	return strings.Join(slice, sep)
}

// ToMap converts a string to a map using the specified separators
func ToMap(s, pairSep, kvSep string) map[string]string {
	result := make(map[string]string)
	pairs := ToSlice(s, pairSep)

	for _, pair := range pairs {
		kv := ToSlice(pair, kvSep)
		if len(kv) == 2 {
			result[strings.TrimSpace(kv[0])] = strings.TrimSpace(kv[1])
		}
	}

	return result
}

// FromMap converts a map to a string using the specified separators
func FromMap(m map[string]string, pairSep, kvSep string) string {
	pairs := make([]string, 0, len(m))
	for k, v := range m {
		pairs = append(pairs, k+kvSep+v)
	}
	return FromSlice(pairs, pairSep)
}
