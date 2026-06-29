package main

// key.go — envoyer une TOUCHE brute à une autre surface (sans taper de texte) :
// valider (enter), effacer une saisie (ctrl+c / escape), interrompre (ctrl+c)…
// Primitive utile pour piloter finement une session (ex. vider un brouillon avant
// d'envoyer un message propre).

import (
	"fmt"
	"strings"
)

func cmdKey(args []string) {
	if len(args) < 2 {
		fail("usage: csend key <cible> <touche>   (enter | escape | ctrl+c | ctrl+u …)")
	}
	tgt, err := resolveTarget(args[0])
	if err != nil {
		fail(err.Error())
	}
	if isSelf(tgt, selfRef()) {
		fail("refus: " + tgt.Ref + " est la session courante (pas d'auto-pilotage)")
	}
	key := strings.ToLower(args[1])
	if err := sendKey(tgt.Ref, tgt.WorkspaceRef, key); err != nil {
		fail(err.Error())
	}
	fmt.Printf("✓ touche « %s » envoyée → %s (%s)\n", key, tgt.Ref, tgt.Workspace)
}
