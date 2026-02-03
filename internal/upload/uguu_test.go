package upload

import (
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
)

func TestUploadSuccess(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Fatalf("expected POST, got %s", r.Method)
		}
		_ = r.ParseMultipartForm(10 << 20)

		_, _, err := r.FormFile("files[]")
		if err != nil {
			t.Fatalf("missing files[]: %v", err)
		}

		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"success":true,"files":[{"url":"https://uguu.se/file.png"}]}`))
	}))
	defer server.Close()

	oldURL := uguuUploadURL
	uguuUploadURL = server.URL

	defer func() { uguuUploadURL = oldURL }()

	filePath := filepath.Join(t.TempDir(), "file.png")
	if err := os.WriteFile(filePath, []byte("data"), 0o600); err != nil {
		t.Fatalf("write temp file: %v", err)
	}

	url, err := Upload(context.Background(), filePath)
	if err != nil {
		t.Fatalf("upload failed: %v", err)
	}

	if url != "https://uguu.se/file.png" {
		t.Fatalf("unexpected url: %s", url)
	}
}

func TestUploadRejectsNonHTTPS(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"success":true,"files":[{"url":"http://uguu.se/file.png"}]}`))
	}))
	defer server.Close()

	oldURL := uguuUploadURL
	uguuUploadURL = server.URL

	defer func() { uguuUploadURL = oldURL }()

	filePath := filepath.Join(t.TempDir(), "file.png")
	if err := os.WriteFile(filePath, []byte("data"), 0o600); err != nil {
		t.Fatalf("write temp file: %v", err)
	}

	_, err := Upload(context.Background(), filePath)
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestUploadHTTPError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte("oops"))
	}))
	defer server.Close()

	oldURL := uguuUploadURL
	uguuUploadURL = server.URL

	defer func() { uguuUploadURL = oldURL }()

	filePath := filepath.Join(t.TempDir(), "file.png")
	if err := os.WriteFile(filePath, []byte("data"), 0o600); err != nil {
		t.Fatalf("write temp file: %v", err)
	}

	_, err := Upload(context.Background(), filePath)
	if err == nil {
		t.Fatal("expected error")
	}
}
