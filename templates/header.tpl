{{- define "header.tpl" -}}
<!DOCTYPE html>
<html lang="en">
	<head>
		<meta charset="utf-8" />
		<meta http-equiv="X-UA-Compatible" content="IE=edge" />
		<meta
			name="viewport"
			content="width=device-width, initial-scale=1, shrink-to-fit=no"
		/>
		<meta name="description" content="{{- .description -}}" />
		<meta name="author" content="{{- .author -}}" />
		<meta name="robots" content="noindex, nofollow" />
		{{- template "favicons.tpl" . -}}
		<!--	<meta name="theme-color" content="#ffffff" /> -->
		<title>{{- if .Title -}}{{- .Title -}}{{- else -}}{{- .titleCommon -}}{{- end -}}{{- if .description -}} | {{- .description -}}{{- end -}}</title>
		<meta http-equiv="x-dns-prefetch-control" content="on" />
		<link rel="preconnect" href="https://cdnjs.cloudflare.com" />
		<link rel="preconnect" href="https://fonts.googleapis.com" />
		<link rel="preconnect" href="https://fonts.gstatic.com" crossorigin />
		<link href="https://fonts.googleapis.com/css2?family=Work+Sans:ital,wght@0,400;0,700;1,400;1,700&display=swap" rel="stylesheet">
		{{- if .hasCode -}}
		<link href="https://fonts.googleapis.com/css2?family=Fira+Code&display=swap" rel="stylesheet" />
		{{- end -}}
		<link rel="stylesheet" href="https://cdnjs.cloudflare.com/ajax/libs/bootstrap-icons/1.10.5/font/bootstrap-icons.min.css" integrity="sha512-ZnR2wlLbSbr8/c9AgLg3jQPAattCUImNsae6NHYnS9KrIwRdcY9DxFotXhNAKIKbAXlRnujIqUWoXXwqyFOeIQ==" crossorigin="anonymous" referrerpolicy="no-referrer" />
		{{- if .hasCode -}}
		<link rel="stylesheet" href="https://cdnjs.cloudflare.com/ajax/libs/highlight.js/11.5.0/styles/nord.min.css" integrity="sha512-U/cZqAAOThvb4J9UCt/DWkkjoJWHXvutFDS/nZmZlirci2ZMuH6qFokOQDuuKgE7pXD+FmhDNH2jT43x0GreCQ==" crossorigin="anonymous" referrerpolicy="no-referrer" />
		{{- end -}}
		<!-- make sure that jQuery is always the first script to be loaded, no matter what! -->
		<script src="https://cdnjs.cloudflare.com/ajax/libs/jquery/3.7.0/jquery.min.js" integrity="sha512-3gJwYpMe3QewGELv8k/BX9vcqhryRdzRMxVfq6ngyWXwo03GFEzjsUm8Q7RZcHPHksttq7/GFoxjCVUjkjvPdw==" crossorigin="anonymous" referrerpolicy="no-referrer"></script>
		<!-- 	To reuse the slim version (more compact, removes useless things), we need to change the xmlhttprequest
				and use HTML5 built-in `fetch()` instead. TODO(gwyneth), 20220329
				<script src="https://cdnjs.cloudflare.com/ajax/libs/jquery/3.6.0/jquery.slim.min.js" integrity="sha512-6ORWJX/LrnSjBzwefdNUyLCMTIsGoNP6NftMy2UAm1JBm6PRZCO1d7OHBStWpVFZLO+RerTvqX/Z9mBFfCJZ4A==" crossorigin="anonymous" referrerpolicy="no-referrer"></script>
		-->
		<script src="https://cdnjs.cloudflare.com/ajax/libs/bootstrap/4.6.2/js/bootstrap.min.js" integrity="sha512-7rusk8kGPFynZWu26OKbTeI+QPoYchtxsmPeBqkHIEXJxeun4yJ4ISYe7C6sz9wdxeE1Gk3VxsIWgCZTc+vX3g==" crossorigin="anonymous" referrerpolicy="no-referrer"></script>
		{{- if .hasEditor -}}
		<link rel="stylesheet" href="https://cdnjs.cloudflare.com/ajax/libs/codemirror/5.65.2/codemirror.min.css" integrity="sha512-uf06llspW44/LZpHzHT6qBOIVODjWtv4MxCricRxkzvopAlSWnTf6hpZTFxuuZcuNE9CBQhqE0Seu1CoRk84nQ==" crossorigin="anonymous" referrerpolicy="no-referrer" />
		<link rel="stylesheet" href="https://cdnjs.cloudflare.com/ajax/libs/codemirror/5.65.2/theme/nord.min.css" integrity="sha512-sPc4jmw78pt6HyMiyrEt3QgURcNRk091l3dZ9M309x4wM2QwnCI7bUtsLnnWXqwBMECE5YZTqV6qCDwmC2FMVA==" crossorigin="anonymous" referrerpolicy="no-referrer" />
		<script src="https://cdnjs.cloudflare.com/ajax/libs/codemirror/5.65.2/codemirror.min.js" integrity="sha512-xwrAU5yhWwdTvvmMNheFn9IyuDbl/Kyghz2J3wQRDR8tyNmT8ZIYOd0V3iPYY/g4XdNPy0n/g0NvqGu9f0fPJQ==" crossorigin="anonymous" referrerpolicy="no-referrer"></script>
		<script src="https://cdnjs.cloudflare.com/ajax/libs/codemirror/5.65.2/keymap/emacs.min.js" integrity="sha512-JRRAFgTvx2rg2AO6juzwLSqaBfA5MVnZAdnWNwgsLIAnjYsMI6liEnBjFgIbskM3oi5hHBLGCzwZUMd6Nsee8Q==" crossorigin="anonymous" referrerpolicy="no-referrer"></script>
		<script src="https://cdnjs.cloudflare.com/ajax/libs/codemirror/5.65.2/mode/lua/lua.min.js" integrity="sha512-MXR/wu8WxkFikybMYGuaR9O0SgRrcSReZUNuherC0XZ7SJN/db3W+qQCh+4rAiBBeNk/yd/NdnQd/s2nO4q4fA==" crossorigin="anonymous" referrerpolicy="no-referrer"></script>
		{{- end -}}
		{{- if (or .gridName .agGridJS) -}}
		<!-- bootstrap3-dialog does not work with Bootstrap 4/5
		<link rel="stylesheet" href="https://cdnjs.cloudflare.com/ajax/libs/bootstrap3-dialog/1.35.4/css/bootstrap-dialog.min.css" integrity="sha512-PvZCtvQ6xGBLWHcXnyHD67NTP+a+bNrToMsIdX/NUqhw+npjLDhlMZ/PhSHZN4s9NdmuumcxKHQqbHlGVqc8ow==" crossorigin="anonymous" referrerpolicy="no-referrer" />
		<script src="https://cdnjs.cloudflare.com/ajax/libs/bootstrap3-dialog/1.35.4/js/bootstrap-dialog.min.js" integrity="sha512-LbO5ZwEjd9FPp4KVKsS6fBk2RRvKcXYcsHatEapmINf8bMe9pONiJbRWTG9CF/WDzUig99yvvpGb64dNQ27Y4g==" crossorigin="anonymous" referrerpolicy="no-referrer"></script>
		-->
		<!-- Trying to use bootstrap bundled with popper, because otherwise, this must come *first* (gwyneth 202230802)
		<script src="https://cdnjs.cloudflare.com/ajax/libs/popper.js/2.11.5/umd/popper.min.js" integrity="sha512-8cU710tp3iH9RniUh6fq5zJsGnjLzOWLWdZqBMLtqaoZUA6AWIE34lwMB3ipUNiTBP5jEZKY95SfbNnQ8cCKvA==" crossorigin="anonymous" referrerpolicy="no-referrer"></script>-->
		<link rel="stylesheet" href="{{- .URLPathPrefix -}}/assets/css/bootstrap4-dialog.css" />
		<script src="{{- .URLPathPrefix -}}/assets/js/bootstrap4-dialog.js"></script>
		<!-- Call agGrid -->
		<link rel="stylesheet" href="https://cdnjs.cloudflare.com/ajax/libs/ag-grid/Docs-27.1.0-20220316/styles/ag-grid.min.css" integrity="sha512-nCEk9jlAm2EovHp0fAuD2ZdW7PuHXa4/2U7RWOae0p8bnFat2DJ77IjTaoY+Nh/Ith8P13iDVOWAvkAEgD6IQQ==" crossorigin="anonymous" referrerpolicy="no-referrer" />
		<link rel="stylesheet" href="https://cdnjs.cloudflare.com/ajax/libs/ag-grid/Docs-27.1.0-20220316/styles/ag-theme-alpine-dark.min.css" integrity="sha512-JP97wY1K1lnrZnyUOg+BviTgUGkfmX7nvfTA9HhsGnkSIGwTp/KmsKiGbZEz3N3JiUZFKlXw3233N0FGGbP3PQ==" crossorigin="anonymous" referrerpolicy="no-referrer" />
		<script src="https://cdnjs.cloudflare.com/ajax/libs/ag-grid/Docs-27.1.0-20220316/ag-grid-community.min.noStyle.js" integrity="sha512-RMhS9dNrbhSpMQyj+Mi/kJqdks8IwkVDI2AUsK7HIFKY+Nb90Ajp96pyvDSY9nPcB2qOEXUw043glP0ObxFlpg==" crossorigin="anonymous" referrerpolicy="no-referrer"></script>
		<!-- we're using an external function to geneate UUIDs in JavaScript: -->
		<script src="https://cdnjs.cloudflare.com/ajax/libs/uuid/8.3.2/uuidv4.min.js" integrity="sha512-BCMqEPl2dokU3T/EFba7jrfL4FxgY6ryUh4rRC9feZw4yWUslZ3Uf/lPZ5/5UlEjn4prlQTRfIPYQkDrLCZJXA==" crossorigin="anonymous" referrerpolicy="no-referrer"></script>
		<script src="{{- .URLPathPrefix -}}/assets/js/random-people.js?cache-buster={{- .cacheBuster -}}"></script>
		<!-- these are our own JS support functions for agGrid -->
		<script src="{{- .URLPathPrefix -}}/assets/js/{{- .agGridJS -}}?cache-buster={{- .cacheBuster -}}"></script>
		{{- end -}}
		<!-- Nord theme comes after the above, so that it might overwrite things in case of need -->
		<link rel="stylesheet" href="{{- .URLPathPrefix -}}/assets/css/nordbootstrap.css?cache-buster={{- .cacheBuster -}}" />
		<!-- our own CSS at the end, so that everything from CodeMirror can be overridden (gwyneth 20220327) -->
		<link rel="stylesheet" href="{{- .URLPathPrefix -}}/assets/css/style.css?cache-buster={{- .cacheBuster -}}" />
	</head>
	<body id="page-top">
		<!-- this allows us to get URLPathPrefix from JS, if needed later -->
		<span id="URLPathPrefix" hidden>{{- .URLPathPrefix -}}</span>
		<div class="container-fluid py-5">
			<!-- A grey horizontal navbar that becomes vertical on small screens.
			Used to have 'fixed-top' but this seems to put everything too transparent for my taste... -->
			<nav class="navbar navbar-expand-md navbar-light bg-secondary text-primary fixed-top border-bottom">
				<a class="navbar-brand" href="{{- .URLPathPrefix -}}">
					<img src="{{- .URLPathPrefix -}}/assets/logos/streamdude-logo-128x128.png" class="svg-logo bg-white rounded shadow-lg" alt="StreamDude Logo - Home" title="StreamDude Logo - Home">
				</a>
				<!-- Toggler/collapsible Button -->
				<button class="navbar-toggler navbar-light" type="button" data-toggle="collapse" data-target="#collapsibleNavbar">
					<span class="navbar-toggler-icon navbar-light"></span>
				</button>
				<div class="collapse navbar-collapse" id="collapsibleNavbar">
					<ul class="navbar-nav">
						<li class="nav-item">
							<a class="nav-link" href="{{- .URLPathPrefix -}}/ping"><i class="bi bi-broadcast-pin" aria-hidden="true"></i>&nbsp;Ping</a>
						</li>
						<li class="nav-item">
							<a class="nav-link" href="{{- .URLPathPrefix -}}/api/auth"><i class="bi bi-person-lock" aria-hidden="true"></i>&nbsp;Authentication</a>
						</li>
						<li class="nav-item">
							<a class="nav-link" href="{{- .URLPathPrefix -}}/api/play"><i class="bi bi-music-note-beamed" aria-hidden="true"></i>&nbsp;Play Stream</a>
						</li>
						<!--
						<li class="nav-item dropdown">
							<a class="nav-link dropdown-toggle" data-toggle="dropdown" href="#"><i class="bi bi-hdd-rack"></i>&nbsp;Database</a>
							<div class="dropdown-menu bg-primary border shadow-lg">
								<a class="dropdown-item" href="{{- .URLPathPrefix -}}/admin/database/agents">Agents</a>
								<a class="dropdown-item" href="{{- .URLPathPrefix -}}/admin/database/inventory">Inventory</a>
								<a class="dropdown-item" href="{{- .URLPathPrefix -}}/admin/database/objects">Objects</a>
							</div>
						</li>
						<li class="nav-item dropdown">
							<a class="nav-link dropdown-toggle" data-toggle="dropdown" href="#"><i class="bi bi-gear"></i>&nbsp;Debug</a>
							<div class="dropdown-menu bg-primary border shadow-lg">
								<a class="dropdown-item" href="{{- .URLPathPrefix -}}/admin/restbot-sessions">List RESTbot Sessions (XML)</a>
								<a class="dropdown-item" href="{{- .URLPathPrefix -}}/admin/exit-bots">Exit all RESTbots</a>
							</li>
						</li>-->
						<li class="nav-item">
							<a class="nav-link" href="{{- .URLPathPrefix -}}/credits"><i class="bi bi-info-circle" aria-hidden="true">&nbsp;</i>Credits</a>
						</li>
					</ul>
				</div>
			</nav>
		</div>	<!-- /container-fluid -->
		<div class="container-fluid">
			<div class="content">
				<div class="row clearfix">
{{ end }}
