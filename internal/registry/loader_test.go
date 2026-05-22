package registry

import (
	"testing"
)

func TestListMCPs_ReturnsAllBundled(t *testing.T) {
	l := New()
	entries := l.ListMCPs()
	if len(entries) == 0 {
		t.Fatal("expected bundled MCP entries, got none")
	}
	// Verify every entry has required fields.
	for _, e := range entries {
		if e.Name == "" {
			t.Errorf("entry has empty name: %+v", e)
		}
		if e.Command == "" {
			t.Errorf("entry %q has empty command", e.Name)
		}
	}
}

func TestGetMCP_Found(t *testing.T) {
	l := New()
	entry, err := l.GetMCP("supabase")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if entry.Name != "supabase" {
		t.Errorf("got name %q, want %q", entry.Name, "supabase")
	}
	if entry.Auth.EnvVar != "SUPABASE_ACCESS_TOKEN" {
		t.Errorf("got envVar %q, want SUPABASE_ACCESS_TOKEN", entry.Auth.EnvVar)
	}
}

func TestGetMCP_NotFound_WithSuggestion(t *testing.T) {
	l := New()
	_, err := l.GetMCP("supabse") // typo
	if err == nil {
		t.Fatal("expected error for unknown MCP, got nil")
	}
	nfe, ok := err.(*NotFoundError)
	if !ok {
		t.Fatalf("expected *NotFoundError, got %T", err)
	}
	if len(nfe.Suggestions) == 0 {
		t.Error("expected at least one suggestion for typo 'supabse'")
	}
	found := false
	for _, s := range nfe.Suggestions {
		if s == "supabase" {
			found = true
		}
	}
	if !found {
		t.Errorf("expected 'supabase' in suggestions, got %v", nfe.Suggestions)
	}
}

func TestGetMCP_NoAuthMCP(t *testing.T) {
	l := New()
	entry, err := l.GetMCP("filesystem")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if entry.Auth.EnvVar != "" {
		t.Errorf("filesystem should have empty envVar, got %q", entry.Auth.EnvVar)
	}
}

func TestLevenshtein(t *testing.T) {
	cases := []struct {
		a, b string
		want int
	}{
		{"supabase", "supabase", 0},
		{"supabse", "supabase", 1},
		{"github", "gitlab", 2},
		{"abc", "", 3},
		{"", "abc", 3},
	}
	for _, c := range cases {
		got := levenshtein(c.a, c.b)
		if got != c.want {
			t.Errorf("levenshtein(%q, %q) = %d, want %d", c.a, c.b, got, c.want)
		}
	}
}

func TestListSkills_ReturnsValidEntries(t *testing.T) {
	l := New()
	entries := l.ListSkills()
	// Skills registry may be empty (built-in skills removed in favour of skills.sh).
	// Validate shape of any entries that do exist.
	for _, e := range entries {
		if e.Name == "" {
			t.Errorf("skill entry has empty name: %+v", e)
		}
		if e.Latest == "" {
			t.Errorf("skill %q has empty latest version", e.Name)
		}
	}
}
