package extractor

import (
	"io"
	"strings"
)

func Extract(r io.Reader) ([]string, error) {

	data, err := io.ReadAll(r)

	if err != nil {
		return nil, err
	}

	text := string(data)
	from := 0

	res := []string{}

	for {
		idx := strings.Index(text[from:], "](")

		if idx == -1 {
			break
		}

		idy := strings.Index(text[from:], ")")

		if idy == -1 {
			break
		}

		url := text[from+idx+2 : from+idy]

		res = append(res, url)

		from = from + idy + 1

	}

	return res, nil
}
