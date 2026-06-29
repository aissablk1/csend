---
title: "csend — Playbook d'adoption open source"
date: 2026-06-27
auteur: Aïssa BELKOUSSA
projet: csend
version: 1.0
statut: validé
tags: [oss, adoption, distribution, securite, csend, go]
---

# csend — Playbook d'adoption open source

> Cap retenu (2026-06-27) : csend est **viable comme outil open source**, pas comme
> business. Objectif = adoption, diffusion, confiance — pas de chiffre d'affaires.
> Ce playbook remplace le « founder pack » de vente par un plan de lancement OSS.

---

## 1. Verdict de viabilité

**Technique : solide.** Phase 1 production-ready (`go build` ✅, `go test ./...` ✅),
32 fichiers Go dont 12 de tests, CI/CD GitHub Actions, **zéro dépendance externe**
(stdlib Go 1.24), licence MIT. Ce n'est pas un prototype.

**Stratégique : dépend du positionnement.** « Orchestrateur Claude Code » = doublon
du natif (Agent Teams, février 2026) et de ruflo → faible. « Bus souverain E2E,
cross-provider, cross-OS, avec vault + recovery » = niche réelle et défendable.

**Risque marché à garder en tête.** La valeur migre vers (a) le natif/plateforme
(Anthropic internalise) et (b) les standards d'interop (A2A/MCP/ACP/ANP, Linux
Foundation, 150+ orgs). Un bus tiers *générique* est comprimé entre les deux ; un bus
*sécurisé et cross-provider* occupe l'angle qu'aucun des deux ne couvre.

---

## 2. Positionnement

**One-liner :** « Le bus de messages chiffré de bout en bout, souverain et
post-quantique, pour agents de code — cross-provider, cross-OS, avec injection dans
les sessions en cours. »

**Comparatif (ce que csend fait seul) :**

| Capacité | csend | Agent Teams (natif) | ruflo | tmux-orchestrator |
|---|:---:|:---:|:---:|:---:|
| Cross-provider (Claude/Codex/Gemini) | ✅ (Claude ✅, autres phase 2) | ❌ Claude only | ◐ | ❌ |
| Cross-OS + mobile (client) | ✅ | ❌ desktop | ❌ | ❌ Unix only |
| Chiffrement E2E des messages | ✅ | ❌ sandbox/vault partagés | ❌ | ❌ clair |
| Post-quantique (ML-KEM-768 hybride) | ✅ | ❌ | ❌ | ❌ |
| Vault + recovery (Shamir, BIP-39) | ✅ | ❌ | ❌ | ❌ |
| Injection dans une session déjà lancée | ✅ (Unix) | ◐ cadre Agent Teams | ◐ | ◐ via panes |
| Zéro dépendance externe | ✅ | n/a (natif) | ❌ npm | ◐ bash |

> Honnêteté (§29) : Windows natif (injection) = ❌ ; mobile/passkey/multi-machine =
> phases 2-4 ; implémentation Shamir = audit crypto externe recommandé avant usage
> critique. Ne pas survendre — le comparatif doit rester vérifiable.

---

## 3. Distribution — où publier (classé par effet / effort)

1. **GitHub, base camp.** README qui prouve la valeur en 60 s, topics
   (`claude-code`, `ai-agents`, `mcp`, `e2e-encryption`, `post-quantum`, `tmux`,
   `cli`), releases taguées avec **binaires pré-compilés** (arm64 + x86_64).
2. **Installation friction-zéro.** `go install github.com/aissablk1/csend@latest`
   (fonctionne déjà) + **tap Homebrew** (`homebrew-csend`) + `install.sh` (fait).
3. **Awesome-lists.** PR sur *awesome-claude-code*, *awesome-go*, et *awesome-mcp*
   (le jour où la surface MCP ship).
4. **Communautés.** Show HN (Hacker News), r/ClaudeAI, r/commandline, Discord
   Anthropic / Claude Code, X (thread dev). Angle = sécurité souveraine, pas
   « énième orchestrateur ».
5. **Contenu.** Vidéo de lancement (skill `brag`), post technique sur le **modèle de
   sécurité** (E2E + PQC hybride + Shamir), site landing (optionnel, fort effort).
6. **Registre MCP.** Publier la surface MCP quand elle est livrée.

---

## 4. Qualité & signaux de confiance (ce qui fait adopter un OSS)

- **README structuré :** valeur → quickstart 60 s → modèle de sécurité → comparatif
  → roadmap honnête (matrice OS sans survente).
- **Confiance :** `SECURITY.md` (modèle crypto + « audit externe en attente »),
  `CONTRIBUTING.md`, `CODE_OF_CONDUCT.md`, templates issues/PR, `good-first-issues`.
- **Badges :** CI, version Go, licence MIT, Go Report Card.
- **Releases :** versioning sémantique + `CHANGELOG.md` + tag `v0.2.0`.

---

## 5. Onboarding — friction zéro, humain ET IA

- **Humain :** une commande (`brew install` / `go install` / `curl … | sh`),
  puis `csend --help` et un quickstart qui marche du premier coup.
- **IA :** surfaces skill + hook + MCP pour qu'un agent s'auto-installe et s'auto-
  enregistre (`csend register`) sans intervention humaine.
- Règle : **chaque étape d'install retirée = de l'adoption gagnée** (§57).

---

## 6. Métrique d'adoption (1 KPI, pas du vanity)

KPI principal : **installs hebdomadaires** (téléchargements de release +
imports pkg.go.dev). Secondaire : nombre d'agents distincts enregistrés. Les étoiles
GitHub = signal faible, à ne pas piloter.

---

## 7. Plan 30 jours (lancement OSS)

| Semaine | Focus | Livrables |
|---|---|---|
| **S1** | Fondations de confiance | README v2 + comparatif + `SECURITY.md` + tag `v0.2.0` (binaires) + tap Homebrew |
| **S2** | Lancement | Vidéo `brag` + Show HN + r/ClaudeAI + thread X + PR awesome-claude-code/awesome-go |
| **S3** | Le moat | Adaptateurs **Codex + Gemini** (phase 2) = preuve vivante du cross-provider |
| **S4** | Amplification | Post « modèle de sécurité PQC/Shamir » + site landing (optionnel) + itération feedback |

---

## 8. TODO — prochaine étape

Choisir la première action concrète :
- [ ] Polir README + `SECURITY.md` + tagger la release `v0.2.0`.
- [ ] Tap Homebrew + vérifier `go install` / `install.sh`.
- [ ] Vidéo de lancement (`brag`).
- [ ] Phase 2 : adaptateurs Codex/Gemini (le vrai argument cross-provider).
- [ ] Site landing (optionnel, niveau Awwwards).

> Garde-fous : matrice OS honnête (§29), pas de fausse preuve d'adoption (§34),
> audit crypto externe avant de présenter Shamir comme « production-grade » (§38).
