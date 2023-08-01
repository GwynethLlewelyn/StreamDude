{{- define "footer.tpl" -}}
			</div> <!-- /row -->
			<div class="row">
				<div class="col-md-12 py-3 border-top">
					<!-- Footer -->
					<footer class="sticky-footer">
						<div class="container-fluid my-auto">
							<div class="row">
								<div class="col-11">
									<div class="copyright text-center my-auto">
										<span>&copy; {{.now }} by <a href="https://gwynethllewelyn.net">Gwyneth Llewelyn</a1>. All rights reserved. Uses some <a href="https://startbootstrap.com/" target=_blank title="Bootstrap"><i class="bi bi-bootstrap-fill" aria-label="Bootstrap"></i></a> love and lots of <a href="https://golang.org/" target=_blank>Go.</a></span>
									</div>	<!-- /copyright -->
								</div>	<!-- /col -->
								<div class="col-1 align-content-end">
									<!-- Scroll to Top Button-->
									<a class="scroll-to-top rounded text-success shadow" href="#page-top">
										<i class="bi bi-arrow-up-circle-fill" aria-label="Scroll to top"></i>
									</a>
								</div>	<!-- /col -->
							</div>	<!-- /row -->
						</div>	<!-- /container-fluid --A
					</footer>	<!-- /footer -->
				</div>	<!-- /col-auto -->
			</div> <!-- /row -->
		</div>	<!-- /content -->
		{{- if .hasCode -}}
		<script src="https://cdnjs.cloudflare.com/ajax/libs/highlight.js/11.5.0/highlight.min.js" integrity="sha512-BNc7saQYlxCL10lykUYhFBcnzdKMnjx5fp5s5wPucDyZ7rKNwCoqJh1GwEAIhuePEK4WM9askJBRsu7ma0Rzvg==" crossorigin="anonymous" referrerpolicy="no-referrer"></script>
		<script src="https://cdnjs.cloudflare.com/ajax/libs/highlight.js/11.5.0/languages/lsl.min.js" integrity="sha512-1ZKgH4N0QMMiPsi5bQ3vNmlWFe0HN0tae+qkGY+XlKHC7J2tkeb7CtCjrbRhCBMa5d+O5hh5OajDqdL3TAuTdQ==" crossorigin="anonymous" referrerpolicy="no-referrer"></script>
		<script>hljs.highlightAll();</script>
		{{- end -}}
	</body>
</html>
{{- end -}}