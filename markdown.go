package main

// Passthrough function for markdown parsing to template shit.

import (
	"strings"

	"github.com/gomarkdown/markdown"
	"github.com/gomarkdown/markdown/html"
	"github.com/gomarkdown/markdown/parser"
)

func ParseMarkdown(md string) string {

	extensions := parser.CommonExtensions | parser.AutoHeadingIDs | parser.NoEmptyLineBeforeBlock
	p := parser.NewWithExtensions(extensions)
	doc := p.Parse([]byte(md))
	// create HTML renderer with extensions
	htmlFlags := html.CommonFlags | html.HrefTargetBlank
	opts := html.RendererOptions{Flags: htmlFlags}
	renderer := html.NewRenderer(opts)
	result := string(markdown.Render(doc, renderer))
	result = strings.Replace(result, "<p>", "", 1)
	result = strings.Replace(result, "</p>", "", 1)
	result = strings.Replace(result, "<ul>", "", 1)
	result = strings.Replace(result, "</ul>", "", 1)
	result = strings.Replace(result, "<li>", "*", 1)
	result = strings.Replace(result, "</li>", "", 1)
	result = strings.Replace(result, "<hr>", " ", 1)
	return result
}
