# csend v0.2.0 — premier socle public (alpha)

**Le bus de messages chiffré de bout en bout qui fait parler tes agents de code entre
eux.** Une session d'agent CLI (Claude Code, Codex, Gemini…) écrit à une **autre session
en cours** — en lisant son état avant d'agir, à travers terminaux, providers et
machines, chiffré et résistant au quantique. Go, **zéro dépendance**, MIT.

> **Alpha.** Le code est testé (`go test -race ./...` vert) et sans dépendance, mais
> l'API peut évoluer et la cryptographie **n'a pas encore reçu d'audit externe**. À
> explorer et éprouver, pas encore à confier des secrets critiques. Voir
> [SECURITY.md](SECURITY.md).

## Points forts

- **Bus coopératif universel.** Inbox par fichiers, durable, sans démon : `register`,
  `inbox`, `recv`, `agents`, `whoami`. Marche sur **tout OS et tout provider**.
- **Injection live state-aware (Unix).** Backends **cmux** + **tmux** ; csend lit l'état
  de la cible (`● idle` / `◐ busy` / `⚠ confirm`) **avant** d'écrire — là où les autres
  injectent à l'aveugle. Graphe familial + broadcasts.
- **Chiffrement E2E hybride post-quantique.** Signé **Ed25519**, chiffré **AES-256-GCM**
  sous **X25519 ⊕ ML-KEM-768** : il faut casser **les deux** échanges de clés pour lire
  (résistance « Harvest Now, Decrypt Later »). Le relais ne voit que du chiffré.
- **Vault + recovery.** Identité dérivée d'une graine maître, scellée AES-256-GCM
  (PBKDF2-SHA256, 600 000 itérations) ; recovery par **Shamir N-sur-M** (GF(2⁸)) et
  **phrase BIP-39** de 24 mots.
- **Réseau machine-à-machine.** `serve`/`remote` avec **TLS 1.3 hybride PQC**
  (`X25519MLKEM768`) et épinglage d'empreinte. Cible : loopback/LAN.
- **Zéro dépendance.** Toute la crypto vient de la stdlib Go 1.24 — rien à auditer en
  amont, rien qui casse.

## Installation

```sh
# Go (toute plateforme avec Go 1.24+)
go install github.com/aissablk1/csend@latest

# Homebrew (macOS · Linux)
brew install aissablk1/tap/csend

# Script (à auditer — tire le binaire des Releases)
curl -fsSL https://csend.dev/install.sh | sh

# Source (binaire universel arm64 + x86_64 sur macOS)
git clone https://github.com/aissablk1/csend && cd csend && make install
```

Binaires fournis par la release : **darwin / linux / windows** × **amd64 / arm64** +
`checksums.txt` (via GoReleaser).

## Démarrage

```sh
csend register                               # rejoindre le bus (tout OS/provider)
csend inbox SACEM "lance le build de prod"   # écrire à une autre session
csend recv                                   # relever son inbox
csend list && csend tree                     # flotte Unix + état + graphe familial
CSEND_VAULT_PASS=… csend id --create         # identité chiffrée
csend recovery split 2 3                      # 3 parts Shamir, seuil 2-sur-3
```

## Limites connues (honnêteté)

- **Injection clavier = Unix uniquement.** Windows natif et mobile sont
  **coopératif-only** (pas de PTY maître partageable / sandbox) — limite physique, pas un
  manque.
- **Cross-provider partiel.** Détection d'état **Claude livrée** ; adaptateurs **Codex /
  Gemini = à venir** (en attente de vrais écrans à calibrer).
- **Mobile = phase ultérieure.** Le mobile rejoindra le bus comme **client** (voir,
  approuver, broadcaster), il n'injecte pas au clavier.
- **Réseau = loopback/LAN de confiance.** TLS hybride + épinglage, mais **sans auth
  mutuelle** des pairs pour l'instant — ne pas exposer hors loopback sans durcir.
- **Crypto non auditée en externe.** Implémentations Shamir / BIP-39 from-scratch testées
  (roundtrip + propriété de seuil) mais **à auditer avant usage critique**.
- Vault par **PBKDF2** (Argon2id ciblé), **pas de passkey** ni d'anti-rejeu crypto
  encore. Détails : [SECURITY.md](SECURITY.md).

## Pour la suite

Adaptateurs Codex/Gemini, passkey WebAuthn, durcissement réseau (auth mutuelle), clients
mobiles, surface MCP, signatures ML-DSA, et audit cryptographique externe. Voir
[CHANGELOG.md](CHANGELOG.md) et [docs/NEXT.md](docs/NEXT.md).

## Signaler une faille

Via **GitHub Private Vulnerability Reporting** (onglet Security → Report a vulnerability),
jamais en issue publique. Voir [SECURITY.md](SECURITY.md).

---

**Auteur** : Aïssa BELKOUSSA · Licence MIT
