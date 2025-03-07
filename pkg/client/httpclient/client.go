package httpclient

import (
	"net/http"
	"time"
)

// Client is a wrapper around http.Client with additional functionality
type Client struct {
	client     *http.Client
	baseURL    string
	headers    map[string]string
	timeout    time.Duration
	maxRetries int
	retryDelay time.Duration
}

// NewClient creates a new HTTP client with the given options
func NewClient(options ...ClientOption) *Client {
	c := &Client{
		client:     &http.Client{},
		headers:    make(map[string]string),
		timeout:    30 * time.Second,
		maxRetries: 3,
		retryDelay: 1 * time.Second,
	}

	for _, option := range options {
		option(c)
	}

	c.client.Timeout = c.timeout
	return c
}

// Response represents an HTTP response
type Response struct {
	StatusCode int
	Headers    http.Header
	Body       []byte
}

// FormFile represents a file to be uploaded in a multipart form
type FormFile struct {
	FieldName string
	FileName  string
	FilePath  string
	FileData  []byte // Optional: Use this or FilePath
}

// Request represents an HTTP request
type Request struct {
	Method  string
	URL     string
	Headers map[string]string
	Body    interface{}
	Query   map[string]string
}

// Helper function to check if a URL is absolute
func isAbsoluteURL(url string) bool {
	return len(url) > 7 && (url[:7] == "http://" || url[:8] == "https://")
}
