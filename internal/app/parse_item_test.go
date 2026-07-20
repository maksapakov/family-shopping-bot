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
		{"comma_space_dot", "Картошка, Морковь Соль. перец",
			[]string{"Картошка", "Морковь", "Соль", "перец"}},
		{"only_separators", "; , . ", []string{}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ParseItemNames(tt.in)
			if len(got) != len(tt.want) {
				t.Errorf("parseItemNames() = %v, want %v", got, tt.want)
			}
			for i := range tt.want {
				if got[i] != tt.want[i] {
					t.Errorf("parseItemNames() = %v, want %v", got, tt.want)
				}
			}
		},
		)
	}
}
