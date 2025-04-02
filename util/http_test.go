package util

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestPostWithJson(t *testing.T) {
	// test server
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("expected method POST, got %s", r.Method)
		}

		if r.Header.Get("Content-Type") != "application/json" {
			t.Errorf("expected Content-Type application/json, got %s", r.Header.Get("Content-Type"))
		}

		body, err := io.ReadAll(r.Body)
		if err != nil {
			t.Fatalf("could not read request body: %v", err)
		}
		defer r.Body.Close()
		if string(body) != `{"key":"value"}` {
			t.Errorf("expected body {\"key\":\"value\"}, got %s", body)
		}
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"response":"ok"}`))
	}))
	defer ts.Close()

	// Prepare test data
	ctx := context.Background()
	data := []byte(`{"key":"value"}`)
	endpoint := ts.URL

	// Call the function
	resp, err := PostWithJson(ctx, data, endpoint)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if !strings.Contains(string(resp), `{"response":"ok"}`) {
		t.Errorf("expected response {\"response\":\"ok\"}, got %s", resp)
	}
}
