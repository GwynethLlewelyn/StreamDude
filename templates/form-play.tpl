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
											<h1 class="h4 text-gray-900 mb-4">{{- if .Title -}}{{- .Title -}}{{- else -}}Play{{- end -}}</h1>
										</div>
										<form class="user" action="{{- .URLPathPrefix -}}/api/play" method="POST">
											<div class="form-group">
												<label for="token">Token received during authentication:</label>
												<input type="text" class="form-control form-control-user" id="token" placeholder="Enter your token here"><br>
												<label for="filename">Enter a file name to play on the server:</label>
												<input type="text" class="form-control form-control-user" id="filename" placeholder="streaming-file.mp4"><br>
												<label for="masterKey">Master key for your LAL server:</label>
												<input type="text" class="form-control form-control-user" id="masterKey" placeholder="only-you-know"><br>
											</div>
											<input type="submit" value="Play" class="btn btn-primary btn-user btn-block">
										</form>
									</div>
								</div>
							</div>
						</div>
					</div>
{{ template "footer.tpl" . }}
{{ end }}