{{- define "streamdir.tpl" -}}
{{- template "header.tpl" . -}}
					<div class="col">
						{{- if .Title -}}
						<h1>{{- .Title -}}</h1>
						{{- end -}}
						{{- if .mediaDirectory -}}
						<div class="alert alert-info" role="info">
						{{ .mediaDirectory }}
						</div>
						{{- end -}}
						{{- if .setBanner -}}
						<div class="alert alert-success alert-dismissible fade show" role="alert">
							{{ .Text }}
							<button type="button" class="close" data-dismiss="alert" aria-label="Close">
								<span aria-hidden="true">&times;</span>
							</button>
						</div>
						{{- end -}}
						<form role="form" class="user" action="{{- .URLPathPrefix -}}/api/stream" method="POST">
							<div class="container d-flex justify-content-center">
								<ul class="list-group mt-5 text-white">
									{{- range $file := .playlist -}}
									<li class="list-group-item d-flex justify-content-between align-content-center">
										<div class="d-flex flex-row">
											{{- if $file.IsDir -}}
											<i class="bi bi-folder-fill" style="font-size: 40px; color: var(--yellow);" aria-hidden="true"></i>
											{{- else -}}
											<i class="bi bi-music-note-beamed" style="font-size: 40px; color: var(--purple);" aria-hidden="true"></i>
											{{- end -}}
											<div class="ml-2 filename-{{- $file.Name -}}">
												<h6 class="mb-0">{{- $file.Name -}}</h6>
												<div class="about">
													<span>
														<integer>{{- $file.Size -}}</integer> bytes
													</span>
													<span><time datetime="{{- formatAsDate $file.ModTime -}}">{{- formatAsDate $file.ModTime -}}</time></span>
												</div>
											</div>
										</div> <!-- /d-flex flex-row -->
										{{- if not $file.IsDir -}}
										<div class="check">
											<input type="checkbox" name="checkbox-{{- $file.Name -}}">
										</div>
										{{- end -}}
									</li>
									{{- end -}}<!-- loop -->
								</ul>
								<input type="submit" value="Stream" class="btn btn-primary btn-user btn-sm">
							</div> <!-- /container d-flex -->
						</form>
						{{- if not .setBanner -}}
						{{ .Text }}
						{{- end -}}
					</div> <!-- /col -->
{{ template "footer.tpl" . }}
{{ end }}