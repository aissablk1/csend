package main

// register.go — faire de TOUTE session un participant first-class du bus.
//
// `csend list` montre les surfaces d'un multiplexeur (cmux/tmux). Mais une session
// bash / Terminal / Windows / Codex / Gemini qui parle au bus par la voie
// COOPÉRATIVE n'a pas de surface. `csend register` l'inscrit dans le registre
// persistant → elle devient visible (`csend agents`) et adressable (`csend inbox
// <id>`), quel que soit le terminal, le provider ou l'OS.

import (
	"fmt"
	"os"
	"path/filepath"
	"text/tabwriter"
)

func providerSuffix(p string) string {
	if p == "" {
		return ""
	}
	return " (" + p + ")"
}

// registerSelf records this session into the persistent registry and ensures its
// inbox exists. Testable without any CLI/terminal.
func registerSelf(s *Store, id, provider, workspace string) error {
	if err := s.UpsertSession(SessionRecord{
		SessionID: id, Provider: provider, Workspace: workspace, State: "registered",
	}); err != nil {
		return err
	}
	return os.MkdirAll(filepath.Join(s.Inbox().Root, safeID(id)), 0o755)
}

func cmdRegister(args []string) {
	id, provider, workspace := selfAgentID(), "", ""
	for i := 0; i < len(args); i++ {
		switch args[i] {
		case "--id":
			if i+1 < len(args) {
				id = args[i+1]
				i++
			}
		case "--provider":
			if i+1 < len(args) {
				provider = args[i+1]
				i++
			}
		case "--workspace":
			if i+1 < len(args) {
				workspace = args[i+1]
				i++
			}
		}
	}
	if workspace == "" {
		if wd, err := os.Getwd(); err == nil {
			workspace = filepath.Base(wd)
		}
	}
	s := mustStore()
	if err := registerSelf(s, id, provider, workspace); err != nil {
		fail(err.Error())
	}
	fmt.Printf("✓ session enregistrée sur le bus : %s%s\n", id, providerSuffix(provider))
	fmt.Printf("  Adressable par : csend inbox %s \"…\"   (et csend recv pour relever)\n", id)
}

func cmdWhoami(args []string) {
	s := mustStore()
	id := selfAgentID()
	if wantJSON(args) {
		fp := ""
		if b, ok := loadPublicBundle(s); ok {
			fp = fingerprint(b)
		}
		n, _ := s.Inbox().Pending(id)
		emitJSON(map[string]any{"agent": id, "fingerprint": fp, "inbox_pending": n})
		return
	}
	fmt.Printf("agent-id : %s\n", id)
	if b, ok := loadPublicBundle(s); ok {
		fmt.Printf("identité : %s\n", fingerprint(b))
	} else {
		fmt.Println("identité : aucune (csend id --create)")
	}
	if n, _ := s.Inbox().Pending(id); n > 0 {
		fmt.Printf("inbox    : %d message(s) en attente — csend recv\n", n)
	} else {
		fmt.Println("inbox    : vide")
	}
}

// cmdAgents shows the cooperative fleet (registry) — agents on ANY terminal /
// provider / OS — complementing `csend list` (live multiplexer surfaces only).
func cmdAgents(args []string) {
	s := mustStore()
	recs, err := s.ListSessions()
	if err != nil {
		fail(err.Error())
	}
	if wantJSON(args) {
		type ja struct {
			Agent     string `json:"agent"`
			Provider  string `json:"provider"`
			Workspace string `json:"workspace"`
			LastSeen  string `json:"last_seen"`
			Inbox     int    `json:"inbox_pending"`
		}
		arr := make([]ja, 0, len(recs))
		for _, r := range recs {
			n, _ := s.Inbox().Pending(r.SessionID)
			arr = append(arr, ja{Agent: r.SessionID, Provider: r.Provider, Workspace: r.Workspace, LastSeen: r.LastSeen, Inbox: n})
		}
		emitJSON(map[string]any{"count": len(arr), "agents": arr})
		return
	}
	if len(recs) == 0 {
		fmt.Println("Aucun agent coopératif enregistré. Dans une session : csend register")
		return
	}
	fmt.Println("Agents coopératifs du bus (tous terminaux / providers / OS) :")
	w := tabwriter.NewWriter(os.Stdout, 0, 2, 2, ' ', 0)
	fmt.Fprintln(w, "AGENT\tPROVIDER\tWORKSPACE\tINBOX\tVU LE")
	for _, r := range recs {
		n, _ := s.Inbox().Pending(r.SessionID)
		prov := r.Provider
		if prov == "" {
			prov = "—"
		}
		fmt.Fprintf(w, "%s\t%s\t%s\t%d\t%s\n", r.SessionID, prov, r.Workspace, n, r.LastSeen)
	}
	w.Flush()
}
