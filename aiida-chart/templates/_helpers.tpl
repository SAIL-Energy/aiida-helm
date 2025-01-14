{{/* Metadata for aiida components */}}
{{- define "aiida.labels" }}
  generator: helm
  date: {{ now | htmlDate }}
  chart: {{ .Chart.Name }}
  version: {{ .Chart.Version }}
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