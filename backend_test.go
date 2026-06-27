package main

import "testing"

func TestParseTmuxPanes(t *testing.T) {
	out := "%0\tdev\tclaude — SACEM\n%3\tbuild\tcodex backend\n\n%7\tdev\t\n"
	got := parseTmuxPanes(out, "%3")
	if len(got) != 3 {
		t.Fatalf("got %d panes, want 3", len(got))
	}
	if got[0].Ref != "%0" || got[0].Workspace != "dev" || got[0].Title != "claude — SACEM" {
		t.Fatalf("pane 0 mismatch: %+v", got[0])
	}
	if !got[1].Here {
		t.Fatalf("pane %%3 should be Here (self)")
	}
	if got[0].Here || got[2].Here {
		t.Fatalf("only %%3 should be Here")
	}
	if got[2].Title != "" {
		t.Fatalf("empty title should stay empty, got %q", got[2].Title)
	}
	for _, s := range got {
		if s.Type != "terminal" {
			t.Fatalf("tmux panes are terminals, got %q", s.Type)
		}
	}
}

func TestTmuxKeyMapping(t *testing.T) {
	cases := map[string]string{
		"enter": "Enter", "Enter": "Enter", "escape": "Escape", "esc": "Escape",
		"ctrl+c": "C-c", "Tab": "Tab", // unknown passes through
	}
	for in, want := range cases {
		if got := tmuxKey(in); got != want {
			t.Fatalf("tmuxKey(%q) = %q, want %q", in, got, want)
		}
	}
}

func TestBackendsImplementInterface(t *testing.T) {
	// Compile-time-ish guarantee that every backend satisfies the contract, and
	// that the no-backend fallback errors instead of panicking.
	var bs = []Backend{cmuxBackend{}, tmuxBackend{}, unavailableBackend{}}
	for _, b := range bs {
		if b.Name() == "" {
			t.Fatalf("backend has empty name")
		}
	}
	if _, err := (unavailableBackend{}).ListSurfaces(); err == nil {
		t.Fatal("unavailable backend must error on ListSurfaces")
	}
}
