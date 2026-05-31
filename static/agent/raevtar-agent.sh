#!/bin/sh

set -eu

RAEVTAR_URL="${RAEVTAR_URL:-${RAEVTAR_API:-http://127.0.0.1:8080}}"
SERVER_ID="${RAEVTAR_SERVER_ID:-${1:-}}"
TOKEN="${RAEVTAR_AGENT_TOKEN:-${RAEVTAR_KEY:-}}"

if [ -z "${SERVER_ID}" ]; then
	echo "RAEVTAR_SERVER_ID is required" >&2
	exit 1
fi

if [ -z "${TOKEN}" ]; then
	echo "RAEVTAR_AGENT_TOKEN is required" >&2
	exit 1
fi

cpu_percent() {
	if [ ! -r /proc/stat ]; then
		echo "0.0"
		return
	fi
	read _ user nice system idle iowait irq softirq steal _ < /proc/stat
	idle_all=$((idle + iowait))
	non_idle=$((user + nice + system + irq + softirq + steal))
	total=$((idle_all + non_idle))
	if [ "${total}" -le 0 ]; then
		echo "0.0"
		return
	fi
	awk "BEGIN { printf \"%.1f\", (${non_idle} / ${total}) * 100 }"
}

ram_values() {
	if [ ! -r /proc/meminfo ]; then
		echo "0 0"
		return
	fi
	awk '
		/^MemTotal:/ { total=$2 }
		/^MemAvailable:/ { available=$2 }
		END {
			if (total == "") total = 0
			if (available == "") available = total
			printf "%d %d", (total - available) / 1024, total / 1024
		}
	' /proc/meminfo
}

disk_used_gb() {
	df -BG / 2>/dev/null | awk 'NR==2 { gsub(/G/, "", $3); print $3 + 0 }'
}

uptime_seconds() {
	awk '{ print int($1) }' /proc/uptime 2>/dev/null || echo "0"
}

CPU="$(cpu_percent)"
set -- $(ram_values)
RAM_USED="${1:-0}"
RAM_TOTAL="${2:-0}"
DISK_USED="$(disk_used_gb)"
UPTIME="$(uptime_seconds)"

JSON=$(printf '{"cpu_percent":%.1f,"ram_used_mb":%d,"ram_total_mb":%d,"disk_used_gb":%d,"uptime_seconds":%d,"online":true}' \
	"${CPU:-0}" "${RAM_USED:-0}" "${RAM_TOTAL:-0}" "${DISK_USED:-0}" "${UPTIME:-0}")

curl -fsS -X POST "${RAEVTAR_URL%/}/api/v1/servers/${SERVER_ID}/ping" \
	-H "Authorization: Bearer ${TOKEN}" \
	-H "Content-Type: application/json" \
	-d "${JSON}"
