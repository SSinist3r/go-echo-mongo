package httpclient

import (
	"context"
	"encoding/json"
	"fmt"
)

// GetJSON sends a GET request and unmarshals the JSON response into the provided target
func (c *Client) GetJSON(ctx context.Context, url string, query map[string]string, target interface{}) error {
	resp, err := c.Get(ctx, url, query, nil)
	if err != nil {
		return err
	}

	if resp.StatusCode >= 400 {
		return fmt.Errorf("request failed with status code %d: %s", resp.StatusCode, string(resp.Body))
	}

	return json.Unmarshal(resp.Body, target)
}

// PostJSON sends a POST request with a JSON body and unmarshals the JSON response into the provided target
func (c *Client) PostJSON(ctx context.Context, url string, body interface{}, target interface{}) error {
	resp, err := c.Post(ctx, url, body, nil)
	if err != nil {
		return err
	}

	if resp.StatusCode >= 400 {
		return fmt.Errorf("request failed with status code %d: %s", resp.StatusCode, string(resp.Body))
	}

	return json.Unmarshal(resp.Body, target)
}

// PutJSON sends a PUT request with a JSON body and unmarshals the JSON response into the provided target
func (c *Client) PutJSON(ctx context.Context, url string, body interface{}, target interface{}) error {
	resp, err := c.Put(ctx, url, body, nil)
	if err != nil {
		return err
	}

	if resp.StatusCode >= 400 {
		return fmt.Errorf("request failed with status code %d: %s", resp.StatusCode, string(resp.Body))
	}

	return json.Unmarshal(resp.Body, target)
}

// DeleteJSON sends a DELETE request and unmarshals the JSON response into the provided target
func (c *Client) DeleteJSON(ctx context.Context, url string, target interface{}) error {
	resp, err := c.Delete(ctx, url, nil)
	if err != nil {
		return err
	}

	if resp.StatusCode >= 400 {
		return fmt.Errorf("request failed with status code %d: %s", resp.StatusCode, string(resp.Body))
	}

	return json.Unmarshal(resp.Body, target)
}

// PatchJSON sends a PATCH request with a JSON body and unmarshals the JSON response into the provided target
func (c *Client) PatchJSON(ctx context.Context, url string, body interface{}, target interface{}) error {
	resp, err := c.Patch(ctx, url, body, nil)
	if err != nil {
		return err
	}

	if resp.StatusCode >= 400 {
		return fmt.Errorf("request failed with status code %d: %s", resp.StatusCode, string(resp.Body))
	}

	return json.Unmarshal(resp.Body, target)
}
