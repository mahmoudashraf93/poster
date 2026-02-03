package graph

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestExchangeToken(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v19.0/oauth/access_token" {
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}

		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"access_token":"longtoken"}`))
	}))
	defer server.Close()

	oldBase := graphBaseURL
	graphBaseURL = server.URL + "/"

	defer func() { graphBaseURL = oldBase }()

	token, err := ExchangeToken(context.Background(), "app", "secret", "short")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if token != "longtoken" {
		t.Fatalf("unexpected token: %s", token)
	}
}

func TestDebugToken(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v19.0/debug_token" {
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}

		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"data":{"app_id":"1","is_valid":true,"expires_at":123}}`))
	}))
	defer server.Close()

	oldBase := graphBaseURL
	graphBaseURL = server.URL + "/"

	defer func() { graphBaseURL = oldBase }()

	info, err := DebugToken(context.Background(), "token")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if info.AppID != "1" || !info.IsValid {
		t.Fatalf("unexpected info: %+v", info)
	}
}

func TestFetchIGUserID(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v19.0/page123" {
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}

		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"instagram_business_account":{"id":"ig123"}}`))
	}))
	defer server.Close()

	oldBase := graphBaseURL
	graphBaseURL = server.URL + "/"

	defer func() { graphBaseURL = oldBase }()

	id, err := FetchIGUserID(context.Background(), "page123", "token")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if id != "ig123" {
		t.Fatalf("unexpected id: %s", id)
	}
}

func TestFetchOwnedPages(t *testing.T) {
	var server *httptest.Server
	server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/v19.0/biz123/owned_pages":
			w.Header().Set("Content-Type", "application/json")
			_, _ = fmt.Fprintf(w, `{"data":[{"id":"p1","name":"Page One","instagram_business_account":{"id":"ig1"}}],"paging":{"next":"%s/next"}}`, server.URL)
		case "/next":
			w.Header().Set("Content-Type", "application/json")
			_, _ = fmt.Fprint(w, `{"data":[{"id":"p2","name":"Page Two"}]}`)
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))

	defer server.Close()

	oldBase := graphBaseURL
	graphBaseURL = server.URL + "/"

	defer func() { graphBaseURL = oldBase }()

	pages, err := FetchOwnedPages(context.Background(), "biz123", "token")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(pages) != 2 {
		t.Fatalf("expected 2 pages, got %d", len(pages))
	}

	if pages[0].ID != "p1" || pages[0].IGUserID != "ig1" {
		t.Fatalf("unexpected page data: %+v", pages[0])
	}

	if pages[1].ID != "p2" || pages[1].Name != "Page Two" {
		t.Fatalf("unexpected page data: %+v", pages[1])
	}
}
