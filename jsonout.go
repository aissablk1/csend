package main

// jsonout.go — sortie machine-lisible (`--json`) pour piloter csend depuis un
// AGENT (le besoin central : un agent ne peut pas parser de la prose). Les
// commandes de lecture (list/recv/agents) émettent un JSON stable quand `--json`
// est présent.

import (
	"encoding/json"
	"fmt"
)

// wantJSON reports whether --json was passed.
func wantJSON(args []string) bool {
	for _, a := range args {
		if a == "--json" {
			return true
		}
	}
	return false
}

// emitJSON prints v as indented JSON on stdout.
func emitJSON(v any) {
	b, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		fail("json: " + err.Error())
	}
	fmt.Println(string(b))
}

// jsonMessage is the stable shape of an inbox message in --json output.
type jsonMessage struct {
	From      string `json:"from"`
	TS        string `json:"ts"`
	Body      string `json:"body"`
	Encrypted bool   `json:"encrypted"`
}
