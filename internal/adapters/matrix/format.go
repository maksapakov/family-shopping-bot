package matrix

import (
	"strings"

	"github.com/maksapakov/family-shopping-bot/internal/domain"
)

func formatList(rendered domain.RenderedList) (string, string) {

	var plain strings.Builder
	var html strings.Builder

	for _, item := range rendered.Items {
		mark := "☐"
		if item.IsChecked {
			mark = "☑"
		}

		plain.WriteString(mark)
		plain.WriteString(" ")
		plain.WriteString(item.Name)
		plain.WriteString("\n")

		html.WriteString(mark)
		html.WriteString(` <a href="`)
		html.WriteString(item.ClickURL)
		html.WriteString(`">`)
		html.WriteString(item.Name)
		html.WriteString("</a><br>")
	}

	plain.WriteString("undo\n")

	html.WriteString(`<a href="`)
	html.WriteString(rendered.UndoURL)
	html.WriteString(`">undo</a>`)

	return plain.String(), html.String()
}
