# Go Echo Mongo

A comprehensive template/blueprint for building REST API servers with Go, Echo framework, and MongoDB. This project provides a solid foundation with working examples of common REST API patterns and features.

## Overview

This project serves as a starting point for developing your own REST APIs with Go. It includes:

- **Ready-to-use architecture**: A clean, modular structure following best practices
- **Working examples**: Practical implementations of handlers, services, repositories, and routes
- **Common API patterns**: Examples of CRUD operations, filtering, pagination, batch processing
- **Standard features**: Authentication, rate limiting, validation, error handling, logging

The project is not a specific application but rather a template that showcases different types of RESTful API implementations. All handlers, services, and server code are provided as examples that you can modify, extend, or replace for your specific use case.

## Features

- **Clean Architecture**: Follows a modular design with clear separation of concerns using repositories, services, and handlers
- **MongoDB Integration**: Complete MongoDB support with connection pooling and automatic retries
- **Redis Support**: Redis integration for caching, session management and rate limiting
- **Authentication**: API key authentication with role-based authorization
- **Rate Limiting**: Multiple rate limiting strategies:
  - Fixed Window Rate Limiting
  - Sliding Window Rate Limiting
  - Token Bucket Rate Limiting
  - Leaky Bucket Rate Limiting
- **Structured Logging**: Structured JSON logging using slog and zerolog
- **Metrics Monitoring**: Prometheus metrics for monitoring application performance
- **Graceful Shutdown**: Handles shutdown gracefully, ensuring all requests are processed
- **Docker Support**: Easy containerization with Docker and Docker Compose
- **Hot Reloading**: Development mode with automatic reloading using Air
- **API Documentation**: Built-in request and response validation
- **Generic Repository Pattern**: Type-safe generic repository for database operations
- **Environment Configuration**: Configuration via environment variables

## Project Structure

```
.
├── .github/               # GitHub CI/CD workflows
├── cmd/                   # Application entry points
│   └── api/               # API server entry point
├── internal/              # Private application code
│   ├── dto/               # Data Transfer Objects
│   ├── handler/           # HTTP handlers
│   ├── model/             # Domain models
│   ├── repository/        # Database repositories
│   ├── server/            # Server initialization and configuration
│   └── service/           # Business logic
├── pkg/                   # Public libraries that can be used by other applications
│   ├── client/            # Client utilities for external services
│   ├── database/          # Database connection utilities
│   ├── ratelimit/         # Rate limiting utilities
│   ├── secutil/           # Security utilities
│   ├── strutil/           # String utilities
│   └── web/               # Web utilities
│       ├── mwutil/        # Middleware utilities
│       └── response/      # Response formatting utilities
├── .air.toml              # Air configuration for hot reloading
├── .env                   # Environment variables (not in git)
├── .env.example           # Example environment variables
├── Dockerfile             # Docker configuration
├── docker-compose.yml     # Docker Compose configuration
├── go.mod                 # Go module definition
├── go.sum                 # Go module checksum
└── Makefile               # Build commands
```

## Prerequisites

- Go 1.24+
- MongoDB
- Redis (optional, but recommended for rate limiting)
- Docker and Docker Compose (optional)

## Getting Started

### Setting up the Environment

1. Clone the repository:

```bash
git clone https://github.com/yourusername/go-echo-mongo.git
cd go-echo-mongo
```

2. Copy the example environment file and modify as needed:

```bash
cp .env.example .env
```

3. Start the MongoDB database (using Docker):

```bash
make docker-run
```

### Running the Application

1. Install dependencies:

```bash
go mod download
```

2. Build and run the application:

```bash
make run
```

3. For development with hot reloading:

```bash
make watch
```

## Detailed Implementation Guide

For detailed instructions on extending this template and implementing specific features, please refer to the [HowTo.md](./HowTo.md) guide. This comprehensive document covers:

- **Step-by-step implementation examples**: Complete walkthrough of adding new routes, from model to API endpoint
- **Rate limiting implementation**: How to use different rate limiting strategies for various scenarios
- **Authentication configuration**: Setting up API key authentication with different role-based permissions
- **Example customizations**: Practical examples for common implementation patterns

The HowTo.md guide is especially useful for developers who want to understand how to implement their own domain logic while leveraging the architecture patterns provided in this template.

### API Endpoints

The template includes the following example API endpoints that demonstrate common REST patterns:

- **User Management Examples**:
  - `POST /api/v1/users` - Example of creating a resource
  - `GET /api/v1/users` - Example of retrieving a collection
  - `GET /api/v1/users/paginated` - Example of pagination implementation
  - `GET /api/v1/users/:id` - Example of retrieving a resource by ID
  - `PUT /api/v1/users/:id` - Example of updating a resource
  - `DELETE /api/v1/users/:id` - Example of deleting a resource
  - `POST /api/v1/users/login` - Example of authentication endpoint

- **Product Management Examples**:
  - `POST /api/v1/products` - Example of resource creation with validation
  - `GET /api/v1/products` - Example of collection retrieval
  - `GET /api/v1/products/paginated` - Example of advanced pagination
  - `GET /api/v1/products/:id` - Example of single resource retrieval
  - `PUT /api/v1/products/:id` - Example of resource updating with validation
  - `DELETE /api/v1/products/:id` - Example of resource deletion
  - `GET /api/v1/products/category/:category` - Example of filtering by parameter

- **Batch Operations Examples**:
  - `POST /api/v1/users/batch` - Example of batch creation
  - `POST /api/v1/users/filter` - Example of filtering with request body
  - `PUT /api/v1/users/batch` - Example of batch updating
  - `DELETE /api/v1/users/batch` - Example of batch deletion
  - `POST /api/v1/products/batch` - Example of batch operations with validation
  - `POST /api/v1/products/filter` - Example of advanced filtering
  - `PUT /api/v1/products/batch` - Example of bulk updates
  - `DELETE /api/v1/products/batch` - Example of bulk deletion

- **Metrics and Health Examples**:
  - `GET /metrics` - Example of Prometheus metrics endpoint
  - `GET /redis/health` - Example of service health check

## Authentication

The API uses API key authentication. To access protected endpoints, include the API key in the request header:

```
X-API-Key: your-api-key
```

API keys are automatically generated for each user and can be used to authenticate API requests.

## Rate Limiting

Multiple rate limiting strategies are available to protect the API from abuse:

- Fixed Window: Limits requests within a fixed time window
- Sliding Window: Provides smoother rate limiting across time boundaries
- Token Bucket: Allows bursts of traffic while maintaining a steady average
- Leaky Bucket: Controls the flow of requests at a constant rate

Rate limits are applied per API key or IP address and can be configured per route or globally.

## Docker Deployment

To deploy the application using Docker:

```bash
docker-compose up -d
```

This will start the API server and MongoDB in detached mode.

## Testing

Run tests with:

```bash
make test
```

For integration tests:

```bash
make itest
```

## Using This Template for Your Project

To use this template as a starting point for your own REST API project:

1. **Clone or download the project**
   ```bash
   git clone https://github.com/yourusername/go-echo-mongo.git my-api-project
   cd my-api-project
   ```

2. **Remove the existing Git history and start fresh**
   ```bash
   rm -rf .git
   git init
   ```

3. **Modify the module name in go.mod**
   ```bash
   # Edit go.mod to change the module name to your own
   go mod edit -module github.com/yourusername/my-api-project
   ```

4. **Customize environment settings**
   ```bash
   cp .env.example .env
   # Edit .env file with your own settings
   ```

5. **Replace or modify the example models, repositories, services, and handlers**
   - Use the examples as a reference to create your own domain-specific implementation
   - You can keep the structure and patterns but replace the business logic
   - Follow the step-by-step examples in [HowTo.md](./HowTo.md) for implementation guidance

6. **Update documentation**
   - Modify this README.md to describe your specific project
   - Update API documentation to reflect your actual endpoints

7. **Run tests to ensure everything works**
   ```bash
   make test
   ```

8. **Start developing your own API!**

This template provides working examples, but for a production application, you should:
- Replace example entities with your own domain models
- Implement proper authentication for your use case
- Configure appropriate rate limits
- Add comprehensive tests for your specific endpoints
- Customize error handling for your specific requirements

The architecture is designed to be modular, so you can easily replace components while maintaining the overall structure. 

## Contributing

Contributions to improve this template are welcome! Whether you want to fix a bug, add a feature, or improve documentation, your help is appreciated.

### How to Contribute

1. **Fork the repository**
   - Create your own fork of the project
   - Set up the development environment locally

2. **Create a feature branch**
   ```bash
   git checkout -b feature/your-feature-name
   ```

3. **Make your changes**
   - Follow the Go coding standards and project conventions
   - Keep changes focused on a single goal
   - Add tests for new functionality
   - Ensure existing tests pass with `make test`

4. **Document your changes**
   - Update documentation to reflect your changes
   - Add comments to your code where necessary
   - If adding a new feature, consider updating the HowTo.md

5. **Commit your changes**
   ```bash
   git commit -m "Add a descriptive message about your changes"
   ```

6. **Push to your fork**
   ```bash
   git push origin feature/your-feature-name
   ```

7. **Create a Pull Request**
   - Submit a PR from your branch to the main repository
   - Describe your changes in detail
   - Link any related issues

### Code Style Guidelines

- Follow [Go best practices](https://golang.org/doc/effective_go)
- Run `go fmt` before committing code
- Use meaningful variable and function names
- Keep functions small and focused
- Write descriptive comments
- Follow the project's existing architecture patterns

### Reporting Issues

- Use the GitHub issue tracker to report bugs
- Provide detailed reproduction steps
- Include information about your environment (Go version, MongoDB version, etc.)
- If possible, provide a minimal failing example

### Feature Requests

- Use the GitHub issue tracker for feature requests
- Clearly describe the feature and its use case
- Be open to discussion about the feature's implementation

Thank you for contributing to make this template better for everyone!

## Note on Documentation

All documentation in this template, including this README.md and the HowTo.md guide, was generated using AI-powered documentation tools. This approach ensures comprehensive, consistent, and well-structured documentation while saving development time.

The documentation was designed to be both informative for understanding the template and practical for implementing your own solutions based on it.
