package extractor

import (
	"io"
	"strings"

	"golang.org/x/net/html"
)

func isEscaped(text string, pos int) bool {
	n := 0
	for i := pos - 1; i >= 0 && text[i] == '\\'; i-- {
		n++
	}
	return n%2 == 1
}

func MarkdownExtract(r io.Reader) ([]string, error) {

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

		if isEscaped(text, from+idx) {
			from = from + idx + 2
			continue
		}

		idy := -1
		cnt := 1

		for i := from + idx + 2; cnt != 0 && i < len(text); i++ {
			if text[i] == '(' {
				cnt += 1
			}

			if text[i] == ')' {
				cnt -= 1
			}

			if cnt == 0 {
				idy = i
				break
			}
		}

		if idy == -1 {
			from = from + idx + 2
			continue
		}

		url := text[from+idx+2 : idy]

		res = append(res, url)

		from = idy + 1

	}

	return res, nil
}

func HTMLExtract(r io.Reader) ([]string, error) {
	links := []string{}
	z := html.NewTokenizer(r)

	for {
		switch z.Next() {
		case html.ErrorToken:
			if err := z.Err(); err != io.EOF {
				return nil, err
			}
			return links, nil
		case html.StartTagToken, html.SelfClosingTagToken:
			for _, attr := range z.Token().Attr {
				if (attr.Key == "href" || attr.Key == "src") && attr.Val != "" {
					links = append(links, attr.Val)
				}
			}
		}
	}
}
