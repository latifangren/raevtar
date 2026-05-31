#!/bin/sh

DIR=$(CDPATH= cd -- "$(dirname -- "$0")/.." && pwd)
exec sh "${DIR}/static/agent/raevtar-agent.sh" "$@"
