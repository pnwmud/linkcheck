package checker

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"errors"
	"fmt"
	"net"
	"net/http"
)

type Outcome int

const (
	OK Outcome = iota
	Redirect
	ClientError
	ServerError
	Timeout
	DNSError
	TLSError
	OtherError
)

func (o Outcome) Broken() bool { return o != OK && o != Redirect }

type Result struct {
	URL     string
	Status  int
	Outcome Outcome
	Err     error
}

func (o Outcome) String() string {
	switch o {
	case OK:
		return "ok"
	case Redirect:
		return "redirect"
	case ClientError:
		return "client error"
	case ServerError:
		return "server error"
	case Timeout:
		return "timeout"
	case DNSError:
		return "dns error"

	case TLSError:
		return "tls error"
	case OtherError:
		return "other error"
	default:
		return "unknown"
	}
}

func classifyStatus(status int) Outcome {
	switch {
	case status >= 200 && status < 300:
		return OK
	case status >= 300 && status < 400:
		return Redirect
	case status >= 400 && status < 500:
		return ClientError
	case status >= 500 && status < 600:
		return ServerError
	default:
		return OtherError

	}
}

func classifyError(err error) Outcome {
	var (
		dnsErr      *net.DNSError
		netErr      net.Error
		unknwnAuth  x509.UnknownAuthorityError
		hostErr     x509.HostnameError
		invalidCert x509.CertificateInvalidError
		verifyErr   *tls.CertificateVerificationError
		recordErr   tls.RecordHeaderError
	)
	switch {
	case errors.As(err, &dnsErr):
		return DNSError
	case errors.As(err, &unknwnAuth), errors.As(err, &hostErr), errors.As(err, &invalidCert), errors.As(err, &verifyErr), errors.As(err, &recordErr):
		return TLSError
	case errors.As(err, &netErr) && netErr.Timeout():
		return Timeout
	case errors.Is(err, context.DeadlineExceeded):
		return Timeout
	default:
		return OtherError
	}
}

func (h *HTTPChecker) Check(ctx context.Context, url string) Result {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)

	if err != nil {
		return Result{URL: url, Status: 0, Outcome: classifyError(err), Err: fmt.Errorf("check %q: %w", url, err)}
	}

	done, err := h.client.Do(req)

	if err != nil {
		return Result{URL: url, Status: 0, Outcome: classifyError(err), Err: fmt.Errorf("check %q: %w", url, err)}
	}

	defer done.Body.Close()

	return Result{URL: url, Status: done.StatusCode, Outcome: classifyStatus(done.StatusCode), Err: nil}
}

type HTTPChecker struct {
	client Doer
}

type Doer interface {
	Do(req *http.Request) (*http.Response, error)
}

func New(d Doer) *HTTPChecker {
	return &HTTPChecker{
		client: d,
	}
}
