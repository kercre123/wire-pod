package wirepod_ttr

import "testing"

func TestMatchKey(t *testing.T) {
	if got := matchKey("Я Ушёл"); got != "я ушел" {
		t.Errorf("matchKey = %q, want %q", got, "я ушел")
	}
}

func TestContainsKeyphrase(t *testing.T) {
	tests := []struct {
		text, keyphrase string
		want            bool
	}{
		// short keyphrases only match as separate words
		{"давай поиграем", "да", false},
		{"когда будет дождь", "да", false},
		{"да", "да", true},
		{"ну да конечно", "да", true},
		{"да, давай", "да", true},
		{"нет спасибо", "нет", true},
		{"иди домой", "дом", false},
		{"иди в дом", "дом", true},
		{"расскажи еще", "ещё", true},
		// longer keyphrases keep the substring behaviour
		{"а сделай громкость потише", "громкость", true},
		{"двигайся вперед быстрее", "двигайся вперёд", true},
		{"какая погода", "погода", true},
		{"", "да", false},
		{"да", "", false},
	}
	for _, tt := range tests {
		if got := containsKeyphrase(tt.text, tt.keyphrase); got != tt.want {
			t.Errorf("containsKeyphrase(%q, %q) = %v, want %v", tt.text, tt.keyphrase, got, tt.want)
		}
	}
}
