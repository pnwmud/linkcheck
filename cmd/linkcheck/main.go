package main

import (
	"flag"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/pnwmud/linkcheck/internal/checker"
	"github.com/pnwmud/linkcheck/internal/extractor"
)

func main() {
	timeoutFlag := flag.Duration("timeout", 5*time.Second, "HTTP request timeout")

	flag.Parse()

	args := flag.Args()

	if len(args) < 1 {
		fmt.Println("Err: no file path")
		os.Exit(2)
	}

	filePath := args[0]

	fmt.Printf("Configuration: timeout is %v, file %s\n", *timeoutFlag, filePath)

	file, err := os.Open(filePath)

	if err != nil {
		fmt.Printf("Err: can't open file: %v\n", err)
		os.Exit(2)
	}

	defer file.Close()

	var links []string

	ext := filepath.Ext(filePath)
	ext = strings.ToLower(ext)

	switch ext {
	case ".md":
		links, err = extractor.MarkdownExtract(file)
	case ".html":
		links, err = extractor.HTMLExtract(file)
	default:
		fmt.Println("Err: wrong file type")
		os.Exit(2)
	}

	if err != nil {
		fmt.Printf("Err: %v\n", err)
		os.Exit(2)
	}

	linksCnt := 0

	fmt.Println("\nFound links:")
	for _, link := range links {
		linksCnt += 1
		fmt.Println(link)
	}

	httpClient := &http.Client{
		Timeout: *timeoutFlag,
	}

	c := checker.New(httpClient)

	fmt.Println("\nChecking links:")

	linksBroken := 0

	for _, link := range links {
		status, err := c.Check(link)

		if err != nil {
			linksBroken += 1
			fmt.Printf("[BROKEN] %s (err: %v)\n", link, err)
		} else {
			if status >= 400 && status <= 599 {
				linksBroken += 1
				fmt.Printf("[BROKEN] %s (status: %d)\n", link, status)
			} else {
				fmt.Printf("[+] %s (status: %d)\n", link, status)
			}
		}
	}

	fmt.Printf("checked: %d, broken: %d", linksCnt, linksBroken)
}
