package memories

import (
	"strings"
	"testing"
)

func setup(t *testing.T) {
	DirOverride = t.TempDir()
	cache = map[string][]Memory{}
}

func TestMemoriesPerRobotAndPersistence(t *testing.T) {
	setup(t)
	if _, added, err := Add("AAAA1111", "Sam likes tea"); err != nil || !added {
		t.Fatal(added, err)
	}
	if _, added, _ := Add("aaaa1111", "sam LIKES tea"); added {
		t.Fatal("case-insensitive duplicate was added")
	}
	if p := PromptBlock("bbbb2222"); p != "" {
		t.Fatalf("robot B saw robot A memories: %q", p)
	}
	if p := PromptBlock("aaaa1111"); !strings.Contains(p, "Things you remember:\n- Sam likes tea") {
		t.Fatalf("robot A missing memory: %q", p)
	}
	// survives a cold cache (reload from disk)
	cache = map[string][]Memory{}
	if list, _ := List("aaaa1111"); len(list) != 1 {
		t.Fatalf("not persisted: %v", list)
	}
}

func TestMemoriesCapAndLimits(t *testing.T) {
	setup(t)
	for i := 0; i < MaxPerRobot+5; i++ {
		if _, _, err := Add("aaaa1111", "fact "+strings.Repeat("x", i)); err != nil {
			t.Fatal(err)
		}
	}
	list, _ := List("aaaa1111")
	if len(list) != MaxPerRobot || list[0].Text != "fact "+strings.Repeat("x", 5) {
		t.Fatalf("cap/oldest-drop wrong: %d %q", len(list), list[0].Text)
	}
	if _, _, err := Add("aaaa1111", strings.Repeat("y", MaxTextLen+1)); err != ErrTooLong {
		t.Fatal(err)
	}
	if _, _, err := Add("aaaa1111", "   "); err != ErrEmpty {
		t.Fatal(err)
	}
	if _, _, err := Add("../etc", "x"); err != ErrBadESN {
		t.Fatal(err)
	}
}

func TestMemoriesDeleteClear(t *testing.T) {
	setup(t)
	a, _, _ := Add("aaaa1111", "one")
	Add("aaaa1111", "two")
	if found, _ := Delete("aaaa1111", a.ID); !found {
		t.Fatal("not found")
	}
	if found, _ := Delete("aaaa1111", "nope"); found {
		t.Fatal("deleted nonexistent")
	}
	if list, _ := List("aaaa1111"); len(list) != 1 || list[0].Text != "two" {
		t.Fatal(list)
	}
	if err := Clear("aaaa1111"); err != nil {
		t.Fatal(err)
	}
	if list, _ := List("aaaa1111"); len(list) != 0 {
		t.Fatal(list)
	}
}
