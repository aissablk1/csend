#!/usr/bin/env bash
# Démo « Green Build Relay » — TROIS éditeurs (Claude + Codex + Gemini) collaborent
# dans UN MÊME bus csend local. Le RELAIS est 100 % réel (vraies livraisons csend,
# cross-vendor, hashées). Le CONTENU des verdicts est SCRIPTÉ ici (driven) tant que
# Codex n'est pas installé / Gemini pas authentifié — honnêteté §2/§29 : on prouve le
# tuyau cross-vendor, pas une autonomie d'IA simulée. Store éphémère, zéro télémétrie.
set -euo pipefail
cd "$(dirname "$0")/.."
BIN="$(mktemp -d)/csend"; go build -o "$BIN" .
STORE="$(mktemp -d)"; export CSEND_STORE_DIR="$STORE"

echo "1) Trois éditeurs RIVAUX rejoignent le même bus local"
CSEND_AGENT_ID=claude-dev "$BIN" register --provider claude >/dev/null
CSEND_AGENT_ID=codex-exec "$BIN" register --provider codex  >/dev/null
CSEND_AGENT_ID=gemini-rev "$BIN" register --provider gemini >/dev/null
"$BIN" agents
echo

echo "2) Claude (orchestrateur) → Codex : lance les tests"
CSEND_AGENT_ID=claude-dev "$BIN" inbox codex-exec "test is_prime prêt — lance pytest, dis si c'est vert" >/dev/null
echo "   Codex reçoit (ce que fait son hook) :"
CSEND_AGENT_ID=codex-exec "$BIN" hook | sed 's/^/   /'
echo

echo "3) Codex (exécuteur) → Gemini : build vert, relis les cas limites"
CSEND_AGENT_ID=codex-exec "$BIN" inbox gemini-rev "build vert sur is_prime — relis le diff, cas limites (0,1,négatifs)" >/dev/null
echo "   Gemini reçoit :"
CSEND_AGENT_ID=gemini-rev "$BIN" hook | sed 's/^/   /'
echo

echo "4) Gemini (relecteur) → Claude : verdict adversarial qui RETOMBE dans la session vivante"
CSEND_AGENT_ID=gemini-rev "$BIN" inbox claude-dev "RÉFUTÉ : is_prime(1) doit renvoyer False — ajoute 'if n <= 1: return False'" >/dev/null
echo "   Claude reçoit (forme hook réelle, cross-vendor visible) :"
CSEND_AGENT_ID=claude-dev "$BIN" hook | sed 's/^/   /'
echo

echo "5) MONEY-FRAME — trois éditeurs sur un bus, corps en HASH (jamais le clair)"
"$BIN" agents
"$BIN" journal
echo
echo "Le relais cross-vendor est RÉEL (livraisons csend, 3 providers, 1 bus). Le contenu"
echo "des verdicts est scripté ici (driven). Pour des verdicts d'IA réels : installe Codex"
echo "(npm i -g @openai/codex) et authentifie Gemini (GEMINI_API_KEY). Réception live"
echo "permanente : câble 'csend hook' (csend hook --install)."
rm -rf "$STORE" "$(dirname "$BIN")"
