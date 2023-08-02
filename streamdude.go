// StreamDude
//
// Environment variables used when launching:
//
// `LAL_MASTER_KEY` - because it's too dangerous to keep it in code and/or files
// `STREAMER_URL` - another way to override the streamer URL; may be useful in scripts
//
// © 2023 by Gwyneth Llewelyn. All rights reserved.
package main

import (
	//	"log"
	"net/http"
	"os"
	"path"
	"path/filepath"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"

	//	"github.com/google/martian/log"
	flag "github.com/karrick/golf" // flag replacement library
	"github.com/mattn/go-isatty"
	"github.com/sirupsen/logrus"
)

// Globals

var (
	help bool				// if set, show usage
	ffmpegPath string		// path to ffmpeg executable
	ginMode *string			// ginMode is `debug` for development, `release` for production.
	host string				// this host — where StreamDude is running.
	serverPort string		// port where StreamDude server is listening
	externalPort string		// external port if using a reverse proxy
	templatePath string		// where templates are held
	pathToStaticFiles string // where static assets are stored
	workingDirectory string // workingDirectory is the result of os.Getwd() or "." if that fails.
	urlPathPrefix string	// URL path prefix
	lslSignaturePIN string	// what we send from LSL
	debug bool				// Set to debug level

	// use a single instance of Validate, it caches struct info
	validate *validator.Validate

	// Global logger using logrus.
	logme = logrus.New()

	// Stuff for the lal streaming server
	streamerURL string		// RTSP streaming URL for lal
	lalMasterKey string		// too dangerous to show, put into LAL_MASTER_KEY environment
)

func main() {
	// Extract things from command line
	flag.BoolVarP(&help,			'h', "help",			false, 			"show command usage")
	flag.StringVarP(&ffmpegPath,	'm', "ffmpeg",			"/usr/local/bin/ffmpeg", "path to ffmpeg executable")
	flag.StringVarP(&host,			'j', "host",			"localhost", 	"server host where we're running")
	flag.StringVarP(&serverPort,	'p', "port", 			":3554", 		"port where StreamDude server is listening")
	flag.StringVarP(&externalPort,	'P', "external",		":80",			"external port if using a reverse proxy")
	flag.StringVarP(&templatePath,	't', "templatepath",	"./templates",	"where templates are held")
	flag.StringVarP(&pathToStaticFiles, 's', "staticpath",	"./assets",		"where static assets are stored")
	flag.StringVarP(&urlPathPrefix,	'u', "urlprefix",		"",				"URL path prefix")
	flag.StringVarP(&lslSignaturePIN, 'l',	"lslpin",		"0000",			"LSL signature PIN")
	flag.BoolVarP(&debug,			'd', "debug",			false, 			"set debug level (omit for normal logs)")
	flag.StringVarP(&streamerURL,	'r', "streamer",		"rtsp://127.0.0.1:554/",	"streamer URL")
	flag.StringVarP(&lalMasterKey,	'k', "masterkey",		"",				"lal server master key")

	flag.Parse()

	if help {
		flag.Usage()
		os.Exit(0)
	}

	// setup a single instance of the validator service.
	validate = validator.New()

	/**
	 * Starting backend web server using Gin Gonic.
	 */
	router := gin.Default()
	router.Delims("{{", "}}") // stick to default delims for Go templates.
	router.SetTrustedProxies(nil)	// as per https://pkg.go.dev/github.com/gin-gonic/gin#readme-don-t-trust-all-proxies (gwyneth 20220111).
	router.TrustedPlatform = gin.PlatformCloudflare	// we're running behind Cloudflare CDN

	// Configure logrus.
	//	log.SetFlags(log.LstdFlags | log.Lshortfile)

	logme.Formatter = new(logrus.TextFormatter)
	logme.Formatter.(*logrus.TextFormatter).DisableColors = false		// keep colors
	logme.Formatter.(*logrus.TextFormatter).DisableTimestamp = false	// keep timestamp

	// set debug level, depending on the argument value
	if (debug) {
		logme.SetLevel(logrus.DebugLevel)
	}
	// logme.Debugf("Output descriptor: %+v\n", logme.Out)
	logme.Debugf("Debug level set to %q\n", logme.GetLevel().String())

	// respect CLICOLOR_FORCE and NO_COLOR in Gin (logrus is already compliant)
	// Figure out if we're running in a terminal, and, if so, apply all the relevant commands
	// See https://bixense.com/clicolors/ and https://no-color.org/ (gwyneth 20230731)

	var isTerm = true	// are we logging to a tty?

	// for the weird type casting, see https://github.com/mattn/go-isatty/issues/80#issuecomment-1470096598 (gwyneth 20230801)
	if os.Getenv("TERM") == "dumb" ||
		(!isatty.IsTerminal(logme.Out.(*os.File).Fd()) && !isatty.IsCygwinTerminal(logme.Out.(*os.File).Fd())) {
			isTerm = false
	}
	if _, ok := os.LookupEnv("NO_COLOR"); ok {
		gin.DisableConsoleColor()
	} else if _, ok := os.LookupEnv("CLICOLOR_FORCE"); ok && isTerm {
		gin.ForceConsoleColor()
	}

	// Override lal master key from environment.
	if temp := os.Getenv("LAL_MASTER_KEY"); temp != "" {
		lalMasterKey = temp
	}
	logme.Debugf("lal key (obfuscated): %q\n", obfuscate(lalMasterKey))

	// Override streamer, if env exists.
	if temp := os.Getenv("STREAMER_URL"); temp != "" {
		streamerURL = temp
	}
	// Validate that the streamer has a valid URL (either from command-line or env var).
	if err := validate.Var(streamerURL, "required,url"); err != nil {
		logme.Fatalf("invalid streamer URL: %q, aborting\n", streamerURL)
	}
	logme.Infof("remote streamer URL set to: %q\n", streamerURL)

	// Setup templating system.

	var err error	// needed for scope issues
	if workingDirectory, err = os.Getwd(); err != nil {
		workingDirectory = "."	// if os.Getwd() fails, use local directory, maybe it works (gwyneth 2022011.
		// no need to panic, this error is 'fixed'!
	}
	// Figure out where the templates are: deal with empty path.
	if templatePath == "" {
		templatePath = filepath.Join(workingDirectory, "/templates")
	}
	htmlGlobFilePath := filepath.Join(templatePath, "/*.tpl")
	logme.Infof("loading templates from pathToStaticFiles: %q, templatePath: %q, final destination: %q\n",
		pathToStaticFiles, templatePath, htmlGlobFilePath)

	router.LoadHTMLGlob(htmlGlobFilePath)

	// Some useful static dirs & files
	router.Static(path.Join(urlPathPrefix, "/assets"), filepath.Join(pathToStaticFiles, "/assets"))

	router.StaticFile(path.Join(urlPathPrefix, "/favicon.ico"), filepath.Join(pathToStaticFiles, "/assets/favicons/favicon.ico"))
	router.StaticFile(path.Join(urlPathPrefix, "/browserconfig.xml"), filepath.Join(pathToStaticFiles, "/assets/favicons/browserconfig.xml"))
	router.StaticFile(path.Join(urlPathPrefix, "/site.webmanifest"), filepath.Join(pathToStaticFiles, "/assets/favicons/site.webmanifest"))

	// Make the router handle these exceptions with better HTTP error codes
	router.HandleMethodNotAllowed = true
	router.RedirectTrailingSlash = true
	router.RedirectFixedPath = true

	// Ping handler (who knows, it might be useful in some contexts... such as Let's Encrypt certificates
	router.Any(path.Join(urlPathPrefix, "/ping"), func(c *gin.Context) {
		payload := "pong back to "
		// check if we're behind Cloudflare
		if c.GetHeader(gin.PlatformCloudflare) != "" {
			payload += c.GetHeader(gin.PlatformCloudflare)	// this is CF-Connecting-IP from Cloudflare
			if c.GetHeader("CF-IPCountry") != "" {			// this will usually be set by Cloudflare, too
				payload += "(from " + c.GetHeader("CF-IPCountry") + ")"
			}
		} else {
			payload += c.RemoteIP()
		}
		logme.Debugf("Ping request had Content-Type set to %q and accepts %q\n", c.ContentType(), c.GetHeader("Accept"))

		contentType := c.GetHeader("Accept")
		if contentType == "" {
			contentType = c.ContentType()
		}

		if contentType == "*/*" {
			contentType = "application/json"
		}

		switch contentType {
			case "application/json":
				c.JSON(http.StatusOK, gin.H{"status": "ok", "message": payload})
			case "text/html":
				c.HTML(http.StatusOK, "generic.tpl", environment(c, gin.H{
					"Title"			: http.StatusMethodNotAllowed,
					"description"	: http.StatusText(http.StatusOK),
					"Text"			: payload,
				}))
			case "text/xml":
			case "application/soap+xml":
			case "application/xml":
				c.XML(http.StatusOK, gin.H{"status": "ok", "message": payload})
			default:
				c.String(http.StatusOK, payload)
		}
	})

	// Lower-leval API for
	apiRoutes := router.Group(path.Join(urlPathPrefix, "/api"))
	{		// base page for complex scripts.
		apiRoutes.POST("/play",	apiStreamFile)
		apiRoutes.POST("/auth",	apiSimpleAuthGenKey)
	}

	// Catch all other routes and send back an error
	router.NoRoute(func(c *gin.Context) {
		errorMessage := "Command " + c.Request.URL.Path + " not found."

		contentType := c.GetHeader("Accept")
		if contentType == "" {
			contentType = c.ContentType()
		}

		if contentType == "*/*" {
			contentType = "application/json"
		}

		switch contentType {
			case "application/json":
				c.JSON(http.StatusNotFound, gin.H{"status": "error", "message": errorMessage})
			case "text/html":
				c.HTML(http.StatusNotFound, "generic.tpl", environment(c, gin.H{
					"Title"			: http.StatusNotFound,
					"description"	: http.StatusText(http.StatusNotFound),
					"Text"			: errorMessage,
				}))
			case "text/xml":
			case "application/soap+xml":
			case "application/xml":
				c.XML(http.StatusNotFound, gin.H{"status": "error", "message": errorMessage})
			default:
				c.String(http.StatusNotFound, errorMessage)
		}
	})

	router.NoMethod(func(c *gin.Context) {
		errorMessage := "Method " + c.Request.Method + " not allowed."

		contentType := c.GetHeader("Accept")
		if contentType == "" {
			contentType = c.ContentType()
		}

		if contentType == "*/*" {
			contentType = "application/json"
		}

		switch contentType {
			case "application/json":
				c.JSON(http.StatusMethodNotAllowed, gin.H{"status": "error", "message": errorMessage})
			case "text/html":
				c.HTML(http.StatusMethodNotAllowed, "generic.tpl", environment(c, gin.H{
					"Title"			: http.StatusMethodNotAllowed,
					"description"	: http.StatusText(http.StatusMethodNotAllowed),
					"Text"			: errorMessage,
				}))
			case "text/xml":
			case "application/soap+xml":
			case "application/xml":
				c.XML(http.StatusMethodNotAllowed, gin.H{"status": "error", "message": errorMessage})
			default:
				c.String(http.StatusMethodNotAllowed, errorMessage)
		}
	})

	/*
	 *  Launch the server (finally) and log an error if it crashes.
	 */

	// this might require another layer to check for https
	logme.Fatal(router.Run(host + serverPort))
}

/*
 * Some special cases
 */

