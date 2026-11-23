package wirepod_wyoming

import "testing"

func TestNormalize(t *testing.T) {
	cases := map[string]string{
		" Która jest godzina?":           "która jest godzina",
		"Ustaw minutnik na 5 minut.":     "ustaw minutnik na 5 minut",
		"Hej, Vector!  Jak się masz?":    "hej vector jak się masz",
		"What's the weather like today?": "what's the weather like today",
		"":                               "",
	}
	for in, want := range cases {
		if got := Normalize(in); got != want {
			t.Errorf("Normalize(%q) = %q, want %q", in, got, want)
		}
	}
}
