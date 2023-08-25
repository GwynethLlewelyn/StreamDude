{{- define "generic.tpl" -}}
{{- template "header.tpl" . -}}
					<div class="col">
						{{- if .Title -}}
						<h1>{{- .Title -}}</h1>
						{{- end -}}
						{{- if .setBanner -}}
						<div class="alert alert-success alert-dismissible fade show" role="alert">
						  {{ .Text }}
						  <button type="button" class="close" data-dismiss="alert" aria-label="Close">
							<span aria-hidden="true">&times;</span>
						  </button>
						</div>
						{{- end -}}
						<div class="container d-flex justify-content-center">
							<ul class="list-group mt-5 text-white">
								{{- range $file := .mediaDirectory -}}
								<li class="list-group-item d-flex justify-content-between align-content-center">
									<div class="d-flex flex-row">
										{{- if $file.IsDir -}}
										<i class="bi bi-folder-fill" style="font-size: 40px; color: var(--yellow);" aria-hidden="true">
										{{- else -}}
										<i class="bi bi-music-note-beamed" style="font-size: 40px; color: var(--purple);" aria-hidden="true">
										{{- end -}}
										<div class="ml-2 filename-{{- $file.Name -}}">
											<h6 class="mb-0">{{- $file.Name -}}</h6>
											<div class="about">
												<span><integer>{{- $file.Size -}}</integer> bytes</span>
												<span><time datetime="{{- /* $file.ModTime */ -}}">{{- /* $file.ModTime */ -}}</time></span>
											</div>
										</div>
									</div>
									{{- if not $file.IsDir -}}
									<div class="check">
										<input type="checkbox" name="checkbox-{{- $file.Name -}}">
									</div>
									{{- end -}}
								</li>
								{{- end -}}<!-- loop -->
							</ul>
						</div>
						{{- if not .setBanner -}}
						{{ .Text }}
						{{- end -}}
					</div>	<!-- /col -->
{{ template "footer.tpl" . }}
{{ end }}