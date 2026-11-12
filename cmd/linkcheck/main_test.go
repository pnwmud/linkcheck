package main

import (
	"bytes"
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func newSite(t *testing.T) *httptest.Server {
	t.Helper()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/ok" {
			w.WriteHeader(http.StatusOK)
			return
		}
		w.WriteHeader(http.StatusNotFound)
	}))
	t.Cleanup(srv.Close)
	return srv
}

func TestRunExitCodes(t *testing.T) {
	srv := newSite(t)

	tests := []struct {
		name, file, content string
		wantCode            int
		wantOut             string
	}{
		{"all links alive", "a.md", "[ok](" + srv.URL + "/ok)", exitOK, "checked: 1, broken: 0"},
		{"broken link", "a.md", "[ok](" + srv.URL + "/ok) [x](" + srv.URL + "/gone)", exitBroken, "checked: 2, broken: 1"},
		{"non-http links are skipped", "a.md", "[m](mailto:a@b.c) [s](#top) [f](./x.md)", exitOK, "checked: 0, broken: 0"},
		{"duplicates are checked once", "a.md", "[a](" + srv.URL + "/ok) [b](" + srv.URL + "/ok)", exitOK, "checked: 1, broken: 0"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			path := filepath.Join(t.TempDir(), tt.file)
			if err := os.WriteFile(path, []byte(tt.content), 0o644); err != nil {
				t.Fatal(err)
			}
			var stdout, stderr bytes.Buffer

			code := run(context.Background(), []string{path}, &stdout, &stderr, srv.Client())

			if code != tt.wantCode {
				t.Errorf("exit code: got %d, want %d", code, tt.wantCode)
			}
			if !strings.Contains(stdout.String(), tt.wantOut) {
				t.Errorf("stdout %q does not contain %q", stdout.String(), tt.wantOut)
			}
		})
	}
}
