package extractor

import (
	"slices"
	"strings"
	"testing"
)

func FuzzMarkdownExtract(f *testing.F) {
	f.Add("[go](https://example.com)")
	f.Add(`\[a\](b) [c](d(e))`)
	f.Fuzz(func(t *testing.T, in string) {
		if _, err := MarkdownExtract(strings.NewReader(in)); err != nil {
			t.Fatal((err))
		}
	})
}

func TestMarkdownSoloNil(t *testing.T) {
	got, err := MarkdownExtract(strings.NewReader(""))

	if err != nil {
		t.Error(err)
	}

	var want []string = nil

	if !slices.Equal(got, want) {
		t.Errorf("got %v, want %v", got, want)
	}
}

func TestMarkdownSoloUrl(t *testing.T) {
	got, err := MarkdownExtract(strings.NewReader("[go](https://example.com)"))

	if err != nil {
		t.Error(err)
	}

	want := []string{"https://example.com"}

	if !slices.Equal(got, want) {
		t.Errorf("got %v, want %v", got, want)
	}
}

func TestMarkdownDoubleUrl(t *testing.T) {
	got, err := MarkdownExtract(strings.NewReader("[a](https://a.com) [b](https://b.com)"))

	if err != nil {
		t.Error(err)
	}

	want := []string{"https://a.com", "https://b.com"}

	if !slices.Equal(got, want) {
		t.Errorf("got %v, want %v", got, want)
	}
}

func TestMarkdownNestedParenthesesUrl(t *testing.T) {
	got, err := MarkdownExtract(strings.NewReader("[article](https://en.wikipedia.org/wiki/Foo_(bar))"))

	if err != nil {
		t.Error(err)
	}

	want := []string{"https://en.wikipedia.org/wiki/Foo_(bar)"}

	if !slices.Equal(got, want) {
		t.Errorf("got %v, want %v", got, want)
	}
}

func TestMarkdownBrokenUrl(t *testing.T) {
	got, err := MarkdownExtract(strings.NewReader("[a](https://broken"))

	if err != nil {
		t.Error(err)
	}

	want := []string{}

	if !slices.Equal(got, want) {
		t.Errorf("got %v, want %v", got, want)
	}
}

func TestMarkdownMailto(t *testing.T) {
	got, err := MarkdownExtract(strings.NewReader("[mail](mailto:someone@example.com)"))

	if err != nil {
		t.Error(err)
	}

	want := []string{"mailto:someone@example.com"}

	if !slices.Equal(got, want) {
		t.Errorf("got %v, want %v", got, want)
	}
}

func TestMarkdownAnchor(t *testing.T) {
	got, err := MarkdownExtract(strings.NewReader("[anchor](#section)"))

	if err != nil {
		t.Error(err)
	}

	want := []string{"#section"}

	if !slices.Equal(got, want) {
		t.Errorf("got %v, want %v", got, want)
	}
}

func TestMarkdownRelative(t *testing.T) {
	got, err := MarkdownExtract(strings.NewReader("[file](./docs/intro.md)"))

	if err != nil {
		t.Error(err)
	}

	want := []string{"./docs/intro.md"}

	if !slices.Equal(got, want) {
		t.Errorf("got %v, want %v", got, want)
	}
}

func TestMarkdownEscaping(t *testing.T) {
	tests := []struct {
		name string
		in   string
		want []string
	}{
		{"single backslash escapes", `That's \[NotAUrl\](https://a.com), real one [Url](https://b.com)`, []string{"https://b.com"}},
		{"double backslash is not escape", `\\](https://a.com)`, []string{"https://a.com"}},
		{"starts with close bracket pattern", `](https://a.com)`, []string{"https://a.com"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := MarkdownExtract(strings.NewReader(tt.in))
			if err != nil {
				t.Fatal(err)
			}

			if !slices.Equal(got, tt.want) {
				t.Errorf("got %v, want %v", got, tt.want)
			}
		})
	}
}

func TestHTMLNil(t *testing.T) {
	got, err := HTMLExtract(strings.NewReader(""))

	if err != nil {
		t.Error(err)
	}

	want := []string{}

	if !slices.Equal(got, want) {
		t.Errorf("got %v, want %v", got, want)
	}
}

func TestHTMLSoloUrl(t *testing.T) {
	got, err := HTMLExtract(strings.NewReader(`<a class="btn" href="https://example.org/page">клик</a>`))

	if err != nil {
		t.Error(err)
	}

	want := []string{"https://example.org/page"}

	if !slices.Equal(got, want) {
		t.Errorf("got %v, want %v", got, want)
	}
}

func TestHTMLSoloLongUrl(t *testing.T) {
	got, err := HTMLExtract(strings.NewReader(`<a class="btn">клик</a><a href="https://a.com">вторая</a>`))

	if err != nil {
		t.Error(err)
	}

	want := []string{"https://a.com"}

	if !slices.Equal(got, want) {
		t.Errorf("got %v, want %v", got, want)
	}
}

func TestHTMLDoubleUrl(t *testing.T) {
	got, err := HTMLExtract(strings.NewReader(`<a href="https://a.com">первая</a> <a href="https://b.com">вторая</a>`))

	if err != nil {
		t.Error(err)
	}

	want := []string{"https://a.com", "https://b.com"}

	if !slices.Equal(got, want) {
		t.Errorf("got %v, want %v", got, want)
	}
}

func TestHTMLBrokenUrl(t *testing.T) {
	got, err := HTMLExtract(strings.NewReader(`<a href="https://ok.com">первая</a> <a href="broken вторая</a>`))

	if err != nil {
		t.Error(err)
	}

	want := []string{"https://ok.com"}

	if !slices.Equal(got, want) {
		t.Errorf("got %v, want %v", got, want)
	}
}

func TestHTMLSrcUrl(t *testing.T) {
	got, err := HTMLExtract(strings.NewReader(`<img src="a.png">`))

	if err != nil {
		t.Error(err)
	}

	want := []string{"a.png"}

	if !slices.Equal(got, want) {
		t.Errorf("got %v, want %v", got, want)
	}
}
