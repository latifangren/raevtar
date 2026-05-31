#!/bin/sh
# Raevtar agent — report system metrics to raevtar.tech
# Usage: RAEVTAR_KEY=<admin_key> ./agent-ping.sh <server_id>
# Cron: */5 * * * * RAEVTAR_KEY=xxx /home/latif/raevtar/cron/agent-ping.sh 1

set -e
SERVER_ID="${1:-1}"
API="${RAEVTAR_API:-https://raevtar.tech}"
KEY="${RAEVTAR_KEY}"

if [ -z "$KEY" ]; then
	echo "RAEVTAR_KEY not set" >&2
	exit 1
fi

# Collect metrics — use /proc directly (always available on Linux)

# CPU: read cpu line from /proc/stat, calc usage % from (user+system)/(total) * 100
CPU=0
if [ -r /proc/stat ]; then
	set -- $(awk '/^cpu /{print $2,$4,$5}' /proc/stat)
	user=$1; system=$2; idle=$3; total=$((user+system+idle))
	[ "$total" -gt 0 ] && CPU=$(awk "BEGIN{printf \"%.1f\", (($user+$system)/$total)*100}")
fi

# RAM from /proc/meminfo
RAM_TOTAL=0; RAM_AVAIL=0
if [ -r /proc/meminfo ]; then
	eval $(awk '/^MemTotal/{print "RAM_TOTAL="$2} /^MemAvailable/{print "RAM_AVAIL="$2}' /proc/meminfo)
	RAM_USED=$(( (RAM_TOTAL - RAM_AVAIL) / 1024 ))
	RAM_TOTAL_MB=$(( RAM_TOTAL / 1024 ))
fi

# Disk
DISK_USED=$(df -BG / 2>/dev/null | awk 'NR==2{print $3}' | tr -d 'G' || echo 0)

# Uptime
UPTIME=$(awk '{print int($1)}' /proc/uptime 2>/dev/null || echo 0)

ONLINE=true

# Build JSON (printf for safe encoding)
JSON=$(printf '{"cpu_percent":%.1f,"ram_used_mb":%d,"ram_total_mb":%d,"disk_used_gb":%d,"uptime_seconds":%d,"online":%s}' \
  "${CPU:-0}" "${RAM_USED:-0}" "${RAM_TOTAL_MB:-0}" "${DISK_USED:-0}" "${UPTIME:-0}" "${ONLINE}")

# Send
AUTH_HEADER="Authorization: Bearer ${KEY}"
curl -s -X POST "${API}/api/v1/servers/${SERVER_ID}/ping" \
  -H "${AUTH_HEADER}" \
  -H "Content-Type: application/json" \
  -d "${JSON}" \
  -o /dev/null -w "%{http_code}" 2>/dev/null || echo "FAIL"
