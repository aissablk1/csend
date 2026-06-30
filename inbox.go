package main

// inbox.go — transport COOPÉRATIF (couche 3a) : la colonne vertébrale universelle.
//
// Une session « coopérante » (qui a le hook/listener/skill csend) reçoit ses
// messages en lisant son inbox. Ce chemin marche PARTOUT — tout OS, tout provider,
// Agent Teams, sous-agents, distant — là où l'injection clavier ne peut pas aller.
//
// Modèle fichier (durable, sans démon requis) :
//   <store>/inbox/<agent-id>/<id>.json        message en attente
//   <store>/inbox/<agent-id>/.read/<id>.json  message lu (conservé, mémoire §6)

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// Inbox is the directory root holding every agent's mailbox.
type Inbox struct {
	Root string
}

// Inbox returns the cooperative mailbox transport for this store.
func (s *Store) Inbox() *Inbox {
	return &Inbox{Root: filepath.Join(s.Dir, "inbox")}
}

// InboxMessage is one cooperative message. Body holds plaintext for local trust;
// Sealed (optional) carries an E2E envelope when sender/recipient identities exist.
type InboxMessage struct {
	ID       string         `json:"id"`
	TS       string         `json:"ts"`
	From     string         `json:"from"`
	To       string         `json:"to"`
	Provider string         `json:"provider,omitempty"` // éditeur de l'expéditeur (claude|codex|gemini…)
	Kind     string         `json:"kind,omitempty"`     // msg | nudge | result
	Body     string         `json:"body,omitempty"`
	Sealed   *SealedMessage `json:"sealed,omitempty"`
}

func (ib *Inbox) agentDir(agentID string) string {
	return filepath.Join(ib.Root, safeID(agentID))
}

// Deliver writes a message into the recipient's mailbox. Atomic (write temp+rename).
func (ib *Inbox) Deliver(m InboxMessage) error {
	if m.To == "" || m.ID == "" {
		return os.ErrInvalid
	}
	dir := ib.agentDir(m.To)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}
	data, err := json.MarshalIndent(m, "", "  ")
	if err != nil {
		return err
	}
	final := filepath.Join(dir, safeID(m.ID)+".json")
	tmp := final + ".tmp"
	if err := os.WriteFile(tmp, data, 0o600); err != nil {
		return err
	}
	return os.Rename(tmp, final)
}

// Pending returns the count of unread messages for an agent.
func (ib *Inbox) Pending(agentID string) (int, error) {
	msgs, err := readDir(ib.agentDir(agentID))
	return len(msgs), err
}

// Receive returns an agent's unread messages oldest→newest. If markRead is true,
// each returned message is moved to the .read/ archive (kept, not deleted — §6).
func (ib *Inbox) Receive(agentID string, markRead bool) ([]InboxMessage, error) {
	dir := ib.agentDir(agentID)
	names, err := readDir(dir)
	if err != nil {
		return nil, err
	}
	var msgs []InboxMessage
	for _, name := range names {
		data, rerr := os.ReadFile(filepath.Join(dir, name))
		if rerr != nil {
			continue
		}
		var m InboxMessage
		if json.Unmarshal(data, &m) != nil {
			continue
		}
		msgs = append(msgs, m)
	}
	sort.Slice(msgs, func(i, j int) bool {
		if msgs[i].TS != msgs[j].TS {
			return msgs[i].TS < msgs[j].TS
		}
		return msgs[i].ID < msgs[j].ID
	})
	if markRead {
		readDirPath := filepath.Join(dir, ".read")
		if err := os.MkdirAll(readDirPath, 0o755); err != nil {
			return msgs, err
		}
		for _, name := range names {
			_ = os.Rename(filepath.Join(dir, name), filepath.Join(readDirPath, name))
		}
	}
	return msgs, nil
}

// readDir lists the *.json message files in an agent dir (not the .read archive).
func readDir(dir string) ([]string, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}
	var out []string
	for _, e := range entries {
		if e.IsDir() {
			continue // skips .read/
		}
		if strings.HasSuffix(e.Name(), ".json") {
			out = append(out, e.Name())
		}
	}
	return out, nil
}

// safeID keeps a handle usable as a single path segment (no traversal, no slashes).
func safeID(id string) string {
	r := strings.NewReplacer("/", "_", "\\", "_", "..", "_", ":", "_", " ", "_")
	s := r.Replace(id)
	if s == "" || s == "." {
		return "_"
	}
	return s
}
