#!/usr/bin/env bash
set -euo pipefail

# Pass build-time config via link flags (no file edits). Go build tags cannot
# carry values like 18; -ldflags -X is the standard way to inject strings.
#
#   ./flash.sh <onHours> [target] [dutyPercent] [dutyPeriodMins] [onWindowReduceMins]
#
# onHours: decimal hours, 0 < n < 24 (e.g. 12, 12.5 for 12h30m on-window)
# target:  tinygo -target=… (default: nano-rp2040)
# dutyPercent: 1–100 (default: 100)
# dutyPeriodMins: 0 = relay held on full on-window; >0 and duty<100 = cycling (default: 0)
# onWindowReduceMins: subtract this many minutes from the on-window after each full 24h cycle; 0 = off (default: 0)

usage() {
	echo "usage: $0 <onHours> [target] [dutyPercent] [dutyPeriodMins] [onWindowReduceMins]" >&2
	echo "  example: $0 12.5 pico" >&2
	echo "  example: $0 17.5 pico 100 0 15   # first on-window 17.5h, then −15m each cycle" >&2
	echo "  example: $0 12 nano-rp2040 50 60   # 50% on, 60 min period" >&2
	exit 1
}

[[ "${1-}" ]] || usage
hours="$1"
if ! awk -v h="$hours" 'BEGIN {
	if (h !~ /^[0-9]+(\.[0-9]*)?$/ && h !~ /^\.[0-9]+$/) exit 1
	x = h + 0
	if (x <= 0 || x >= 24) exit 1
	exit 0
}'; then
	echo "error: onHours must be a number with 0 < onHours < 24 (e.g. 12 or 12.5)" >&2
	exit 1
fi

target="${2:-nano-rp2040}"
dpercent="${3:-100}"
dperiod="${4:-0}"
reduceMins="${5:-0}"

[[ "$dpercent" =~ ^[0-9]+$ ]] || usage
[[ "$dperiod" =~ ^[0-9]+$ ]] || usage
[[ "$reduceMins" =~ ^[0-9]+$ ]] || usage
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
# Values must stay one shell word inside the single-quoted -ldflags list.
ldflags_x="-X main.onHours=${hours} -X main.dutyPercent=${dpercent} -X main.dutyPeriodMins=${dperiod} -X main.onWindowReduceMins=${reduceMins}"

echo
echo "copy: tinygo flash -ldflags='${ldflags_x}' -target ${target} -monitor $(printf %q "$root")"
echo
tinygo flash -ldflags="$ldflags_x" -target "$target" "$root"
echo
offv=$(awk -v h="$hours" 'BEGIN { printf "%.10g", 24 - h }')
echo "OK: flashed target=${target} onHours=${hours} (off ${offv}h first cycle) duty%=${dpercent} periodMins=${dperiod} onWindowReduceMins=${reduceMins}"
echo