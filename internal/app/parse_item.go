package app

import "strings"

func ParseItemNames(text string) []string {
	r := strings.FieldsFunc(text, func(r rune) bool {
		return r == ',' || r == ';' || r == '\n'
	})

	var out []string
	for _, s := range r {
		s = strings.TrimSpace(s)
		if s != "" {
			out = append(out, s)
		}
	}
	return out
}
