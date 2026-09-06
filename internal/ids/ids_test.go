package ids

import (
	"strings"
	"testing"
)

func TestNewIsValidAndUnique(t *testing.T) {
	seen := map[string]bool{}
	for range 2000 {
		id := New()
		if !Valid(id) {
			t.Fatalf("New() = %q não passa em Valid", id)
		}
		if seen[id] {
			t.Fatalf("id repetido: %s", id)
		}
		seen[id] = true
	}
}

func TestValidRejectsOtherShapes(t *testing.T) {
	for _, bad := range []string{"", "abc", strings.Repeat("a", 15), strings.Repeat("A", 16), strings.Repeat("a", 15) + "1", strings.Repeat("a", 17), "abcdefghij23456-"} {
		if Valid(bad) {
			t.Errorf("Valid(%q) = true", bad)
		}
	}
}
