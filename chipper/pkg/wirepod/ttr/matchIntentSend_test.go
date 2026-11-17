package wirepod_ttr

import "testing"

func TestMatchKey(t *testing.T) {
	if got := matchKey("Я Ушёл"); got != "я ушел" {
		t.Errorf("matchKey = %q, want %q", got, "я ушел")
	}
}
