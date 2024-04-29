{{- define "php-version" -}}
{{- printf "%s" .Values.phpVersion -}}
{{- end -}}

{{- define "build-image-suffix" -}}
-debian-bullseye-rootless
{{- end -}}