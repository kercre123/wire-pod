package sdkapp

import (
	"testing"
	"time"
)

func at(h, m int) time.Time {
	return time.Date(2024, 1, 1, h, m, 0, 0, time.Local)
}

func TestInActiveHours(t *testing.T) {
	cases := []struct {
		start, end string
		h, m       int
		want       bool
	}{
		{"10:00", "20:00", 9, 59, false},
		{"10:00", "20:00", 10, 0, true},
		{"10:00", "20:00", 19, 59, true},
		{"10:00", "20:00", 20, 0, false},
		// spans midnight
		{"22:00", "07:00", 23, 30, true},
		{"22:00", "07:00", 6, 59, true},
		{"22:00", "07:00", 12, 0, false},
		// equal or invalid windows never put a robot to sleep
		{"10:00", "10:00", 3, 0, true},
		{"bogus", "20:00", 23, 0, true},
	}
	for _, c := range cases {
		if got := InActiveHours(at(c.h, c.m), c.start, c.end); got != c.want {
			t.Errorf("InActiveHours(%02d:%02d, %s-%s) = %v, want %v", c.h, c.m, c.start, c.end, got, c.want)
		}
	}
}

func TestParseClock(t *testing.T) {
	if m, err := ParseClock("08:30"); err != nil || m != 510 {
		t.Errorf("ParseClock(08:30) = %d, %v", m, err)
	}
	for _, bad := range []string{"24:00", "12:60", "noon", ""} {
		if _, err := ParseClock(bad); err == nil {
			t.Errorf("ParseClock(%q) should fail", bad)
		}
	}
}
