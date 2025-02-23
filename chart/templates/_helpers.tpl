{{/* Metadata for aiida components */}}
{{- define "aiida.labels" }}
  generator: helm
  date: {{ now | htmlDate }}
  chart: {{ .Chart.Name }}
  version: {{ .Chart.Version }}
{{- end }}

{{/* Metadata for core components */}}
{{- define "aiida.core.labels" }}
labels:
  app: {{ .Release.Name }}-core
{{- include "aiida.labels" . }}
{{- end }}

{{/* Metadata for timescaledb components */}}
{{- define "aiida.timescaledb.labels" }}
labels:
  app: {{ .Release.Name }}-timescaledb
{{- include "aiida.labels" . }}
{{- end }}

{{/* Metadata for nanomq components */}}
{{- define "aiida.nanomq.labels" }}
labels:
  app: {{ .Release.Name }}-nanomq
{{- include "aiida.labels" . }}
{{- end }}

{{/* Metadata for teleinfo2mqtt components */}}
{{- define "aiida.teleinfo2mqtt.labels" }}
labels:
  app: {{ .Release.Name }}-teleinfo2mqtt
{{- include "aiida.labels" . }}
{{- end }}

{{/* Metadata for keycloak components */}}
{{- define "aiida.keycloak.labels" }}
labels:
  app: {{ .Release.Name }}-keycloak
{{- include "aiida.labels" . }}
{{- end }}

{{/* Helper for http protocol */}}
{{- define "aiida.ingress.protocol" }}
{{- if .Values.ingress.tlsEnabled }}https{{- else }}http{{- end }}
{{- end }}

{{/* Helper to build the postgres url */}}
{{- define "aiida.timescaledb.jdbc-url" -}}
"jdbc:postgresql://{{ .Release.Name }}-timescaledb:{{ .Values.timescaledb.port }}/{{ .Values.timescaledb.database }}"
{{- end }}

{{/* Helper to build the external core host */}}
{{- define "aiida.core.external-host" -}}
"{{ include "aiida.ingress.protocol" . }}://{{ .Values.ingress.host }}{{ .Values.core.path }}"
{{- end }}

{{/* Helper to build the internal keycloak host */}}
{{- define "aiida.keycloak.external-host" -}}
"{{ include "aiida.ingress.protocol" . }}://{{ .Values.ingress.host }}{{ .Values.keycloak.path }}"
{{- end }}

{{/* Helper to build the internal keycloak host */}}
{{- define "aiida.keycloak.internal-host" -}}
"http://{{ .Release.Name }}-keycloak:{{ .Values.keycloak.port }}{{ .Values.keycloak.path }}"
{{- end }}
