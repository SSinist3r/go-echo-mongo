# String Utilities (strutil)

A comprehensive collection of string manipulation, validation, and conversion utilities for Go applications.

## Features

- Random string generation
- String formatting and case conversion
- String validation
- Type conversion
- Common string operations

## Usage

### Random String Generation

```go
import "yourproject/pkg/strutil"

// Generate a random string
random, err := strutil.GenerateRandom(16, true, true, true, false)

// Generate an API key with prefix
apiKey, err := strutil.GenerateKey(32, "api-")

// Generate a secure password
password, err := strutil.GeneratePassword(12, true, true, true, true)
```

### String Formatting

```go
// Case conversion
snake := strutil.ToSnake("MyVariableName")     // my_variable_name
camel := strutil.ToCamel("my_variable_name")   // myVariableName
pascal := strutil.ToPascal("my_variable_name") // MyVariableName
kebab := strutil.ToKebab("MyVariableName")     // my-variable-name

// Number formatting
formatted := strutil.FormatNumber(1234567.89, 2) // 1,234,567.89
bytes := strutil.FormatBytes(1234567)            // 1.2 MB

// String manipulation
truncated := strutil.Truncate("Long text...", 10)      // "Long te..."
cleaned := strutil.RemoveSpecialChars("Hello, World!") // HelloWorld
```

### String Validation

```go
// Basic validation
isEmail := strutil.IsEmail("user@example.com")
isPhone := strutil.IsPhone("+1234567890")
isURL := strutil.IsURL("https://example.com")
isUUID := strutil.IsUUID("550e8400-e29b-41d4-a716-446655440000")

// Password strength
isStrong := strutil.IsStrongPassword("MyP@ssw0rd")

// Length validation
hasMin := strutil.HasMinLength("test", 3)
hasMax := strutil.HasMaxLength("test", 5)

// Character validation
isAlpha := strutil.IsAlphanumeric("Test123")
valid := strutil.ContainsOnly("123", "0123456789")
```

### Type Conversion

```go
// String to primitive types
i, err := strutil.ToInt("123")
f, err := strutil.ToFloat64("123.45")
b, err := strutil.ToBool("true")
t, err := strutil.ToTime("2006-01-02", "2006-01-02")

// Encoding/Decoding
base64Str := strutil.ToBase64("Hello")
original, err := strutil.FromBase64(base64Str)

// JSON conversion
jsonStr, err := strutil.ToJSON(map[string]string{"key": "value"})
var data map[string]string
err = strutil.FromJSON(jsonStr, &data)

// Collection conversion
slice := strutil.ToSlice("a,b,c", ",")
str := strutil.FromSlice([]string{"a", "b", "c"}, ",")

map1 := strutil.ToMap("key1=value1;key2=value2", ";", "=")
str = strutil.FromMap(map1, ";", "=")
```

## Best Practices

1. Always check for errors when using functions that return them
2. Use appropriate validation functions for specific use cases
3. Consider performance implications when using regular expressions
4. Use string builders for complex string manipulations
5. Handle empty strings and edge cases appropriately

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request. 