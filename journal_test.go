package main

import (
	"path/filepath"
	"testing"
)

// Le journal trace les messages SANS jamais imprimer le clair : on vérifie qu'un
// message scellé ressort « chiffré » (hash du ciphertext) et qu'un message clair
// ressort « clair », les deux réduits à un hash — le corps n'apparaît jamais.
func TestScanJournalHashesNeverLeaksPlaintext(t *testing.T) {
	dir := t.TempDir()
	ib := &Inbox{Root: filepath.Join(dir, "inbox")}
	if err := ib.Deliver(InboxMessage{ID: "1", TS: "2026-06-30T10:00:00Z", From: "gemini-rev", To: "claude-dev", Body: "secret en clair"}); err != nil {
		t.Fatal(err)
	}
	if err := ib.Deliver(InboxMessage{ID: "2", TS: "2026-06-30T10:00:01Z", From: "gemini-rev", To: "claude-dev", Sealed: &SealedMessage{Ct: []byte("CIPHERTEXT-OPAQUE")}}); err != nil {
		t.Fatal(err)
	}
	entries := scanJournal(ib.Root)
	if len(entries) != 2 {
		t.Fatalf("attendu 2 entrées, obtenu %d", len(entries))
	}
	if entries[0].State != "clair" || entries[1].State != "chiffré" {
		t.Fatalf("états attendus clair/chiffré, obtenu %q/%q", entries[0].State, entries[1].State)
	}
	for _, e := range entries {
		if len(e.Sum) != 16 {
			t.Fatalf("hash tronqué invalide: %q", e.Sum)
		}
		if e.From != "gemini-rev" || e.To != "claude-dev" {
			t.Fatalf("métadonnées corrompues: %+v", e)
		}
	}
}
