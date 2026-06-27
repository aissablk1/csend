#!/bin/sh
# csend — installeur. POSIX sh, auditable (lis-le avant de l'exécuter).
#   curl -fsSL https://csend.dev/install.sh | sh
# Télécharge le binaire pré-construit correspondant à ta plateforme depuis les
# GitHub Releases, le rend exécutable et le place dans ~/.local/bin.
# Aucun sudo, aucune télémétrie, aucune dépendance autre que curl/tar.
set -eu

REPO="aissablk1/csend"
BINDIR="${CSEND_BINDIR:-$HOME/.local/bin}"

os="$(uname -s | tr '[:upper:]' '[:lower:]')"
arch="$(uname -m)"
case "$arch" in
  x86_64|amd64) arch="amd64" ;;
  arm64|aarch64) arch="arm64" ;;
  *) echo "csend: architecture non supportée: $arch" >&2; exit 1 ;;
esac
case "$os" in
  darwin|linux) ;;
  *) echo "csend: OS non supporté par le binaire ($os). Essaie: go install github.com/$REPO@latest" >&2; exit 1 ;;
esac

asset="csend_${os}_${arch}.tar.gz"
url="https://github.com/$REPO/releases/latest/download/$asset"

echo "csend: téléchargement $os/$arch…"
tmp="$(mktemp -d)"
trap 'rm -rf "$tmp"' EXIT
if ! curl -fsSL "$url" -o "$tmp/csend.tar.gz"; then
  echo "csend: aucune release trouvée. Installe depuis la source :" >&2
  echo "  go install github.com/$REPO@latest" >&2
  exit 1
fi
tar -xzf "$tmp/csend.tar.gz" -C "$tmp"
mkdir -p "$BINDIR"
mv "$tmp/csend" "$BINDIR/csend"
chmod +x "$BINDIR/csend"

echo "✓ csend installé : $BINDIR/csend"
case ":$PATH:" in
  *":$BINDIR:"*) ;;
  *) echo "  Ajoute $BINDIR à ton PATH :  export PATH=\"$BINDIR:\$PATH\"" ;;
esac
"$BINDIR/csend" help >/dev/null 2>&1 && echo "  Prêt. Lance : csend help"
