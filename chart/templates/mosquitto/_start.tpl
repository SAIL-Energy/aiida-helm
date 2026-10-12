{{/* SPDX-License-Identifier: Apache-2.0. Runtime content for a ConfigMap. */}}
{{- define "aiida.mosquitto.start" -}}
#!/bin/sh
set -eu
umask 077

if [ -z "${MOSQUITTO_DB_PASSWORD:-}" ] || [ -z "${AIIDA_MQTT_PASSWORD:-}" ]; then
    echo "Error: MOSQUITTO_DB_PASSWORD and AIIDA_MQTT_PASSWORD must be set." >&2
    exit 1
fi

# Reading the environment in awk avoids evaluating passwords as code. The
# plugin inserts this option into a libpq connection string, so quote it for
# that second parser as well, preserving spaces, quotes, and backslashes.
awk '
    BEGIN {
        password = ENVIRON["MOSQUITTO_DB_PASSWORD"]
        if (password ~ /[\r\n]/) {
            print "Error: MOSQUITTO_DB_PASSWORD cannot contain line breaks." > "/dev/stderr"
            exit 1
        }
        escaped_password = ""
        for (character_index = 1; character_index <= length(password); character_index++) {
            character = substr(password, character_index, 1)
            if (character == "\\" || character == "\047") {
                escaped_password = escaped_password "\\"
            }
            escaped_password = escaped_password character
        }
    }
    /^auth_opt_pg_password([[:space:]]|$)/ {
        print "auth_opt_pg_password \047" escaped_password "\047"
        next
    }
    { print }
' /config/mosquitto.conf > /tmp/mosquitto.conf

/mosquitto/pw -h bcrypt -p "${AIIDA_MQTT_PASSWORD}" > /tmp/mosquitto-password-hash
{
    printf 'aiida:'
    cat /tmp/mosquitto-password-hash
} > /tmp/mosquitto-passwords

cat > /tmp/mosquitto-acls <<'EOF'
user aiida
topic readwrite #
EOF

if [ "${MOSQUITTO_TLS_ENABLED:-false}" != "true" ]; then
    exec /usr/sbin/mosquitto -c /tmp/mosquitto.conf
fi

# The Secret volume updates its files when cert-manager renews the certificate.
# Mosquitto reloads both certificate and key when it receives SIGHUP.
certificate_fingerprint() {
    cksum /etc/mqtt-tls/tls.crt /etc/mqtt-tls/tls.key
}

previous_fingerprint=$(certificate_fingerprint)
broker_pid=''
watcher_pid=''

cleanup() {
    if [ -n "$watcher_pid" ]; then
        kill -TERM "$watcher_pid" 2>/dev/null || true
        wait "$watcher_pid" 2>/dev/null || true
    fi
}

stop_broker() {
    trap - TERM INT
    if [ -n "$broker_pid" ]; then
        kill -TERM "$broker_pid" 2>/dev/null || true
        wait "$broker_pid" 2>/dev/null || true
    fi
    exit 0
}

trap cleanup EXIT
trap stop_broker TERM INT

/usr/sbin/mosquitto -c /tmp/mosquitto.conf &
broker_pid=$!
(
    while kill -0 "$broker_pid" 2>/dev/null; do
        sleep 30
        if current_fingerprint=$(certificate_fingerprint 2>/dev/null); then
            if [ "$current_fingerprint" != "$previous_fingerprint" ]; then
                if kill -HUP "$broker_pid" 2>/dev/null; then
                    previous_fingerprint=$current_fingerprint
                    echo "MQTT TLS certificate changed; requested Mosquitto reload."
                fi
            fi
        fi
    done
) &
watcher_pid=$!

if wait "$broker_pid"; then
    exit 0
else
    exit "$?"
fi
{{- end -}}
