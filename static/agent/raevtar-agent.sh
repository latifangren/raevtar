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

cpu_load_values() {
	if [ ! -r /proc/loadavg ]; then
		echo "0 0 0"
		return
	fi
	awk '{ printf "%.2f %.2f %.2f", $1, $2, $3 }' /proc/loadavg
}

cpu_cores() {
	cores="$(getconf _NPROCESSORS_ONLN 2>/dev/null || true)"
	case "${cores}" in
		''|*[!0-9]*) cores=0 ;;
	esac
	if [ "${cores}" -eq 0 ] && [ -r /proc/cpuinfo ]; then
		cores="$(awk '/^processor[[:space:]]*:/ { count++ } END { print count + 0 }' /proc/cpuinfo)"
	fi
	echo "${cores}"
}

disk_values() {
	df -Pk / 2>/dev/null | awk 'NR==2 { printf "%.2f %.2f", $(NF - 3) / 1048576, $(NF - 4) / 1048576 }' || echo "0 0"
}

temperature_values() {
	for zone in /sys/class/thermal/thermal_zone*/temp; do
		if [ -r "${zone}" ]; then
			temperature="$(awk '/^-?[0-9]+([.][0-9]+)?$/ { printf "%.2f", $1 / 1000 }' "${zone}")"
			if [ -n "${temperature}" ]; then
				echo "${temperature} true"
				return
			fi
		fi
	done
	echo "0 false"
}

uptime_seconds() {
	awk '{ print int($1) }' /proc/uptime 2>/dev/null || echo "0"
}

CPU="$(cpu_percent)"
set -- $(cpu_load_values)
CPU_LOAD_1="${1:-0}"
CPU_LOAD_5="${2:-0}"
CPU_LOAD_15="${3:-0}"
CPU_CORES="$(cpu_cores)"
set -- $(ram_values)
RAM_USED="${1:-0}"
RAM_TOTAL="${2:-0}"
set -- $(disk_values)
DISK_USED="${1:-0}"
DISK_TOTAL="${2:-0}"
set -- $(temperature_values)
TEMPERATURE="${1:-0}"
TEMPERATURE_AVAILABLE="${2:-false}"
UPTIME="$(uptime_seconds)"

JSON=$(printf '{"cpu_percent":%.1f,"cpu_load_1":%.2f,"cpu_load_5":%.2f,"cpu_load_15":%.2f,"cpu_cores":%d,"ram_used_mb":%d,"ram_total_mb":%d,"disk_used_gb":%.2f,"disk_total_gb":%.2f,"temperature_c":%.2f,"temperature_available":%s,"uptime_seconds":%d,"online":true}' \
	"${CPU:-0}" "${CPU_LOAD_1:-0}" "${CPU_LOAD_5:-0}" "${CPU_LOAD_15:-0}" "${CPU_CORES:-0}" "${RAM_USED:-0}" "${RAM_TOTAL:-0}" "${DISK_USED:-0}" "${DISK_TOTAL:-0}" "${TEMPERATURE:-0}" "${TEMPERATURE_AVAILABLE:-false}" "${UPTIME:-0}")

curl -fsS -X POST "${RAEVTAR_URL%/}/api/v1/servers/${SERVER_ID}/ping" \
	-H "Authorization: Bearer ${TOKEN}" \
	-H "Content-Type: application/json" \
	-d "${JSON}"
