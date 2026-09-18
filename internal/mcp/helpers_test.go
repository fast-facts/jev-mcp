package mcp

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/fast-facts/jev-mcp/internal/jev"
)

func newTestHandlers(t *testing.T, answers map[string]jev.Answer) *handlers {
	t.Helper()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != jev.PathSystemOne {
			t.Errorf("path %s", r.URL.Path)
		}
		body, err := io.ReadAll(r.Body)
		if err != nil {
			t.Errorf("read body: %v", err)
			return
		}
		if !strings.Contains(string(body), `"model":"jev-latest"`) {
			t.Errorf("missing model in %s", body)
		}
		resp := map[string]any{
			"model":   "jev-latest",
			"answers": answers,
			"usage":   jev.Usage{InputTokens: 5, OutputTokens: 1},
		}
		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(resp); err != nil {
			t.Errorf("encode: %v", err)
		}
	}))
	t.Cleanup(srv.Close)
	return &handlers{client: jev.New(srv.URL, "ts_test", "jev-latest").WithHTTP(srv.Client())}
}
