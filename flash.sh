#!/usr/bin/env bash
set -euo pipefail

usage() {
	echo "usage: $0 <OnHours>" >&2
	echo "  OnHours: integer 1–23 (written to main.go, then tinygo flash)" >&2
	exit 1
}

[[ "${1-}" ]] || usage
[[ "$1" =~ ^[0-9]+$ ]] || usage
hours="$1"
if (( hours < 1 || hours > 23 )); then
	echo "error: OnHours must be between 1 and 23 (got $hours)" >&2
	exit 1
fi

root="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
main="$root/main.go"

tmp="$(mktemp)"
trap 'rm -f "$tmp"' EXIT
sed "s/^const OnHours = [0-9][0-9]*$/const OnHours = $hours/" "$main" >"$tmp"
if ! grep -q "^const OnHours = $hours\$" "$tmp"; then
	echo "error: failed to patch const OnHours in $main" >&2
	exit 1
fi
mv "$tmp" "$main"
trap - EXIT

export GOWORK=off
tinygo flash -target=nano-rp2040 "$root"
off=$((24 - hours))
echo "OK: flashed OnHours=${hours} (D2 high ${hours}h, low ${off}h per 24h cycle)"
