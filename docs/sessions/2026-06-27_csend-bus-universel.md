---
session_id: 6cb99775-csend-bus-universel
date_debut: 2026-06-27
date_fin: 2026-06-27
workspace: /Volumes/Professionnel/Projets/Développement/Outils/csend
auteur: Aïssa BELKOUSSA
statut: phases 0 et 1 livrées (vert, committé) — phases 2+ à venir
tags: [csend, inter-agent, architecture, crypto, pqc, memoire]
---

# Session 2026-06-27 — csend : du fix à la refonte « bus universel »

## QQOQCCP

- **Qui** : Aïssa BELKOUSSA (décisionnaire) ; exécution assistée.
- **Quoi** : analyser csend (outil d'injection inter-sessions Claude CLI), corriger ses
  faiblesses, le comparer au marché, puis le refondre en **bus de messagerie inter-agents
  universel** (cross-session, cross-provider, cross-terminal, cross-OS) avec mémoire persistante,
  crypto, vault passkey et recovery Shamir/BIP-39.
- **Où** : `…/Outils/csend` (binaire Go), config `~/.claude/skills/csend`, `~/.claude/commands/csend.md`.
- **Quand** : 2026-06-27.
- **Comment** : analyse de code + 2 agents de recherche (GitHub/Reddit) → brainstorming
  (skill superpowers) → spec → git → implémentation TDD par phases.
- **Combien** : chantier XL, livré par phases ; v1 = Phase 0 + Phase 1.
- **Pourquoi** : aucun outil du marché ne combine injection-dans-session-externe + state-aware
  + cross-provider + mémoire persistante. Trou de marché réel.

## Actions analysées

- csend actuel : Go, backend cmux (socket JSON-RPC), state-aware (idle/busy/confirm), graphe
  familial, audit PII-safe. Vérifié en live (`csend list` sur 9 sessions / 6 workspaces).
- Faiblesses confirmées : bug d'identité self (UUID env vs `surface:N` → garde anti-auto-injection
  contournable) ; couplage dur cmux ; détection fragile (3/6 sessions « unknown ») ; pas de git ;
  pas d'async/offline.
- Marché (vérifié) : injecteurs aveugles (Tmux-Orchestrator, MCP tmux) vs managers state-aware
  qui possèdent la session (ccmanager, CAO, agtx, tmai). Aucun ne couvre le croisement visé.
- Contraintes physiques : TIOCSTI mort, PTY slave ≠ stdin, seul le détenteur du PTY maître injecte.

## Actions réalisées

- 2026-06-27 — Analyse complète du code csend + log + skill + tests (lecture).
- 2026-06-27 — Vérif live : env cmux, `csend list`, `cmux tree --json` → bug self confirmé.
- 2026-06-27 — 2 agents de recherche (paysage GitHub ; Reddit + mécanismes techniques).
- 2026-06-27 — Brainstorming (skill) : architecture 8 couches validée par Aïssa.
- 2026-06-27 — `git init` du dépôt ; arborescence `docs/`.
- 2026-06-27 — Spec écrite : `docs/superpowers/specs/2026-06-27-csend-bus-universel-design.md`.
- 2026-06-27 — Journal de session créé (ce fichier).
- 2026-06-27 — Commit `4f198d1` : socle + spec sous git.
- 2026-06-27 — Phase 0 : fix bug self (TDD `callerRef`/`isSelf`) → commit `6a101b5`.
- 2026-06-27 — Phase 0 : Makefile + CI GitHub Actions + PROJECT.nfo + README v2 → `fa33244`.
- 2026-06-27 — Phase 1 : couche crypto E2E hybride (Ed25519+X25519+ML-KEM) + vault → `e625e7e`.
- 2026-06-27 — Phase 1 : pilier mémoire (registre + journal interrogeable) → `8469b16`.
- 2026-06-27 — Phase 1 : backbone coopératif (inbox + router + CLI recv/inbox/id, hors cmux),
  smoke E2E vérifié → `24b7e23`.
- 2026-06-27 — Stretch : recovery Shamir GF(256) + CLI split/combine, smoke E2E → `a20dd40`.
- 2026-06-27 — Couche providers pluggable (Claude réel, extension Codex/Gemini) → `08a3db2`.
- 2026-06-27 — Suite verte tout du long (`go test -race`), 8 commits (cœur).
- 2026-06-27 — Workflow frontend §10 : anti-slop lu, direction nommée (#3 rétro-futuriste
  terminal × #11 industriel), checklist §9 cochée.
- 2026-06-27 — Site `site/` (HTML/CSS/JS + GSAP) : hero terminal animé, séquence scroll
  8 couches, crypto, install, statut honnête (zéro fausse preuve). **Vérifié au screenshot**
  (hero + page complète), favicon corrigé, console propre.
- 2026-06-27 — Install facilitée : module path → `go install`, `.goreleaser.yaml` (Homebrew),
  `site/install.sh`, README réécrit (humains + agents) → commit `bda3e2b`.
- 2026-06-27 — `docs/PUBLISHING.md` : recommandation de publication + distribution.

## Actions à mener à l'avenir (honnête — non fait cette session)

- **Backend tmux/screen** (injection live hors cmux) : nécessite d'abstraire le contrat
  transport ; reporté pour ne pas fragiliser le vert en fin de session (§32).
- **Encodage mot des parts** (BIP-39 / SLIP-39 wordlist) : la math Shamir est faite ;
  l'habillage « phrase » reste à brancher (pas de wordlist inventée, §2).
- **Adaptateurs Codex/Gemini réels** : à calibrer sur de vrais écrans capturés (§2).
- **Phase 2+** : passkey WebAuthn ; Phase 3 réseau multi-machine (TLS hybride PQC) ;
  Phase 4 mobile + Windows + bridge Agent Teams ; Phase 5 PQC complet + audit externe.

## Notes / Décisions / Blocages

- Décision : bus coopératif = backbone universel ; injection live = repli Unix-only.
- Décision : crypto = primitives auditées (stdlib Go 1.24 : ed25519/ecdh/mlkem/aes-gcm/hkdf/pbkdf2),
  jamais maison (§38). Recovery = BIP-39 + Shamir GF(256).
- Honnêteté (§29) : Windows/mobile = coopératif-only (pas d'injection). Livré par phases ;
  rapport final distingue testé vs conçu.
