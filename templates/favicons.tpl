{{- define "favicons.tpl" -}}
		<link rel="apple-touch-icon" sizes="180x180" href="{{- .URLPathPrefix -}}assets/favicons/apple-touch-icon.png">
		<link rel="icon" type="image/png" sizes="32x32" href="{{- .URLPathPrefix -}}assets/favicons/favicon-32x32.png">
		<link rel="icon" type="image/png" sizes="16x16" href="{{- .URLPathPrefix -}}assets/favicons/favicon-16x16.png">
		<link rel="manifest" href="{{- .URLPathPrefix -}}assets/favicons/site.webmanifest">
		<link rel="mask-icon" href="{{- .URLPathPrefix -}}assets/favicons/safari-pinned-tab.svg" color="#5bbad5">
		<link rel="shortcut icon" href="{{- .URLPathPrefix -}}assets/favicons/favicon.ico">
		<meta name="msapplication-TileColor" content="#da532c">
		<meta name="msapplication-config" content="{{- .URLPathPrefix -}}assets/favicons/browserconfig.xml">
		<meta name="theme-color" content="#ffffff">
{{- end -}}