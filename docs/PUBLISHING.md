# Où publier csend — recommandation

Objectif : rendre csend **trivial à installer** (humains + agents) et **visible**.
Tout est gratuit et souverain-compatible (§58). Ordre = priorité.

## 1. Le code — GitHub (canonique)

- **Repo public** : `github.com/aissablk1/csend` (MIT). C'est la source de vérité.
- Authorship Git en `noreply` GitHub — **jamais** le Gmail perso (§35). Contact public =
  `contact@aissabelkoussa.fr`.
- Active **GitHub Releases** : un `git tag vX.Y.Z` déclenche GoReleaser
  (`.goreleaser.yaml`) → binaires darwin/linux/windows × amd64/arm64 + `checksums.txt`.

## 2. Les canaux d'installation (du plus simple au plus brut)

| Canal | Commande | Mise en place |
|---|---|---|
| **Homebrew** | `brew install aissablk1/tap/csend` | créer le repo `github.com/aissablk1/homebrew-tap` ; GoReleaser y pousse la formule |
| **Go** | `go install github.com/aissablk1/csend@latest` | déjà prêt (module path posé) |
| **Script** | `curl -fsSL https://csend.dev/install.sh \| sh` | `site/install.sh` tire le binaire des Releases |
| **Source** | `git clone … && make install` | déjà prêt |

> `go install` marche **dès** que le repo est public. Homebrew et le script marchent dès
> la **première Release** taguée.

## 3. Le site — hébergement statique

`site/` est 100 % statique (HTML/CSS/JS + GSAP CDN) → héberge où tu veux :

- **Cloudflare Pages** (recommandé, §58 — gratuit, rapide, domaine custom `csend.dev`).
- **GitHub Pages** (zéro infra : `docs/` ou branche `gh-pages` ; pointer sur `site/`).
- **Vercel** / **Netlify** (drag-and-drop du dossier `site/`).

Domaine suggéré : **`csend.dev`** (ou `csend.sh`). Mets l'URL réelle dans `index.html`
(`canonical`, `og:`) et dans `install.sh` (`https://csend.dev/install.sh`).

## 4. Distribution (faire connaître)

- **Reddit** : r/ClaudeAI, r/ClaudeCode, r/commandline, r/golang.
- **Hacker News** : « Show HN: csend — message a running Claude/Codex/Gemini CLI session ».
- **Listes curées** : PR vers `awesome-claude-skills` (ComposioHQ), `awesome-claude-code`,
  `awesome-go`, `awesome-cli-apps`.
- **Product Hunt** (le jour d'une vraie release stable).
- **X / LinkedIn** : la vidéo de lancement via le skill `brag` (§60) à partir de `site/`.

## 5. Honnêteté (§2/§29/§34)

csend est **alpha**. Sur le site et les annonces : pas de témoignage/étoile/compteur
inventé. La preuve = code open source + architecture vérifiable + roadmap honnête.
Annoncer « alpha », lister ce qui marche vs ce qui arrive (déjà fait sur le site).

**Auteur** : Aïssa BELKOUSSA
