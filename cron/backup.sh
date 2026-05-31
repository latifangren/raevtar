#!/bin/sh
# SQLite backup — copy file, simpan 7 hari terakhir
# Usage: ./cron/backup.sh [/path/to/backup/dir]

set -e

DB="${RAEVTAR_DB:-$HOME/.raevtar/data.db}"
BACKUP_DIR="${1:-$HOME/.raevtar/backups}"
TIMESTAMP="$(date +%Y%m%d_%H%M%S)"

mkdir -p "$BACKUP_DIR"

cp "$DB" "$BACKUP_DIR/data_$TIMESTAMP.db"

# Hapus backup lebih dari 7 hari
find "$BACKUP_DIR" -name 'data_*.db' -mtime +7 -delete

echo "backup: $BACKUP_DIR/data_$TIMESTAMP.db"
echo "size: $(du -h "$BACKUP_DIR/data_$TIMESTAMP.db" | cut -f1)"
