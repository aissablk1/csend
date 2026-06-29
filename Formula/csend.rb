# Formule Homebrew — csend (build Go depuis les sources).
#
# Cette formule COMPILE csend depuis le tarball source d'une release GitHub, plutôt
# que de télécharger un binaire pré-construit. Avantages : reproductible, auditable,
# pas de binaire opaque. Elle vise le tap personnel `aissablk1/homebrew-tap`, où elle
# cohabite avec la formule binaire générée automatiquement par GoReleaser
# (`.goreleaser.yaml`) ; garde celle qui convient à ton usage.
#
# AVANT PUBLICATION (à faire une fois la release v0.2.0 taguée) :
#   1. `git tag v0.2.0 && git push --tags`  (déclenche la release GitHub).
#   2. Récupérer le sha256 du tarball source :
#        curl -fsSL https://github.com/aissablk1/csend/archive/refs/tags/v0.2.0.tar.gz \
#          | shasum -a 256
#   3. Remplacer le PLACEHOLDER `sha256` ci-dessous par cette valeur.
#   4. `brew install --build-from-source ./Formula/csend.rb` pour vérifier en local,
#      puis `brew test csend` et `brew audit --strict csend`.
#
# Statut : alpha (v0.2.0-dev). Tant que la release n'est pas taguée, installe plutôt
# via `go install github.com/aissablk1/csend@latest` ou `make install`.

class Csend < Formula
  desc "Bus de messages inter-agents pour CLI (state-aware, chiffré E2E, zéro dépendance)"
  homepage "https://csend.dev"
  url "https://github.com/aissablk1/csend/archive/refs/tags/v0.2.0.tar.gz"
  sha256 "REMPLACER_PAR_LE_SHA256_DU_TARBALL_SOURCE_v0.2.0"
  license "MIT"
  head "https://github.com/aissablk1/csend.git", branch: "main"

  # csend ne dépend QUE de la toolchain Go pour compiler ; aucune dépendance runtime.
  depends_on "go" => :build

  def install
    # Build reproductible, sans CGO, symboles strippés. -X main.version est posé
    # pour le jour où le binaire exposera sa version (variable à câbler côté code).
    ldflags = "-s -w -X main.version=#{version}"
    system "go", "build", *std_go_args(ldflags: ldflags), "."
  end

  test do
    # `csend help` doit répondre et mentionner le nom de l'outil. Pas de réseau,
    # pas d'état requis : test hermétique, valable en sandbox Homebrew.
    assert_match "csend", shell_output("#{bin}/csend help")
  end
end
