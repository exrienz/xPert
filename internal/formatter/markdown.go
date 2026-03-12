package formatter

type MarkdownFormatter struct{}

func (f MarkdownFormatter) Format(content string) string {
	return content
}
