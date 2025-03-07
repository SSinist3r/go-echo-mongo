# Security Utilities (secutil)

A collection of security-related utilities for Go applications, providing simple and secure implementations of common cryptographic operations.

## Features

- Password hashing and verification
- General-purpose hashing (MD5, SHA256, SHA512)
- HMAC creation and verification
- AES encryption and decryption
- Secure random key generation

## Usage

### Password Operations

```go
import "yourproject/pkg/secutil"

// Hash a password
hashedPassword, err := secutil.HashPassword("mypassword123")

// Verify a password
err = secutil.VerifyPassword(hashedPassword, "mypassword123")
```

### Hashing

```go
// Create different types of hashes
sha256Hash, err := secutil.HashString("my data", "sha256")
sha512Hash, err := secutil.HashString("my data", "sha512")
md5Hash, err := secutil.HashString("my data", "md5")

// Create and verify HMAC
hmac, err := secutil.CreateHMAC("message", "secret-key", "sha256")
isValid, err := secutil.VerifyHMAC("message", "secret-key", hmac, "sha256")

// Compare hashes safely (constant-time comparison)
isMatch := secutil.CompareHashes(hash1, hash2)
```

### Encryption

```go
// Generate a random encryption key (16, 24, or 32 bytes for AES-128, AES-192, or AES-256)
key, err := secutil.GenerateKey(32)

// Encrypt data
encrypted, err := secutil.Encrypt([]byte("sensitive data"), key)

// Decrypt data
decrypted, err := secutil.Decrypt(encrypted, key)

// String convenience methods
encryptedStr, err := secutil.EncryptString("sensitive data", key)
decryptedStr, err := secutil.DecryptString(encryptedStr, key)

// Generate random bytes
randomBytes, err := secutil.GenerateRandomBytes(32)
```

## Best Practices

1. **Password Hashing**
   - Always use `HashPassword` for password storage
   - Never store plain-text passwords
   - Use `VerifyPassword` for password verification

2. **Encryption**
   - Use appropriate key sizes (16, 24, or 32 bytes)
   - Keep encryption keys secure
   - Use `GenerateKey` for key generation
   - Don't reuse nonces/IVs

3. **Hashing**
   - Use SHA-256 or SHA-512 for secure hashing
   - Only use MD5 for non-security-critical operations
   - Use HMAC for message authentication
   - Use `CompareHashes` for secure hash comparison

4. **Error Handling**
   - Always check for errors
   - Handle encryption/decryption errors gracefully
   - Don't expose error details to end users

## Security Considerations

1. **Key Management**
   - Store encryption keys securely
   - Rotate keys periodically
   - Use environment variables or secure key management systems

2. **Algorithm Choice**
   - AES-GCM for encryption (authenticated encryption)
   - bcrypt for password hashing
   - SHA-256/SHA-512 for general hashing
   - HMAC with SHA-256 for message authentication

3. **Implementation Details**
   - Uses secure random number generation
   - Implements constant-time comparisons
   - Uses standard Go crypto packages

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request. 