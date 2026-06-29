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
- 2026-06-27 — **Backend tmux** + abstraction transport (interface `Backend`, cmux+tmux
  interchangeables). **E2E réel vérifié** (injection dans un pane tmux 3.6a exécutée) ;
  self-fix confirmé en live (`(ici)` apparaît). Commit `89f4db2`. README+site MAJ (tmux livré).
- 2026-06-27 — **BIP-39** : identité dérivée d'UNE graine maître → **récupérable par phrase
  de 24 mots** ; wordlist officielle, vecteurs de test ; CLI `recovery phrase/from-phrase`.
  Smoke E2E (phrase → même identité). Commit `…BIP-39`.
- 2026-06-27 — **Réseau** `serve`/`remote` (loopback, payload E2E-ready) ; smoke 2 process
  (`remote → serve → recv`). Commit réseau.
- 2026-06-27 — **Durcissement vault** : passphrase via `CSEND_VAULT_PASS_FILE` (évite la fuite
  `ps`, §38). Commit vault-file.
- 2026-06-27 — **`docs/NEXT.md`** : triage honnête du reste (passkey, mobile, Codex/Gemini,
  Agent Teams, screen, TLS/PQC) avec la raison réelle de chaque blocage.
- 2026-06-27 — Recherche vérifiée **Sakana Fugu** (Sakana AI, 22 juin 2026) : routeur de
  modèles cloud OpenAI-compatible → **catégorie différente** de csend (pas un concurrent).
- 2026-06-27 — **E2E sur le fil** : keyring (clés publiques des pairs), `id --export`,
  `contact add/list` ; `inbox`/`remote` scellent **auto** quand le contact est connu, `recv`
  déchiffre. Le « chiffré de bout en bout » est branché sur les vrais messages (plus dormant).
  Tests cross-store + smoke (0 clair dans le fichier déposé). Commit `…E2E`.
- 2026-06-27 — **Aide** accessible via `h`, `-h`, `--h`, `help`, `-help`, `--help`, `-?`
  (demande d'Aïssa). 6 formes vérifiées.
- 2026-06-27 — **Flotte coopérative cross-OS/provider** : `csend register` (rejoindre le bus
  depuis tout terminal/provider/OS sans multiplexeur), `csend agents` (vue flotte), `csend
  whoami`. Smoke 3 providers (bash/codex/gemini). Commit flotte.
- 2026-06-29 — Session parallèle détectée (§7) : l'autre session traite distribution/projet
  (LICENSE, SECURITY.md, adapters Codex/Gemini, CHANGELOG, Formula Homebrew, site, version).
  Je travaille dans **mes** fichiers ; `main.go` committé sur autorisation explicite d'Aïssa.
- 2026-06-29 — **M1 `--json`** (recv/agents/whoami) → pilotage par agents. Smoke. Commit.
- 2026-06-29 — **M2 réception live** : `csend hook` (UserPromptSubmit → messages dans le
  contexte, sans polling) + `csend watch` + `--install`. Câblage `main.go` finalisé. Smoke.
- 2026-06-29 — **M3 auth + file offline** : `serve --authz` (n'accepte qu'un message E2E signé
  d'un expéditeur allowlisté) ; `remote` met en file si injoignable et rejoue. Tests + smoke.
- 2026-06-29 — **N verrou fichier** (`UpsertSession` anti lost-update) ; test `-race` (25
  upserts concurrents = 25 agents). **O anti-replay** (dédup sur le nonce signé). Tests.
- 2026-06-29 — **P CI cross-OS** (Linux/macOS/Windows) + race + **fuzz** du parser. **Q**
  unités launchd/systemd + **démo runnable 2 agents** (collaboration via hook, prouvée).
- 2026-06-29 — **Viabilité (workflow /startup + /brainstorming)** : analyse marché réelle
  (marché agents IA ~8 Md$ en 2026 ; Agent Teams natif depuis févr. 2026 = doublon du cœur
  « messagerie inter-sessions »). Verdict validé par Aïssa : csend **viable en open source**,
  pas en business → `docs/adoption/playbook-adoption-oss.md`.
- 2026-06-29 — **Build « tout réaliser » (4 agents parallèles)** : (1) docs/confiance — README
  v2, SECURITY, CHANGELOG, **LICENSE MIT** (manquait alors que revendiqué partout), Formula
  Homebrew, release ; (2) adaptateurs Codex/Gemini + tests ; (3) site landing terminal-souverain
  (anti-slop, zéro fausse preuve) ; (4) vidéo de lancement `brag` (sans audio, publiable). Vert.
- 2026-06-29 — **Calibration adaptateurs sur SOURCE PRIMAIRE (§2/§29)** : Gemini sur le bundle
  installé `@google/gemini-cli` 0.40.1 (vérifié par grep local), Codex sur `openai/codex`
  `rust-v0.142.3` (contre-vérifié par fetch). Suppositions fausses retirées (`context left`,
  `gpt-\d`, `⏎ send`). `go test ./...` vert. Commit `3bac2fd`.
- 2026-06-29 — **Prep publication** : `csend version` câblé (+ injection ldflags
  goreleaser/Homebrew), `_backup/` et `brag-output/` ignorés, statut Codex/Gemini mis à jour
  (README + site). Commits `e3dfc23` (docs) + `eb13d6c` (site).
- 2026-06-29 — **Revert effet de bord §5** : `hyperframes init` (lancé pour la vidéo) avait semé
  des liens dans 9 outils d'agents tiers (`.codex`/`.cursor`/`.continue`/`.hermes`/…) + 8
  dossiers dans `~/.agents`. Révertés (72 liens + 8 dossiers) avec manifestes de backup ;
  `~/.claude` et le reste du store `~/.agents` (skills vidéo utilisables) préservés (choix Aïssa).

## Actions à mener à l'avenir

Triage complet et honnête dans **`docs/NEXT.md`** (raison réelle de chaque blocage) :
- Reporté par choix (faible ROI) : backend `screen`.
- Bloqué par prérequis physique : passkey WebAuthn (authentificateur), TLS hybride PQC,
  clients mobiles, Windows natif (coop-only — limite physique assumée).
- Adaptateurs Codex/Gemini : **calibrés sur source officielle** (gemini-cli 0.40.1, codex
  rust-v0.142.3) ; reste la **confirmation par capture d'écran live** (auth Gemini / install
  Codex requis). Bridge Agent Teams : toujours à concevoir (§2).
- Publication (GitHub/Homebrew/site) : en attente du feu vert d'Aïssa.

## Notes / Décisions / Blocages

- Décision : bus coopératif = backbone universel ; injection live = repli Unix-only.
- Décision : crypto = primitives auditées (stdlib Go 1.24 : ed25519/ecdh/mlkem/aes-gcm/hkdf/pbkdf2),
  jamais maison (§38). Recovery = BIP-39 + Shamir GF(256).
- Honnêteté (§29) : Windows/mobile = coopératif-only (pas d'injection). Livré par phases ;
  rapport final distingue testé vs conçu.
