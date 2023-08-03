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

## Compile & launch

You can get [here the full API for Postman](assets/StreamDude.postman_collection).

Currently, the only streaming server supported is [lal (Live And Live)](https://github.com/q191201771/lal).

1. Make sure you have the streaming server running.
2. `go build`
3. `LAL_MASTER_KEY = ./StreamDude -d` (if you wish debugging to console, or redirect it to a log file)
4. `/usr/bin/curl --header "Content-Type: application/json" --header "Accept: application/json" --request GET    http://127.0.0.1:3554/ping` — should give `{"message":"pong back to 127.0.0.1","status":"ok"}`
5. `/usr/bin/curl --header "Content-Type: application/json" --header "Accept: application/json" --request POST   --data '{ "objectPIN": "0000" }' http://127.0.0.1:3554/api/auth` — should give you an authentication token, e.g. `ZmFrZXRva2Vu`
6. `/usr/bin/curl --header "Content-Type: application/json" --header "Accept: application/json" --request POST   --data '{ "token": "ZmFrZXRva2Vu", "filename": "/path/to/video.mp4"  }' http://127.0.0.1:3554/api/play` — should launch ffmpeg and send `video.mp4` to be streamed

**Note:** `objectPIN` and `token` are not really, really being enforced — there is no database/KV store backend yet, but as soon as there is one, I've put the validation code in place, so you should fill in those fields.

Also note that there are further fields for Second Life®/OpenSimulator, all of which are being ignored right now.

## Backoffice

Currently not implemented. There is just a home page and a Ping test. The remaining menu areas are not really working (they will become forms for testing purposes).

The home page, properly speaking, will become an instance of [MusicFolderPlayer](https://github.com/ltguillaume/music-folder-player/) (the inspiration for this project).