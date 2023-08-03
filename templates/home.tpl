{{- define "generic.tpl" -}}
{{- template "header.tpl" . -}}
					<div class="col">
						{{ .Text }}
					</div>	<!-- /col -->
{{ template "footer.tpl" . }}
{{ end }}