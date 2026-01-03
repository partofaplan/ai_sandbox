{{- define "ai-image-stack.name" -}}
{{- default .Chart.Name .Values.nameOverride | trunc 63 | trimSuffix "-" -}}
{{- end -}}

{{- define "ai-image-stack.fullname" -}}
{{- if .Values.fullnameOverride -}}
{{- .Values.fullnameOverride | trunc 63 | trimSuffix "-" -}}
{{- else -}}
{{- .Release.Name | trunc 63 | trimSuffix "-" -}}
{{- end -}}
{{- end -}}

{{- define "ai-image-stack.labels" -}}
app.kubernetes.io/name: {{ include "ai-image-stack.name" . }}
helm.sh/chart: {{ printf "%s-%s" .Chart.Name .Chart.Version | replace "+" "_" }}
app.kubernetes.io/instance: {{ .Release.Name }}
app.kubernetes.io/managed-by: {{ .Release.Service }}
{{- end -}}

{{- define "ai-image-stack.selectorLabels" -}}
app.kubernetes.io/name: {{ include "ai-image-stack.name" . }}
app.kubernetes.io/instance: {{ .Release.Name }}
{{- end -}}

{{- define "ai-image-stack.genName" -}}
{{- printf "%s-gen-ui" (include "ai-image-stack.fullname" .) | trunc 63 | trimSuffix "-" -}}
{{- end -}}

{{- define "ai-image-stack.galleryName" -}}
{{- printf "%s-gallery" (include "ai-image-stack.fullname" .) | trunc 63 | trimSuffix "-" -}}
{{- end -}}
