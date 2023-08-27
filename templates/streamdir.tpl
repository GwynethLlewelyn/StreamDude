{{- define "streamdir.tpl" -}}
{{- template "header.tpl" . -}}
					<div class="col">
						{{- if .Title -}}
						<h1>{{- .Title -}}</h1>
						{{- end -}}
						{{- if .mediaDirectory -}}
						<div class="row">
							<div class="alert alert-info" role="info">
							{{ .mediaDirectory }}
							</div>
						</div>
						{{- end -}}
						{{- if .setBanner -}}
						<div class="row">
							<div class="alert alert-success alert-dismissible fade show" role="alert">
								{{ .Text }}
								<button type="button" class="close" data-dismiss="alert" aria-label="Close">
									<span aria-hidden="true">&times;</span>
								</button>
							</div>
						</div>
						{{- end -}}
						<div class="row">
							<div class="col-lg-5 d-none d-lg-block bg-register-image"></div>
							<div class="col-lg-7">
								<div class="p-5">
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
														<!-- note: all checkboxes checked & disabled for now -->
														<div class="check">
															<input type="checkbox" id="checkbox-{{- $file.Name -}}" name="{{- $file.Name -}}" disabled="disabled" checked>
														</div>
														{{- end -}}
													</li>
													{{- end -}}<!-- loop -->
												</ul>
											</div> <!-- /container d-flex -->
											<input type="submit" value="Stream" class="btn btn-primary btn-user btn-sm">
									</form>
								</div> <!-- /p-5 -->
							</div> <!-- /col lg-7 -->
						</div> <!-- row -->
						{{- if not .setBanner -}}
						<div class="row">
							{{ .Text }}
						</div>
						{{- end -}}
					</div> <!-- /col -->
{{ template "footer.tpl" . }}
{{ end }}