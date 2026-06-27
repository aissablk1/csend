package main

import (
	"encoding/base64"
	"encoding/json"
	"os"
	"os/exec"
	"regexp"
	"time"
)

// cmux.go — transport BACKEND for the cmux / Vibe Island app.
//
// Topology (the workspace/surface tree) is read via `cmux tree --json`. Surface
// I/O (read screen, send text, send key) goes straight to the cmux Unix socket
// as JSON-RPC (see socket.go), passing each surface's REAL workspace_id — which
// is what makes cross-workspace messaging work with zero flicker.
//
// This is one backend behind a surface-addressed contract; tmux / AppleScript /
// ConPTY backends will sit beside it. Nothing here is Claude-specific — provider
// detection lives in state.go.

var reSessionID = regexp.MustCompile(`[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{2,4}`)

// Surface is one addressable terminal/browser pane in cmux.
type Surface struct {
	Ref          string // e.g. "surface:42"
	Type         string // "terminal" | "browser"
	Title        string // tab title (carries the agent session-id for CLIs)
	Workspace    string // human workspace title, e.g. "SACEM"
	WorkspaceRef string // workspace ref, e.g. "workspace:2" — REQUIRED for socket I/O
	Here         bool   // the surface this csend process runs in (never a target)
}

// cmuxAvailable reports whether the backend can be used: the `cmux` CLI (for
// topology) and a reachable socket (for I/O).
func cmuxAvailable() bool {
	if _, err := exec.LookPath("cmux"); err != nil {
		return false
	}
	_, err := os.Stat(socketPath())
	return err == nil
}

// --- topology via `cmux tree --json` ---

type treeJSON struct {
	Caller struct {
		SurfaceRef string `json:"surface_ref"`
	} `json:"caller"`
	Windows []struct {
		Workspaces []struct {
			Ref   string `json:"ref"`
			Title string `json:"title"`
			Panes []struct {
				Surfaces []struct {
					Ref   string `json:"ref"`
					Type  string `json:"type"`
					Title string `json:"title"`
					Here  bool   `json:"here"`
				} `json:"surfaces"`
			} `json:"panes"`
		} `json:"workspaces"`
	} `json:"windows"`
}

func loadTree() (*treeJSON, error) {
	out, err := exec.Command("cmux", "tree", "--json").Output()
	if err != nil {
		return nil, err
	}
	var t treeJSON
	if err := json.Unmarshal(out, &t); err != nil {
		return nil, err
	}
	return &t, nil
}

// callerRef returns the calling surface's ref in the SAME "surface:N" space as
// every target ref. The tree's caller.surface_ref is AUTHORITATIVE; CMUX_SURFACE_ID
// is only a fallback because some cmux builds set it to a UUID — which would never
// equal a "surface:N" target ref and would thus silently defeat the never-inject-
// into-self guard (the Phase-0 safety bug).
func callerRef(t *treeJSON) string {
	if t != nil && t.Caller.SurfaceRef != "" {
		return t.Caller.SurfaceRef
	}
	return os.Getenv("CMUX_SURFACE_ID")
}

// cmuxSelfRef returns the surface this process runs in, so we never target ourselves.
func cmuxSelfRef() string {
	t, err := loadTree()
	if err != nil {
		t = nil
	}
	return callerRef(t)
}

// isSelf reports whether surface s is the calling session. cmux already marks the
// caller with Here=true; we also cross-check the ref against the authoritative self
// ref. Either signal is sufficient — Here is what cmux computes, ref-equality is the
// cross-check that the old guard relied on alone (and that the UUID mismatch broke).
func isSelf(s Surface, self string) bool {
	return s.Here || (self != "" && s.Ref == self)
}

// cmuxListSurfaces flattens the tree, tagging each surface with its workspace ref+name.
func cmuxListSurfaces() ([]Surface, error) {
	t, err := loadTree()
	if err != nil {
		return nil, err
	}
	var out []Surface
	for _, w := range t.Windows {
		for _, ws := range w.Workspaces {
			for _, p := range ws.Panes {
				for _, s := range p.Surfaces {
					out = append(out, Surface{
						Ref: s.Ref, Type: s.Type, Title: s.Title,
						Workspace: ws.Title, WorkspaceRef: ws.Ref, Here: s.Here,
					})
				}
			}
		}
	}
	return out, nil
}

// --- surface I/O via socket (cross-workspace safe) ---

// readScreen returns the surface text via the socket. We prefer the base64 field
// (exact bytes) over the plain "text" field.
func cmuxReadScreen(ref, wsRef string, lines int) (string, error) {
	res, err := rpcCall("surface.read_text", map[string]any{
		"surface_id": ref, "workspace_id": wsRef, "lines": lines, "scrollback": true,
	})
	if err != nil {
		return "", err
	}
	if b64, ok := res["base64"].(string); ok && b64 != "" {
		if dec, derr := base64.StdEncoding.DecodeString(b64); derr == nil {
			return string(dec), nil
		}
	}
	txt, _ := res["text"].(string)
	return txt, nil
}

// readScreenRetry rides over a brief lag (e.g. a just-created surface).
func readScreenRetry(ref, wsRef string, attempts int) (string, error) {
	var out string
	var err error
	for i := 0; i < attempts; i++ {
		if out, err = readScreen(ref, wsRef, stateTailLines); err == nil {
			return out, nil
		}
		time.Sleep(150 * time.Millisecond)
	}
	return "", err
}

// sendText types text into a surface WITHOUT submitting (never embeds a newline:
// cmux turns "\n"/"\r" into Enter and Claude's Ink TUI would submit early).
func cmuxSendText(ref, wsRef, text string) error {
	_, err := rpcCall("surface.send_text", map[string]any{
		"surface_id": ref, "workspace_id": wsRef, "text": text,
	})
	return err
}

// sendKey sends one key event (lowercase: "enter", "escape", "ctrl+c").
func cmuxSendKey(ref, wsRef, key string) error {
	_, err := rpcCall("surface.send_key", map[string]any{
		"surface_id": ref, "workspace_id": wsRef, "key": key,
	})
	return err
}
