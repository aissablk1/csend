# Politique de sécurité — csend

csend est un **bus de messages chiffré de bout en bout** pour agents de code. La
sécurité est le cœur du projet, pas une option. Ce document décrit le modèle de menace,
les primitives réellement employées, leurs limites honnêtes, et **comment signaler une
faille**.

> [!IMPORTANT]
> **Statut : alpha (v0.2.0-dev).** Les primitives sont auditées et standard, mais
> **l'assemblage cryptographique de csend n'a pas encore reçu d'audit externe**.
> N'utilisez pas csend — en particulier le partage Shamir — pour protéger des secrets
> critiques (clés de production, fonds, données vitales) **avant cet audit**.

---

## Signaler une vulnérabilité

**N'ouvrez pas d'issue publique pour une faille de sécurité.** Utilisez la
**divulgation responsable privée de GitHub** :

> Onglet **Security** du dépôt → **Report a vulnerability** (GitHub Private Vulnerability
> Reporting).

C'est le canal privilégié : il garde l'échange privé jusqu'à un correctif et un avis
coordonné. Merci de **ne pas** divulguer publiquement avant qu'un correctif soit
disponible.

Ce qui aide un bon rapport : version (`git rev-parse HEAD` ou tag), OS/arch, étapes de
reproduction, impact estimé, et — si possible — un test qui échoue. Nous visons un
accusé de réception rapide et un échange suivi jusqu'à résolution.

### Périmètre

Sont **dans le périmètre** : la couche cryptographique (`crypto.go`, `shamir.go`,
`bip39.go`, `tlsbus.go`), la gestion du vault et des identités (`bus.go`, `keyring.go`,
`recovery.go`), le transport réseau (`net.go`), et toute fuite de plaintext, de clé
privée ou de secret de vault.

Sont **hors périmètre** (limites assumées, pas des failles) : l'injection clavier sur
Windows/mobile (impossible par conception), la compromission d'un terminal déjà sous
contrôle de l'attaquant, et les fonctionnalités non encore livrées (voir « Limites
connues »).

---

## Modèle de menace

### Ce que csend protège

| Bien | Contre qui | Comment |
|---|---|---|
| **Confidentialité des messages** | un relais/bus curieux ou malveillant, un sniffer réseau | chiffrement E2E hybride (X25519 ⊕ ML-KEM-768) → AES-256-GCM ; le relais ne voit que du chiffré |
| **Intégrité & authenticité** | un relais qui altère ou rejoue un message falsifié | signature Ed25519 sur l'intégralité de l'enveloppe, **vérifiée avant déchiffrement** |
| **Résistance « Harvest Now, Decrypt Later »** | un adversaire qui capture aujourd'hui pour déchiffrer avec un futur ordinateur quantique | KEM **hybride** : casser le secret exige de casser **X25519 ET ML-KEM-768** |
| **Clés privées au repos** | accès au disque / sauvegarde volée | vault scellé AES-256-GCM, clé dérivée par PBKDF2-SHA256 (600 000 itérations) |
| **Perte d'un appareil** | un seul device perdu/volé | recovery Shamir K-sur-N : K-1 parts ne révèlent rien (sûreté de seuil) |

### Hypothèses de confiance

- L'**endpoint local** est de confiance : si la machine qui détient le vault **et** la
  passphrase est compromise, l'attaquant a l'identité. csend protège les messages **en
  transit** et les clés **au repos**, pas un poste déjà sous contrôle adverse.
- La **passphrase de vault** a une entropie suffisante : PBKDF2 ralentit le bruteforce,
  il ne sauve pas un mot de passe trivial.
- Sur le **réseau**, la version actuelle vise le **loopback / LAN de confiance**
  (voir « Limites »).

---

## Primitives cryptographiques (réellement employées)

Toutes proviennent de la **bibliothèque standard de Go 1.24** — aucune dépendance
externe, aucune crypto maison.

| Rôle | Primitive | Paquet stdlib | Notes |
|---|---|---|---|
| Signature | **Ed25519** | `crypto/ed25519` | signe l'enveloppe `(eph‖mlkem_ct‖nonce‖ct)`, vérifiée avant déchiffrement |
| KEM classique | **X25519** | `crypto/ecdh` | clé **éphémère par message** (confidentialité persistante côté classique) |
| KEM post-quantique | **ML-KEM-768** | `crypto/mlkem` | FIPS 203 ; encapsulation vers la clé statique du destinataire |
| Chiffrement | **AES-256-GCM** | `crypto/aes` + `crypto/cipher` | nonce 96 bits aléatoire ; clé **fraîche par message** (pas de réutilisation de nonce) |
| Dérivation de clé | **HKDF-SHA256** | `crypto/hkdf` | dérive la clé AEAD de `X25519_shared ‖ ML-KEM_shared`, et l'identité de la graine maître (séparation de domaine) |
| Dérivation de vault | **PBKDF2-SHA256** | `crypto/pbkdf2` | 600 000 itérations, sel 128 bits |
| Aléa | **CSPRNG** | `crypto/rand` | graines, nonces, sels, polynômes Shamir |
| Recovery (seuil) | **Shamir N-sur-M** sur GF(2⁸) | `shamir.go` (from-scratch, testé) | corps AES (0x11b), interpolation de Lagrange en 0 |
| Recovery (phrase) | **BIP-39** (24 mots) | `bip39.go` (from-scratch, testé) | wordlist anglaise officielle, checksum SHA-256 |
| Transport réseau | **TLS 1.3 hybride PQC** | `crypto/tls` (`X25519MLKEM768`) | cert self-signed Ed25519 + épinglage d'empreinte SHA-256 |

### Construction du message scellé

1. Le secret de session = `HKDF-SHA256(X25519_shared ‖ ML-KEM_shared)` → clé AES-256.
2. `AES-256-GCM(plaintext)` avec nonce aléatoire.
3. **Signature Ed25519** de l'expéditeur sur `(eph_pub ‖ mlkem_ct ‖ nonce ‖ ct)`.
4. À l'ouverture : la signature est **vérifiée d'abord** (sinon rejet), puis double
   décapsulation X25519 + ML-KEM, puis déchiffrement GCM. Une enveloppe falsifiée est
   rejetée avant tout déchiffrement.

### Identité et vault

Toute l'identité (Ed25519 + X25519 + ML-KEM) **dérive d'une seule graine maître de 32
octets** par HKDF à domaines séparés. Cette graine est l'unique secret : elle est
scellée dans le vault, découpable en parts Shamir, ou encodable en phrase BIP-39. Le
vault sur disque a les permissions `0600`.

---

## Limites connues (honnêteté)

Ces points sont des **limites assumées de l'alpha**, documentés ici plutôt que cachés :

- **Pas d'audit cryptographique externe.** Les primitives sont standard, mais leur
  assemblage (et les implémentations Shamir / BIP-39 from-scratch, quoique testées par
  roundtrip et propriété de seuil) **doivent être audités** avant tout usage critique.
- **PBKDF2, pas encore Argon2id.** Le vault utilise PBKDF2-SHA256 (600 000 itérations).
  Argon2id (résistant au matériel dédié) est la cible, mais exige `golang.org/x/crypto`
  — donc une première dépendance, à arbitrer.
- **Pas de passkey/WebAuthn.** Le déverrouillage du vault repose sur une passphrase
  (fichier via `CSEND_VAULT_PASS_FILE`, recommandé, ou variable `CSEND_VAULT_PASS`). Le
  MFA résistant au phishing (FIDO2/PRF) est planifié.
- **Réseau : confiance loopback/LAN.** `serve`/`remote` chiffrent le transport en TLS
  1.3 hybride PQC avec épinglage d'empreinte, mais **sans authentification mutuelle des
  pairs** et avec `InsecureSkipVerify` (la vérification se fait par épinglage manuel ;
  un pin **vide** sur loopback accepte n'importe quel certificat). **Ne pas exposer hors
  loopback/LAN de confiance** sans le durcissement de la phase 3.
- **Pas d'anti-rejeu au niveau crypto.** L'enveloppe scellée ne lie ni horodatage ni
  compteur ; la déduplication se fait au niveau de l'inbox (par identifiant de message),
  pas par la signature. Un anti-rejeu explicite est à ajouter.
- **Confidentialité persistante partielle.** Le X25519 est éphémère par message
  (forward secrecy côté classique) ; la décapsulation ML-KEM vise la clé **statique** du
  destinataire (pas de forward secrecy côté post-quantique).
- **Repli en clair en local.** Sans contact connu, un message coopératif local part en
  clair (confiance locale assumée). Le scellement n'est appliqué que si la clé publique
  du destinataire est enregistrée et le vault déverrouillable.

---

## Bonnes pratiques d'usage

- Préférez **`CSEND_VAULT_PASS_FILE`** à `CSEND_VAULT_PASS` (un fichier évite de fuiter
  le secret dans `ps`/dumps d'environnement).
- **Sauvegardez** la phrase BIP-39 **hors ligne** (papier, coffre) : quiconque la
  possède possède l'identité.
- **Répartissez** les parts Shamir sur des supports distincts (téléphone, clé matérielle,
  proche, papier) — la perte d'un support n'est pas la perte de l'identité.
- N'exposez **jamais** `csend serve` hors d'un réseau de confiance tant que la phase 3
  (auth mutuelle) n'est pas livrée.

---

## Versions supportées

csend est en alpha : seule la **branche `main`** (dernier commit) reçoit des correctifs
de sécurité. Le support par version stable commencera avec la première série taguée
stable.

| Version | Support sécurité |
|---|:---:|
| `main` (0.2.0-dev) | ✅ |
| versions antérieures non taguées | ❌ |

---

**Auteur** : Aïssa BELKOUSSA · signalement via GitHub Private Vulnerability Reporting
