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

		idy := -1
		cnt := 1

		for i := from + idx + 2; cnt != 0; i++ {
			if string(text[i]) == "(" {
				cnt += 1
			}

			if string(text[i]) == ")" {
				cnt -= 1
			}

			if cnt == 0 {
				idy = i
				break
			}
		}

		if idy == -1 {
			break
		}

		url := text[from+idx+2 : idy]

		res = append(res, url)

		from = idy + 1

	}

	return res, nil
}
