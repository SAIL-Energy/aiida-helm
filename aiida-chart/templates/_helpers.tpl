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
{{- include "aiida.nanomq" . }}
{{- end }}