# Package Directory

This directory contains reusable packages that provide various utilities and functionality for the application.

## Packages

### client

The `client` package provides HTTP client utilities for making requests to external services.

- **httpclient**: A flexible HTTP client with support for retries, timeouts, and convenience methods for common HTTP operations.

### database

The `database` package provides database connection utilities for MongoDB and Redis.

- MongoDB connection with configurable settings, connection pooling, and health checks
- Redis connection with configurable settings, connection pooling, and health statistics monitoring

### ratelimit

The `ratelimit` package provides rate limiting utilities for controlling the rate of requests to your API.

- Repository interface for rate limiting operations
- Multiple rate limiting strategies (Fixed Window, Sliding Window, Token Bucket, Leaky Bucket)
- Configurable rate limits and time windows

### secutil

The `secutil` package provides security-related utilities for cryptographic operations.

- Password hashing and verification
- General-purpose hashing (MD5, SHA256, SHA512)
- HMAC creation and verification
- AES encryption and decryption
- Secure random key generation

### strutil

The `strutil` package provides string manipulation, validation, and conversion utilities.

- Random string generation
- String formatting and case conversion
- String validation
- Type conversion
- Common string operations

### web

The `web` package provides utilities for building web applications with the Echo framework.

- **response**: Standardized API response utilities
- **mwutil**: Custom middleware functions
- **validator**: Request validation utilities

## Usage

Each package has its own README.md file with detailed usage instructions and examples.

## Best Practices

1. **Modularity**: Each package should have a single responsibility and be independent of other packages when possible.
2. **Documentation**: All exported functions, types, and constants should have proper GoDoc comments.
3. **Testing**: Each package should have comprehensive unit tests.
4. **Error Handling**: Use descriptive error messages and proper error wrapping.
5. **Configuration**: Use configuration structs with sensible defaults.
6. **Interfaces**: Define interfaces for external dependencies to facilitate testing and flexibility. 