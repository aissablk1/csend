package main

// backend.go — couche 3 (transports) abstraite : cmux ⟷ tmux interchangeables.
//
// Tout le moteur (engine.go) parle à un Backend via des dispatchers (listSurfaces,
// readScreen, sendText, sendKey, selfRef) ; le backend concret est choisi au
// runtime (CSEND_BACKEND, sinon auto-détection). Ajouter screen/ConPTY = un struct.

import (
	"errors"
	"os"
	"os/exec"
	"strings"
)

var errNoBackend = errors.New("aucun backend de terminal (cmux/tmux) disponible — " +
	"lance cmux ou tmux, ou utilise la voie coopérative (csend inbox/recv)")

// Backend is a surface-addressed terminal transport: it can enumerate live agent
// surfaces and read/inject into any of them.
type Backend interface {
	Name() string
	Available() bool
	ListSurfaces() ([]Surface, error)
	ReadScreen(ref, wsRef string, lines int) (string, error)
	SendText(ref, wsRef, text string) error
	SendKey(ref, wsRef, key string) error
	SelfRef() string
}

// allBackends is the preference order for auto-detection.
var allBackends = []Backend{cmuxBackend{}, tmuxBackend{}}

var cachedBackend Backend

func backend() Backend {
	if cachedBackend != nil {
		return cachedBackend
	}
	if name := os.Getenv("CSEND_BACKEND"); name != "" {
		for _, b := range allBackends {
			if b.Name() == name {
				cachedBackend = b
				return b
			}
		}
	}
	for _, b := range allBackends {
		if b.Available() {
			cachedBackend = b
			return b
		}
	}
	cachedBackend = unavailableBackend{}
	return cachedBackend
}

// backendAvailable reports whether any (or the forced) backend can be used. It does
// not cache, so it is safe to call before committing to a backend.
func backendAvailable() bool {
	if name := os.Getenv("CSEND_BACKEND"); name != "" {
		for _, b := range allBackends {
			if b.Name() == name {
				return b.Available()
			}
		}
		return false
	}
	for _, b := range allBackends {
		if b.Available() {
			return true
		}
	}
	return false
}

// --- dispatchers (the names engine.go calls) ---

func listSurfaces() ([]Surface, error)                       { return backend().ListSurfaces() }
func readScreen(ref, wsRef string, lines int) (string, error) { return backend().ReadScreen(ref, wsRef, lines) }
func sendText(ref, wsRef, text string) error                 { return backend().SendText(ref, wsRef, text) }
func sendKey(ref, wsRef, key string) error                   { return backend().SendKey(ref, wsRef, key) }
func selfRef() string                                        { return backend().SelfRef() }

// --- cmux backend (delegates to the existing socket implementation) ---

type cmuxBackend struct{}

func (cmuxBackend) Name() string                                              { return "cmux" }
func (cmuxBackend) Available() bool                                           { return cmuxAvailable() }
func (cmuxBackend) ListSurfaces() ([]Surface, error)                          { return cmuxListSurfaces() }
func (cmuxBackend) ReadScreen(ref, wsRef string, lines int) (string, error)   { return cmuxReadScreen(ref, wsRef, lines) }
func (cmuxBackend) SendText(ref, wsRef, text string) error                    { return cmuxSendText(ref, wsRef, text) }
func (cmuxBackend) SendKey(ref, wsRef, key string) error                      { return cmuxSendKey(ref, wsRef, key) }
func (cmuxBackend) SelfRef() string                                           { return cmuxSelfRef() }

// --- tmux backend (live injection via send-keys / capture-pane) ---

type tmuxBackend struct{}

func (tmuxBackend) Name() string { return "tmux" }

func (tmuxBackend) Available() bool {
	if _, err := exec.LookPath("tmux"); err != nil {
		return false
	}
	// A server must be running for panes to exist.
	return exec.Command("tmux", "list-panes", "-a").Run() == nil
}

func (tmuxBackend) ListSurfaces() ([]Surface, error) {
	out, err := exec.Command("tmux", "list-panes", "-a", "-F",
		"#{pane_id}\t#{session_name}\t#{pane_title}").Output()
	if err != nil {
		return nil, err
	}
	return parseTmuxPanes(string(out), os.Getenv("TMUX_PANE")), nil
}

func (tmuxBackend) ReadScreen(ref, wsRef string, lines int) (string, error) {
	out, err := exec.Command("tmux", "capture-pane", "-p", "-t", ref).Output()
	if err != nil {
		return "", err
	}
	return string(out), nil
}

func (tmuxBackend) SendText(ref, wsRef, text string) error {
	// -l = literal: tmux types the text verbatim, never interpreting it as keys.
	return exec.Command("tmux", "send-keys", "-t", ref, "-l", text).Run()
}

func (tmuxBackend) SendKey(ref, wsRef, key string) error {
	return exec.Command("tmux", "send-keys", "-t", ref, tmuxKey(key)).Run()
}

func (tmuxBackend) SelfRef() string { return os.Getenv("TMUX_PANE") }

// parseTmuxPanes turns `tmux list-panes -a -F '#{pane_id}\t#{session_name}\t#{pane_title}'`
// output into Surfaces. Pure (no exec) so it is unit-testable.
func parseTmuxPanes(out, self string) []Surface {
	var res []Surface
	for _, line := range strings.Split(strings.TrimRight(out, "\n"), "\n") {
		if line == "" {
			continue
		}
		f := strings.SplitN(line, "\t", 3)
		if len(f) < 2 {
			continue
		}
		title := ""
		if len(f) >= 3 {
			title = f[2]
		}
		res = append(res, Surface{
			Ref: f[0], Type: "terminal", Title: title,
			Workspace: f[1], WorkspaceRef: f[1], Here: f[0] == self,
		})
	}
	return res
}

// tmuxKey maps csend's lowercase key names to tmux key syntax. Pure/testable.
func tmuxKey(key string) string {
	switch strings.ToLower(key) {
	case "enter":
		return "Enter"
	case "escape", "esc":
		return "Escape"
	case "ctrl+c":
		return "C-c"
	default:
		return key
	}
}

// --- no-backend fallback ---

type unavailableBackend struct{}

func (unavailableBackend) Name() string                                            { return "none" }
func (unavailableBackend) Available() bool                                         { return false }
func (unavailableBackend) ListSurfaces() ([]Surface, error)                        { return nil, errNoBackend }
func (unavailableBackend) ReadScreen(ref, wsRef string, lines int) (string, error) { return "", errNoBackend }
func (unavailableBackend) SendText(ref, wsRef, text string) error                  { return errNoBackend }
func (unavailableBackend) SendKey(ref, wsRef, key string) error                    { return errNoBackend }
func (unavailableBackend) SelfRef() string                                         { return "" }
