# csend

**Le bus de messagerie entre tous tes agents CLI.** `csend` injecte un message dans une
**autre** session d'agent (Claude Code, Codex, Gemini…) **en cours** — en lisant son état
avant d'agir (idle / busy / confirmation), à travers terminaux, providers et machines,
**chiffré de bout en bout**. Go, **zéro dépendance**, MIT.

> Aucune API officielle n'injecte un prompt dans une session d'agent CLI vivante. `csend`
> comble ce trou : **inbox coopératif** = colonne vertébrale universelle (tout OS, tout
> provider) ; **injection clavier live** = repli pour les TUI Unix.

/ Statut : **alpha**. Voir « Ce qui marche / ce qui arrive » plus bas et la spec complète. /

## Installation

```sh
# Homebrew (macOS · Linux)
brew install aissablk1/tap/csend

# Go (toute plateforme)
go install github.com/aissablk1/csend@latest

# Script (audite-le d'abord : site/install.sh)
curl -fsSL https://csend.dev/install.sh | sh

# Source (zéro confiance)
git clone https://github.com/aissablk1/csend && cd csend && make install
```

## Démarrage

```sh
# --- Voie coopérative : marche PARTOUT, sans multiplexeur ---
csend inbox SACEM "lance le build de prod"   # une session écrit à une autre
csend recv                                   # cette session relève son inbox

# --- Voie cmux : pilotage state-aware d'une flotte ---
csend list                  # sessions agent + état (idle/busy/confirm)
csend tree                  # graphe familial père → enfants
csend send SACEM "go"       # message gouverné (valide seulement si idle)
csend send --down "build"   # broadcast aux enfants

# --- Identité & recovery par seuil ---
CSEND_VAULT_PASS=… csend id --create     # identité hybride en vault chiffré
csend recovery split 2 3                 # 3 parts Shamir, seuil 2-sur-3
csend recovery combine <part1> <part2>   # reconstitue depuis ≥ 2 parts
```

## Pour les agents IA

csend est exposé comme **skill Claude Code** (`~/.claude/skills/csend`) : une session peut
**écrire à / relever / coordonner** ses paires en langage naturel (« dis à SACEM de… »,
« relève mon inbox », « broadcast à mes enfants »). La voie coopérative (`inbox`/`recv`) ne
dépend d'aucun terminal particulier — idéale pour des flottes d'agents hétérogènes.

## Comment ça marche

8 couches, chacune une responsabilité :

```
0 Sécurité/crypto · 1 Identité+registre · 2 Providers · 3 Transports
4 Routeur · 5 Réseau · 6 Mémoire · 7 Surfaces (CLI/skill/hook/MCP)
```

- **Transports** : inbox coopératif (universel) · injection cmux (✓) / tmux (à venir) · bridges.
- **Routeur** : `inbox › bridge › injection › file` — l'inbox gagne, car durable et sans course.
- **Mémoire** : journal interrogeable des messages (hash, jamais le clair) + registre des sessions.

## Sécurité

Primitives **auditées, jamais maison** (stdlib Go 1.24, zéro dépendance) :

- **E2E hybride PQC** : signé Ed25519, chiffré **X25519 ⊕ ML-KEM-768** (casser *les deux* KEM
  pour lire — anti « Harvest Now, Decrypt Later »).
- **Vault** chiffré AES-256-GCM (passkey WebAuthn à venir).
- **Recovery Shamir** N-sur-M sur GF(2⁸) : K-1 parts ne révèlent *rien*.

## Ce qui marche / ce qui arrive

**Livré & testé** : inbox coopératif · cmux state-aware + graphe familial · crypto E2E + vault ·
recovery Shamir · mémoire/journal.
**En route** : backend tmux/screen · adaptateurs Codex/Gemini · passkey · réseau multi-machine
(TLS hybride PQC) · mobile · bridge Agent Teams · recovery en phrases (BIP-39).

## Liens

- Design & roadmap : [`docs/superpowers/specs/2026-06-27-csend-bus-universel-design.md`](docs/superpowers/specs/2026-06-27-csend-bus-universel-design.md)
- Où publier : [`docs/PUBLISHING.md`](docs/PUBLISHING.md)
- Site : `site/` (statique, hébergeable partout)

---

**Auteur** : Aïssa BELKOUSSA · contact@aissabelkoussa.fr · MIT
