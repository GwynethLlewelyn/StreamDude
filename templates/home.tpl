{{- define "generic.tpl" -}}
{{- template "header.tpl" . -}}
					<div class="col">
						{{- if .Title -}}
						<h1>{{- .Title -}}</h1>
						{{- end -}}
						{{ .Text }}
					</div>	<!-- /col -->
{{ template "footer.tpl" . }}
{{ end }}