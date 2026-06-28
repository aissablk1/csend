package main

import (
	"encoding/json"
	"testing"
)

func TestWantJSON(t *testing.T) {
	if !wantJSON([]string{"--peek", "--json"}) {
		t.Fatal("--json devrait être détecté")
	}
	if wantJSON([]string{"--peek"}) {
		t.Fatal("--json ne devrait pas être détecté")
	}
}

func TestJSONMessageShape(t *testing.T) {
	b, _ := json.Marshal(jsonMessage{From: "A", TS: "t", Body: "hi", Encrypted: true})
	var got map[string]any
	if err := json.Unmarshal(b, &got); err != nil {
		t.Fatal(err)
	}
	for _, k := range []string{"from", "ts", "body", "encrypted"} {
		if _, ok := got[k]; !ok {
			t.Fatalf("clé JSON %q manquante", k)
		}
	}
}
