package posting

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
	"unicode/utf8"
)

const (
	defaultHTTPTimeout = 30 * time.Second
	maxErrorBodyRunes  = 500
)

var httpClient = &http.Client{Timeout: defaultHTTPTimeout}

func truncateToken(token string) string {
	if len(token) > 8 {
		return token[:4] + "****"
	}
	return "****"
}

func doJSONRequest(method, url string, headers map[string]string, body any) ([]byte, int, error) {
	var bodyReader io.Reader
	if body != nil {
		payload, err := json.Marshal(body)
		if err != nil {
			return nil, 0, fmt.Errorf("marshal request body: %w", err)
		}
		bodyReader = bytes.NewReader(payload)
	}

	req, err := http.NewRequest(method, url, bodyReader)
	if err != nil {
		return nil, 0, fmt.Errorf("create request: %w", err)
	}

	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	for key, value := range headers {
		req.Header.Set(key, value)
	}

	resp, err := httpClient.Do(req)
	if err != nil {
		return nil, 0, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	responseBody, err := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	if err != nil {
		return nil, resp.StatusCode, fmt.Errorf("read response body: %w", err)
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return responseBody, resp.StatusCode, fmt.Errorf("unexpected status %d: %s", resp.StatusCode, sanitizeErrorBody(responseBody))
	}

	return responseBody, resp.StatusCode, nil
}

func sanitizeErrorBody(body []byte) string {
	text := string(body)
	if utf8.RuneCountInString(text) > maxErrorBodyRunes {
		return truncateRunes(text, maxErrorBodyRunes)
	}
	return text
}
