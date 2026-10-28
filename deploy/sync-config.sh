#!/bin/sh
# Derives the container config from the local dev config. Everything is copied
# verbatim except what cannot work inside the container network:
#   hosts (localhost -> service names), console logs -> json, LLM on the host,
#   and the LLM debug dump directory.
set -eu

src=${1:-backend/configs/config.yaml}
dst=${2:-deploy/config.yaml}

if [ ! -f "$src" ]; then
  echo "sync-config: $src not found (run: make config)" >&2
  exit 1
fi

sed \
  -e 's|@localhost:5432/|@db:5432/|' \
  -e 's|- localhost:9094|- kafka:9092|' \
  -e 's|^  format: console$|  format: json|' \
  -e 's|^  base_url: http://localhost:|  base_url: http://host.docker.internal:|' \
  -e 's|^  base_url: http://127.0.0.1:|  base_url: http://host.docker.internal:|' \
  -e 's|^  debug_dir: .*$|  debug_dir: ""|' \
  -e 's|^  api_key: ""$|  api_key: "REPLACE_WITH_OPENROUTER_KEY"|' \
  "$src" >"$dst"

if grep -q 'api_key: "REPLACE_WITH_OPENROUTER_KEY"' "$dst"; then
  echo "sync-config: llm.api_key is empty in $src, placeholder written to $dst" >&2
fi

echo "sync-config: $src -> $dst"
echo "sync-config: llm.debug_dir disabled in the container (relative host path is not writable there)" >&2
