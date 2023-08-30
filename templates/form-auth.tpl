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
											<h1 class="h4 text-gray-900 mb-4"><i class="bi bi-person-lock" aria-hidden="true"></i>&nbsp;{{- if .Title -}}{{- .Title -}}{{- else -}}Authentication{{- end -}}</h1>
										</div>
										<form role="form" class="user" action="{{- .URLPathPrefix -}}api/auth" method="POST">
											<div class="form-group input-group">
												<label for="objectPIN" class="col-form-label">4-digit Object PIN:</label>
												<input type="number" max=9999 min=0 maxlength=4 minlength=4 size=4 class="form-control form-control-user" id="objectPIN" name="objectPIN" placeholder="0000" autofocus required>
											</div>
											<div class="form-group input-group">
												<label for="masterKey" class="col-form-label">Master key for your LAL server:</label>
												<input type="text" class="form-control form-control-user" id="masterKey" name="masterKey" placeholder="only you know" size=32 required>
											</div>
											<input type="submit" value="Get Your Token!" class="btn btn-primary btn-user btn-sm">
										</form>
									</div>
								</div>
							</div>
						</div>
					</div>
{{ template "footer.tpl" . }}
{{ end }}