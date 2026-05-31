#!/bin/sh
# Raevtar health check
# Returns non-empty output only on failure (silent = healthy).
URL="http://localhost:8080"
STATUS=$(curl -s -o /dev/null -w "%{http_code}" --connect-timeout 5 --max-time 10 "$URL" 2>/dev/null)

case "$STATUS" in
  2*|3*)
    # Healthy — silent (watchdog pattern)
    exit 0
    ;;
  *)
    echo "⚠️ Raevtar DOWN at $(date '+%Y-%m-%d %H:%M:%S') — HTTP $STATUS"
    exit 1
    ;;
esac
