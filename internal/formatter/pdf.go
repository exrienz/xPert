package formatter

import "html"

type PDFFormatter struct{}

func (f PDFFormatter) Format(title string, content string) string {
	return "<!DOCTYPE html><html><head><meta charset=\"utf-8\"><meta name=\"viewport\" content=\"width=device-width, initial-scale=1\"><title>" + html.EscapeString(title) + "</title><style>body{font-family:serif;max-width:860px;margin:40px auto;padding:0 20px;line-height:1.5}pre{white-space:pre-wrap;word-break:break-word}</style></head><body><pre>" + html.EscapeString(content) + "</pre></body></html>"
}
