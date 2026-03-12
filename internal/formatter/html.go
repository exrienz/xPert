package formatter

import "html"

type HTMLFormatter struct{}

func (f HTMLFormatter) Format(title string, content string) string {
	return "<!DOCTYPE html><html><html><head><meta charset=\"utf-8\"><meta name=\"viewport\" content=\"width=device-width, initial-scale=1\"><title>" + html.EscapeString(title) + "</title><style>body{font-family:system-ui,sans-serif;background:#f5f1e8;color:#1f2933;max-width:960px;margin:0 auto;padding:32px 20px;line-height:1.6}pre{white-space:pre-wrap;word-break:break-word;background:#fff;padding:24px;border:1px solid #d5d9e0;border-radius:16px;box-shadow:0 10px 30px rgba(15,23,42,.08)}</style></head><body><pre>" + html.EscapeString(content) + "</pre></body></html>"
}
