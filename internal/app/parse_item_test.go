package app

import "testing"

func TestParseItemNames(t *testing.T) {
	tests := []struct {
		name string
		in   string
		want []string
	}{
		{"empty", "", []string{}},
		{"single", "молоко", []string{"молоко"}},
		{"compound_with_space", "зубная паста", []string{"зубная паста"}},
		{"comma_three", "Молоко, зубная паста, хлеб", []string{"Молоко", "зубная паста", "хлеб"}},
		{"comma_two", "Картошка, Морковь", []string{"Картошка", "Морковь"}},
		{"space_not_split", "Картошка Морковь", []string{"Картошка Морковь"}},
		{"semicolon", "Молоко; хлеб", []string{"Молоко", "хлеб"}},
		{"newline", "Молоко\nхлеб", []string{"Молоко", "хлеб"}},
		{"do_not_split", "зуб.паста", []string{"зуб.паста"}},
		{"only_separators", "; , \n ", []string{}},
		{"trim_spaces", " Молоко , хлеб ", []string{"Молоко", "хлеб"}},
		{"comma_with_gap", " Молоко, , хлеб ", []string{"Молоко", "хлеб"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ParseItemNames(tt.in)
			if len(got) != len(tt.want) {
				t.Errorf("parseItemNames() = %q, want %q", got, tt.want)
			}
			for i := range tt.want {
				if got[i] != tt.want[i] {
					t.Errorf("parseItemNames() = %q, want %q", got, tt.want)
				}
			}
		},
		)
	}
}
