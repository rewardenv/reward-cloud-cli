{{- if eq .Values.projectType "magento" -}}
{{- include "magento-dockerfile-start" . | nindent 0 -}}
{{- else if eq .Values.projectType "wordpress" -}}
{{- include "wordpress-dockerfile-start" . | nindent 0 -}}
{{- else if eq .Values.projectType "shopware" -}}
{{- include "shopware-dockerfile-start" . | nindent 0 -}}
{{- else if eq .Values.projectType "magento-pwa" -}}
{{- include "magento-pwa-dockerfile-start" . | nindent 0 -}}
{{- else if eq .Values.projectType "shopware-pwa" -}}
{{- include "shopware-pwa-dockerfile-start" . | nindent 0 -}}
{{- end }}

{{- range $key, $value := .Values.envVars }}
ARG {{ printf "%s=%s" $key $value }}
ENV {{ printf "%s $%s" $key $key }}
{{- end }}

{{- if eq .Values.projectType "magento" -}}
{{- include "magento-dockerfile-end" . | nindent 0 -}}
{{- else if eq .Values.projectType "wordpress" -}}
{{- include "wordpress-dockerfile-end" . | nindent 0 -}}
{{- else if eq .Values.projectType "shopware" -}}
{{- include "shopware-dockerfile-end" . | nindent 0 -}}
{{- else if eq .Values.projectType "magento-pwa" -}}
{{- include "magento-pwa-dockerfile-end" . | nindent 0 -}}
{{- else if eq .Values.projectType "shopware-pwa" -}}
{{- include "shopware-pwa-dockerfile-end" . | nindent 0 -}}
{{- end }}
