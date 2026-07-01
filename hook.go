package main

// hook.go — RÉCEPTION LIVE : faire qu'une session reçoive ses messages sans polling.
//
// `csend hook` est conçu pour être câblé comme hook `UserPromptSubmit` de Claude
// Code : à chaque prompt, il draine l'inbox de la session et imprime les messages
// → ils surgissent dans le contexte de l'agent. C'est ce qui transforme la
// plomberie en conversation. `csend watch` est l'équivalent « tail live » pour un
// humain / un pane.

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"time"
)

const hookInstall = `Câble csend comme hook de réception. Dans ~/.claude/settings.json :

  "hooks": {
    "UserPromptSubmit": [
      { "hooks": [ { "type": "command", "command": "csend hook" } ] }
    ]
  }

→ à chaque prompt, les messages reçus via le bus inter-agents apparaissent dans le
  contexte de la session. (L'identité est dérivée du session_id du hook ; pose
  CSEND_AGENT_ID pour la forcer.)`

// hookInstallFor rend le snippet de câblage adapté à l'éditeur. Pur affichage (aucune
// écriture hors workspace, §5) — l'utilisateur applique le snippet lui-même. Les formats
// de hook évoluent : on renvoie vers la doc de chaque CLI plutôt que d'affirmer (§29).
func hookInstallFor(provider string) string {
	switch provider {
	case "codex":
		return `Câble csend dans Codex CLI (~/.codex/ — vérifie la doc Codex, le format évolue) :
  hook UserPromptSubmit (et Stop pour l'async) → commande « csend hook ».
  IMPORTANT : Codex exige d'APPROUVER le hook (commande /hooks, ou
  --dangerously-bypass-hook-trust) sinon le message n'arrive jamais.
  Au besoin : « csend hook --provider codex » si l'auto-détection échoue.`
	case "gemini":
		return `Câble csend dans Gemini CLI (~/.gemini/settings.json — vérifie la doc Gemini) :
  event BeforeAgent → commande « csend hook »
  (Gemini reçoit le contexte additionnel via le stdout brut de csend).`
	default:
		return hookInstall // Claude (UserPromptSubmit, hookSpecificOutput)
	}
}

// hookPayload est le sous-ensemble du JSON que les CLIs passent au hook sur stdin
// (Claude/Codex : UserPromptSubmit ; Gemini : BeforeAgent). On ne l'EXIGE pas :
// absent → on retombe sur CSEND_AGENT_ID / l'identité courante (rétro-compat).
type hookPayload struct {
	HookEventName string `json:"hook_event_name"`
	SessionID     string `json:"session_id"`
	Cwd           string `json:"cwd"`
}

// hookIdentity dérive une identité d'agent STABLE depuis le payload (session_id),
// pour que la réception « marche toute seule » sans poser CSEND_AGENT_ID à la main.
// C'est LE gap « zéro-config » du chemin hook.
func hookIdentity(p hookPayload) string {
	if env := os.Getenv("CSEND_AGENT_ID"); env != "" {
		return env
	}
	if p.SessionID != "" {
		base := p.SessionID
		if len(base) > 8 {
			base = base[:8]
		}
		return "sess-" + base
	}
	return selfAgentID()
}

// readHookStdin lit (sans bloquer sur un terminal interactif) le payload du hook.
func readHookStdin() hookPayload {
	var p hookPayload
	if fi, _ := os.Stdin.Stat(); fi != nil && (fi.Mode()&os.ModeCharDevice) == 0 {
		if data, err := io.ReadAll(io.LimitReader(os.Stdin, 1<<20)); err == nil && len(data) > 0 {
			_ = json.Unmarshal(data, &p)
		}
	}
	return p
}

// cmdHook drains this session's inbox and emits the messages as the agent's added
// context (consuming them so each surfaces once). Prints nothing when empty — zéro
// bruit dans le hook. Provider-aware : émet la forme attendue par chaque CLI.
func cmdHook(args []string) {
	forceProvider := ""
	for i, a := range args {
		if a == "--install" {
			prov := "claude"
			if i+1 < len(args) {
				prov = args[i+1]
			}
			fmt.Println(hookInstallFor(prov))
			return
		}
		if a == "--provider" && i+1 < len(args) {
			forceProvider = args[i+1]
		}
	}
	p := readHookStdin()
	if forceProvider != "" { // override l'auto-détection quand le CLI ne passe pas hook_event_name
		if forceProvider == "gemini" {
			p.HookEventName = "BeforeAgent"
		} else {
			p.HookEventName = "UserPromptSubmit"
		}
	}
	s := mustStore()
	agent := hookIdentity(p)
	msgs, err := s.Inbox().Receive(agent, true)
	if err != nil || len(msgs) == 0 {
		return
	}
	text := fmt.Sprintf("[csend] %d message(s) reçus via le bus inter-agents :\n", len(msgs))
	for _, m := range msgs {
		who := m.From
		if m.Provider != "" {
			who = m.Provider + ":" + m.From // cross-vendor visible dans le contexte injecté
		}
		text += fmt.Sprintf("  • de %s : %s\n", who, openBody(s, m))
	}
	switch p.HookEventName {
	case "BeforeAgent": // Gemini : contexte additionnel via stdout brut
		fmt.Print(text)
	default: // Claude / Codex : hookSpecificOutput.additionalContext
		b, _ := json.Marshal(map[string]any{
			"hookSpecificOutput": map[string]any{
				"hookEventName":     "UserPromptSubmit",
				"additionalContext": text,
			},
		})
		fmt.Println(string(b))
	}
}

// cmdWatch tails the inbox live, printing new messages as they arrive.
func cmdWatch(args []string) {
	interval := 2 * time.Second
	for i := 0; i < len(args); i++ {
		if args[i] == "--interval" && i+1 < len(args) {
			var n int
			if _, e := fmt.Sscanf(args[i+1], "%d", &n); e == nil && n > 0 {
				interval = time.Duration(n) * time.Second
			}
			i++
		}
	}
	s := mustStore()
	agent := selfAgentID()
	fmt.Printf("csend watch — inbox de %s (Ctrl+C pour arrêter)\n", agent)
	for {
		if msgs, _ := s.Inbox().Receive(agent, true); len(msgs) > 0 {
			for _, m := range msgs {
				fmt.Printf("[%s] %s : %s\n", m.TS, m.From, openBody(s, m))
			}
		}
		time.Sleep(interval)
	}
}
