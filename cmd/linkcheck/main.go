package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/pnwmud/linkcheck/internal/checker"
	"github.com/pnwmud/linkcheck/internal/extractor"
)

const (
	exitOK     = 0
	exitBroken = 1
	exitUsage  = 2
)

var errUnsupportedType = errors.New("unsupported file type")

func extractFile(path string) ([]string, error) {
	f, err := os.Open(path)

	if err != nil {
		return nil, fmt.Errorf("open %s: %w", path, err)
	}

	defer f.Close()

	switch strings.ToLower(filepath.Ext(path)) {
	case ".md":
		return extractor.MarkdownExtract(f)
	case ".html":
		return extractor.HTMLExtract(f)
	default:
		return nil, fmt.Errorf("%w: %q", errUnsupportedType, filepath.Ext(path))
	}

}

func unique(links []string) []string {
	seen := make(map[string]struct{}, len(links))
	out := make([]string, 0, len(links))

	for _, l := range links {
		if _, ok := seen[l]; ok {
			continue
		}
		seen[l] = struct{}{}
		out = append(out, l)
	}

	return out
}

func isHTTP(link string) bool {
	u, err := url.Parse(link)

	if err != nil {
		return false
	}

	if (u.Scheme == "http" || u.Scheme == "https") && u.Host != "" {
		return true
	}

	return false
}

func describe(res checker.Result) string {
	s := res.Outcome.String()
	if res.Status > 0 {
		s += fmt.Sprintf(", status %d", res.Status)
	}
	if res.Err != nil {
		s += fmt.Sprintf(": %v", res.Err)
	}
	return s
}

func run(ctx context.Context, args []string, stdout, stderr io.Writer, client checker.Doer) int {
	fs := flag.NewFlagSet("linkcheck", flag.ContinueOnError)
	fs.SetOutput(stderr)
	timeout := fs.Duration("timeout", 5*time.Second, "HTTP request timeout")

	if err := fs.Parse(args); err != nil {
		return exitUsage
	}

	if fs.NArg() != 1 {
		fmt.Fprintln(stderr, "usage: linkcheck [flags] <file.md|file.html>")
		return exitUsage
	}

	links, err := extractFile(fs.Arg(0))
	if err != nil {
		fmt.Fprintf(stderr, "linkcheck: %v\n", err)
		return exitUsage
	}

	if client == nil {
		client = &http.Client{
			Timeout:       *timeout,
			CheckRedirect: func(_ *http.Request, _ []*http.Request) error { return http.ErrUseLastResponse },
		}
	}

	c := checker.New(client)

	checked, broken := 0, 0
	for _, link := range unique(links) {
		if !isHTTP(link) {
			fmt.Fprintf(stdout, "[SKIP] %s\n", link)
			continue
		}

		res := c.Check(ctx, link)
		checked++

		if res.Outcome.Broken() {
			broken++
			fmt.Fprintf(stdout, "[BROKEN] %s (%s)\n", link, describe(res))
		} else {
			fmt.Fprintf(stdout, "[OK] %s (%d)\n", link, res.Status)
		}
	}

	fmt.Fprintf(stdout, "checked: %d, broken: %d\n", checked, broken)

	if broken > 0 {
		return exitBroken
	}
	return exitOK

}

func main() {
	os.Exit(run(context.Background(), os.Args[1:], os.Stdout, os.Stderr, nil))
}
