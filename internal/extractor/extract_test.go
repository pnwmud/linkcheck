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

func TestMarkdownNestedParenthesesUrl(t *testing.T) {
	got, err := Extract(strings.NewReader("[article](https://en.wikipedia.org/wiki/Foo_(bar))"))

	if err != nil {
		t.Error(err)
	}

	want := []string{"https://en.wikipedia.org/wiki/Foo_(bar)"}

	if !slices.Equal(got, want) {
		t.Errorf("got %v, want %v", got, want)
	}
}

func TestMarkdownBrokenUrl(t *testing.T) {
	got, err := Extract(strings.NewReader("[a](https://broken"))

	if err != nil {
		t.Error(err)
	}

	want := []string{}

	if !slices.Equal(got, want) {
		t.Errorf("got %v, want %v", got, want)
	}
}

func TestMarkdownMailto(t *testing.T) {
	got, err := Extract(strings.NewReader("[mail](mailto:someone@example.com)"))

	if err != nil {
		t.Error(err)
	}

	want := []string{"mailto:someone@example.com"}

	if !slices.Equal(got, want) {
		t.Errorf("got %v, want %v", got, want)
	}
}

func TestMarkdownAnchor(t *testing.T) {
	got, err := Extract(strings.NewReader("[anchor](#section)"))

	if err != nil {
		t.Error(err)
	}

	want := []string{"#section"}

	if !slices.Equal(got, want) {
		t.Errorf("got %v, want %v", got, want)
	}
}

func TestMarkdownRelative(t *testing.T) {
	got, err := Extract(strings.NewReader("[file](./docs/intro.md)"))

	if err != nil {
		t.Error(err)
	}

	want := []string{"./docs/intro.md"}

	if !slices.Equal(got, want) {
		t.Errorf("got %v, want %v", got, want)
	}
}

func TestMarkdownEscaping(t *testing.T) {
	got, err := Extract(strings.NewReader("That's \\[NotAUrl\\](https://a.com), real one [Url](https://b.com)"))

	if err != nil {
		t.Error(err)
	}

	want := []string{"https://b.com"}

	if !slices.Equal(got, want) {
		t.Errorf("got %v, want %v", got, want)
	}
}
