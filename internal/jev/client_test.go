package jev

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"
)

func Test_Evaluate_posts_systemone_and_returns_answers(t *testing.T) {
	var gotAuth, gotPath string
	var payload requestBody
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotAuth = r.Header.Get("Authorization")
		gotPath = r.URL.Path
		body, err := io.ReadAll(r.Body)
		if err != nil {
			t.Errorf("read body: %v", err)
		}
		if err := json.Unmarshal(body, &payload); err != nil {
			t.Errorf("decode: %v", err)
		}
		w.Header().Set("Content-Type", "application/json")
		if _, err := w.Write([]byte(`{
			"model": "jev-1.13.0",
			"answers": {"is_urgent": {"type": "noul", "noul": 0.92}},
			"usage": {"input_tokens": 10, "output_tokens": 2}
		}`)); err != nil {
			t.Errorf("write: %v", err)
		}
	}))
	t.Cleanup(srv.Close)

	client := New(srv.URL, "ts_test", "jev-latest").WithHTTP(srv.Client())
	got, err := client.Evaluate(context.Background(), "help", map[string]Question{
		"is_urgent": Noul("Does this convey urgency?", "yes", "no"),
	})
	if err != nil {
		t.Fatal(err)
	}
	if gotAuth != "Bearer ts_test" {
		t.Fatalf("auth %q", gotAuth)
	}
	if gotPath != PathSystemOne {
		t.Fatalf("path %q", gotPath)
	}
	if payload.Model != "jev-latest" {
		t.Fatalf("model %q", payload.Model)
	}
	if got.Answers["is_urgent"].Noul != 0.92 {
		t.Fatalf("noul %v", got.Answers["is_urgent"].Noul)
	}
	if got.Usage.InputTokens != 10 {
		t.Fatalf("usage %+v", got.Usage)
	}
}

func Test_Evaluate_retries_429_then_succeeds(t *testing.T) {
	var calls atomic.Int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		n := calls.Add(1)
		if n == 1 {
			w.Header().Set("Retry-After", "0")
			w.WriteHeader(http.StatusTooManyRequests)
			if _, err := w.Write([]byte(`{"error":"rate"}`)); err != nil {
				t.Errorf("write: %v", err)
			}
			return
		}
		w.Header().Set("Content-Type", "application/json")
		if _, err := w.Write([]byte(`{"model":"jev-latest","answers":{"q":{"type":"noul","noul":1}},"usage":{}}`)); err != nil {
			t.Errorf("write: %v", err)
		}
	}))
	t.Cleanup(srv.Close)

	client := New(srv.URL, "ts_test", "jev-latest").WithHTTP(srv.Client())
	got, err := client.Evaluate(context.Background(), "s", map[string]Question{"q": Noul("yes?", "y", "n")})
	if err != nil {
		t.Fatal(err)
	}
	if calls.Load() != 2 {
		t.Fatalf("calls %d", calls.Load())
	}
	if got.Answers["q"].Noul != 1 {
		t.Fatalf("noul %v", got.Answers["q"].Noul)
	}
}

func Test_Evaluate_does_not_retry_client_errors(t *testing.T) {
	for _, status := range []int{http.StatusUnauthorized, http.StatusUnprocessableEntity} {
		t.Run(http.StatusText(status), func(t *testing.T) {
			var calls atomic.Int32
			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				calls.Add(1)
				w.WriteHeader(status)
				if _, err := w.Write([]byte(`{"error":"no"}`)); err != nil {
					t.Errorf("write: %v", err)
				}
			}))
			t.Cleanup(srv.Close)

			client := New(srv.URL, "k", "jev-latest").WithHTTP(srv.Client())
			_, err := client.Evaluate(context.Background(), "s", map[string]Question{"q": Noul("yes?", "y", "n")})
			var api *APIError
			if !errors.As(err, &api) || api.Status != status {
				t.Fatalf("err=%v", err)
			}
			if calls.Load() != 1 {
				t.Fatalf("calls %d", calls.Load())
			}
		})
	}
}

func Test_Evaluate_rejects_empty_questions(t *testing.T) {
	client := New("http://127.0.0.1", "k", "jev-latest")
	_, err := client.Evaluate(context.Background(), "s", nil)
	if err == nil {
		t.Fatal("expected error")
	}
}

func Test_sleep_cancels_with_context(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if err := sleep(ctx, time.Second); err == nil {
		t.Fatal("expected cancel")
	}
}
