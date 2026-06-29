package main

import (
	"encoding/hex"
	"encoding/json"
	"testing"
)

// FuzzFrameParse fuzzes the network frame parser: arbitrary bytes must never panic
// the same code path handleBusConn runs (JSON decode + dedup-key derivation). Seeds
// run under plain `go test`; `go test -fuzz=FuzzFrameParse` explores further.
func FuzzFrameParse(f *testing.F) {
	f.Add([]byte(`{"to":"B","body":"x"}`))
	f.Add([]byte(`{"id":"1","sealed":{"nonce":"AAAA","ct":"BB"}}`))
	f.Add([]byte(`{"sealed":{}}`))
	f.Add([]byte(`not json at all`))
	f.Add([]byte(``))
	f.Add([]byte(`{"to":`))

	f.Fuzz(func(t *testing.T, data []byte) {
		var m InboxMessage
		_ = json.Unmarshal(data, &m) // ne doit jamais paniquer
		// même dérivation de clé de dédup que handleBusConn
		if m.Sealed != nil && len(m.Sealed.Nonce) > 0 {
			_ = "n:" + hex.EncodeToString(m.Sealed.Nonce)
		}
	})
}
