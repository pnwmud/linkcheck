package checker

import (
	"errors"
	"io"
	"net/http"
	"strings"
	"testing"
)

type fakeDoer struct {
	status int
	err    error
}

func (f fakeDoer) Do(req *http.Request) (*http.Response, error) {
	resp := &http.Response{StatusCode: f.status, Body: io.NopCloser(strings.NewReader(""))}
	return resp, f.err
}

func TestCheck_OK(t *testing.T) {
	f := fakeDoer{status: 200, err: nil}
	h := HTTPChecker{client: f}
	wantS := 200

	status, err := h.Check("http://example.com")

	if status != wantS {
		t.Errorf("got %d, want %d", status, wantS)
	}

	if err != nil {
		t.Errorf("got %v, want %v", err, nil)
	}
}

func TestCheck_NotFound(t *testing.T) {
	f := fakeDoer{status: 404, err: nil}
	h := HTTPChecker{client: f}
	wantS := 404

	status, err := h.Check("http://example.com")

	if status != wantS {
		t.Errorf("got %d, want %d", status, wantS)
	}

	if err != nil {
		t.Errorf("got %v, want %v", err, nil)
	}
}

func TestCheck_NetworkError(t *testing.T) {
	NetErr := errors.New("network down: connection refused")
	f := fakeDoer{status: 0, err: NetErr}
	h := HTTPChecker{client: f}
	wantS := 0

	status, err := h.Check("http://example.com")

	if status != wantS {
		t.Errorf("got %d, want %d", status, wantS)
	}

	if err != NetErr {
		t.Errorf("got %v, want %v", err, NetErr)
	}
}
