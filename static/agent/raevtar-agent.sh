#!/bin/sh
#
# raevtar-agent.sh — Monitoring agent for Raevtar
# Collects system metrics and reports to Raevtar server.
# Supports Linux (full) and macOS (via sysctl/top).
#
# Usage:
#   RAEVTAR_URL=https://raevtar.tech RAEVTAR_SERVER_ID=1 RAEVTAR_AGENT_TOKEN=abc123 ./raevtar-agent.sh
#
# Single-line install via bootstrap:
#   curl -fsSL https://raevtar.tech/api/v1/bootstrap/<id>/<token> | sh
#

set -eu

RAEVTAR_URL="${RAEVTAR_URL:-${RAEVTAR_API:-http://127.0.0.1:8080}}"
SERVER_ID="${RAEVTAR_SERVER_ID:-${1:-}}"
TOKEN="${RAEVTAR_AGENT_TOKEN:-${RAEVTAR_KEY:-}}"
VERBOSE="${RAEVTAR_VERBOSE:-1}"

info()  { [ "${VERBOSE}" != "0" ] && printf "[INFO]  %s\n" "$@"; }
ok()    { [ "${VERBOSE}" != "0" ] && printf "[OK]    %s\n" "$@"; }
warn()  { [ "${VERBOSE}" != "0" ] && printf "[WARN]  %s\n" "$@" >&2; }
fail()  { [ "${VERBOSE}" != "0" ] && printf "[FAIL]  %s\n" "$@" >&2; }
detail(){ [ "${VERBOSE}" != "0" ] && printf "        %s\n" "$@"; }

if [ -z "${SERVER_ID}" ]; then
	fail "RAEVTAR_SERVER_ID is required"
	exit 1
fi

if [ -z "${TOKEN}" ]; then
	fail "RAEVTAR_AGENT_TOKEN is required"
	exit 1
fi

# --- Detect OS ---
detect_os() {
	case "$(uname -s)" in
		Linux*)  echo "linux"  ;;
		Darwin*) echo "darwin" ;;
		*)       echo "unknown" ;;
	esac
}

OS="$(detect_os)"
info "Raevtar Agent starting"
detail "Server ID : ${SERVER_ID}"
detail "Server URL: ${RAEVTAR_URL}"
detail "OS        : ${OS}"
detail ""

# --- CPU percent ---
cpu_percent() {
	case "${OS}" in
		linux)
			if [ ! -r /proc/stat ]; then
				warn "/proc/stat not found — skipping CPU"
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
			;;
		darwin)
			# Get CPU usage from top sampling (macOS doesn't expose /proc/stat)
			top -l 2 -n 0 2>/dev/null | awk '/CPU usage/ { if (n++ == 0) next; print $3 }' | tr -d '%' || echo "0.0"
			;;
		*)
			echo "0.0"
			;;
	esac
}

CPU="$(cpu_percent)"
ok "CPU usage: ${CPU}%"
detail ""

# --- CPU load ---
cpu_load_values() {
	case "${OS}" in
		linux)
			if [ ! -r /proc/loadavg ]; then
				echo "0 0 0"
				return
			fi
			awk '{ printf "%.2f %.2f %.2f", $1, $2, $3 }' /proc/loadavg
			;;
		darwin)
			sysctl -n vm.loadavg 2>/dev/null | awk '{ printf "%.2f %.2f %.2f", $1, $2, $3 }' || echo "0 0 0"
			;;
		*)
			echo "0 0 0"
			;;
	esac
}

set -- $(cpu_load_values)
CPU_LOAD_1="${1:-0}"
CPU_LOAD_5="${2:-0}"
CPU_LOAD_15="${3:-0}"
ok "CPU load: 1m=${CPU_LOAD_1} 5m=${CPU_LOAD_5} 15m=${CPU_LOAD_15}"
detail ""

# --- CPU cores ---
cpu_cores() {
	case "${OS}" in
		linux)
			cores="$(getconf _NPROCESSORS_ONLN 2>/dev/null || true)"
			case "${cores}" in
				''|*[!0-9]*) cores=0 ;;
			esac
			if [ "${cores}" -eq 0 ] && [ -r /proc/cpuinfo ]; then
				cores="$(awk '/^processor[[:space:]]*:/ { count++ } END { print count + 0 }' /proc/cpuinfo)"
			fi
			echo "${cores}"
			;;
		darwin)
			sysctl -n hw.ncpu 2>/dev/null || echo "0"
			;;
		*)
			getconf _NPROCESSORS_ONLN 2>/dev/null || echo "0"
			;;
	esac
}

CPU_CORES="$(cpu_cores)"
ok "CPU cores: ${CPU_CORES}"
detail ""

# --- RAM ---
ram_values() {
	case "${OS}" in
		linux)
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
			;;
		darwin)
			# macOS: hw.memsize in bytes, vm_stat pages in bytes
			total_bytes=$(sysctl -n hw.memsize 2>/dev/null || echo "0")
			page_size=$(sysctl -n hw.pagesize 2>/dev/null || echo "4096")
			# Get free + inactive + speculative from vm_stat (in pages)
			vm_free=$(vm_stat 2>/dev/null | awk '/Pages free/ { gsub(/\./,"",$NF); print $NF }' || echo "0")
			vm_inactive=$(vm_stat 2>/dev/null | awk '/Pages inactive/ { gsub(/\./,"",$NF); print $NF }' || echo "0")
			vm_active=$(vm_stat 2>/dev/null | awk '/Pages active/ { gsub(/\./,"",$NF); print $NF }' || echo "0")
			total_mb=$((total_bytes / 1048576))
			available_mb=$(( (vm_free + vm_inactive) * page_size / 1048576 ))
			used_mb=$((vm_active * page_size / 1048576))
			echo "${used_mb} ${total_mb}"
			;;
		*)
			echo "0 0"
			;;
	esac
}

set -- $(ram_values)
RAM_USED="${1:-0}"
RAM_TOTAL="${2:-0}"
RAM_PCT=0
if [ "${RAM_TOTAL}" -gt 0 ]; then
	RAM_PCT=$((RAM_USED * 100 / RAM_TOTAL))
fi
ok "RAM: ${RAM_USED}MB / ${RAM_TOTAL}MB (${RAM_PCT}%)"
detail ""

# --- Disk ---
disk_values() {
	case "${OS}" in
		linux)
			df -Pk / 2>/dev/null | awk 'NR==2 { printf "%.2f %.2f", $(NF - 3) / 1048576, $(NF - 4) / 1048576 }' || echo "0 0"
			;;
		darwin)
			df -Pk / 2>/dev/null | awk 'NR==2 { printf "%.2f %.2f", $(NF - 3) / 1048576, $(NF - 4) / 1048576 }' || echo "0 0"
			;;
		*)
			echo "0 0"
			;;
	esac
}

set -- $(disk_values)
DISK_USED="${1:-0}"
DISK_TOTAL="${2:-0}"
ok "Disk: ${DISK_USED}GB / ${DISK_TOTAL}GB"
detail ""

# --- Temperature ---
temperature_values() {
	case "${OS}" in
		linux)
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
			;;
		darwin)
			# macOS: no standard thermal zone; try powermetrics (needs sudo, skip)
			echo "0 false"
			;;
		*)
			echo "0 false"
			;;
	esac
}

set -- $(temperature_values)
TEMPERATURE="${1:-0}"
TEMPERATURE_AVAILABLE="${2:-false}"
if [ "${TEMPERATURE_AVAILABLE}" = "true" ]; then
	ok "Temperature: ${TEMPERATURE}°C"
else
	detail "Temperature: not available"
fi
detail ""

# --- Uptime ---
uptime_seconds() {
	case "${OS}" in
		linux)
			awk '{ print int($1) }' /proc/uptime 2>/dev/null || echo "0"
			;;
		darwin)
			# macOS: kern.boottime is seconds since epoch
			boot_sec=$(sysctl -n kern.boottime 2>/dev/null | awk '{ print $4 }' | tr -d ',')
			now=$(date +%s)
			if [ -n "${boot_sec}" ] && [ "${boot_sec}" -gt 0 ] 2>/dev/null; then
				echo $((now - boot_sec))
			else
				echo "0"
			fi
			;;
		*)
			echo "0"
			;;
	esac
}

UPTIME="$(uptime_seconds)"
UPTIME_HUMAN=""
if [ "${UPTIME}" -gt 86400 ]; then
	UPTIME_HUMAN="$((UPTIME / 86400))d $(( (UPTIME % 86400) / 3600 ))h"
elif [ "${UPTIME}" -gt 3600 ]; then
	UPTIME_HUMAN="$((UPTIME / 3600))h $(( (UPTIME % 3600) / 60 ))m"
else
	UPTIME_HUMAN="${UPTIME}s"
fi
ok "Uptime: ${UPTIME_HUMAN}"
detail ""

# --- Build JSON payload ---
JSON=$(printf '{"cpu_percent":%.1f,"cpu_load_1":%.2f,"cpu_load_5":%.2f,"cpu_load_15":%.2f,"cpu_cores":%d,"ram_used_mb":%d,"ram_total_mb":%d,"disk_used_gb":%.2f,"disk_total_gb":%.2f,"temperature_c":%.2f,"temperature_available":%s,"uptime_seconds":%d,"online":true}' \
	"${CPU:-0}" "${CPU_LOAD_1:-0}" "${CPU_LOAD_5:-0}" "${CPU_LOAD_15:-0}" "${CPU_CORES:-0}" "${RAM_USED:-0}" "${RAM_TOTAL:-0}" "${DISK_USED:-0}" "${DISK_TOTAL:-0}" "${TEMPERATURE:-0}" "${TEMPERATURE_AVAILABLE:-false}" "${UPTIME:-0}")

# --- Send metrics ---
info "Sending metrics to ${RAEVTAR_URL}"
PING_URL="${RAEVTAR_URL%/}/api/v1/servers/${SERVER_ID}/ping"
HTTP_CODE=$(curl -fsS -o /dev/null -w "%{http_code}" -X POST "${PING_URL}" \
	-H "Authorization: Bearer ${TOKEN}" \
	-H "Content-Type: application/json" \
	-d "${JSON}" 2>/dev/null || echo "000")

case "${HTTP_CODE}" in
	200|201|202|204)
		ok "Metrics sent (HTTP ${HTTP_CODE})"
		;;
	401|403)
		fail "Metrics rejected (HTTP ${HTTP_CODE}) — check agent token"
		;;
	404)
		fail "Metrics rejected (HTTP ${HTTP_CODE}) — server not found"
		;;
	000)
		fail "Metrics failed — network error or timeout"
		detail "URL: ${PING_URL}"
		;;
	*)
		warn "Metrics responded (HTTP ${HTTP_CODE})"
		;;
esac
detail ""

# --- Poll for pending commands ---
info "Polling for pending commands..."
COMMANDS_JSON=$(curl -fsS "${RAEVTAR_URL%/}/api/v1/servers/${SERVER_ID}/commands" \
	-H "Authorization: Bearer ${TOKEN}" 2>/dev/null || echo "[]")

if [ "${COMMANDS_JSON}" = "[]" ] || [ -z "${COMMANDS_JSON}" ]; then
	ok "No pending commands"
else
	# Count commands
	CMD_COUNT=$(echo "${COMMANDS_JSON}" | grep -o '"id"' | wc -l)
	ok "${CMD_COUNT} pending command(s) found"
	detail ""

	# Split JSON array and process each
	echo "${COMMANDS_JSON}" | sed 's/^\[//;s/\]$//' | sed 's/},{/}\n{/g' | while read -r CMD_JSON; do
		[ -z "${CMD_JSON}" ] && continue
		CMD_ID=$(echo "${CMD_JSON}" | grep -o '"id":[0-9]*' | cut -d: -f2)
		CMD_NAME=$(echo "${CMD_JSON}" | grep -o '"command":"[^"]*"' | cut -d'"' -f4)
		if [ -z "${CMD_NAME}" ]; then
			continue
		fi

		info "Executing command: ${CMD_NAME} (id=${CMD_ID})"
		RESULT=""
		FAILED="false"

		case "${CMD_NAME}" in
			RESTART_AGENT)
				detail "Agent restart requested for this node."
				detail "Restart the agent process manually or via init/systemd."
				RESULT="Agent restart requested."
				;;
			CLEAR_CACHE)
				if command -v sync >/dev/null 2>&1; then
					sync
					detail "System sync completed."
				else
					detail "sync command not available."
				fi
				RESULT="Cache clear: system sync done."
				;;
			REBOOT_NODE)
				detail "Reboot requested for this node."
				detail "Requires root privileges or manual intervention."
				RESULT="Reboot requested. Node requires manual reboot or sudo."
				;;
			UPDATE_AGENT)
				detail "Updating agent script..."
				SCRIPT_PATH="${0}"
				TMP_SCRIPT="/tmp/raevtar-agent-$$.sh"
				if curl -fsS -o "${TMP_SCRIPT}" "${RAEVTAR_URL%/}/static/agent/raevtar-agent.sh" 2>/dev/null; then
					chmod +x "${TMP_SCRIPT}"
					mv "${TMP_SCRIPT}" "${SCRIPT_PATH}"
					ok "Agent updated successfully (${SCRIPT_PATH})"
					RESULT="Agent script updated to latest version."
				else
					fail "Failed to download updated agent from ${RAEVTAR_URL}"
					RESULT="Update failed: could not download agent script."
					FAILED="true"
				fi
				;;
			*)
				fail "Unknown command: ${CMD_NAME}"
				RESULT="Unknown command: ${CMD_NAME}"
				FAILED="true"
				;;
		esac

		detail ""
		info "Reporting result for command ${CMD_ID}..."
		RESULT_ESC=$(printf '%s\n' "${RESULT}" | sed 's/"/\\"/g')
		RESULT_JSON=$(printf '{"command_id":%d,"result":"%s","failed":%s}' \
			"${CMD_ID}" "${RESULT_ESC}" "${FAILED}")
		REPORT_CODE=$(curl -fsS -o /dev/null -w "%{http_code}" -X POST "${RAEVTAR_URL%/}/api/v1/servers/${SERVER_ID}/commands/result" \
			-H "Authorization: Bearer ${TOKEN}" \
			-H "Content-Type: application/json" \
			-d "${RESULT_JSON}" 2>/dev/null || echo "000")
		if [ "${REPORT_CODE}" = "200" ] || [ "${REPORT_CODE}" = "201" ] || [ "${REPORT_CODE}" = "202" ]; then
			ok "Result reported (HTTP ${REPORT_CODE})"
		else
			warn "Result report failed (HTTP ${REPORT_CODE})"
		fi
		detail ""
	done
fi

detail ""
ok "Agent cycle complete"
