# Changelog

Toutes les évolutions notables de csend sont consignées ici.

Le format suit [Keep a Changelog](https://keepachangelog.com/fr/1.1.0/), et le projet
adopte le [versionnage sémantique](https://semver.org/lang/fr/).

## [Non publié]

Voir [`docs/NEXT.md`](docs/NEXT.md) pour le détail de ce qui reste et pourquoi.
Prochains jalons : adaptateurs Codex/Gemini, passkey WebAuthn, auth mutuelle réseau,
clients mobiles, surface MCP, audit cryptographique externe.

## [0.2.0] — 2026-06-27

Première base documentée — **alpha**. csend passe du simple injecteur inter-sessions
cmux à un **bus de messages universel** : coopératif, cross-OS/provider, chiffré de bout
en bout et résistant au quantique, avec mémoire persistante et recovery. Tout l'ensemble
ci-dessous est dans l'arbre et couvert par la suite de tests (`go test -race ./...`
vert) ; rien n'est promis sans être livré.

### Ajouté

- **Bus coopératif (colonne vertébrale universelle).** Inbox par fichiers, durable et
  sans démon : `register`, `inbox`, `recv`, `agents`, `whoami`. Marche sur tout OS et
  tout provider, là où l'injection clavier ne peut pas aller.
- **Injection live state-aware (Unix).** Backends **cmux** et **tmux** interchangeables
  derrière une interface `Backend` ; lecture de l'état (`● idle` / `◐ busy` /
  `⚠ confirm`) **avant** d'écrire ; modes de livraison gouvernés (`--stage`, `--send`,
  `--force`). Détection d'état **Claude** calibrée sur de vrais écrans.
- **Graphe familial.** `tree`, `link`/`unlink`, broadcasts `--up`/`--down`/
  `--to-siblings`/`--to-descendants`, avec garde anti-cycle.
- **Chiffrement E2E hybride post-quantique.** Messages signés **Ed25519**, chiffrés
  **AES-256-GCM** sous une clé dérivée de **X25519 ⊕ ML-KEM-768** (HKDF-SHA256).
  Scellement automatique sur le fil dès qu'un contact est connu (`contact add`/`list`,
  jetons `csend id --export`).
- **Identités & vault.** `id --create`/`--export` ; identité hybride dérivée d'une graine
  maître de 32 octets ; vault au repos scellé **AES-256-GCM** avec clé **PBKDF2-SHA256**
  (600 000 itérations). Passphrase via `CSEND_VAULT_PASS_FILE` (recommandé) ou
  `CSEND_VAULT_PASS`.
- **Recovery.** Partage par seuil **Shamir N-sur-M** sur GF(2⁸) (`recovery split`/
  `combine`) et **phrase BIP-39** de 24 mots (`recovery phrase`/`from-phrase`), wordlist
  anglaise officielle embarquée.
- **Réseau machine-à-machine.** `serve`/`remote` (frame JSON) avec payload chiffrable
  E2E, et **TLS 1.3 hybride PQC** (`X25519MLKEM768`) via cert self-signed Ed25519 +
  épinglage d'empreinte (`--tls`, `--pin`). Cible : loopback/LAN.
- **Mémoire.** Journal append-only interrogeable (hash du contenu, jamais le clair) +
  registre persistant des sessions.
- **Aide accessible** sous toutes les formes (`h`, `-h`, `--h`, `help`, `-help`,
  `--help`, `-?`).
- **Outillage de publication.** CI GitHub Actions (`go vet` + `go test -race` + build),
  `Makefile`, `install.sh` (binaire universel arm64 + x86_64), configuration GoReleaser
  (binaires multi-plateformes + formule Homebrew), `PROJECT.nfo`, site statique.

### Corrigé

- **Anti-auto-injection.** `selfRef()` comparait un UUID (`CMUX_SURFACE_ID`) à des cibles
  en forme `surface:N`, si bien que le garde « ne pas s'écrire à soi-même » ne matchait
  jamais. Source d'identité faisant désormais autorité (`cmux tree` / `here=true`), avec
  test de non-régression.

### Sécurité

- Toutes les primitives proviennent de la **stdlib Go 1.24** — **zéro dépendance
  externe**, aucune crypto maison.
- L'assemblage cryptographique **n'a pas encore été audité en externe** ; voir
  [`SECURITY.md`](SECURITY.md) pour le modèle de menace et les limites connues (PBKDF2 vs
  Argon2id, auth réseau, anti-rejeu).

### Limites connues

- Injection clavier **Unix uniquement** (Windows natif et mobile = coopératif-only).
- Adaptateurs Codex/Gemini, passkey, auth mutuelle réseau, clients mobiles et surface
  MCP : **non livrés** (feuille de route).

[Non publié]: https://github.com/aissablk1/csend/compare/v0.2.0...HEAD
[0.2.0]: https://github.com/aissablk1/csend/releases/tag/v0.2.0
