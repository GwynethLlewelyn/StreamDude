{{- define "generic.tpl" -}}
{{- template "header.tpl" . -}}
					<div class="col">
						{{- if .Title -}}
						<h1>{{- .Title -}}</h1>
						{{- end -}}
						<div class="container d-flex justify-content-center">
							<ul class="list-group mt-5 text-white">
								{{- range $file := .mediaDirectory -}}
								<li class="list-group-item d-flex justify-content-between align-content-center">
									<div class="d-flex flex-row">
										<i class="bi bi-music-note-beamed" style="font-size: 40px; color: var(--yellow);" aria-hidden="true">
										<!-- width 40 -->
										<div class="ml-2">
											<h6 class="mb-0">Turbine parts</h6>
											<div class="about">
												<span>802 Files</span>
												<span>Jan 29, 2020</span>
											</div>
										</div>
									</div>
									<div class="check">
										<input type="checkbox" name="a">
									</div>
								</li>
								{{- end -}}
							</ul>
						</div>
						{{ .Text }}
					</div>	<!-- /col -->
{{ template "footer.tpl" . }}
{{ end }}