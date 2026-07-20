package app

import "strings"

func ParseItemNames(text string) []string {
	r := strings.NewReplacer(
		",", " ",
		";", " ",
		".", " ",
		"\n", " ",
	)
	rr := r.Replace(text)
	return strings.Fields(rr)
}
