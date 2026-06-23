#!/bin/bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_DIR="$(dirname "$SCRIPT_DIR")"

if [ $# -lt 1 ]; then
    echo "Usage: $0 <backup-file>"
    echo ""
    echo "Backups are stored in: ${PROJECT_DIR}/backups/"
    echo ""
    echo "Available backups:"
    ls -1 "${PROJECT_DIR}/backups/"*.sql.gz 2>/dev/null || echo "  No backups found"
    exit 1
fi

BACKUP_FILE="$1"

if [ ! -f "$BACKUP_FILE" ]; then
    echo "Error: Backup file not found: ${BACKUP_FILE}"
    exit 1
fi

# Load .env
if [ -f "${PROJECT_DIR}/.env" ]; then
    set -a
    source "${PROJECT_DIR}/.env"
    set +a
fi

echo "Starting PostgreSQL restore..."
echo "Database: ${POSTGRES_DB:-postgres_api}"
echo "Backup: ${BACKUP_FILE}"

echo "WARNING: This will overwrite the current database!"
read -rp "Are you sure? (y/N): " CONFIRM
if [ "$CONFIRM" != "y" ] && [ "$CONFIRM" != "Y" ]; then
    echo "Restore cancelled."
    exit 0
fi

gunzip -c "$BACKUP_FILE" | docker exec -i pg-api-postgres psql \
    -U "${POSTGRES_USER:-api_admin}" \
    -d "${POSTGRES_DB:-postgres_api}" \
    --quiet

echo "Restore completed: ${BACKUP_FILE}"
