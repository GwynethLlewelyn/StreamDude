{{- define "favicons.tpl" -}}
		<link
			rel="icon"
			type="image/svg+xml"
			href="{{- .URLPathPrefix -}}/assets/favicons/favicon.svg"
		/>
		<link
			rel="icon"
			type="image/png"
			sizes="48x48"
			href="{{- .URLPathPrefix -}}/assets/favicons/favicon.png"
		/>
		<link
			rel="apple-touch-icon"
			sizes="180x180"
			href="{{- .URLPathPrefix -}}/assets/favicons/apple-touch-icon.png"
		/>
		<link
			rel="icon"
			type="image/png"
			sizes="32x32"
			href="{{- .URLPathPrefix -}}/assets/favicons/favicon-32x32.png"
		/>
		<link
			rel="icon"
			type="image/png"
			sizes="16x16"
			href="{{- .URLPathPrefix -}}/assets/favicons/favicon-16x16.png"
		/>
		<link rel="manifest" href="{{- .URLPathPrefix -}}/assets/favicons/site.webmanifest" />
		<link
			rel="mask-icon"
			href="{{- .URLPathPrefix -}}/assets/favicons/safari-pinned-tab.svg"
			color="#5bbad5"
		/>
		<link rel="shortcut icon" href="{{- .URLPathPrefix -}}/assets/favicons/favicon.ico" />
		<meta name="msapplication-TileColor" content="#00a300" />
		<meta
			name="msapplication-TileImage"
			content="{{- .URLPathPrefix -}}/assets/favicons/mstile-144x144.png"
		/>
		<meta
			name="msapplication-config"
			content="{{- .URLPathPrefix -}}/assets/favicons/browserconfig.xml"
		/>
{{- end -}}