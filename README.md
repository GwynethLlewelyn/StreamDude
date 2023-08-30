![StreamDude Logo](assets/logos/streamdude-logo-128x128.png)

# StreamDude

[![Go](https://github.com/GwynethLlewelyn/StreamDude/actions/workflows/go.yml/badge.svg)](https://github.com/GwynethLlewelyn/StreamDude/actions/workflows/go.yml) [![Codacy Security Scan](https://github.com/GwynethLlewelyn/StreamDude/actions/workflows/codacy.yml/badge.svg)](https://github.com/GwynethLlewelyn/StreamDude/actions/workflows/codacy.yml)

A work in progress simple playlist manager to remotely send files to a streaming server.

Licensed under a [MIT License](https://gwyneth-llewelyn.mit-license.org/).

## Purpose

There are a million gorgeous tools out there which allow you to manage your local playlist and send it to a streaming server; there are million more — such as OBS — which, while not _stricly_ designed to do just that, can very easily be adapted for that purpose. And there are a *b*illion that even allow you to stream (using so-called progressive download) files from a webpage to your friends?

But what if your music is not stored locally, or you wish to avoid the inevitable delays (latency and/or jitter) for sending the files from your computer to a remote server, and stream from _there_? Think of an Internet radio station, managed by several parties, with a common repository of music (or video!), all stored and managed remotely. Oh, of course: and for free.

Well, you're out of luck.

The oldest tool I know that does just that is Apple's own Darwin Streaming Server — which is even free and open-source — but it's overkill if you already have an external streaming server that you connect with.

There are a few complex ones out there, some even for free, but installing them is a pain.

Thus, StreamDude.

## No, really?

Well, no, I'm lying.

Actually what I _really_ need **right now** is a simple way to get scripts from inside the [Second Life®](https://secondlife.com) and [OpenSimulator](http://opensimulator.org) to be able to remotely select what streams to play. I could do it with a single Python command line argument (no need to run or compile anything or write any code whatsoever). But what would be the fun in that?

Instead, I'm sort of joining three separate projects into one.

## Environment variables

-   `LAL_MASTER_KEY` - because it's too dangerous to keep it in code and/or files
-   `STREAMER_URL` - another way to override the streamer URL; may be useful in scripts

Also, StreamDude attempts to comply with the informal `CLICOLOR_FORCE` and `NO_COLOR` conventions. See https://bixense.com/clicolors/ and https://no-color.org/.

## Compile & launch

You can get [here the full API for Postman](extras/StreamDude.postman_collection.json).

Currently, the only streaming server supported is [lal (Live And Live)](https://github.com/q191201771/lal).

1. Make sure you have the streaming server running first!
2. `go install github.com/GwynethLLewelyn/StreamDude@latest` or, if you prefer, `git clone https://github.com/GwynethLLewelyn/StreamDude`.
3. If you cloned the repo, then run `go build` (and possibly with `go install` you'll get the compiled binary under `~/go/bin`, which, hopefully, is part of your `$PATH`)
4. `LAL_MASTER_KEY=blahblehblih ./StreamDude -d` (if you wish debugging to console, or redirect it to a log file)
5. `/usr/bin/curl --header "Content-Type: application/json" --header "Accept: application/json" --request GET    http://127.0.0.1:3554/ping` — should give `{"message":"pong back to 127.0.0.1","status":"ok"}`
6. `/usr/bin/curl --header "Content-Type: application/json" --header "Accept: application/json" --request POST   --data '{ "objectPIN": "0000" }' http://127.0.0.1:3554/api/auth` — should give you an authentication token, e.g. `ZmFrZXRva2Vu`
7. `/usr/bin/curl --header "Content-Type: application/json" --header "Accept: application/json" --request POST   --data '{ "token": "ZmFrZXRva2Vu", "filename": "/path/to/video.mp4"  }' http://127.0.0.1:3554/api/play` — should launch ffmpeg and send `video.mp4` to be streamed
8. For streaming a whole playlist, you will need to have the ALSA utils installed — currently, streaming a playlist requires the [VLC libraries](https://www.videolan.org/vlc/) as well as the `alsa-utils` package (on Linux and FreeBSD).
9. For security issues, you should only expose the `/media` directory for playlist streaming purposes; you _can_ place a symbolic link in there, pointing to your media library, but be aware of the issues when doing that.

**Note:** `objectPIN` and `token` are not really, really being enforced — there is no database/KV store backend yet, but as soon as there is one, I've put the validation code in place, so you should fill in those fields.

Also note that there are further fields for Second Life®/OpenSimulator, all of which are being ignored right now.

## Backoffice

Under construction. Authentication, of course, is fake.

The home page, properly speaking, may become an instance of [MusicFolderPlayer](https://github.com/ltguillaume/music-folder-player/) (the inspiration for this project).

## Nginx conf sample

### Assumptions

1. You're running Unix-like environment (WSL2 will possibly work, too);
2. `nginx` is pointing to `http(s)://my.streaming.server` under `/var/www/my.streaming.server`;
3. **StreamDude** is installed under `/StreamDude` (i.e. `/templates`, `/assets` and even `/media` are there);
4. It runs as a service under `localhost:3445` (default)

Then you will need something like this:

```nginx
location ~* /StreamDude/assets {
	rewrite ^/StreamDude/assets/(.*)$ /$1 break;
	root /var/www/my.streaming.server/StreamDude/assets;
	expires max;
	add_header Cache-Control public;
	fastcgi_hide_header Set-Cookie;
	try_files $uri =404;
}

location ~* /StreamDude/templates {
	rewrite ^/StreamDude/templates/(.*)$ /$1 break;
	root /var/www/my.streaming.server/StreamDude/templates;
	expires max;
	add_header Cache-Control public;
	fastcgi_hide_header Set-Cookie;
	try_files $uri =404;
}

location /StreamDude/ {
	proxy_pass_request_headers on;
	proxy_pass_request_body on;
	proxy_set_header X-Forwarded-For   $remote_addr;
	proxy_set_header X-Forwarded-Proto $scheme;
	proxy_set_header Host              $host;
	proxy_set_header X-Real-IP         $remote_addr;
	proxy_pass_header CF-Connecting-IP;
	proxy_pass_header CF-IPCountry;
	proxy_pass_header Set-Cookie;
	proxy_buffering off;
	proxy_ssl_server_name on;
	proxy_read_timeout 5m;
	proxy_set_header Access-Control-Allow-Credentials true;
	proxy_set_header Content-Encoding gzip;
	proxy_pass http://127.0.0.1:3554;
}

location ~* ^.+\.(xml|ogg|ogv|svg|svgz|eot|otf|woff|mp4|ttf|css|rss|atom|js|jpg|jpeg|gif|png|ico|zip|tgz|gz|rar|bz2|doc|xls|exe|ppt|tar|mid|midi|wav|bmp|rtf|lsl|lua)$ {
		access_log off; log_not_found off; expires max;
		add_header Cache-Control public;
		fastcgi_hide_header Set-Cookie;
}
```

and launch StreamDude (in debug mode) with:

```bash
$ ./StreamDude -d -r rtsp://127.0.0.1:5544/ -u /StreamDude -x my.streaming.server
```

Add `-P ":443"` if your front-end server is running HTTPS.

If you're launching StreamDude directly from the root of your virtual host (i.e. no `/StreamDude` subfolder), then you might need to add a trailing slash on `proxy_pass http://127.0.0.1:3554/;`. Getting the slashes to match properly is always messy.

## Launching from `systemd`

If you're running a Unix version supporting `systemd`, you can grab a [sample unit service file](extras/StreamDude.service.sample) to adapt to your needs. StreamDude complies with the [`sd_notify`](https://www.man7.org/linux/man-pages/man3/sd_notify.3.html) specifications and tries to play nicely with `systemd`.

Coloured `journald` logs are yet to be implemented, but at least you can get them using `journalctl -u StreamDude -f`.

## Third-party dependencies and thanks

-   [Gin](https://gin-gonic.com/), of course.
-   Thanks as well to the team that provided [access to the VLC libraries using Go](https://github.com/adrg/libvlc-go), or else I couldn't stream from a disk directory stored on the server :-P

Music sample files for testing, under `./media/Kevin MacLeod/Rock harder`, are licensed under the following license:

> "Big Rock", "Cool Hard Facts", "El Magicia", "Gearhead", "Neolith", "Sax Rock and Roll" and "What You Want"  
> by Kevin MacLeod (incompetech.com)  
> Licensed under Creative Commons: By Attribution 3.0  
> https://creativecommons.org/licenses/by/3.0/  

## Release notes

See [CHANGELOG.md](CHANGELOG.md).
