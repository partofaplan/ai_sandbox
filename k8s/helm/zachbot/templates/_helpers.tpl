{{- define "zachbot.name" -}}
{{- .Chart.Name -}}
{{- end -}}

{{- define "zachbot.fullname" -}}
{{- printf "%s-%s" .Release.Name .Chart.Name -}}
{{- end -}}
