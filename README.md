![StreamDude Logo](assets/logos/streamdude-logo-128x128.png)
# StreamDude

A work in progress simple playlist manager to remotely send files to a streaming server.

Licensed under a [MIT License](https://gwyneth-llewelyn.mit-license.org/).

## Purpose

There are a million gorgeous tools out there which allow you to manage your local playlist and send it to a streaming server; there are million more — such as OBS — which, while not *stricly* designed to do just that, can very easily be adapted for that purpose. And there are a *b*illion that even allow you to stream (using so-called progressive download) files from a webpage to your friends?

But what if your music is not stored locally, or you wish to avoid the inevitable delays (latency and/or jitter) for sending the files from your computer to a remote server, and stream from *there*? Think of an Internet radio station, managed by several parties, with a common repository of music (or video!), all stored and managed remotely. Oh, of course: and for free.

Well, you're out of luck.

The oldest tool I know that does just that is Apple's own Darwin Streaming Server — which is even free and open-source — but it's overkill if you already have an external streaming server that you connect with.

There are a few complex ones out there, some even for free, but installing them is a pain.

Thus, StreamDude.

## No, really?

Well, no, I'm lying.

Actually what I *really* need **right now** is a simple way to get scripts from inside the [Second Life®](https://secondlife.com) and [OpenSimulator](http://opensimulator.org) to be able to remotely select what streams to play. I could do it with a single Python command line argument (no need to run or compile anything or write any code whatsoever). But what would be the fun in that?

Instead, I'm sort of joining three separate projects into one.