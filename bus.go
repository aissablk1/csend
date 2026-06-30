package main

// bus.go — câblage CLI de la voie COOPÉRATIVE (inbox + identité + mémoire).
//
// Contrairement aux commandes cmux (list/tree/send/read), ces commandes ne
// dépendent PAS d'un multiplexeur : elles marchent sur tout OS / tout shell.
// C'est le premier pas du découplage cmux → bus universel.

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

func nowRFC3339() string { return time.Now().UTC().Format(time.RFC3339) }

// selfAgentID is the stable handle of the calling session for cooperative routing.
// Priority: explicit env > cmux session key (if any) > hostname > "local".
func selfAgentID() string {
	if v := os.Getenv("CSEND_AGENT_ID"); v != "" {
		return v
	}
	if k, err := selfKey(); err == nil && k != "" {
		return k
	}
	if h, err := os.Hostname(); err == nil && h != "" {
		return h
	}
	return "local"
}

func mustStore() *Store {
	s, err := OpenStore(DefaultStoreDir())
	if err != nil {
		fail(err.Error())
	}
	return s
}

// resolveVaultPass returns the vault passphrase. CSEND_VAULT_PASS_FILE wins (a file
// avoids leaking the secret in `ps`/env dumps — §38) ; CSEND_VAULT_PASS is the
// fallback. Returns (nil,false) if neither is set.
func resolveVaultPass() ([]byte, bool) {
	if f := os.Getenv("CSEND_VAULT_PASS_FILE"); f != "" {
		if data, err := os.ReadFile(f); err == nil {
			return []byte(strings.TrimRight(string(data), "\r\n")), true
		}
	}
	if p := os.Getenv("CSEND_VAULT_PASS"); p != "" {
		return []byte(p), true
	}
	return nil, false
}

func sha6(text string) string {
	sum := sha256.Sum256([]byte(text))
	return hex.EncodeToString(sum[:6])
}

// fingerprint is a short, stable display id for a public identity.
func fingerprint(b PublicBundle) string {
	sum := sha256.Sum256(b.SignPub)
	return hex.EncodeToString(sum[:8])
}

// --- identity persistence (secret sealed by a passphrase, §38) ---

func identityVaultPath(s *Store) string { return filepath.Join(s.Dir, "identity.vault") }
func identityPubPath(s *Store) string   { return filepath.Join(s.Dir, "identity.pub.json") }

func saveIdentity(s *Store, id *Identity, pass []byte) error {
	secret, err := id.MarshalSecret()
	if err != nil {
		return err
	}
	blob, err := SealVault(secret, pass)
	if err != nil {
		return err
	}
	bjson, err := json.MarshalIndent(blob, "", "  ")
	if err != nil {
		return err
	}
	if err := os.WriteFile(identityVaultPath(s), bjson, 0o600); err != nil {
		return err
	}
	pub, err := json.MarshalIndent(id.Public(), "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(identityPubPath(s), pub, 0o644)
}

func loadPublicBundle(s *Store) (PublicBundle, bool) {
	data, err := os.ReadFile(identityPubPath(s))
	if err != nil {
		return PublicBundle{}, false
	}
	var b PublicBundle
	if json.Unmarshal(data, &b) != nil {
		return PublicBundle{}, false
	}
	return b, true
}

// cmdID shows the local identity, or creates one with `--create` (needs a vault
// passphrase in CSEND_VAULT_PASS — we never persist a secret key in clear, §38).
func cmdID(args []string) {
	create, export := false, false
	for _, a := range args {
		switch a {
		case "--create":
			create = true
		case "--export":
			export = true
		}
	}
	s := mustStore()
	if export {
		b, ok := loadPublicBundle(s)
		if !ok {
			fail("aucune identité locale — crée-en une : csend id --create")
		}
		fmt.Printf("Jeton de clé PUBLIQUE (fingerprint %s) — partage-le avec tes pairs :\n\n%s\n", fingerprint(b), encodeBundle(b))
		return
	}
	if create {
		pass, ok := resolveVaultPass()
		if !ok {
			fail("définis CSEND_VAULT_PASS_FILE (recommandé) ou CSEND_VAULT_PASS pour chiffrer le vault (§38)")
		}
		id, err := NewIdentity()
		if err != nil {
			fail(err.Error())
		}
		if err := saveIdentity(s, id, pass); err != nil {
			fail(err.Error())
		}
		fmt.Printf("✓ identité créée (hybride Ed25519 + X25519 + ML-KEM-768)\n  fingerprint: %s\n  vault: %s\n",
			fingerprint(id.Public()), identityVaultPath(s))
		return
	}
	b, ok := loadPublicBundle(s)
	if !ok {
		fmt.Println("Aucune identité locale. Crée-en une : CSEND_VAULT_PASS=… csend id --create")
		return
	}
	fmt.Printf("identité locale  fingerprint: %s\n  agent-id: %s\n", fingerprint(b), selfAgentID())
}

// selfProvider renvoie l'éditeur (claude|codex|gemini…) de la session courante,
// lu depuis le registre — peuple l'enveloppe pour que le bus soit « provider-aware »
// (c'est ce qui rend la collaboration cross-vendor visible : qui parle, et de quel CLI).
func selfProvider(s *Store) string {
	if r, ok, _ := s.GetSession(selfAgentID()); ok {
		return r.Provider
	}
	return ""
}

// cmdInbox delivers a message COOPERATIVELY to a recipient's mailbox (no cmux).
//
//	csend inbox <destinataire> <message…>
func cmdInbox(args []string) {
	if len(args) < 2 {
		fail("usage: csend inbox <destinataire> <message…>")
	}
	to := args[0]
	body := strings.Join(args[1:], " ")
	s := mustStore()
	m := InboxMessage{ID: newID(), TS: nowRFC3339(), From: selfAgentID(), To: to, Provider: selfProvider(s), Kind: "msg"}
	enc := ""
	if sealed, ok := maybeSeal(s, to, body); ok {
		m.Sealed = sealed
		enc = " · chiffré E2E"
	} else {
		m.Body = body
	}
	if err := s.Inbox().Deliver(m); err != nil {
		fail(err.Error())
	}
	_ = s.AppendJournal(JournalEntry{
		ID: m.ID, TS: m.TS, From: m.From, To: to, Channel: "inbox",
		Action: "submitted", TextSHA: sha6(body), TextLen: len(body),
	})
	fmt.Printf("✓ déposé dans l'inbox de %s [inbox]%s\n", to, enc)
}

// cmdRecv drains the calling agent's cooperative mailbox.
func cmdRecv(args []string) {
	markRead := true
	for _, a := range args {
		if a == "--peek" {
			markRead = false
		}
	}
	s := mustStore()
	agent := selfAgentID()
	msgs, err := s.Inbox().Receive(agent, markRead)
	if err != nil {
		fail(err.Error())
	}
	if wantJSON(args) {
		arr := make([]jsonMessage, 0, len(msgs))
		for _, m := range msgs {
			arr = append(arr, jsonMessage{From: m.From, TS: m.TS, Body: openBody(s, m), Encrypted: m.Sealed != nil})
		}
		emitJSON(map[string]any{"agent": agent, "count": len(arr), "messages": arr})
		return
	}
	if len(msgs) == 0 {
		fmt.Printf("Inbox vide pour %s.\n", agent)
		return
	}
	fmt.Printf("%d message(s) pour %s%s :\n", len(msgs), agent, peekSuffix(markRead))
	for _, m := range msgs {
		fmt.Printf("  • [%s] de %s : %s\n", m.TS, m.From, openBody(s, m))
	}
}

func peekSuffix(markRead bool) string {
	if markRead {
		return ""
	}
	return " (peek — non consommés)"
}

// cmdContact manages the public-key keyring of peers (enables E2E sealing).
func cmdContact(args []string) {
	s := mustStore()
	if len(args) == 0 || args[0] == "list" {
		cs, _ := s.ListContacts()
		if len(cs) == 0 {
			fmt.Println("Aucun contact. Ajoute la clé publique d'un pair : csend contact add <agent> <jeton>")
			return
		}
		fmt.Println("Contacts (clés publiques connues) :")
		for _, c := range cs {
			fmt.Printf("  • %s\n", c)
		}
		return
	}
	if args[0] == "add" {
		if len(args) < 3 {
			fail("usage: csend contact add <agent> <jeton>")
		}
		b, err := decodeBundle(args[2])
		if err != nil {
			fail("jeton invalide: " + err.Error())
		}
		if err := s.SaveContact(args[1], b); err != nil {
			fail(err.Error())
		}
		fmt.Printf("✓ contact « %s » ajouté (fingerprint %s) — les messages vers lui seront chiffrés E2E\n", args[1], fingerprint(b))
		return
	}
	fail("usage: csend contact add <agent> <jeton>  |  csend contact list")
}
