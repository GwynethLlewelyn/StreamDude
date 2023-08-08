{{- define "form-play.tpl" -}}
{{- template "header.tpl" . -}}
					<div class="card o-hidden border-0 shadow-lg my-5">
						<div class="card-body p-0">
							<!-- Nested Row within Card Body -->
							<div class="row">
								<div class="col-lg-5 d-none d-lg-block bg-register-image"></div>
								<div class="col-lg-7">
									<div class="p-5">
										<div class="text-center">
											<h1 class="h4 text-gray-900 mb-4"><i class="bi bi-music-note-beamed" aria-hidden="true"></i>&nbsp;{{- if .Title -}}{{- .Title -}}{{- else -}}Play{{- end -}}</h1>
										</div>
										<form role="form" class="user" action="{{- .URLPathPrefix -}}/api/play" method="POST">
											<div class="form-group input-group">
												<label for="token" class="col-form-label">Token received during authentication:</label>
												<input type="text" class="form-control form-control-user" id="token" name="token" placeholder="Enter your token here" size=32 autofocus required>
											</div>
											<div class="form-group input-group">
												<label for="filename" class="col-form-label">Enter a file name to play on the server:</label>
												<input type="text" class="form-control form-control-user" id="filename" name="filename" placeholder="~/videos/streaming-file.mp4" size=64 required>
											</div>
											<div class="form-group input-group">
												<label for="masterKey" class="col-form-label">Master key for your LAL server:</label>
												<input type="text" class="form-control form-control-user" id="masterKey" name="masterKey" placeholder="only-you-know" size=32 required>
											</div>
											<input type="submit" value="Play" class="btn btn-primary btn-user btn-sm">
										</form>
									</div>
								</div>
							</div>
						</div>
					</div>
{{ template "footer.tpl" . }}
{{ end }}