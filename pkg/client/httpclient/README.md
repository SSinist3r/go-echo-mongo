# HTTP Client

A flexible HTTP client for Go applications with support for retries, timeouts, and convenience methods for common HTTP operations.

## Features

- Configurable client with options for timeout, retries, and headers
- Support for all common HTTP methods (GET, POST, PUT, DELETE, PATCH)
- JSON request and response handling
- Form data submission
- Multipart form data and file uploads
- Context support for cancellation and timeouts
- Automatic retries with configurable delay

## Usage

### Creating a Client

```go
import (
    "time"
    "github.com/yourusername/go-echo-mongo/pkg/client/httpclient"
)

// Create a client with default options
client := httpclient.NewClient()

// Create a client with custom options
client := httpclient.NewClient(
    httpclient.WithBaseURL("https://api.example.com"),
    httpclient.WithHeader("Authorization", "Bearer token123"),
    httpclient.WithTimeout(10 * time.Second),
    httpclient.WithRetryCount(3),
    httpclient.WithRetryWaitTime(1 * time.Second),
)
```

### Making Basic Requests

```go
import (
    "context"
    "github.com/yourusername/go-echo-mongo/pkg/client/httpclient"
)

ctx := context.Background()

// GET request
resp, err := client.Get(ctx, "users", nil, nil)
if err != nil {
    // Handle error
}

// POST request with JSON body
user := map[string]interface{}{
    "name": "John Doe",
    "email": "john@example.com",
}
resp, err := client.Post(ctx, "users", user, nil)

// PUT request
resp, err := client.Put(ctx, "users/123", user, nil)

// DELETE request
resp, err := client.Delete(ctx, "users/123", nil)

// PATCH request
updates := map[string]interface{}{
    "name": "John Smith",
}
resp, err := client.Patch(ctx, "users/123", updates, nil)
```

### Working with JSON

```go
// Define a struct for your data
type User struct {
    ID    string `json:"id"`
    Name  string `json:"name"`
    Email string `json:"email"`
}

// GET and unmarshal JSON response
var users []User
err := client.GetJSON(ctx, "users", nil, &users)

// POST JSON and unmarshal response
newUser := User{Name: "John Doe", Email: "john@example.com"}
var createdUser User
err := client.PostJSON(ctx, "users", newUser, &createdUser)

// PUT JSON and unmarshal response
updatedUser := User{Name: "John Smith", Email: "john@example.com"}
var result User
err := client.PutJSON(ctx, "users/123", updatedUser, &result)

// DELETE JSON and unmarshal response
var deleteResult struct {
    Success bool `json:"success"`
}
err := client.DeleteJSON(ctx, "users/123", &deleteResult)

// PATCH JSON and unmarshal response
patch := map[string]string{"name": "John Smith"}
var patchedUser User
err := client.PatchJSON(ctx, "users/123", patch, &patchedUser)
```

### Form Data and File Uploads

```go
// POST form data
formData := map[string]string{
    "name": "John Doe",
    "email": "john@example.com",
}
resp, err := client.PostForm(ctx, "users", formData, nil)

// POST multipart form with file upload
formData := map[string]string{
    "name": "John Doe",
    "email": "john@example.com",
}
files := []httpclient.FormFile{
    {
        FieldName: "avatar",
        FileName: "profile.jpg",
        FilePath: "/path/to/profile.jpg",
    },
}
resp, err := client.PostMultipartForm(ctx, "users", formData, files, nil)
```

### Advanced Usage with Request Object

```go
// Create a custom request
req := httpclient.Request{
    Method: "GET",
    URL: "users",
    Headers: map[string]string{
        "X-Custom-Header": "value",
    },
    Query: map[string]string{
        "page": "1",
        "limit": "10",
    },
}

// Send the request
resp, err := client.Do(ctx, req)
if err != nil {
    // Handle error
}

// Process the response
fmt.Printf("Status: %d\n", resp.StatusCode)
fmt.Printf("Body: %s\n", string(resp.Body))
``` 