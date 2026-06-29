# csend

[![CI](https://github.com/aissablk1/csend/actions/workflows/ci.yml/badge.svg)](https://github.com/aissablk1/csend/actions/workflows/ci.yml)
[![Licence : MIT](https://img.shields.io/badge/licence-MIT-blue.svg)](LICENSE)
[![Go 1.24](https://img.shields.io/badge/Go-1.24-00ADD8?logo=go&logoColor=white)](https://go.dev/dl/)
[![Go Report Card](https://goreportcard.com/badge/github.com/aissablk1/csend)](https://goreportcard.com/report/github.com/aissablk1/csend)
[![Zéro dépendance](https://img.shields.io/badge/dépendances-0%20(stdlib)-success)](go.mod)

**Le bus de messages qui fait parler tes agents de code entre eux.** Une session
d'agent CLI (Claude Code, Codex, Gemini…) écrit à une **autre session en cours** — en
lisant son état (idle / busy / confirmation) **avant** d'agir, à travers terminaux,
providers et machines, **chiffré de bout en bout** et résistant au quantique. Go,
**zéro dépendance**, MIT.

> **Le trou que csend comble.** Aucune API officielle n'injecte un prompt dans une
> session d'agent CLI **vivante** (cf. les feature requests Claude Code fermées/ouvertes
> [#24947](https://github.com/anthropics/claude-code/issues/24947),
> [#27441](https://github.com/anthropics/claude-code/issues/27441)). csend combine deux
> chemins : un **inbox coopératif** = colonne vertébrale universelle (tout OS, tout
> provider) et une **injection clavier live** = repli pour les TUI Unix. Une seule
> adresse, un seul outil, et le message trouve toujours un chemin.

> [!WARNING]
> Statut : **alpha** (v0.2.0-dev). Le code est testé (`go test -race ./...` vert) et
> sans dépendance, mais l'API peut bouger et la crypto **n'a pas encore reçu d'audit
> externe** — voir [SECURITY.md](SECURITY.md). Lis « Feuille de route » avant de
> compter dessus en production.

---

## Pourquoi

Tu pilotes plusieurs agents de code en parallèle (une fenêtre par tâche, parfois
plusieurs machines). Il manque le câble entre eux : qu'une session puisse en
**prévenir**, **relancer** ou **coordonner** une autre, sans copier-coller à la main,
sans tout réécrire dans un orchestrateur propriétaire, et **sans confier le contenu de
tes prompts à un relais en clair**. csend est ce câble : souverain, chiffré, sans
dépendance.

## Installation

```sh
# Homebrew (macOS · Linux) — dès la première release taguée
brew install aissablk1/tap/csend

# Go (toute plateforme avec une toolchain Go 1.24+) — marche dès que le repo est public
go install github.com/aissablk1/csend@latest

# Script (audite-le d'abord : il ne fait que tirer le binaire des GitHub Releases)
curl -fsSL https://csend.dev/install.sh | sh

# Source (zéro confiance, binaire universel arm64 + x86_64 sur macOS)
git clone https://github.com/aissablk1/csend && cd csend && make install
```

## En 60 secondes

```sh
# --- Voie coopérative : marche PARTOUT, sans multiplexeur ---
csend register                               # cette session rejoint le bus (tout OS/provider)
csend inbox SACEM "lance le build de prod"   # une session écrit à une autre
csend recv                                   # cette session relève (et vide) son inbox

# --- Voie cmux/tmux : pilotage state-aware d'une flotte Unix ---
csend list                  # sessions agent + état : ● idle  ◐ busy  ⚠ confirm
csend tree                  # graphe familial père → enfants
csend send SACEM "go"       # message gouverné (ne valide QUE si la cible est idle)
csend send --down "build"   # broadcast aux sessions enfants

# --- Identité chiffrée & recovery par seuil ---
CSEND_VAULT_PASS=… csend id --create     # identité hybride (Ed25519 + X25519 + ML-KEM-768) en vault chiffré
csend recovery split 2 3                 # 3 parts Shamir, seuil 2-sur-3
csend recovery combine <part1> <part2>   # reconstitue depuis ≥ 2 parts
```

**Chiffrement de bout en bout, concrètement :** échange les jetons de clé publique
(`csend id --export` → `csend contact add <pair> <jeton>`), et tout message vers ce
pair part **scellé** — l'inbox, le journal et le relais réseau ne voient que du
chiffré. Sans contact connu, csend retombe en clair (confiance locale, jamais sur le
réseau ouvert sans durcissement).

## Pour les agents IA

csend est exposé comme **skill Claude Code** (`~/.claude/skills/csend`) : une session
peut **écrire à / relever / coordonner** ses paires en langage naturel (« dis à SACEM
de relancer le build », « relève mon inbox », « broadcast à mes enfants »). La voie
coopérative (`inbox` / `recv` / `register`) ne dépend d'aucun terminal — idéale pour
des flottes d'agents hétérogènes.

## Comparatif honnête

Ce que csend fait, face aux outils voisins. `✅` = livré · `◐` = partiel · `❌` =
absent ou hors-modèle.

| Capacité | **csend** | Agent Teams (natif) | ruflo (claude-flow) | tmux-orchestrator |
|---|:---:|:---:|:---:|:---:|
| Cross-provider (Claude / Codex / Gemini) | ✅ <sup>1</sup> | ❌ Claude only | ◐ | ❌ |
| Cross-OS + mobile | ✅ <sup>2</sup> | ❌ desktop | ❌ | ❌ Unix only |
| Chiffrement E2E des messages | ✅ | ❌ sandbox/vault partagés | ❌ | ❌ clair |
| Post-quantique (ML-KEM-768 hybride) | ✅ | ❌ | ❌ | ❌ |
| Vault + recovery (Shamir, BIP-39) | ✅ | ❌ | ❌ | ❌ |
| Injection dans une session **déjà lancée** | ✅ <sup>3</sup> | ◐ cadre Agent Teams | ◐ | ◐ panes, à l'aveugle |
| Zéro dépendance externe | ✅ stdlib Go | — natif | ❌ npm | ◐ bash |

<sup>1</sup> Détection d'état **Claude** calibrée sur de vrais écrans ; adaptateurs
**Codex et Gemini livrés**, calibrés sur source officielle (bundle `@google/gemini-cli`
0.40.1 et `openai/codex` `rust-v0.142.3`) — confirmation par capture d'écran live en
attente. La voie coopérative, elle, est déjà provider-agnostique.
<sup>2</sup> Voie coopérative livrée **sur tout OS** ; clients **mobiles = phase 4**
(le mobile rejoint le bus comme client — voir, approuver, broadcaster —, il **n'injecte
pas** au clavier : sandbox).
<sup>3</sup> Injection live **Unix uniquement** (cmux / tmux) ; **Windows natif = ❌**
(pas de PTY maître partageable). csend lit l'état de la cible **avant** d'écrire, là où
tmux-orchestrator envoie à l'aveugle.

> Honnêteté (le comparatif doit rester vérifiable) : Agent Teams est une **fonction
> native** d'un seul écosystème (Claude), pas un défaut ; ruflo (rebrand de
> `claude-flow`) est un meta-harness npm bien plus large mais non chiffré et lourd ;
> tmux-orchestrator est minimal, clair et Unix. csend occupe l'angle qu'aucun ne
> couvre : **souverain, chiffré E2E, post-quantique, cross-provider**.

## Modèle de sécurité (résumé)

Primitives **auditées, jamais maison** — toutes dans la stdlib Go 1.24 (zéro
dépendance) :

- **Messages E2E hybrides PQC** : signés **Ed25519**, chiffrés **AES-256-GCM** sous une
  clé dérivée de **X25519 ⊕ ML-KEM-768**. Il faut casser **les deux** échanges de clés
  pour lire — résistance « Harvest Now, Decrypt Later ». Le bus/relais ne voit que du
  chiffré signé (zero-trust).
- **Vault au repos** : la graine maître (32 octets, dont **tout** dérive) est scellée en
  **AES-256-GCM**, clé via **PBKDF2-SHA256** (600 000 itérations). Aucune clé privée
  n'est jamais écrite en clair. Déverrouillage par passkey WebAuthn = à venir.
- **Recovery par seuil** : **Shamir N-sur-M** sur GF(2⁸) — K-1 parts ne révèlent *rien*
  — et **phrase BIP-39** (24 mots) pour une sauvegarde papier.
- **Réseau** : `serve`/`remote` en loopback/LAN, avec **TLS 1.3 hybride post-quantique**
  (`X25519MLKEM768`) et épinglage d'empreinte. Pas d'exposition hors loopback sans
  durcir.

Détails complets, menaces couvertes et **limites honnêtes** (audit externe en attente,
PBKDF2 vs Argon2id, auth réseau) : **[SECURITY.md](SECURITY.md)**.

## Matrice OS — la vérité

| Plateforme | Bus coopératif | Injection live |
|---|:---:|:---:|
| macOS / Linux / WSL / Crostini | ✅ | ✅ (cmux / tmux) |
| Windows natif | ✅ | ❌ pas de PTY maître partageable |
| iOS / iPadOS | ✅ (client, phase 4) | ❌ sandbox |
| Android | ✅ (client, phase 4) | ❌ live cross-app |
| BSD / conteneurs / CI | ✅ s'ils parlent au bus | selon le multiplexeur présent |

L'injection clavier est **Unix-only** par contrainte physique (`TIOCSTI` désactivé,
écrire dans le PTY slave part vers l'affichage). Sur les plateformes sans injection, la
voie coopérative reste pleine et entière.

## Comment ça marche

8 couches, chacune une responsabilité :

```
0 Sécurité/crypto · 1 Identité+registre · 2 Providers · 3 Transports
4 Routeur · 5 Réseau · 6 Mémoire · 7 Surfaces (CLI/skill/hook/MCP)
```

- **Transports** : inbox coopératif (universel) · injection **cmux + tmux** state-aware
  (✅) · bridges (à venir).
- **Routeur** : `inbox › bridge › injection › file` — l'inbox gagne, car durable et sans
  course.
- **Mémoire** : journal interrogeable des messages (hash du contenu, jamais le clair) +
  registre des sessions.

Conception détaillée :
[`docs/superpowers/specs/2026-06-27-csend-bus-universel-design.md`](docs/superpowers/specs/2026-06-27-csend-bus-universel-design.md).

## Feuille de route

L'implémentation a parfois devancé le plan : ce tableau reflète **l'arbre réel**, pas
les intentions.

| Phase | Livré | En route |
|---|---|---|
| **0–1 · Fondations** | ✅ inbox coopératif · registre + mémoire/journal · crypto E2E hybride PQC · vault AES-256-GCM | — |
| **2 · Terminaux & providers** | ✅ injection **cmux + tmux** state-aware + graphe familial · détection **Claude** · adaptateurs **Codex + Gemini** (calibrés sur source, confirmation live en attente) | ❌ backend `screen` · passkey WebAuthn |
| **3 · Identité & réseau** | ✅ recovery **Shamir** · phrase **BIP-39** · réseau loopback/LAN + **TLS hybride PQC** | ❌ auth mutuelle réseau · durcissement hors-LAN |
| **4 · Portée** | — | ❌ clients **mobiles** · **Windows** (coop) · bridge **Agent Teams** · surface **MCP** |
| **5 · Durcissement** | — | ❌ signatures **ML-DSA** · **audit crypto externe** · autres OS |

## Développement

```sh
make test     # go test ./...
make vet      # go vet ./...
make build    # binaire local
make install  # binaire universel arm64+x86_64 → ~/.local/bin (macOS)
```

CI : `go vet` + `go test -race ./...` + build, à chaque push
([`.github/workflows/ci.yml`](.github/workflows/ci.yml)).

## Liens

- Modèle de sécurité : [`SECURITY.md`](SECURITY.md)
- Journal des versions : [`CHANGELOG.md`](CHANGELOG.md)
- Design & feuille de route : [`docs/superpowers/specs/2026-06-27-csend-bus-universel-design.md`](docs/superpowers/specs/2026-06-27-csend-bus-universel-design.md)
- Ce qui reste, et pourquoi : [`docs/NEXT.md`](docs/NEXT.md)
- Où publier : [`docs/PUBLISHING.md`](docs/PUBLISHING.md)

---

**Auteur** : Aïssa BELKOUSSA · contact@aissabelkoussa.fr · Licence MIT
