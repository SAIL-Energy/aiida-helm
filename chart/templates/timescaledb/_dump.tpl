{{/* SPDX-License-Identifier: Apache-2.0. Runtime content for a ConfigMap. */}}
{{- define "aiida.timescaledb.dump" -}}
#!/bin/sh
set -eu
umask 077

: "${DB_HOST:?DB_HOST must be set}"
: "${DB_PORT:?DB_PORT must be set}"
: "${DB_NAME:?DB_NAME must be set}"
: "${DB_USERNAME:?DB_USERNAME must be set}"
: "${DB_PASSWORD:?DB_PASSWORD must be set}"
: "${BACKUP_NAME:?BACKUP_NAME must be set}"

timestamp=$(date -u +%Y%m%dT%H%M%SZ)
filename="${BACKUP_NAME}-${timestamp}.dump"
backup_path="/backup/${filename}"
export PGPASSWORD="$DB_PASSWORD"

# Use pg_dump's archive compression directly. A failed dump must never become
# an apparently successful gzip stream or reach the upload container.
if ! pg_dump --host="$DB_HOST" --port="$DB_PORT" --username="$DB_USERNAME" \
  --dbname="$DB_NAME" --format=custom --compress=6 --no-password \
  --file="$backup_path"; then
  rm -f "$backup_path"
  printf '%s\n' 'Database dump failed; no backup will be uploaded.' >&2
  exit 1
fi

if [ ! -s "$backup_path" ]; then
  printf '%s\n' 'Database dump produced an empty archive; refusing upload.' >&2
  exit 1
fi

printf '%s\n' "$filename" > /backup/filename
printf '%s\n' 'Database archive completed successfully.'
{{- end -}}
