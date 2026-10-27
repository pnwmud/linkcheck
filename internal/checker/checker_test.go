package checker

import (
	"context"
	"crypto/x509"
	"errors"
	"io"
	"net"
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

	res := h.Check(context.Background(), "http://example.com")

	if res.Status != wantS {
		t.Errorf("got %d, want %d", res.Status, wantS)
	}

	if res.Outcome != OK {
		t.Errorf("got %v, want %v", res.Outcome, OK)
	}

	if res.Err != nil {
		t.Errorf("got %v, want %v", res.Err, nil)
	}
}

func TestCheck_NotFound(t *testing.T) {
	f := fakeDoer{status: 404, err: nil}
	h := HTTPChecker{client: f}
	wantS := 404

	res := h.Check(context.Background(), "http://example.com")

	if res.Status != wantS {
		t.Errorf("got %d, want %d", res.Status, wantS)
	}

	if res.Outcome != ClientError {
		t.Errorf("got %v, want %v", res.Outcome, ClientError)
	}

	if res.Err != nil {
		t.Errorf("got %v, want %v", res.Err, nil)
	}
}

func TestCheck_NetworkError(t *testing.T) {
	NetErr := errors.New("network down: connection refused")
	f := fakeDoer{status: 0, err: NetErr}
	h := HTTPChecker{client: f}
	wantS := 0

	res := h.Check(context.Background(), "http://example.com")

	if res.Status != wantS {
		t.Errorf("got %d, want %d", res.Status, wantS)
	}

	if res.Outcome != OtherError {
		t.Errorf("got %v, want %v", res.Outcome, OtherError)
	}

	if !errors.Is(res.Err, NetErr) {
		t.Errorf("got %v, want %v", res.Err, NetErr)
	}
}

func TestOutcome_Broken(t *testing.T) {
	tests := []struct {
		name    string
		outcome Outcome
		want    bool
	}{
		{"ok not broken", OK, false},
		{"redirect not broken", Redirect, false},
		{"client error broken", ClientError, true},
		{"server error broken", ServerError, true},
		{"timeout broken", Timeout, true},
		{"dns error broken", DNSError, true},
		{"tls error broken", TLSError, true},
		{"other error broken", OtherError, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.outcome.Broken(); got != tt.want {
				t.Errorf("got %v, want %v", got, tt.want)
			}
		})
	}
}

func TestClassifyStatus(t *testing.T) {
	tests := []struct {
		name   string
		status int
		want   Outcome
	}{
		{"StatusOK", 200, OK},
		{"StatusRedirect", 300, Redirect},
		{"StatusClientErr", 400, ClientError},
		{"StatusServerErr", 500, ServerError},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := classifyStatus(tt.status); got != tt.want {
				t.Errorf("got %v, want %v", got, tt.want)
			}
		})
	}
}

func TestClassifyError(t *testing.T) {
	tests := []struct {
		name string
		err  error
		want Outcome
	}{
		{"nil is other", nil, OtherError},
		{"dns", &net.DNSError{Err: "no such host"}, DNSError},
		{"deadline", context.DeadlineExceeded, Timeout},
		{"tls unknown authority", x509.UnknownAuthorityError{}, TLSError},
		{"other", errors.New("hoba"), OtherError},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := classifyError(tt.err); got != tt.want {
				t.Errorf("got %v, want %v", got, tt.want)
			}
		})
	}
}
