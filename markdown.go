package main

import (
	"regexp"

	"github.com/JohannesKaufmann/html-to-markdown/v2/converter"
	"github.com/JohannesKaufmann/html-to-markdown/v2/plugin/base"
	"github.com/JohannesKaufmann/html-to-markdown/v2/plugin/commonmark"
	"github.com/JohannesKaufmann/html-to-markdown/v2/plugin/table"
)

// Description formats.
const (
	HTMLFormat     = "html"
	MarkdownFormat = "markdown"
)

var (
	supRegex    = regexp.MustCompile(`(?s)<sup>(.*?)</sup>`)
	subRegex    = regexp.MustCompile(`(?s)<sub>(.*?)</sub>`)
	simpleRegex = regexp.MustCompile(`^\w+$`)

	markdownConverter = converter.NewConverter(converter.WithPlugins(
		base.NewBasePlugin(),
		commonmark.NewCommonmarkPlugin(),
		table.NewTablePlugin(),
	))
)

// htmlToMarkdown converts a problem description to Markdown.
// Superscripts and subscripts become ^x and _x (parenthesized when complex),
// since Markdown has no syntax for them and they often appear inside code spans.
func htmlToMarkdown(html string) (string, error) {
	html = supRegex.ReplaceAllStringFunc(html, func(s string) string {
		return scriptNotation("^", supRegex.FindStringSubmatch(s)[1])
	})
	html = subRegex.ReplaceAllStringFunc(html, func(s string) string {
		return scriptNotation("_", subRegex.FindStringSubmatch(s)[1])
	})
	return markdownConverter.ConvertString(html)
}

func scriptNotation(marker, text string) string {
	if simpleRegex.MatchString(text) {
		return marker + text
	}
	return marker + "(" + text + ")"
}
