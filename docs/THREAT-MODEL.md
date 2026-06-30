# csend — Modèle de menace (honnête)

> Ce document dit ce que csend protège, contre qui, et **ce qu'il ne protège pas**.
> Principe : zéro fausse promesse de sécurité (cf. Kerckhoffs ; aucune sécurité par
> l'obscurité). La cryptographie est **implémentée mais N'A PAS été auditée par un
> tiers** — ne pas la traiter comme « Signal-grade » tant qu'un audit n'a pas eu lieu.

## Ce que csend EST

Un bus de messagerie inter-agents **local**, qui transporte des messages entre sessions
de CLI d'agents (Claude Code, Codex, Gemini…) sur la **même machine** (voie coopérative
par fichiers) ou entre machines (`serve`/`remote`).

## Actifs à protéger

1. **Le contenu des messages** entre agents (peut contenir du code, des décisions, des secrets transitoires).
2. **L'authenticité de l'expéditeur** (savoir quel agent/identité a réellement envoyé un message).
3. **L'intégrité de l'enveloppe** (un message ne doit pas être altéré ni ré-attribué).

## Mécanismes en place

| Menace (STRIDE) | Mitigation dans csend |
|---|---|
| **Spoofing** (usurper un expéditeur) | Signature **Ed25519** sur le transcript ; identité dérivée d'une graine maître. Allowlist cryptographique (`serve --authz`) : seuls les expéditeurs signés et autorisés passent. |
| **Tampering** (altérer le message) | AEAD **AES-256-GCM** + signature sur le transcript (eph keys, nonce, ciphertext, **SenderPub**, **AAD from→to**). Toute altération invalide la signature/AEAD. |
| **Repudiation** | Journal append-only (hash + longueur, jamais le clair) ; `csend journal` trace de→à sans révéler le corps. |
| **Information disclosure** | Chiffrement **E2E hybride post-quantique** : KEM X25519 ⊕ ML-KEM-768 → HKDF → AES-GCM (confidentialité tient sauf si **les deux** KEM sont cassés). Vault PBKDF2→AES-GCM. |
| **Replay / ré-emballage** | Anti-replay (dédup sur le **nonce signé**). **AAD from→to** liée dans l'AEAD ET la signature (§41) : un payload ré-emballé sous un autre couple expéditeur→destinataire est rejeté. |
| **Réseau** | TLS 1.3 **hybride PQC** (X25519MLKEM768) + pinning d'empreinte sur `serve`/`remote`. |

## Limites HONNÊTES (ce que csend NE protège PAS)

- **Crypto NON auditée.** Implémentation soignée (stdlib Go, primitives NIST), mais **aucun audit
  indépendant**. Ne pas vendre « le Signal des agents » : Signal est audité, csend non.
- **Même machine = même utilisateur.** Pour deux process **sous le même UID**, la crypto E2E
  apporte peu : un attaquant qui exécute déjà du code sous cet UID lit la clé, le vault, la mémoire
  du process. La crypto ne prend toute sa valeur qu'en franchissant une **frontière de confiance**
  (inter-hôtes). Sur localhost, c'est surtout du *defense-in-depth*, pas une barrière.
- **L'agent récepteur est une surface d'attaque (prompt injection).** csend **transporte** un message ;
  il ne garantit pas que l'agent qui le reçoit ne sera pas détourné par son contenu. La défense
  prompt-injection est au niveau du **harnais/modèle** (Claude Code…), **pas** du bus. Un champ
  « ceci est de la donnée » dans l'enveloppe n'empêche pas l'agent de le concaténer dans son prompt.
- **Le bus ne voit pas tout le trafic.** Les agents communiquent aussi par fichiers partagés, par le
  canal natif de l'orchestrateur, par MCP, par stdout. csend n'est pas un moniteur de référence.
- **Métadonnées.** Le journal/registre expose qui parle à qui et quand (pas le contenu). Pas d'anonymat.
- **Quantique.** La couche KEM est hybride PQC (bon contre « Harvest Now, Decrypt Later »), mais les
  **signatures** restent Ed25519 (classique) ; migrer en ML-DSA si le modèle de menace l'exige.

## Recommandations d'usage

- Garder `serve` en **loopback** sauf besoin ; hors loopback, **`--authz` + allowlist** obligatoires.
- Ne **jamais** mettre de secret durable dans un message (le bus chiffre le transit, pas le stockage de l'agent).
- Avant tout usage « production / sensible » : **faire auditer la crypto** et ne pas survendre la garantie.

**Auteur** : Aïssa BELKOUSSA
