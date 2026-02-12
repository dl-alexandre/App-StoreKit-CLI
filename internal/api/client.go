package api

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/dl-alexandre/App-StoreKit-CLI/internal/auth"
	"github.com/dl-alexandre/App-StoreKit-CLI/internal/retry"
)

type Client struct {
	BaseURL   string
	HTTP      *http.Client
	Signer    auth.Signer
	Retry     retry.Config
	RequestID string
	UserAgent string
	Debug     bool
	DebugOut  io.Writer
}

type Response struct {
	Status  int
	Headers http.Header
	Body    []byte
	JSON    any
}

type APIError struct {
	Status  int
	Code    string
	Message string
	Raw     string
}

func (e APIError) Error() string {
	if e.Code != "" || e.Message != "" {
		return fmt.Sprintf("api error (%d): %s %s", e.Status, e.Code, e.Message)
	}
	return fmt.Sprintf("api error (%d)", e.Status)
}

func (c Client) Do(ctx context.Context, method, path string, query map[string][]string, body []byte, contentType string) (Response, error) {
	if c.HTTP == nil {
		c.HTTP = &http.Client{Timeout: 30 * time.Second}
	}
	if c.BaseURL == "" {
		return Response{}, errors.New("base url is required")
	}

	bodyBytes := body
	if bodyBytes == nil {
		bodyBytes = []byte{}
	}

	res := retry.Do(ctx, c.Retry, func() retry.Result {
		req, err := c.newRequest(ctx, method, path, query, bodyBytes, contentType)
		if err != nil {
			return retry.Result{Err: err}
		}
		resp, err := c.HTTP.Do(req)
		if err != nil {
			c.debugf("%s %s error: %v", method, req.URL.Path, err)
			return retry.Result{Err: err, Retry: isTemporary(err)}
		}
		defer resp.Body.Close()
		raw, readErr := io.ReadAll(resp.Body)
		if readErr != nil {
			c.debugf("%s %s read error: %v", method, req.URL.Path, readErr)
			return retry.Result{Err: readErr}
		}

		parsed, _ := parseJSON(raw)
		response := Response{Status: resp.StatusCode, Headers: resp.Header, Body: raw, JSON: parsed}
		c.debugResponse(method, req.URL.Path, resp, response)
		if resp.StatusCode >= 500 || resp.StatusCode == 429 {
			return retry.Result{
				Err:   responseError{status: resp.StatusCode, headers: resp.Header, body: raw, json: parsed},
				Retry: true,
				Value: response,
				Wait:  retryAfter(resp),
			}
		}
		if resp.StatusCode < 200 || resp.StatusCode >= 300 {
			return retry.Result{Err: responseError{status: resp.StatusCode, headers: resp.Header, body: raw, json: parsed}, Retry: false, Value: response}
		}
		return retry.Result{Err: nil, Retry: false, Value: response}
	})

	if res.Err == nil {
		if response, ok := res.Value.(Response); ok {
			return response, nil
		}
		return Response{}, nil
	}

	if apiErr, ok := res.Err.(responseError); ok {
		response := Response{
			Status:  apiErr.status,
			Headers: apiErr.headers,
			Body:    apiErr.body,
			JSON:    apiErr.json,
		}
		return response, apiErr.toAPIError()
	}

	if response, ok := res.Value.(Response); ok {
		return response, res.Err
	}

	return Response{}, res.Err
}

func (c Client) newRequest(ctx context.Context, method, path string, query map[string][]string, body []byte, contentType string) (*http.Request, error) {
	token, err := c.Signer.Token()
	if err != nil {
		return nil, err
	}

	base := strings.TrimRight(c.BaseURL, "/")
	urlPath := strings.TrimLeft(path, "/")
	full := fmt.Sprintf("%s/%s", base, urlPath)

	parsed, err := url.Parse(full)
	if err != nil {
		return nil, err
	}

	q := parsed.Query()
	for key, values := range query {
		for _, value := range values {
			if value != "" {
				q.Add(key, value)
			}
		}
	}
	parsed.RawQuery = q.Encode()

	req, err := http.NewRequestWithContext(ctx, method, parsed.String(), bytes.NewReader(body))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Accept", "application/json")
	if c.UserAgent != "" {
		req.Header.Set("User-Agent", c.UserAgent)
	}
	if len(body) > 0 {
		if contentType == "" {
			contentType = "application/json"
		}
		req.Header.Set("Content-Type", contentType)
	}
	if c.RequestID != "" {
		req.Header.Set("X-Request-ID", c.RequestID)
	}

	return req, nil
}

func parseJSON(body []byte) (any, error) {
	if len(body) == 0 {
		return nil, nil
	}
	var parsed any
	if err := json.Unmarshal(body, &parsed); err != nil {
		return nil, err
	}
	return parsed, nil
}

type responseError struct {
	status  int
	headers http.Header
	body    []byte
	json    any
}

func (e responseError) Error() string {
	return fmt.Sprintf("response status %d", e.status)
}

func (e responseError) toAPIError() APIError {
	apiErr := APIError{Status: e.status, Raw: string(e.body)}
	if m, ok := e.json.(map[string]any); ok {
		if code, ok := m["errorCode"].(string); ok {
			apiErr.Code = code
		}
		if msg, ok := m["errorMessage"].(string); ok {
			apiErr.Message = msg
		}
	}
	return apiErr
}

func isTemporary(err error) bool {
	var netErr net.Error
	if errors.As(err, &netErr) {
		return netErr.Timeout()
	}
	return false
}

func (c Client) debugf(format string, args ...any) {
	if !c.Debug {
		return
	}
	out := c.DebugOut
	if out == nil {
		out = io.Discard
	}
	_, _ = fmt.Fprintf(out, format+"\n", args...)
}

func (c Client) debugResponse(method, path string, resp *http.Response, response Response) {
	if !c.Debug || resp == nil {
		return
	}
	message := fmt.Sprintf("%s %s -> %d", method, path, response.Status)
	if retryAfterValue := strings.TrimSpace(resp.Header.Get("Retry-After")); retryAfterValue != "" {
		message += " retry-after=" + retryAfterValue
	}
	if requestID := strings.TrimSpace(resp.Header.Get("x-request-id")); requestID != "" {
		message += " request-id=" + requestID
	}
	c.debugf(message)
}

func retryAfter(resp *http.Response) time.Duration {
	if resp == nil {
		return 0
	}
	value := strings.TrimSpace(resp.Header.Get("Retry-After"))
	if value == "" {
		return 0
	}
	if seconds, err := strconv.Atoi(value); err == nil {
		if seconds <= 0 {
			return 0
		}
		return time.Duration(seconds) * time.Second
	}
	if t, err := http.ParseTime(value); err == nil {
		wait := time.Until(t)
		if wait < 0 {
			return 0
		}
		return wait
	}
	return 0
}
