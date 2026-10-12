{{/* SPDX-License-Identifier: Apache-2.0. Runtime content for a ConfigMap. */}}
{{- define "aiida.timescaledb.initialize" -}}
#!/bin/sh
set -eu

: "${POSTGRES_USER:?POSTGRES_USER must be set}"
: "${POSTGRES_DB:?POSTGRES_DB must be set}"
: "${MQTT_DATABASE_PASSWORD:?MQTT_DATABASE_PASSWORD must be set}"

if [ "$POSTGRES_USER" = emqx ]; then
  printf '%s\n' 'The bootstrap user must differ from the emqx broker role.' >&2
  exit 1
fi

# AIIDA requires this fixed broker role. Read the password from the environment
# inside psql so quotes and shell metacharacters remain data.
psql --no-psqlrc --username "$POSTGRES_USER" --dbname "$POSTGRES_DB" \
  --set=ON_ERROR_STOP=1 <<'SQL'
\getenv broker_password MQTT_DATABASE_PASSWORD
CREATE ROLE emqx WITH LOGIN PASSWORD :'broker_password';
SQL
{{- end -}}
