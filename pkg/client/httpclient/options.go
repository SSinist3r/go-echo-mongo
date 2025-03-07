package httpclient

import (
	"time"
)

// ClientOption is a function that configures a Client
type ClientOption func(*Client)

// WithBaseURL sets the base URL for the client
func WithBaseURL(baseURL string) ClientOption {
	return func(c *Client) {
		c.baseURL = baseURL
	}
}

// WithHeader sets a header for all requests
func WithHeader(key, value string) ClientOption {
	return func(c *Client) {
		c.headers[key] = value
	}
}

// WithTimeout sets the timeout for all requests
func WithTimeout(timeout time.Duration) ClientOption {
	return func(c *Client) {
		c.timeout = timeout
	}
}

// WithRetryCount sets the maximum number of retry attempts
func WithRetryCount(maxRetries int) ClientOption {
	return func(c *Client) {
		c.maxRetries = maxRetries
	}
}

// WithRetryWaitTime sets the delay between retry attempts
func WithRetryWaitTime(retryDelay time.Duration) ClientOption {
	return func(c *Client) {
		c.retryDelay = retryDelay
	}
}

// WithRetryMaxWaitTime sets both retry count and delay
func WithRetryMaxWaitTime(maxWaitTime time.Duration) ClientOption {
	return func(c *Client) {
		// This is a placeholder - in a real implementation, you might
		// calculate an appropriate retry delay based on the max wait time
		// For now, we'll just set a reasonable default
		c.retryDelay = maxWaitTime / 5
	}
}
