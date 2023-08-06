{{- define "form-auth.tpl" -}}
{{- template "header.tpl" . -}}
					<div class="card o-hidden border-0 shadow-lg my-5">
						<div class="card-body p-0">
							<!-- Nested Row within Card Body -->
							<div class="row">
								<div class="col-lg-5 d-none d-lg-block bg-register-image"></div>
								<div class="col-lg-7">
									<div class="p-5">
										<div class="text-center">
											<h1 class="h4 text-gray-900 mb-4">{{- if .Title -}}{{- .Title -}}{{- else -}}{{- end -}}</h1>
										</div>
										<form class="user" action="{{- .URLPathPrefix -}}/api/auth" method="POST">
											<div class="form-group">
												Object PIN: <input type="text" class="form-control form-control-user" id="objectPIN" placeholder="0000">
												LAL Master Key: <input type="text" class="form-control form-control-user" id="masterKey" placeholder="only you know">
											</div>
											<input type="submit" value="Get Your Token!" class="btn btn-primary btn-user btn-block">
										</form>
									</div>
								</div>
							</div>
						</div>
					</div>
{{ template "footer.tpl" . }}
{{ end }}