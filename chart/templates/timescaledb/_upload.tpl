{{/* SPDX-License-Identifier: Apache-2.0. Runtime content for a ConfigMap. */}}
{{- define "aiida.timescaledb.upload" -}}
#!/bin/sh
set -eu

: "${S3_REGION:?S3_REGION must be set}"
: "${S3_BUCKET:?S3_BUCKET must be set}"
filename=$(cat /backup/filename)
case "$filename" in
  ''|*/*|*..*)
    printf '%s\n' 'Invalid backup filename; refusing upload.' >&2
    exit 1
    ;;
esac

backup_path="/backup/${filename}"
[ -s "$backup_path" ] || { printf '%s\n' 'Backup archive is missing.' >&2; exit 1; }
prefix=${S3_PREFIX:-}
prefix=${prefix#/}
prefix=${prefix%/}
object_key=${prefix:+${prefix}/}${filename}

# A custom endpoint is optional for AWS S3. Quote every argument so values
# containing shell metacharacters are never evaluated as shell commands.
set -- s3 cp "$backup_path" "s3://${S3_BUCKET}/${object_key}" \
  --region "$S3_REGION" --only-show-errors
if [ -n "${S3_ENDPOINT:-}" ]; then
  set -- "$@" --endpoint-url "$S3_ENDPOINT"
fi
aws "$@"
printf '%s\n' 'Backup upload completed successfully.'
{{- end -}}
