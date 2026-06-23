#!/bin/bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_DIR="$(dirname "$SCRIPT_DIR")"
BACKUP_DIR="${PROJECT_DIR}/backups"
TIMESTAMP="$(date +%Y%m%d_%H%M%S)"
BACKUP_FILE="${BACKUP_DIR}/postgres_api_${TIMESTAMP}.sql.gz"

# Load .env
if [ -f "${PROJECT_DIR}/.env" ]; then
    set -a
    source "${PROJECT_DIR}/.env"
    set +a
fi

mkdir -p "$BACKUP_DIR"

echo "Starting PostgreSQL backup..."
echo "Database: ${POSTGRES_DB:-postgres_api}"
echo "Backup: ${BACKUP_FILE}"

docker exec pg-api-postgres pg_dump \
    -U "${POSTGRES_USER:-api_admin}" \
    -d "${POSTGRES_DB:-postgres_api}" \
    --clean \
    --if-exists \
    --no-owner \
    --no-acl \
    --verbose \
    2>"${BACKUP_DIR}/backup_${TIMESTAMP}.log" \
    | gzip > "$BACKUP_FILE"

echo "Backup completed: ${BACKUP_FILE}"
echo "Log: ${BACKUP_DIR}/backup_${TIMESTAMP}.log"

# Clean backups older than 30 days
find "$BACKUP_DIR" -name "postgres_api_*.sql.gz" -mtime +30 -delete
find "$BACKUP_DIR" -name "backup_*.log" -mtime +30 -delete

echo "Old backups cleaned (retention: 30 days)"
