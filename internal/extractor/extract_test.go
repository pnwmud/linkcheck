package extractor

import (
	"slices"
	"strings"
	"testing"
)

func TestMarkdownSoloNil(t *testing.T) {
	got, err := Extract(strings.NewReader(""))

	if err != nil {
		t.Error(err)
	}

	var want []string = nil

	if !slices.Equal(got, want) {
		t.Errorf("got %v, want %v", got, want)
	}
}

func TestMarkdownSoloUrl(t *testing.T) {
	got, err := Extract(strings.NewReader("[go](https://example.com)"))

	if err != nil {
		t.Error(err)
	}

	want := []string{"https://example.com"}

	if !slices.Equal(got, want) {
		t.Errorf("got %v, want %v", got, want)
	}
}

func TestMarkdownDoubleUrl(t *testing.T) {
	got, err := Extract(strings.NewReader("[a](https://a.com) [b](https://b.com)"))

	if err != nil {
		t.Error(err)
	}

	want := []string{"https://a.com", "https://b.com"}

	if !slices.Equal(got, want) {
		t.Errorf("got %v, want %v", got, want)
	}
}

func TestMarkdownIncludedUrl(t *testing.T) {
	got, err := Extract(strings.NewReader("[article](https://en.wikipedia.org/wiki/Foo_(bar))"))

	if err != nil {
		t.Error(err)
	}

	want := []string{"https://en.wikipedia.org/wiki/Foo_(bar)"}

	if !slices.Equal(got, want) {
		t.Errorf("got %v, want %v", got, want)
	}
}
