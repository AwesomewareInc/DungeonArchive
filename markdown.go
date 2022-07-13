package main

// Passthrough function for markdown parsing to template shit.

import (
	"strings"
	"github.com/gomarkdown/markdown"
)

func ParseMarkdown(value string) (string) {
	result := string(markdown.ToHTML([]byte(value), nil, nil))
	result = strings.Replace(result, "<p>","",1)
	result = strings.Replace(result, "</p>","",1)
	return result
}