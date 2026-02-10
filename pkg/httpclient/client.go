package httpclient

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

type Client struct {
	httpClient *http.Client
	maxRetries int
	baseDelay  time.Duration
}

type Config struct {
	Timeout    time.Duration
	MaxRetries int
	BaseDelay  time.Duration
}

func New(cfg Config) (client *Client) {
	client = &Client{
		httpClient: &http.Client{
			Timeout: cfg.Timeout,
		},
		maxRetries: cfg.MaxRetries,
		baseDelay:  cfg.BaseDelay,
	}
	return
}

func (c *Client) PostJSON(ctx context.Context, url string, payload interface{}, headers map[string]string) (statusCode int, body []byte, err error) {
	var jsonData []byte
	jsonData, err = json.Marshal(payload)
	if err != nil {
		err = fmt.Errorf("failed to marshal payload: %w", err)
		return
	}

	statusCode, body, err = c.doWithRetry(ctx, url, jsonData, headers)
	return
}

func (c *Client) doWithRetry(ctx context.Context, url string, body []byte, headers map[string]string) (statusCode int, responseBody []byte, err error) {
	var lastErr error

	for attempt := 0; attempt <= c.maxRetries; attempt++ {
		if attempt > 0 {
			delay := c.baseDelay * time.Duration(1<<uint(attempt-1))
			select {
			case <-time.After(delay):
			case <-ctx.Done():
				err = ctx.Err()
				return
			}
		}

		statusCode, responseBody, lastErr = c.doRequest(ctx, url, body, headers)
		if lastErr == nil && statusCode >= 200 && statusCode < 300 {
			return
		}

		if lastErr != nil && (ctx.Err() != nil) {
			err = lastErr
			return
		}
	}

	if lastErr != nil {
		err = lastErr
	} else {
		err = fmt.Errorf("request failed with status code %d after %d retries", statusCode, c.maxRetries)
	}
	return
}

func (c *Client) doRequest(ctx context.Context, url string, body []byte, headers map[string]string) (statusCode int, responseBody []byte, err error) {
	var req *http.Request
	req, err = http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		err = fmt.Errorf("failed to create request: %w", err)
		return
	}

	req.Header.Set("Content-Type", "application/json")
	for key, value := range headers {
		req.Header.Set(key, value)
	}

	var resp *http.Response
	resp, err = c.httpClient.Do(req)
	if err != nil {
		err = fmt.Errorf("request failed: %w", err)
		return
	}
	defer resp.Body.Close()

	statusCode = resp.StatusCode

	responseBody, err = io.ReadAll(resp.Body)
	if err != nil {
		err = fmt.Errorf("failed to read response body: %w", err)
		return
	}

	return
}
