#!/usr/bin/env bash
set -euo pipefail

# Pass build-time config via link flags (no file edits). Go build tags cannot
# carry values like 18; -ldflags -X is the standard way to inject strings.
#
#   ./flash.sh <onHours> [target] [dutyPercent] [dutyPeriodMins]
#
# onHours: 1–23
# target:  tinygo -target=… (default: nano-rp2040)
# dutyPercent: 1–100 (default: 100)
# dutyPeriodMins: 0 = relay held on full on-window; >0 and duty<100 = cycling (default: 0)

usage() {
	echo "usage: $0 <onHours> [target] [dutyPercent] [dutyPeriodMins]" >&2
	echo "  example: $0 12 pico" >&2
	echo "  example: $0 12 nano-rp2040 50 60   # 50% on, 60 min period" >&2
	exit 1
}

[[ "${1-}" ]] || usage
[[ "$1" =~ ^[0-9]+$ ]] || usage
hours="$1"
if (( hours < 1 || hours > 23 )); then
	echo "error: onHours must be 1–23 (got $hours)" >&2
	exit 1
fi

target="${2:-nano-rp2040}"
dpercent="${3:-100}"
dperiod="${4:-0}"

[[ "$dpercent" =~ ^[0-9]+$ ]] || usage
[[ "$dperiod" =~ ^[0-9]+$ ]] || usage
if (( dpercent < 1 || dpercent > 100 )); then
	echo "error: dutyPercent must be 1–100 (got $dpercent)" >&2
	exit 1
fi
if (( dpercent < 100 && dperiod < 1 )); then
	echo "error: duty<100 needs dutyPeriodMins >= 1 (got period=$dperiod)" >&2
	exit 1
fi

root="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
# One shell word for the whole -X list. Do not use printf %q for that string — it
# inserts backslashes (e.g. -X\ main...) and a copy-paste breaks into multiple args.
# Values here are digits only, safe inside single quotes.
ldflags_x="-X main.onHours=${hours} -X main.dutyPercent=${dpercent} -X main.dutyPeriodMins=${dperiod}"

echo
echo "copy: tinygo flash -ldflags='${ldflags_x}' -target ${target} -monitor $(printf %q "$root")"
echo
tinygo flash -ldflags="$ldflags_x" -target "$target" "$root"
echo
off=$((24 - hours))
echo "OK: flashed target=${target} onHours=${hours} (on-window ${hours}h, off ${off}h) duty%=${dpercent} periodMins=${dperiod}"
echo