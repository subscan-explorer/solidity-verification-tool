package main

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"runtime"
	"testing"
)

func Test_downloadLatestResolc(t *testing.T) {
	// Create a mock server
	mux := http.NewServeMux()

	mux.HandleFunc("/resolc-universal-apple-darwin.tar.gz", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("mock binary content"))
	})

	mux.HandleFunc("/resolc-x86_64-unknown-linux-musl.tar.gz", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("mock binary content"))
	})

	mux.HandleFunc("/resolc-x86_64-unknown-linux-musl", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("mock binary content"))
	})
	mux.HandleFunc("/resolc-universal-apple-darwin", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("mock binary content"))
	})

	// Create a mock server with the ServeMux
	mockServer := httptest.NewServer(mux)
	// Register multiple handlers
	mux.HandleFunc("/repos/paritytech/revive/releases/latest", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(fmt.Sprintf(`{
			"tag_name": "v1.0.0",
			"assets": [
				{
					"name": "resolc-universal-apple-darwin.tar.gz",
					"browser_download_url": "%s/resolc-universal-apple-darwin.tar.gz"
				},
				{
					"name": "resolc-x86_64-unknown-linux-musl.tar.gz",
					"browser_download_url": "%s/resolc-x86_64-unknown-linux-musl.tar.gz"
				},
				{
					"name": "resolc-universal-apple-darwin",
					"browser_download_url": "%s/resolc-universal-apple-darwin"
				},
				{
					"name": "resolc-x86_64-unknown-linux-musl",
					"browser_download_url": "%s/resolc-x86_64-unknown-linux-musl"
				}
			]
		}`, mockServer.URL, mockServer.URL, mockServer.URL, mockServer.URL)))
	})
	defer mockServer.Close()

	// Call the function with the mock server URL
	fileName := downloadLatestResolc(mockServer.URL + "/repos/paritytech/revive/releases/latest")

	// Check if the file was downloaded correctly
	if runtime.GOOS == "darwin" && (fileName != "resolc-universal-apple-darwin.tar.gz" && fileName != "resolc-universal-apple-darwin") {
		t.Fatalf("expected file name to be resolc-universal-apple-darwin.tar.gz, got %s", fileName)
	}
	if runtime.GOOS == "linux" && (fileName != "resolc-x86_64-unknown-linux-musl.tar.gz" && fileName != "resolc-x86_64-unknown-linux-musl") {
		t.Fatalf("expected file name to be resolc-x86_64-unknown-linux-musl.tar.gz, got %s", fileName)
	}
}

func Test_extractAndSetExec(t *testing.T) {
	// Create a sample tar.gz file
	var buf bytes.Buffer
	gzw := gzip.NewWriter(&buf)
	tw := tar.NewWriter(gzw)

	// Add a file to the tar archive
	execFile := "test-exec-file"
	content := []byte("test content")
	hdr := &tar.Header{
		Name: execFile,
		Mode: 0755,
		Size: int64(len(content)),
	}
	if err := tw.WriteHeader(hdr); err != nil {
		t.Fatalf("failed to write tar header: %v", err)
	}
	if _, err := tw.Write(content); err != nil {
		t.Fatalf("failed to write file content: %v", err)
	}
	if err := tw.Close(); err != nil {
		t.Fatalf("failed to close tar writer: %v", err)
	}
	if err := gzw.Close(); err != nil {
		t.Fatalf("failed to close gzip writer: %v", err)
	}

	// Write the tar.gz file to disk
	src := "test.tar.gz"
	if err := os.WriteFile(src, buf.Bytes(), 0644); err != nil {
		t.Fatalf("failed to write tar.gz file: %v", err)
	}
	defer os.Remove(src)

	// Create a destination directory
	dest := "test-dest"
	if err := os.Mkdir(dest, 0755); err != nil {
		t.Fatalf("failed to create destination directory: %v", err)
	}
	defer os.RemoveAll(dest)

	// Call the function
	if err := extractAndSetExec(src, dest, execFile, "renamed-exec-file"); err != nil {
		t.Fatalf("extractAndSetExec failed: %v", err)
	}

	// Check if the file was extracted and renamed correctly
	extractedFile := filepath.Join(dest, "renamed-exec-file")
	if _, err := os.Stat(extractedFile); os.IsNotExist(err) {
		t.Fatalf("expected file %s to exist, but it does not", extractedFile)
	}

	// Check if the file content is correct
	extractedContent, err := os.ReadFile(extractedFile)
	if err != nil {
		t.Fatalf("failed to read extracted file: %v", err)
	}
	if !bytes.Equal(extractedContent, content) {
		t.Fatalf("expected file content to be %s, got %s", content, extractedContent)
	}
}
