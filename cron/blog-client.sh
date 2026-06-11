#!/bin/bash
# Raevtar Blog API Client wrapper for Hermes agent
# Avoid raw token exposure in prompt & execution logs.
set -euo pipefail

ENV_FILE="/home/latif/raevtar/.env.production"

if [ ! -f "$ENV_FILE" ]; then
  echo "Error: Env file $ENV_FILE not found." >&2
  exit 1
fi

# Extract key from .env.production file directly and strip wrapping quotes
RAEVTAR_ADMIN_KEY=$(grep -E "^RAEVTAR_ADMIN_KEY=" "$ENV_FILE" | cut -d'=' -f2- | tr -d '\r' | tr -d '"' | tr -d "'")

if [ -z "$RAEVTAR_ADMIN_KEY" ]; then
  echo "Error: RAEVTAR_ADMIN_KEY not found in $ENV_FILE" >&2
  exit 1
fi

BASE_URL="http://localhost:8080/api/v1"
COMMAND="${1:-}"

case "$COMMAND" in
  get-inbox)
    curl -s -f \
      -H "Authorization: Bearer $RAEVTAR_ADMIN_KEY" \
      -H "Content-Type: application/json" \
      "$BASE_URL/editorial-inbox?ready=true"
    ;;
  get-posts)
    curl -s -f \
      "$BASE_URL/posts"
    ;;
  create-post)
    PAYLOAD="${2:-}"
    if [ -z "$PAYLOAD" ]; then
      PAYLOAD=$(cat)
    fi
    curl -s -f -X POST \
      -H "Authorization: Bearer $RAEVTAR_ADMIN_KEY" \
      -H "Content-Type: application/json" \
      -d "$PAYLOAD" \
      "$BASE_URL/posts"
    ;;
  mark-done)
    ITEM_ID="${2:-}"
    POST_ID="${3:-}"
    if [ -z "$ITEM_ID" ] || [ -z "$POST_ID" ]; then
      echo "Usage: $0 mark-done <item_id> <published_post_id>" >&2
      exit 1
    fi

    # Read the item first to get all properties
    ITEM_JSON=$(curl -s -f \
      -H "Authorization: Bearer $RAEVTAR_ADMIN_KEY" \
      -H "Content-Type: application/json" \
      "$BASE_URL/editorial-inbox/$ITEM_ID")

    if [ -z "$ITEM_JSON" ]; then
      echo "Error: Unable to fetch inbox item $ITEM_ID" >&2
      exit 1
    fi

    # Build update payload with status=done and published_post_id
    # Using python3 to safely construct JSON without jq
    UPDATED_JSON=$(python3 -c "
import sys, json
item = json.loads(sys.argv[1])
update = {
    'source_type': item.get('source_type', ''),
    'source_value': item.get('source_value', ''),
    'category_hint': item.get('category_hint', ''),
    'priority': item.get('priority', 0),
    'not_before': item.get('not_before', ''),
    'deadline': item.get('deadline'),
    'note': item.get('note', ''),
    'mode': item.get('mode', ''),
    'status': 'done',
    'published_post_id': int(sys.argv[2]),
    'failure_note': item.get('failure_note', ''),
    'failure_meta': item.get('failure_meta', '')
}
print(json.dumps(update))
" "$ITEM_JSON" "$POST_ID")

    curl -s -f -X POST \
      -H "Authorization: Bearer $RAEVTAR_ADMIN_KEY" \
      -H "Content-Type: application/json" \
      -d "$UPDATED_JSON" \
      "$BASE_URL/editorial-inbox/$ITEM_ID"
    ;;
  *)
    echo "Unknown command: $COMMAND" >&2
    echo "Usage: $0 {get-inbox|get-posts|create-post|mark-done}" >&2
    exit 1
    ;;
esac
