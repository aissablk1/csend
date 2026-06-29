package main

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// SubmitMode controls how hard csend pushes a message in.
type SubmitMode int

const (
	// ModeAuto (default, "intelligent"): submit only on a confident Idle; stage
	// (type without Enter) when Busy; refuse on AwaitConfirm/Unknown.
	ModeAuto SubmitMode = iota
	// ModeStage: type the text, never press Enter (still refuses confirm/unknown).
	ModeStage
	// ModeSend: submit even when Busy (still refuses confirm/unknown).
	ModeSend
	// ModeForce: override every guard — type and Enter no matter the state.
	ModeForce
)

// Session is a discovered agent surface plus its detected state.
type Session struct {
	Surface
	SessionID string
	State     State
}

// sessionKey is the stable identity of a session for the family graph: its agent
// session-id when the tab title carries one, else the (less stable) surface ref.
func sessionKey(s Surface) string {
	if sid := reSessionID.FindString(s.Title); sid != "" {
		return sid
	}
	return s.Ref
}

// Outcome is the result of a delivery, for reporting and auditing.
type Outcome struct {
	Ref       string
	Workspace string
	State     State
	Action    string // "submitted" | "staged" | "refused"
	Reason    string
}

// listSessions returns terminal surfaces that look like live agent sessions
// (state detectable), each annotated with its detected state. Read-only.
func listSessions() ([]Session, error) {
	surfaces, err := listSurfaces()
	if err != nil {
		return nil, err
	}
	var sessions []Session
	for _, s := range surfaces {
		if s.Type != "terminal" {
			continue
		}
		st := StateUnknown
		if screen, err := readScreen(s.Ref, s.WorkspaceRef, stateTailLines); err != nil {
			st = StateUnreachable // socket read failed (rare: app/socket down)
		} else {
			st, _ = DetectAnyState(screen) // provider-pluggable (couche 2)
		}
		sid := reSessionID.FindString(s.Title)
		// Keep recognized live agent states, or any surface carrying a session-id
		// (an agent we know about even if currently unreachable). Skip plain shells.
		isAgentState := st == StateIdle || st == StateBusy || st == StateAwaitConfirm
		if isAgentState || sid != "" {
			sessions = append(sessions, Session{Surface: s, SessionID: sid, State: st})
		}
	}
	return sessions, nil
}

// resolveTarget maps a user handle to a single surface ref. Accepted handles:
// an explicit ref ("surface:42"), a workspace name ("SACEM"), a session-id or
// any unique substring of a tab title.
func resolveTarget(handle string) (Surface, error) {
	surfaces, err := listSurfaces()
	if err != nil {
		return Surface{}, err
	}
	// Explicit surface ref wins.
	for _, s := range surfaces {
		if s.Ref == handle {
			return s, nil
		}
	}
	h := strings.ToLower(handle)
	var matches []Surface
	for _, s := range surfaces {
		if s.Type != "terminal" {
			continue
		}
		if strings.EqualFold(s.Workspace, handle) ||
			strings.Contains(strings.ToLower(s.Title), h) ||
			strings.Contains(strings.ToLower(s.Ref), h) {
			matches = append(matches, s)
		}
	}
	switch len(matches) {
	case 0:
		return Surface{}, fmt.Errorf("aucune session ne correspond à %q (essaie `csend list`)", handle)
	case 1:
		return matches[0], nil
	default:
		var refs []string
		for _, m := range matches {
			refs = append(refs, fmt.Sprintf("%s (%s)", m.Ref, m.Workspace))
		}
		return Surface{}, fmt.Errorf("%q est ambigu — précise: %s", handle, strings.Join(refs, ", "))
	}
}

// sendMessage resolves a handle then delivers (state-aware guards + audit).
func sendMessage(handle, text string, mode SubmitMode) (Outcome, error) {
	tgt, err := resolveTarget(handle)
	if err != nil {
		return Outcome{}, err
	}
	return deliverTo(tgt, text, mode)
}

// deliverTo applies the guarded, state-aware delivery to an already-resolved
// surface. Shared by single sends and family broadcasts.
func deliverTo(tgt Surface, text string, mode SubmitMode) (Outcome, error) {
	// Never inject into our own session (Here flag + authoritative ref cross-check).
	if isSelf(tgt, selfRef()) {
		return Outcome{}, fmt.Errorf("refus: %s est la session courante (pas d'auto-injection)", tgt.Ref)
	}

	screen, readErr := readScreenRetry(tgt.Ref, tgt.WorkspaceRef, 8)
	state := StateUnreachable
	if readErr == nil {
		state, _ = DetectAnyState(screen)
	}
	out := Outcome{Ref: tgt.Ref, Workspace: tgt.Workspace, State: state}

	// Guard: only a confident Idle (submitted) or Busy (staged) is safely driven.
	// A confirmation dialog, an unrecognized prompt, or an unreadable/unmaterialized
	// surface is refused unless the caller passes --force.
	if mode != ModeForce && state != StateIdle && state != StateBusy {
		out.Action = "refused"
		out.Reason = fmt.Sprintf("état %s — utilise --force pour passer outre", state)
		audit(tgt, text, out)
		return out, nil
	}

	// Deliver the text (no newline embedded).
	if err := sendText(tgt.Ref, tgt.WorkspaceRef, text); err != nil {
		return Outcome{}, err
	}

	submit := false
	switch {
	case mode == ModeForce, mode == ModeSend:
		submit = true
	case mode == ModeStage:
		submit = false
	case mode == ModeAuto:
		submit = state == StateIdle // busy → staged, not submitted
	}

	if submit {
		// Évite la course de collage : un gros texte suivi d'Entrée immédiate laisse
		// le message TAPÉ mais NON soumis (Claude Code en mode paste traite Entrée
		// comme un saut de ligne). On laisse la saisie se poser, on valide, puis on
		// VÉRIFIE réellement — au lieu d'annoncer « submitted » à l'aveugle.
		time.Sleep(250 * time.Millisecond)
		_ = sendKey(tgt.Ref, tgt.WorkspaceRef, "enter")
		if submitConfirmed(tgt) {
			out.Action = "submitted"
		} else {
			time.Sleep(400 * time.Millisecond)
			_ = sendKey(tgt.Ref, tgt.WorkspaceRef, "enter") // 2e tentative
			if submitConfirmed(tgt) {
				out.Action = "submitted"
			} else {
				out.Action = "staged"
				out.Reason = "Entrée n'a pas validé (saisie restée en place) — déposé, à valider"
			}
		}
	} else {
		out.Action = "staged"
		if state == StateBusy {
			out.Reason = "cible occupée — déposé sans valider"
		}
	}
	audit(tgt, text, out)
	return out, nil
}

// submitConfirmed re-reads the target after an Enter and reports whether the input
// actually cleared (empty idle prompt) or the agent's turn started (busy). If our
// text is still sitting in the input box, the submission did not take (paste race).
func submitConfirmed(tgt Surface) bool {
	time.Sleep(250 * time.Millisecond)
	screen, err := readScreen(tgt.Ref, tgt.WorkspaceRef, stateTailLines)
	if err != nil {
		return false
	}
	st, _ := DetectAnyState(screen)
	return st == StateIdle || st == StateBusy
}

// audit appends an append-only JSONL record. We log a content HASH + length,
// never the raw message body (§35 PII): an audit trail without leaking text.
func audit(tgt Surface, text string, out Outcome) {
	sum := sha256.Sum256([]byte(text))
	rec := map[string]any{
		"ts":        time.Now().UTC().Format(time.RFC3339),
		"from":      selfRef(),
		"to":        tgt.Ref,
		"workspace": tgt.Workspace,
		"state":     out.State.String(),
		"action":    out.Action,
		"reason":    out.Reason,
		"text_sha":  fmt.Sprintf("%x", sum[:6]),
		"text_len":  len(text),
	}
	line, _ := json.Marshal(rec)
	path := auditPath()
	_ = os.MkdirAll(filepath.Dir(path), 0o755)
	f, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o600)
	if err != nil {
		return // auditing must never block delivery
	}
	defer f.Close()
	_, _ = f.Write(append(line, '\n'))
}

func auditPath() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".claude", "logs", "csend.jsonl")
}

// --- family broadcast: resolve a set of targets up/down the relation graph ---

// Direction selects which relatives of a base session a broadcast addresses.
type Direction int

const (
	DirParent Direction = iota
	DirChildren
	DirSiblings
	DirDescendants
)

func (d Direction) String() string {
	switch d {
	case DirParent:
		return "parent"
	case DirChildren:
		return "enfants"
	case DirSiblings:
		return "frères"
	case DirDescendants:
		return "descendants"
	default:
		return "?"
	}
}

// selfKey returns the family-graph key of the calling session.
func selfKey() (string, error) {
	self := selfRef()
	if self == "" {
		return "", fmt.Errorf("session courante inconnue (hors cmux ?) — utilise --from <session>")
	}
	surfaces, err := listSurfaces()
	if err != nil {
		return "", err
	}
	for _, s := range surfaces {
		if s.Ref == self {
			return sessionKey(s), nil
		}
	}
	return self, nil
}

// resolveFamily returns the LIVE target surfaces in the given direction relative
// to baseKey, excluding self and offline nodes.
func resolveFamily(baseKey string, dir Direction) ([]Surface, error) {
	surfaces, err := listSurfaces()
	if err != nil {
		return nil, err
	}
	byKey := map[string]Surface{}
	for _, s := range surfaces {
		if s.Type == "terminal" {
			byKey[sessionKey(s)] = s
		}
	}
	rel := loadRelations()

	var keys []string
	switch dir {
	case DirParent:
		if p, ok := rel.parentOf(baseKey); ok {
			keys = []string{p}
		}
	case DirChildren:
		keys = rel.childrenOf(baseKey)
	case DirSiblings:
		if p, ok := rel.parentOf(baseKey); ok {
			for _, c := range rel.childrenOf(p) {
				if c != baseKey {
					keys = append(keys, c)
				}
			}
		}
	case DirDescendants:
		keys = descendants(rel, baseKey)
	}

	self := selfRef()
	seen := map[string]bool{}
	var out []Surface
	for _, k := range keys {
		if seen[k] {
			continue
		}
		seen[k] = true
		s, ok := byKey[k]
		if !ok || isSelf(s, self) { // offline or self
			continue
		}
		out = append(out, s)
	}
	return out, nil
}

// descendants returns every node in the subtree under base (excluding base).
func descendants(rel *Relations, base string) []string {
	var out []string
	seen := map[string]bool{}
	var dfs func(string)
	dfs = func(id string) {
		for _, c := range rel.childrenOf(id) {
			if seen[c] {
				continue
			}
			seen[c] = true
			out = append(out, c)
			dfs(c)
		}
	}
	dfs(base)
	return out
}
