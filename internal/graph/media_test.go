package graph

import (
	"context"
	"net/http"
	"net/http/httptest"
	"net/url"
	"sync/atomic"
	"testing"
	"time"

	"github.com/mahmoud/igpostercli/internal/config"
)

func TestCreatePhotoContainer(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Fatalf("expected POST, got %s", r.Method)
		}

		if r.URL.Path != "/v19.0/123/media" {
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
		_ = r.ParseForm()
		assertFormValue(t, r.Form, "image_url", "https://example.com/photo.jpg")
		assertFormValue(t, r.Form, "caption", "hello")
		assertFormValue(t, r.Form, "access_token", "token")
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"id":"111"}`))
	}))
	defer server.Close()

	client := newTestClient(server)

	id, err := client.CreatePhotoContainer(context.Background(), "https://example.com/photo.jpg", "hello")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if id != "111" {
		t.Fatalf("unexpected id: %s", id)
	}
}

func TestCreateReelContainer(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = r.ParseForm()
		assertFormValue(t, r.Form, "media_type", "REELS")
		assertFormValue(t, r.Form, "video_url", "https://example.com/reel.mp4")
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"id":"222"}`))
	}))
	defer server.Close()

	client := newTestClient(server)

	id, err := client.CreateReelContainer(context.Background(), "https://example.com/reel.mp4", "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if id != "222" {
		t.Fatalf("unexpected id: %s", id)
	}
}

func TestCreateCarouselChild(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = r.ParseForm()
		assertFormValue(t, r.Form, "is_carousel_item", "true")
		assertFormValue(t, r.Form, "video_url", "https://example.com/clip.mp4")
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"id":"333"}`))
	}))
	defer server.Close()

	client := newTestClient(server)

	id, err := client.CreateCarouselChild(context.Background(), "https://example.com/clip.mp4", true)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if id != "333" {
		t.Fatalf("unexpected id: %s", id)
	}
}

func TestCreateCarouselContainer(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = r.ParseForm()
		assertFormValue(t, r.Form, "media_type", "CAROUSEL")
		assertFormValue(t, r.Form, "children", "1,2,3")
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"id":"444"}`))
	}))

	defer server.Close()

	client := newTestClient(server)

	id, err := client.CreateCarouselContainer(context.Background(), []string{"1", "2", "3"}, "caption")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if id != "444" {
		t.Fatalf("unexpected id: %s", id)
	}
}

func TestPollStatus(t *testing.T) {
	var calls int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&calls, 1)
		w.Header().Set("Content-Type", "application/json")

		if atomic.LoadInt32(&calls) < 2 {
			_, _ = w.Write([]byte(`{"status_code":"IN_PROGRESS"}`))
			return
		}
		_, _ = w.Write([]byte(`{"status_code":"FINISHED"}`))
	}))

	defer server.Close()

	client := newTestClient(server)

	err := client.PollStatus(context.Background(), "555", 10*time.Millisecond, time.Second)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestPublish(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = r.ParseForm()
		assertFormValue(t, r.Form, "creation_id", "999")
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"id":"published"}`))
	}))

	defer server.Close()

	client := newTestClient(server)

	id, err := client.Publish(context.Background(), "999")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if id != "published" {
		t.Fatalf("unexpected id: %s", id)
	}
}

func newTestClient(server *httptest.Server) *Client {
	cfg := &config.Config{
		AccessToken:  "token",
		IGUserID:     "123",
		GraphVersion: "v19.0",
	}
	client := NewClient(cfg)
	client.baseURL = server.URL + "/"
	client.httpClient = server.Client()

	return client
}

func assertFormValue(t *testing.T, form url.Values, key, expected string) {
	t.Helper()

	if form.Get(key) != expected {
		t.Fatalf("expected %s=%s, got %s", key, expected, form.Get(key))
	}
}
