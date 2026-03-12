package formatter

import (
	"encoding/json"
	"strings"
)

type Set struct {
	markdown MarkdownFormatter
	html     HTMLFormatter
	pdf      PDFFormatter
}

func NewFormatterSet() *Set {
	return &Set{
		markdown: MarkdownFormatter{},
		html:     HTMLFormatter{},
		pdf:      PDFFormatter{},
	}
}

func (s *Set) Format(format string, title string, markdown string, reviewSummary []string) string {
	switch format {
	case "html":
		return s.html.Format(title, markdown)
	case "pdf":
		return s.pdf.Format(title, markdown)
	case "notion":
		return notionExport(title, markdown, reviewSummary)
	case "confluence":
		return confluenceExport(title, markdown, reviewSummary)
	case "json":
		payload, _ := json.MarshalIndent(map[string]any{
			"title":          title,
			"content":        markdown,
			"review_summary": reviewSummary,
		}, "", "  ")
		return string(payload)
	default:
		return s.markdown.Format(markdown)
	}
}

func notionExport(title, markdown string, reviewSummary []string) string {
	var builder strings.Builder
	builder.WriteString("NOTION_PAGE\n")
	builder.WriteString("title: " + title + "\n")
	builder.WriteString("quality_summary:\n")
	for _, note := range reviewSummary {
		builder.WriteString("- " + note + "\n")
	}
	builder.WriteString("content:\n")
	builder.WriteString(markdown)
	return builder.String()
}

func confluenceExport(title, markdown string, reviewSummary []string) string {
	var builder strings.Builder
	builder.WriteString("<ac:structured-macro ac:name=\"info\"><ac:rich-text-body>")
	for _, note := range reviewSummary {
		builder.WriteString("<p>" + escapeXML(note) + "</p>")
	}
	builder.WriteString("</ac:rich-text-body></ac:structured-macro>")
	builder.WriteString("<h1>" + escapeXML(title) + "</h1>")
	for _, line := range strings.Split(markdown, "\n") {
		if strings.HasPrefix(line, "# ") {
			builder.WriteString("<h2>" + escapeXML(strings.TrimPrefix(line, "# ")) + "</h2>")
			continue
		}
		if strings.HasPrefix(line, "## ") {
			builder.WriteString("<h3>" + escapeXML(strings.TrimPrefix(line, "## ")) + "</h3>")
			continue
		}
		if strings.HasPrefix(line, "### ") {
			builder.WriteString("<h4>" + escapeXML(strings.TrimPrefix(line, "### ")) + "</h4>")
			continue
		}
		if strings.TrimSpace(line) == "" {
			continue
		}
		builder.WriteString("<p>" + escapeXML(line) + "</p>")
	}
	return builder.String()
}

func escapeXML(value string) string {
	replacer := strings.NewReplacer("&", "&amp;", "<", "&lt;", ">", "&gt;", "\"", "&quot;", "'", "&apos;")
	return replacer.Replace(value)
}
