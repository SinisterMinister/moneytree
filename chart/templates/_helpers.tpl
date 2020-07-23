{{/* vim: set filetype=mustache: */}}
{{/*
Expand the name of the chart.
*/}}
{{- define "moneytree.name" -}}
{{- default .Chart.Name .Values.moneytree.nameOverride | trunc 63 | trimSuffix "-" }}
{{- end }}

{{/*
Create a default fully qualified app name.
We truncate at 63 chars because some Kubernetes name fields are limited to this (by the DNS naming spec).
If release name contains chart name it will be used as a full name.
*/}}
{{- define "moneytree.fullname" -}}
{{- if .Values.moneytree.fullnameOverride }}
{{- .Values.moneytree.fullnameOverride | trunc 63 | trimSuffix "-" }}
{{- else }}
{{- $name := default .Chart.Name .Values.moneytree.nameOverride }}
{{- if contains $name .Release.Name }}
{{- .Release.Name | trunc 63 | trimSuffix "-" }}
{{- else }}
{{- printf "%s-%s" .Release.Name $name | trunc 63 | trimSuffix "-" }}
{{- end }}
{{- end }}
{{- end }}

{{/*
Create chart name and version as used by the chart label.
*/}}
{{- define "moneytree.chart" -}}
{{- printf "%s-%s" .Chart.Name .Chart.Version | replace "+" "_" | trunc 63 | trimSuffix "-" }}
{{- end }}

{{/*
Common labels
*/}}
{{- define "moneytree.labels" -}}
helm.sh/chart: {{ include "moneytree.chart" . }}
{{ include "moneytree.selectorLabels" . }}
{{- if .Chart.AppVersion }}
app.kubernetes.io/version: {{ .Chart.AppVersion | quote }}
{{- end }}
app.kubernetes.io/managed-by: {{ .Release.Service }}
{{- end }}

{{/*
Selector labels
*/}}
{{- define "moneytree.selectorLabels" -}}
app.kubernetes.io/name: {{ include "moneytree.name" . }}
app.kubernetes.io/instance: {{ .Release.Name }}
{{- end }}

{{/*
Create the name of the service account to use
*/}}
{{- define "moneytree.serviceAccountName" -}}
{{- if .Values.moneytree.serviceAccount.create }}
{{- default (include "moneytree.fullname" .) .Values.moneytree.serviceAccount.name }}
{{- else }}
{{- default "default" .Values.moneytree.serviceAccount.name }}
{{- end }}
{{- end }}
