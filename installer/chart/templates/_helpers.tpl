{{/* Metadata for aiida installer components */}}
{{- define "installer.labels" }}
  generator: helm
  date: {{ now | htmlDate }}
  chart: {{ .Chart.Name }}
  version: {{ .Chart.Version }}
{{- end }}

{{/* Metadata for core components */}}
{{- define "installer.core.labels" }}
labels:
  app: {{ .Release.Name }}
{{- include "installer.labels" . }}
{{- end }}