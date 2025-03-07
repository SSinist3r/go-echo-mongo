package httpclient

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// Get sends a GET request to the specified URL
func (c *Client) Get(ctx context.Context, url string, query map[string]string, headers map[string]string) (*Response, error) {
	req := Request{
		Method:  http.MethodGet,
		URL:     url,
		Headers: headers,
		Query:   query,
	}
	return c.Do(ctx, req)
}

// Post sends a POST request to the specified URL with the given body
func (c *Client) Post(ctx context.Context, url string, body interface{}, headers map[string]string) (*Response, error) {
	req := Request{
		Method:  http.MethodPost,
		URL:     url,
		Headers: headers,
		Body:    body,
	}
	return c.Do(ctx, req)
}

// Put sends a PUT request to the specified URL with the given body
func (c *Client) Put(ctx context.Context, url string, body interface{}, headers map[string]string) (*Response, error) {
	req := Request{
		Method:  http.MethodPut,
		URL:     url,
		Headers: headers,
		Body:    body,
	}
	return c.Do(ctx, req)
}

// Delete sends a DELETE request to the specified URL
func (c *Client) Delete(ctx context.Context, url string, headers map[string]string) (*Response, error) {
	req := Request{
		Method:  http.MethodDelete,
		URL:     url,
		Headers: headers,
	}
	return c.Do(ctx, req)
}

// Patch sends a PATCH request to the specified URL with the given body
func (c *Client) Patch(ctx context.Context, url string, body interface{}, headers map[string]string) (*Response, error) {
	req := Request{
		Method:  http.MethodPatch,
		URL:     url,
		Headers: headers,
		Body:    body,
	}
	return c.Do(ctx, req)
}

// PostForm sends a POST request with form data
func (c *Client) PostForm(ctx context.Context, urlStr string, formData map[string]string, headers map[string]string) (*Response, error) {
	// Create form data
	data := url.Values{}
	for key, value := range formData {
		data.Add(key, value)
	}
	encodedData := data.Encode()

	// Create request
	fullURL := urlStr
	if c.baseURL != "" && !isAbsoluteURL(urlStr) {
		fullURL = fmt.Sprintf("%s/%s", c.baseURL, urlStr)
	}

	// Create the HTTP request
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, fullURL, strings.NewReader(encodedData))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Set content type for form data
	httpReq.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	httpReq.Header.Set("Accept", "application/json")

	// Add client-level headers
	for key, value := range c.headers {
		httpReq.Header.Set(key, value)
	}

	// Add request-specific headers
	for key, value := range headers {
		httpReq.Header.Set(key, value)
	}

	return c.executeRequest(ctx, httpReq)
}

// PostMultipartForm sends a POST request with multipart form data and file uploads
func (c *Client) PostMultipartForm(ctx context.Context, url string, formData map[string]string, files []FormFile, headers map[string]string) (*Response, error) {
	// Create a buffer to store the multipart form
	var buf bytes.Buffer
	writer := multipart.NewWriter(&buf)

	// Add form fields
	for key, value := range formData {
		if err := writer.WriteField(key, value); err != nil {
			return nil, fmt.Errorf("failed to write form field: %w", err)
		}
	}

	// Add files
	for _, file := range files {
		var fileReader io.Reader

		if len(file.FileData) > 0 {
			// Use provided file data
			fileReader = bytes.NewReader(file.FileData)
		} else if file.FilePath != "" {
			// Open file from path
			f, err := os.Open(file.FilePath)
			if err != nil {
				return nil, fmt.Errorf("failed to open file %s: %w", file.FilePath, err)
			}
			defer f.Close()

			fileReader = f

			// Use base name if filename not provided
			if file.FileName == "" {
				file.FileName = filepath.Base(file.FilePath)
			}
		} else {
			return nil, fmt.Errorf("no file data or file path provided for field %s", file.FieldName)
		}

		// Create form file
		part, err := writer.CreateFormFile(file.FieldName, file.FileName)
		if err != nil {
			return nil, fmt.Errorf("failed to create form file: %w", err)
		}

		// Copy file data to form
		if _, err := io.Copy(part, fileReader); err != nil {
			return nil, fmt.Errorf("failed to copy file data: %w", err)
		}
	}

	// Close the writer to finalize the form
	if err := writer.Close(); err != nil {
		return nil, fmt.Errorf("failed to close multipart writer: %w", err)
	}

	// Create request
	fullURL := url
	if c.baseURL != "" && !isAbsoluteURL(url) {
		fullURL = fmt.Sprintf("%s/%s", c.baseURL, url)
	}

	// Create the HTTP request
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, fullURL, &buf)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Set content type for multipart form
	httpReq.Header.Set("Content-Type", writer.FormDataContentType())
	httpReq.Header.Set("Accept", "application/json")

	// Add client-level headers
	for key, value := range c.headers {
		httpReq.Header.Set(key, value)
	}

	// Add request-specific headers
	for key, value := range headers {
		httpReq.Header.Set(key, value)
	}

	return c.executeRequest(ctx, httpReq)
}

// Do sends an HTTP request and returns the response
func (c *Client) Do(ctx context.Context, req Request) (*Response, error) {
	var bodyReader io.Reader
	if req.Body != nil {
		bodyBytes, err := json.Marshal(req.Body)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal request body: %w", err)
		}
		bodyReader = bytes.NewReader(bodyBytes)
	}

	// Construct the full URL
	fullURL := req.URL
	if c.baseURL != "" && !isAbsoluteURL(req.URL) {
		fullURL = fmt.Sprintf("%s/%s", c.baseURL, req.URL)
	}

	// Create the HTTP request
	httpReq, err := http.NewRequestWithContext(ctx, req.Method, fullURL, bodyReader)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Add query parameters
	if len(req.Query) > 0 {
		q := httpReq.URL.Query()
		for key, value := range req.Query {
			q.Add(key, value)
		}
		httpReq.URL.RawQuery = q.Encode()
	}

	// Set default headers
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Accept", "application/json")

	// Add client-level headers
	for key, value := range c.headers {
		httpReq.Header.Set(key, value)
	}

	// Add request-specific headers
	for key, value := range req.Headers {
		httpReq.Header.Set(key, value)
	}

	return c.executeRequest(ctx, httpReq)
}

// executeRequest executes an HTTP request with retries
func (c *Client) executeRequest(ctx context.Context, httpReq *http.Request) (*Response, error) {
	// Execute the request with retries
	var resp *http.Response
	var lastErr error

	for attempt := 0; attempt <= c.maxRetries; attempt++ {
		if attempt > 0 {
			select {
			case <-ctx.Done():
				return nil, ctx.Err()
			case <-time.After(c.retryDelay):
				// Wait before retrying
			}
		}

		resp, lastErr = c.client.Do(httpReq)
		if lastErr == nil {
			break
		}
	}

	if lastErr != nil {
		return nil, fmt.Errorf("request failed after %d attempts: %w", c.maxRetries+1, lastErr)
	}

	defer resp.Body.Close()

	// Read the response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	return &Response{
		StatusCode: resp.StatusCode,
		Headers:    resp.Header,
		Body:       body,
	}, nil
}
