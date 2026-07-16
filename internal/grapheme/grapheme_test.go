package grapheme

import (
	"strings"
	"testing"
)

func TestProtectRoundTripsTextAndANSI(t *testing.T) {
	tests := []struct {
		name string
		view string
	}{
		{name: "plain", view: "plain text"},
		{name: "combining", view: "Cafe\u0301"},
		{name: "styled combining", view: "\x1b[31me\u0301\x1b[0m"},
		{name: "wide combining", view: "界\u0301"},
		{name: "emoji", view: "👩🏽‍💻"},
		{name: "existing sentinel", view: "\ue000 e\u0301"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			protected, replacements := Protect(tt.view)
			if got := restore(protected, replacements); got != tt.view {
				t.Fatalf("restored view = %q, want %q", got, tt.view)
			}
			for marker := range replacements {
				if strings.Contains(tt.view, marker) {
					t.Fatalf("allocated marker %q already occurs in source", marker)
				}
			}
		})
	}
}

func TestProtectManySharesUniqueSentinels(t *testing.T) {
	views := []string{"\ue000 e\u0301", "a\u0308", "plain"}
	protected, replacements := ProtectMany(views...)

	if len(protected) != len(views) {
		t.Fatalf("protected view count = %d, want %d", len(protected), len(views))
	}
	if len(replacements) != 2 {
		t.Fatalf("replacement count = %d, want 2", len(replacements))
	}
	for i := range views {
		if got := restore(protected[i], replacements); got != views[i] {
			t.Errorf("restored view %d = %q, want %q", i, got, views[i])
		}
	}
	for marker := range replacements {
		for _, view := range views {
			if strings.Contains(view, marker) {
				t.Fatalf("allocated marker %q already occurs in shared source", marker)
			}
		}
	}
}

func restore(view string, replacements map[string]string) string {
	pairs := make([]string, 0, len(replacements)*2)
	for marker, original := range replacements {
		pairs = append(pairs, marker, original)
	}
	return strings.NewReplacer(pairs...).Replace(view)
}
