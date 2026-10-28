{{/* SPDX-License-Identifier: Apache-2.0 */}}
{{/* Leave room for generated job/volume names and keep every suffix intact. */}}
{{- define "aiida.component-name" -}}
{{- $prefixLength := int (sub 45 (add 1 (len .component))) -}}
{{- $prefix := .root.Release.Name -}}
{{- if gt (len $prefix) $prefixLength -}}
{{- $prefix = printf "%s-%s" ($prefix | trunc (int (sub $prefixLength 9)) | trimSuffix "-") ($prefix | sha256sum | trunc 8) -}}
{{- end -}}
{{- printf "%s-%s" $prefix .component -}}
{{- end -}}
{{- define "aiida.core.name" -}}{{ include "aiida.component-name" (dict "root" . "component" "core") }}{{- end -}}
{{- define "aiida.timescaledb.name" -}}{{ include "aiida.component-name" (dict "root" . "component" "timescaledb") }}{{- end -}}
{{- define "aiida.mosquitto.name" -}}{{ include "aiida.component-name" (dict "root" . "component" "mosquitto") }}{{- end -}}
{{- define "aiida.hivemq.name" -}}{{ include "aiida.component-name" (dict "root" . "component" "hivemq") }}{{- end -}}
{{- define "aiida.ingress.name" -}}{{ include "aiida.component-name" (dict "root" . "component" "ingress") }}{{- end -}}
{{- define "aiida.mqtt.service-name" -}}{{ include "aiida.component-name" (dict "root" . "component" "mqtt") }}{{- end -}}
{{- define "aiida.mqtt.load-balancer-name" -}}{{ include "aiida.component-name" (dict "root" . "component" "mqtt-external") }}{{- end -}}

{{/* Common labels are deterministic; rendering alone must never cause a rollout. */}}
{{- define "aiida.labels" -}}
helm.sh/chart: {{ printf "%s-%s" .Chart.Name .Chart.Version | replace "+" "_" | trunc 63 | trimSuffix "-" | quote }}
app.kubernetes.io/name: {{ .Chart.Name | quote }}
app.kubernetes.io/instance: {{ .Release.Name | quote }}
app.kubernetes.io/version: {{ .Chart.AppVersion | quote }}
app.kubernetes.io/managed-by: {{ .Release.Service | quote }}
{{- end -}}
{{- define "aiida.core.labels" -}}
labels:
  app: {{ include "aiida.core.name" . }}
  app.kubernetes.io/component: core
  {{- include "aiida.labels" . | nindent 2 }}
{{- end -}}
{{- define "aiida.timescaledb.labels" -}}
labels:
  app: {{ include "aiida.timescaledb.name" . }}
  app.kubernetes.io/component: database
  {{- include "aiida.labels" . | nindent 2 }}
{{- end -}}
{{- define "aiida.mosquitto.labels" -}}
labels:
  app: {{ include "aiida.mosquitto.name" . }}
  app.kubernetes.io/component: mosquitto
  {{- include "aiida.labels" . | nindent 2 }}
{{- end -}}
{{- define "aiida.hivemq.labels" -}}
labels:
  app: {{ include "aiida.hivemq.name" . }}
  app.kubernetes.io/component: hivemq
  {{- include "aiida.labels" . | nindent 2 }}
{{- end -}}
{{- define "aiida.mqtt.labels" -}}
labels:
  app: {{ include "aiida.mqtt.service-name" . }}
  app.kubernetes.io/component: mqtt
  {{- include "aiida.labels" . | nindent 2 }}
{{- end -}}
{{- define "aiida.ingress.labels" -}}
labels:
  app: {{ include "aiida.ingress.name" . }}
  app.kubernetes.io/component: ingress
  {{- include "aiida.labels" . | nindent 2 }}
{{- end -}}

{{/* Accept either an image map or a dict containing an image map. */}}
{{- define "aiida.image" -}}
{{- $image := . -}}
{{- if hasKey . "image" -}}{{- $image = .image -}}{{- end -}}
{{- $repository := required "An image repository is required" $image.repository -}}
{{- if $image.digest -}}
{{- printf "%s@%s" $repository $image.digest -}}
{{- else -}}
{{- printf "%s:%s" $repository (required "An image tag or digest is required" $image.tag) -}}
{{- end -}}
{{- end -}}

{{- define "aiida.core.keystore-secret-name" -}}
{{- default (printf "%s-keystore" (include "aiida.core.name" .)) .Values.core.keystore.existingSecret -}}
{{- end -}}
{{- define "aiida.core.pvc-name" -}}
{{- default (printf "%s-data" (include "aiida.core.name" .)) .Values.core.persistence.existingClaim -}}
{{- end -}}
{{- define "aiida.timescaledb.secret-name" -}}
{{- default (printf "%s-secret" (include "aiida.timescaledb.name" .)) .Values.timescaledb.existingSecret -}}
{{- end -}}
{{- define "aiida.mqtt.secret-name" -}}
{{- default (printf "%s-secret" (include "aiida.mqtt.service-name" .)) .Values.mqtt.existingSecret -}}
{{- end -}}
{{- define "aiida.timescaledb.backup-secret-name" -}}
{{- default (printf "%s-backup-secret" (include "aiida.timescaledb.name" .)) .Values.timescaledb.backup.s3.existingSecret -}}
{{- end -}}
{{- define "aiida.timescaledb.host" -}}
{{- if .Values.timescaledb.enabled -}}
{{- include "aiida.timescaledb.name" . -}}
{{- else -}}
{{- required "timescaledb.host is required when timescaledb.enabled=false" .Values.timescaledb.host -}}
{{- end -}}
{{- end -}}
{{- define "aiida.timescaledb.jdbc-url" -}}
jdbc:postgresql://{{ include "aiida.timescaledb.host" . }}:{{ .Values.timescaledb.port }}/{{ include "aiida.timescaledb.database" . }}?sslmode={{ .Values.timescaledb.sslMode }}{{ if .Values.timescaledb.tls.existingSecret }}&sslrootcert=/etc/database-tls/ca.crt{{ end }}
{{- end -}}
{{/* Fixed names expected by AIIDA's PostgreSQL integration. */}}
{{- define "aiida.timescaledb.database" -}}aiida{{- end -}}
{{- define "aiida.mqtt.database-user" -}}emqx{{- end -}}
{{- define "aiida.ingress.protocol" -}}
{{- if .Values.ingress.tls.enabled -}}https{{- else -}}http{{- end -}}
{{- end -}}
{{- define "aiida.core.external-host" -}}
{{- if .Values.ingress.enabled -}}
{{- printf "%s://%s" (include "aiida.ingress.protocol" .) (required "ingress.host is required when ingress.enabled=true" .Values.ingress.host) -}}
{{- else -}}
{{- required "core.publicUrl is required when ingress.enabled=false" .Values.core.publicUrl | trimSuffix "/" -}}
{{- end -}}
{{- end -}}
{{/* Public application URL includes the ingress path; CORS uses the origin above. */}}
{{- define "aiida.core.public-url" -}}
{{- if .Values.ingress.enabled -}}
{{- printf "%s%s" (include "aiida.core.external-host" .) (default "" .Values.ingress.subPath) -}}
{{- else -}}
{{- required "core.publicUrl is required when ingress.enabled=false" .Values.core.publicUrl | trimSuffix "/" -}}
{{- end -}}
{{- end -}}
{{- define "aiida.ingress.tls-secret-name" -}}
{{- default (printf "%s-tls" (include "aiida.component-name" (dict "root" . "component" "web"))) .Values.ingress.tls.existingSecret -}}
{{- end -}}
{{- define "aiida.mqtt.broker" -}}
{{- if not (has .Values.mqtt.broker (list "mosquitto" "hivemq")) -}}
{{- fail "mqtt.broker must be mosquitto or hivemq; EMQX is not supported" -}}
{{- end -}}
{{- .Values.mqtt.broker -}}
{{- end -}}
{{- define "aiida.mqtt.tls-enabled" -}}
{{- if and .Values.mqtt.externalAccess.enabled .Values.mqtt.externalAccess.tls.enabled -}}true{{- else -}}false{{- end -}}
{{- end -}}
{{- define "aiida.mqtt.tls-secret-name" -}}
{{- default (include "aiida.ingress.tls-secret-name" .) .Values.mqtt.externalAccess.tls.existingSecret -}}
{{- end -}}
{{- define "aiida.mqtt.internal-host" -}}
tcp://{{ include "aiida.mqtt.service-name" . }}:1883
{{- end -}}
{{- define "aiida.mqtt.external-hostname" -}}
{{- default .Values.ingress.host .Values.mqtt.externalAccess.host -}}
{{- end -}}
{{- define "aiida.mqtt.external-host" -}}
{{- if not .Values.mqtt.externalAccess.enabled -}}
{{- include "aiida.mqtt.internal-host" . -}}
{{- else if .Values.mqtt.externalAccess.tls.enabled -}}
ssl://{{ include "aiida.mqtt.external-hostname" . }}:8883
{{- else -}}
tcp://{{ include "aiida.mqtt.external-hostname" . }}:1883
{{- end -}}
{{- end -}}
{{- define "aiida.keycloak.external-host" -}}
{{- required "keycloak.externalHost is required" .Values.keycloak.externalHost | trimSuffix "/" -}}
{{- end -}}
{{- define "aiida.keycloak.internal-host" -}}
{{- default (include "aiida.keycloak.external-host" .) .Values.keycloak.internalHost | trimSuffix "/" -}}
{{- end -}}

{{/* Cross-field checks accompany values.schema.json and avoid unusable installs. */}}
{{- define "aiida.validate" -}}
{{- if not .Values.timescaledb.existingSecret -}}
{{- $_ := required "Database credentials require timescaledb.existingSecret or timescaledb.username and timescaledb.password" .Values.timescaledb.username -}}
{{- $_ := required "Database credentials require timescaledb.existingSecret or timescaledb.username and timescaledb.password" .Values.timescaledb.password -}}
{{- if and .Values.timescaledb.enabled (eq .Values.timescaledb.username "emqx") -}}
{{- fail "The bundled database bootstrap username must differ from the emqx broker role" -}}
{{- end -}}
{{- end -}}
{{- if not .Values.mqtt.existingSecret -}}
{{- $_ := required "MQTT credentials require mqtt.existingSecret or mqtt.superuserPassword and mqtt.db.password" .Values.mqtt.superuserPassword -}}
{{- $_ := required "MQTT credentials require mqtt.existingSecret or mqtt.superuserPassword and mqtt.db.password" .Values.mqtt.db.password -}}
{{- end -}}
{{- if and .Values.timescaledb.enabled (ne .Values.timescaledb.sslMode "disable") -}}
{{- fail "The bundled database does not configure TLS; use an external database for timescaledb.sslMode other than disable" -}}
{{- end -}}
{{- if and (has .Values.timescaledb.sslMode (list "verify-ca" "verify-full")) (not .Values.timescaledb.tls.existingSecret) -}}
{{- fail "Verified database TLS requires timescaledb.tls.existingSecret with the trusted CA certificate" -}}
{{- end -}}
{{- if and .Values.timescaledb.enabled (ne (int .Values.timescaledb.port) 5432) -}}
{{- fail "timescaledb.port must be 5432 for the bundled database; it configures external databases only" -}}
{{- end -}}
{{- if and .Values.timescaledb.enabled .Values.timescaledb.host -}}
{{- fail "Set timescaledb.enabled=false when configuring timescaledb.host" -}}
{{- end -}}
{{- if and .Values.ingress.enabled .Values.core.publicUrl -}}
{{- fail "core.publicUrl is only used when ingress.enabled=false" -}}
{{- end -}}
{{- if and .Values.ingress.enabled .Values.ingress.tls.enabled (not (or .Values.ingress.tls.existingSecret .Values.ingress.tls.clusterIssuer)) -}}
{{- fail "HTTPS ingress requires ingress.tls.existingSecret or ingress.tls.clusterIssuer" -}}
{{- end -}}
{{- if .Values.mqtt.externalAccess.enabled -}}
{{- $_ := required "External MQTT requires mqtt.externalAccess.host or ingress.host" (include "aiida.mqtt.external-hostname" .) -}}
{{- if not (or .Values.mqtt.externalAccess.tls.enabled .Values.mqtt.externalAccess.allowPlaintext) -}}
{{- fail "External MQTT requires TLS or an explicit allowPlaintext=true" -}}
{{- end -}}
{{- if and .Values.mqtt.externalAccess.tls.enabled (not .Values.mqtt.externalAccess.tls.existingSecret) (not (and .Values.ingress.enabled .Values.ingress.tls.enabled)) -}}
{{- fail "External MQTT TLS without HTTPS ingress requires mqtt.externalAccess.tls.existingSecret" -}}
{{- end -}}
{{- end -}}
{{- if .Values.timescaledb.backup.enabled -}}
{{- if not .Values.timescaledb.backup.s3.existingSecret -}}
{{- $_ := required "S3 credentials require timescaledb.backup.s3.existingSecret or accessKey and secretKey" .Values.timescaledb.backup.s3.accessKey -}}
{{- $_ := required "S3 credentials require timescaledb.backup.s3.existingSecret or accessKey and secretKey" .Values.timescaledb.backup.s3.secretKey -}}
{{- end -}}
{{- $_ := required "Backups require timescaledb.backup.s3.bucket" .Values.timescaledb.backup.s3.bucket -}}
{{- $_ := required "Backups require timescaledb.backup.s3.region" .Values.timescaledb.backup.s3.region -}}
{{- end -}}
{{- end -}}
