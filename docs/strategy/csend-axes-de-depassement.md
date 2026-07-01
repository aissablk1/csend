---
title: "csend — Tous les axes de dépassement (analyse concurrentielle)"
date: 2026-07-01
auteur: Aïssa BELKOUSSA
projet: csend
version: 1.0
statut: validé
tags: [stratégie, concurrence, positionnement, claudekit, hcom]
---

# csend — Tous les axes de dépassement

> Teardown concurrentiel (skill `competitive-teardown`), ancré sur des faits vérifiés le
> 2026-07-01 (§29) : deux agents de recherche + vérification directe des sources primaires.

## 0. Bottom line (à lire en premier)

Deux vérités qui **recadrent la mission** :

1. **« Surpasser ClaudeKit » est un faux combat.** ClaudeKit (theclaudekit.com) et les repos
   `claudekit`/awesome-skills sont des **packs de capacités intra-session** (slash-commands, skills,
   subagents *reviewers read-only*). Leur « team » est une **métaphore de packaging** — verbatim :
   *« They inspect and report. They do not orchestrate other agents… There are no reviewer gates that
   block completion. »* csend (un **bus de coordination**) n'est pas dans leur catégorie ; il ne les
   « bat » pas, il résout un problème qu'ils ne touchent pas.

2. **Le vrai concurrent existe : `hcom` (aannoo/hcom).** Le créneau « bus inter-sessions,
   cross-provider, chiffré » **n'est pas vide**. hcom fait déjà : live inter-sessions, **10 CLIs**
   (Claude, Codex, Gemini, Cursor, Copilot…), cross-machine (relay MQTT), **chiffré E2E**
   (XChaCha20-Poly1305), MIT, ~363★, v0.7.22. Plus un petit peloton en clair/localhost (MCP Agent
   Mail, murmur, claude-code-inter-session, Agent Teams natif).

**Donc le dépassement ne se joue pas contre ClaudeKit, mais contre hcom — et sur un axe précis : la
CONFIANCE cryptographique.** hcom admet lui-même sa faiblesse : *« Sender identity is routing
metadata, not authorization »* (PSK partagé, pas d'auth d'expéditeur, pas de post-quantique, pas de
recovery). **C'est là que csend est structurellement supérieur.**

---

## 1. Clarification de catégorie (ne pas confondre)

| Catégorie | Ce que c'est | Exemples | csend concurrent ? |
|---|---|---|---|
| **Packs de capacités** | *ce que l'agent sait faire* (skills/commandes/subagents intra-session) | ClaudeKit (6 kits, 14,99–49,99 $/mois, closed-source), duthaho/carlrannaberg/mrgoonie, awesome-lists (337→1000+ skills) | **Non** — autre catégorie |
| **Bus de coordination** | *comment des sessions vivantes se parlent* | **csend**, **hcom**, MCP Agent Mail, murmur, Agent Teams natif | **Oui** — c'est le terrain de csend |

Corollaire : accumuler des skills pour « dépasser » les kits est le **piège** (§57) — un solo OSS ne
gagne pas une guerre de contenu contre une boîte + 1000+ repos. csend gagne sur le **transport**.

## 2. Le vrai champ concurrentiel (bus inter-sessions)

| Outil | Cross-provider | Cross-machine | Chiffré | **Auth expéditeur** | Post-Q | Recovery | Licence | Maturité |
|---|:---:|:---:|:---:|:---:|:---:|:---:|---|---|
| **csend** | ◐ (Claude ✅ ; Codex/Gemini calibrés) | ✅ (LAN, TLS hybride PQC) | ✅ **asym. par destinataire** | ✅ **Ed25519 signé** | ✅ **ML-KEM-768** | ✅ **Shamir + BIP-39** | Apache-2.0 | alpha v0.2.0 |
| **hcom** | ✅ **10 CLIs** | ✅ MQTT | ✅ (PSK sym.) | ❌ *« not authorization »* | ❌ | ❌ | MIT | ~363★, v0.7.22 |
| MCP Agent Mail | ✅ | UNKNOWN | ❌ | ❌ | ❌ | ❌ | MIT | — |
| murmur | ✅ | ❌ localhost | ❌ HTTP | ❌ | ❌ | ❌ | Apache-2.0 | ~24★ |
| Agent Teams (natif) | ❌ | ❌ inbox fichiers | ❌ | ❌ | ❌ | ❌ | — | expérimental |

**Lecture :** hcom domine la **largeur** et la **maturité** ; csend domine la **confiance
cryptographique** (auth, PQC, recovery) et la **souveraineté**. Ce sont deux profils miroirs.

## 3. Tous les axes, scorés (csend vs hcom, 1-5, honnête)

| # | Axe | csend | hcom | Note |
|---|---|:---:|:---:|---|
| 1 | Modèle de coordination (live inter-sessions) | 5 | 5 | Les deux le font vraiment |
| 2 | **Largeur cross-provider** | 3 | **5** | hcom = 10 CLIs auto ; csend = Claude solide, Codex/Gemini calibrés non-confirmés live |
| 3 | Cross-machine | 4 | 5 | hcom MQTT mûr ; csend LAN + TLS hybride PQC (phase 3) |
| 4 | Chiffrement des messages | **5** | 4 | csend asymétrique **par destinataire** ; hcom **PSK partagé** (un secret = tout le relay) |
| 5 | **Auth d'expéditeur (provenance)** | **5** | **2** | **Killer axis.** csend signe Ed25519 ; hcom : *« sender identity is not authorization »* |
| 6 | Post-quantique | **5** | 1 | csend ML-KEM-768 hybride ; hcom : non mentionné |
| 7 | Gestion de clés + recovery | **5** | 2 | csend identités/vault/Shamir/BIP-39 ; hcom PSK unique, pas de recovery |
| 8 | Intégration turnkey (livraison) | 3 | **5** | hcom hooks auto sur 10 CLIs ; csend inbox universel + injection clavier Unix (plus de friction) |
| 9 | Livraison state-aware / sûreté | 4 | 4 | csend safety-first (jamais valider un y/N) ; hcom statut active/blocked + inject via hooks |
| 10 | Souveraineté / coût | **5** | 4 | Les deux libres ; csend relay auto-hébergé + PQC, cadrage souverain |
| 11 | **Maturité / adoption** | 2 | **4** | hcom établi (~363★, 10 CLIs) ; csend alpha d'un jour, zéro adoption |
| 12 | Brand / docs / polish | 3 | 4 | hcom rodé ; csend docs bonnes mais neuf |

**Profil csend :** exceptionnel sur **4-5-6-7-10** (confiance/PQC/souveraineté), en **retard sur
2-8-11-12** (largeur, intégration turnkey, maturité, notoriété).

## 4. SWOT (csend vs hcom)

- **Forces** : seul bus avec **auth cryptographique d'expéditeur + chiffrement par destinataire +
  post-quantique + recovery souveraine**. Zéro dépendance, Apache-2.0.
- **Faiblesses** : largeur providers (Codex/Gemini non confirmés live), intégration plus friction
  (injection clavier vs hooks), **alpha sans adoption**, pas de relay hébergé clé-en-main.
- **Opportunités** : hcom **admet** le trou d'auth ; aucun bus cross-provider n'a **PQC + provenance
  signée**. Niche « haute assurance / zéro-confiance / multi-tenant / régulé / adversarial ».
- **Menaces** : hcom peut ajouter de la signature/PQC ; l'avance de largeur+maturité de hcom se
  creuse ; Agent Teams natif d'Anthropic tire le marché vers le gratuit-intégré (Claude-only).

## 5. Positionnement — le wedge

**hcom = le bus multi-CLI facile « à mot de passe partagé ».**
**csend = le bus d'agents cryptographiquement authentifié, post-quantique et souverain** — pour qui
**ne peut pas** faire confiance à un PSK partagé (environnements régulés, multi-tenant, sensibles,
adversariaux). *« Un mot de passe partagé ne dit pas QUI parle. csend, si — signé, chiffré par
destinataire, résistant au quantique. »*

```
        Haute assurance crypto
                 ▲
          csend ●│
                 │
 Étroit ◀────────┼────────▶ Large (providers/maturité)
                 │● hcom
                 │  ● Agent Mail / murmur / Agent Teams
                 ▼
        Confiance « mot de passe »
```

## 6. Plan d'action — « Tous » (mappé aux 4 directions demandées)

| Horizon | Action | Direction |
|---|---|---|
| **Quick (0-2 sem)** | Recadrer le pitch csend autour de **provenance signée + PQC + souveraineté** (PAS « vs ClaudeKit »). Page de comparaison **honnête vs hcom** (forces ET retards). Rendre lisible l'axe « auth d'expéditeur ». | Moat |
| **Quick** | Cesser de se mesurer à ClaudeKit (mirage) ; citer hcom comme la vraie référence. | Moat |
| **Medium (1-3 mois)** | **Fermer le gap de largeur** : livraison par **hooks** (à la hcom) en plus de l'injection clavier → auto-delivery multi-CLI. **Confirmer Codex/Gemini en live**. | Frontal (vs hcom) |
| **Medium** | Couche kit **mince** : quelques commandes csend-natives d'orchestration (dispatch, hand-off, coordination inter-sessions signée). Pas 5 verticales. | Kit mince |
| **Strategic (3-12 mois)** | Posséder la niche **haute assurance** : modèle de menace publié, permissions par session, ordre/anti-rejeu, relay souverain PQC clé-en-main. Cible : équipes sécurité/régulées/multi-tenant. | Moat + frontal |

## 7. Verdict honnête

csend **peut** surpasser — mais **sur l'axe confiance/sécurité/souveraineté**, pas sur la largeur ni
le contenu. Deux corrections de cap indispensables (§29/§66) :

1. **Arrêter de viser ClaudeKit** : catégorie différente, combat qui ne prouve rien.
2. **Mesurer csend contre `hcom`**, le vrai concurrent — et **assumer les retards** (largeur,
   maturité, intégration) pour les combler, tout en martelant l'avance unique (provenance signée +
   PQC + recovery souveraine, que hcom n'a **pas** et admet ne pas avoir).

**Le dépassement réel de csend = devenir le bus d'agents que l'on peut auditer et auquel on peut faire
confiance dans un environnement hostile — le seul avec une vraie identité cryptographique.**
