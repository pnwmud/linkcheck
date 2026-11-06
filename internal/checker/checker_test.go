package checker

import (
	"context"
	"crypto/x509"
	"errors"
	"net"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"
)

type fakeDoer struct {
	err error
}

func (f fakeDoer) Do(*http.Request) (*http.Response, error) { return nil, f.err }

func TestCheckStatuses(t *testing.T) {
	tests := []struct {
		name   string
		status int
		want   Outcome
	}{
		{"200 is ok", http.StatusOK, OK},
		{"301 is redirect", http.StatusMovedPermanently, Redirect},
		{"404 is client error", http.StatusNotFound, ClientError},
		{"500 is server error", http.StatusInternalServerError, ServerError},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			srv := httptest.NewServer(http.HandlerFunc(
				func(w http.ResponseWriter, r *http.Request) {
					if tt.status >= 300 && tt.status < 400 {
						w.Header().Set("Location", "/elsewhere")
					}
					w.WriteHeader(tt.status)
				}))

			defer srv.Close()

			client := srv.Client()

			noRedirect := func(req *http.Request, via []*http.Request) error {
				return http.ErrUseLastResponse
			}

			client.CheckRedirect = noRedirect

			got := New(client).Check(context.Background(), srv.URL)

			if got.Status != tt.status {
				t.Errorf("status: got %d, want %d", got.Status, tt.status)
			}
			if got.Outcome != tt.want {
				t.Errorf("outcome: got %v, want %v", got.Outcome, tt.want)
			}
		})
	}
}

func TestCheck_Timeout(t *testing.T) {
	release := make(chan struct{})
	srv := httptest.NewServer(http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			<-release
		}))
	defer srv.Close()
	defer close(release)

	client := srv.Client()
	client.Timeout = 50 * time.Millisecond

	got := New(client).Check(context.Background(), srv.URL)

	if got.Outcome != Timeout {
		t.Errorf("outcome: got %v (err: %v), want %v", got.Outcome, got.Err, Timeout)
	}
}

func TestCheck_TLS(t *testing.T) {
	srv := httptest.NewTLSServer(http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {}))
	defer srv.Close()

	client := &http.Client{Transport: &http.Transport{}}

	got := New(client).Check(context.Background(), srv.URL)

	if got.Outcome != TLSError {
		t.Errorf("outcome: got %v (err: %v), want %v", got.Outcome, got.Err, TLSError)
	}
}

func TestCheck_DNS(t *testing.T) {
	dnsErr := &net.DNSError{Err: "no such host", Name: "nope.invalid", IsNotFound: true}

	doer := fakeDoer{err: &url.Error{Op: "Get", URL: "http://nope.invalid", Err: dnsErr}}

	got := New(doer).Check(context.Background(), "http://nope.invalid")

	if got.Outcome != DNSError {
		t.Errorf("outcome: %v (err: %v), want %v", got.Outcome, got.Err, DNSError)
	}

	var target *net.DNSError
	if !errors.As(got.Err, &target) {
		t.Errorf("error chain lost *net.DNSError: %v", got.Err)
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
