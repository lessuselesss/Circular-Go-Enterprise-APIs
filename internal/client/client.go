package client

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

// Client represents an HTTP client for interacting with the Circular Protocol API.
// It provides robust network communication with built-in retry logic and error handling.
type Client struct {
	baseURL       string
	httpClient    *http.Client
	timeout       time.Duration
	retryAttempts int
	retryDelay    time.Duration
}

// NewClient creates a new HTTP client instance with default configuration.
// The baseURL parameter specifies the base URL for all API requests.
func NewClient(baseURL string) *Client {
	return &Client{
		baseURL: baseURL,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
		timeout:       30 * time.Second,
		retryAttempts: 3,
		retryDelay:    1 * time.Second,
	}
}

// SetTimeout configures the timeout duration for HTTP requests.
func (c *Client) SetTimeout(timeout time.Duration) {
	c.timeout = timeout
	c.httpClient.Timeout = timeout
}

// SetRetryAttempts configures the number of retry attempts for failed requests.
func (c *Client) SetRetryAttempts(attempts int) {
	c.retryAttempts = attempts
}

// SetRetryDelay configures the delay between retry attempts.
func (c *Client) SetRetryDelay(delay time.Duration) {
	c.retryDelay = delay
}

// POST sends a POST request to the specified endpoint with JSON payload.
// It includes built-in retry logic for transient failures.
func (c *Client) POST(ctx context.Context, endpoint string, payload interface{}) ([]byte, error) {
	url := c.buildURL(endpoint)
	
	// Marshal payload to JSON
	jsonData, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal payload: %w", err)
	}
	
	var lastErr error
	for attempt := 0; attempt <= c.retryAttempts; attempt++ {
		if attempt > 0 {
			// Wait before retry
			select {
			case <-ctx.Done():
				return nil, ctx.Err()
			case <-time.After(c.retryDelay):
			}
		}
		
		req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(jsonData))
		if err != nil {
			lastErr = fmt.Errorf("failed to create request: %w", err)
			continue
		}
		
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Accept", "application/json")
		
		resp, err := c.httpClient.Do(req)
		if err != nil {
			lastErr = fmt.Errorf("request failed: %w", err)
			continue
		}
		
		defer resp.Body.Close()
		
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			lastErr = fmt.Errorf("failed to read response body: %w", err)
			continue
		}
		
		if resp.StatusCode >= 200 && resp.StatusCode < 300 {
			return body, nil
		}
		
		// Handle non-2xx status codes
		if resp.StatusCode >= 500 {
			// Server error - retry
			lastErr = fmt.Errorf("server error (status %d): %s", resp.StatusCode, string(body))
			continue
		} else {
			// Client error - don't retry
			return nil, fmt.Errorf("client error (status %d): %s", resp.StatusCode, string(body))
		}
	}
	
	return nil, fmt.Errorf("request failed after %d attempts: %w", c.retryAttempts+1, lastErr)
}

// GET sends a GET request to the specified endpoint.
// It includes built-in retry logic for transient failures.
func (c *Client) GET(ctx context.Context, endpoint string) ([]byte, error) {
	url := c.buildURL(endpoint)
	
	var lastErr error
	for attempt := 0; attempt <= c.retryAttempts; attempt++ {
		if attempt > 0 {
			// Wait before retry
			select {
			case <-ctx.Done():
				return nil, ctx.Err()
			case <-time.After(c.retryDelay):
			}
		}
		
		req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
		if err != nil {
			lastErr = fmt.Errorf("failed to create request: %w", err)
			continue
		}
		
		req.Header.Set("Accept", "application/json")
		
		resp, err := c.httpClient.Do(req)
		if err != nil {
			lastErr = fmt.Errorf("request failed: %w", err)
			continue
		}
		
		defer resp.Body.Close()
		
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			lastErr = fmt.Errorf("failed to read response body: %w", err)
			continue
		}
		
		if resp.StatusCode >= 200 && resp.StatusCode < 300 {
			return body, nil
		}
		
		// Handle non-2xx status codes
		if resp.StatusCode >= 500 {
			// Server error - retry
			lastErr = fmt.Errorf("server error (status %d): %s", resp.StatusCode, string(body))
			continue
		} else {
			// Client error - don't retry
			return nil, fmt.Errorf("client error (status %d): %s", resp.StatusCode, string(body))
		}
	}
	
	return nil, fmt.Errorf("request failed after %d attempts: %w", c.retryAttempts+1, lastErr)
}

// buildURL constructs the full URL by combining base URL and endpoint.
func (c *Client) buildURL(endpoint string) string {
	if strings.HasSuffix(c.baseURL, "/") && strings.HasPrefix(endpoint, "/") {
		return c.baseURL + endpoint[1:]
	}
	if !strings.HasSuffix(c.baseURL, "/") && !strings.HasPrefix(endpoint, "/") {
		return c.baseURL + "/" + endpoint
	}
	return c.baseURL + endpoint
}