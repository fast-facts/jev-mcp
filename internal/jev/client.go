package jev

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"
)

const maxAttempts = 3

var retryStatuses = map[int]struct{}{
	http.StatusRequestTimeout:      {},
	http.StatusTooManyRequests:     {},
	http.StatusInternalServerError: {},
	http.StatusBadGateway:          {},
	http.StatusServiceUnavailable:  {},
	529:                            {},
}

type APIError struct {
	Status  int
	Message string
}

func (e *APIError) Error() string {
	return fmt.Sprintf("typesafe api status %d: %s", e.Status, e.Message)
}

type Client struct {
	http    *http.Client
	baseURL string
	apiKey  string
	model   string
}

func New(baseURL, apiKey, model string) *Client {
	return &Client{
		http: &http.Client{
			Timeout: 30 * time.Second,
			Transport: &http.Transport{
				MaxIdleConns:        20,
				MaxIdleConnsPerHost: 10,
				IdleConnTimeout:     90 * time.Second,
				ForceAttemptHTTP2:   true,
			},
		},
		baseURL: strings.TrimRight(baseURL, "/"),
		apiKey:  apiKey,
		model:   model,
	}
}

func (c *Client) WithHTTP(h *http.Client) *Client {
	out := *c
	out.http = h
	return &out
}

func (c *Client) Evaluate(ctx context.Context, state any, questions map[string]Question) (Result, error) {
	if len(questions) == 0 {
		return Result{}, errors.New("jev: questions must not be empty")
	}
	body, err := json.Marshal(requestBody{Model: c.model, State: state, Questions: questions})
	if err != nil {
		return Result{}, fmt.Errorf("jev encode: %w", err)
	}
	var last error
	var retryAfter time.Duration
	for attempt := range maxAttempts {
		if attempt > 0 {
			wait := retryAfter
			if wait <= 0 {
				wait = time.Duration(attempt) * time.Second
			}
			if err := sleep(ctx, wait); err != nil {
				return Result{}, err
			}
		}
		result, retry, ra, err := c.once(ctx, body)
		if err == nil {
			return result, nil
		}
		last = err
		if !retry {
			return Result{}, err
		}
		retryAfter = ra
	}
	return Result{}, last
}

func (c *Client) once(ctx context.Context, body []byte) (Result, bool, time.Duration, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+PathSystemOne, bytes.NewReader(body))
	if err != nil {
		return Result{}, false, 0, fmt.Errorf("jev request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+c.apiKey)
	req.Header.Set("Content-Type", "application/json")
	resp, err := c.http.Do(req)
	if err != nil {
		return Result{}, ctx.Err() == nil, 0, fmt.Errorf("jev http: %w", err)
	}
	defer resp.Body.Close()
	raw, err := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	if err != nil {
		return Result{}, false, 0, fmt.Errorf("jev read: %w", err)
	}
	if resp.StatusCode >= 300 {
		_, retry := retryStatuses[resp.StatusCode]
		return Result{}, retry, parseRetryAfter(resp.Header.Get("Retry-After")), &APIError{
			Status:  resp.StatusCode,
			Message: trimBody(raw),
		}
	}
	var parsed struct {
		Model   string            `json:"model"`
		Answers map[string]Answer `json:"answers"`
		Usage   Usage             `json:"usage"`
	}
	if err := json.Unmarshal(raw, &parsed); err != nil {
		return Result{}, false, 0, fmt.Errorf("jev decode: %w", err)
	}
	return Result{Model: parsed.Model, Answers: parsed.Answers, Usage: parsed.Usage}, false, 0, nil
}

func sleep(ctx context.Context, d time.Duration) error {
	timer := time.NewTimer(d)
	defer timer.Stop()
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-timer.C:
		return nil
	}
}

func parseRetryAfter(v string) time.Duration {
	if v == "" {
		return 0
	}
	n, err := strconv.Atoi(v)
	if err != nil || n <= 0 {
		return 0
	}
	return time.Duration(n) * time.Second
}

func trimBody(raw []byte) string {
	s := strings.TrimSpace(string(raw))
	if len(s) > 200 {
		return s[:200]
	}
	if s == "" {
		return "empty body"
	}
	return s
}
