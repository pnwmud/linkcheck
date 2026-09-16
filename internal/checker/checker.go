package checker

import (
	"net/http"
)

func (h *HTTPChecker) Check(url string) (status int, err error) {
	req, err := http.NewRequest("GET", url, nil)

	if err != nil {
		return status, err
	}

	done, err := h.client.Do(req)

	if err != nil {
		return status, err
	}

	defer done.Body.Close()

	status = done.StatusCode

	return status, nil
}

type HTTPChecker struct {
	client Doer
}

type Doer interface {
	Do(req *http.Request) (*http.Response, error)
}
