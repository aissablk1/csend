package main

// journal.go — `csend journal` : trace des messages livrés sur le bus, corps en
// HASH uniquement (jamais le clair). Prouve que le contenu est chiffré de bout en
// bout sans jamais le révéler — le « money-frame » de la démo cross-vendor :
//
//   de → à : sha256:…  (chiffré)
//
// Aucun corps en clair n'est imprimé : on hash le ciphertext scellé (ou, à défaut
// de contact E2E, le corps clair, marqué « clair » — honnêteté §29).

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
)

type journalEntry struct {
	TS    string `json:"ts"`
	From  string `json:"from"`
	To    string `json:"to"`
	Sum   string `json:"sha256"`
	State string `json:"state"` // "chiffré" | "clair"
}

// scanJournal parcourt toutes les boîtes (en attente + archivées .read) et renvoie
// une ligne par message, le corps réduit à un hash tronqué. Testable sans I/O console.
func scanJournal(root string) []journalEntry {
	var out []journalEntry
	dirs, _ := os.ReadDir(root)
	for _, d := range dirs {
		if !d.IsDir() {
			continue
		}
		for _, sub := range []string{".", ".read"} {
			base := filepath.Join(root, d.Name(), sub)
			files, _ := os.ReadDir(base)
			for _, f := range files {
				if f.IsDir() || filepath.Ext(f.Name()) != ".json" {
					continue
				}
				data, err := os.ReadFile(filepath.Join(base, f.Name()))
				if err != nil {
					continue
				}
				var m InboxMessage
				if json.Unmarshal(data, &m) != nil {
					continue
				}
				raw, state := []byte(m.Body), "clair"
				if m.Sealed != nil {
					raw, state = m.Sealed.Ct, "chiffré" // on hash le ciphertext, jamais le clair
				}
				sum := sha256.Sum256(raw)
				out = append(out, journalEntry{m.TS, m.From, m.To, hex.EncodeToString(sum[:])[:16], state})
			}
		}
	}
	sort.Slice(out, func(i, j int) bool {
		if out[i].TS != out[j].TS {
			return out[i].TS < out[j].TS
		}
		return out[i].From < out[j].From
	})
	return out
}

func cmdJournal(args []string) {
	out := scanJournal(mustStore().Inbox().Root)
	if wantJSON(args) {
		emitJSON(out)
		return
	}
	if len(out) == 0 {
		fmt.Println("(aucun message sur le bus)")
		return
	}
	fmt.Printf("Journal du bus — %d message(s), corps en HASH (jamais le clair) :\n", len(out))
	for _, e := range out {
		fmt.Printf("  %s  %s → %s : sha256:%s…  (%s)\n", e.TS, e.From, e.To, e.Sum, e.State)
	}
}
