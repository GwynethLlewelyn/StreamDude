// StreamDude

package main

import (
	//	"log"
	"net/http"
	"os"
	"path"
	"path/filepath"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	flag "github.com/karrick/golf" // flag replacement library
	"github.com/mattn/go-isatty"
	log "github.com/sirupsen/logrus"
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

	// use a single instance of Validate, it caches struct info
	validate *validator.Validate
)

func main() {
	// Extract things from command line
	flag.BoolVarP(&help,			'h', "help",			false, "show command usage")
	flag.StringVarP(&ffmpegPath,	'm', "ffmpeg",			"/usr/local/bin/ffmpeg", "path to ffmpeg executable")
	flag.StringVarP(&host,			'j', "host",			"localhost", 	"server host where we're running")
	flag.StringVarP(&serverPort,	'p', "port", 			":3554", 		"port where StreamDude server is listening")
	flag.StringVarP(&externalPort,	'P', "external",		":80",			"external port if using a reverse proxy")
	flag.StringVarP(&templatePath,	't', "templatepath",	"./templates",	"where templates are held")
	flag.StringVarP(&pathToStaticFiles, 's', "staticpath",	"./assets",		"where static assets are stored")
	flag.StringVarP(&urlPathPrefix,	'u', "urlprefix",		"",				"URL path prefix")
	flag.StringVarP(&lslSignaturePIN, 'l',	"lslpin",		"0000",			"LSL signature PIN")

	flag.Parse()

	if help {
		flag.Usage()
		os.Exit(0)
	}

	// setup logrus logger
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	// setup a single instance of the validator service
	validate = validator.New()

	/**
	 * Starting backend web server using Gin Gonic.
	 */
	router := gin.Default()
	router.Delims("{{", "}}") // stick to default delims for Go templates.
	router.SetTrustedProxies(nil)	// as per https://pkg.go.dev/github.com/gin-gonic/gin#readme-don-t-trust-all-proxies (gwyneth 20220111).
	router.TrustedPlatform = gin.PlatformCloudflare	// we're running behind Cloudflare CDN

	// respect CLICOLOR_FORCE and NO_COLOR in Gin (logrus is already compliant)
	// Figure out if we're running in a terminal, and, if so, apply all the relevant commands
	// See https://bixense.com/clicolors/ and https://no-color.org/ (gwyneth 20230731)

	var isTerm = true	// are we logging to a tty?

	if os.Getenv("TERM") == "dumb" ||
		(!isatty.IsTerminal(log.Out) && !isatty.IsCygwinTerminal(log.Out)) {
		isTerm = false
	}
	if _, ok := os.LookupEnv("NO_COLOR"); ok {
		gin.DisableConsoleColor()
	} else if _, ok := os.LookupEnv("CLICOLOR_FORCE"); ok && isTerm {
		gin.ForceConsoleColor()
	}

	// Setup templating system
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
	log.Printf("loading templates from pathToStaticFiles: %q, templatePath: %q, final destination: %q\n",
		pathToStaticFiles, templatePath, htmlGlobFilePath)

	router.LoadHTMLGlob(htmlGlobFilePath)

	// Some useful static dirs & files
	router.Static(path.Join(urlPathPrefix, "/assets"), filepath.Join(pathToStaticFiles, "/assets"))

	router.StaticFile(path.Join(urlPathPrefix, "/favicon.ico"), filepath.Join(pathToStaticFiles, "/assets/favicons/favicon.ico"))
	router.StaticFile(path.Join(urlPathPrefix, "/browserconfig.xml"), filepath.Join(pathToStaticFiles, "/assets/favicons/browserconfig.xml"))
	router.StaticFile(path.Join(urlPathPrefix, "/site.webmanifest"), filepath.Join(pathToStaticFiles, "/assets/favicons/site.webmanifest"))

	// Ping handler (who knows, it might be useful in some contexts... such as Let's Encrypt certificates
	router.GET(path.Join(urlPathPrefix, "/ping"), func(c *gin.Context) {
		c.String(http.StatusOK, "pong to " + c.RemoteIP())
	})

	// Lower-leval API for
	apiRoutes := router.Group(path.Join(urlPathPrefix, "/api"))
	{		// base page for complex scripts.
		apiRoutes.POST("/play",	apiStreamFile)
		apiRoutes.POST("/auth",	apiSimpleAuthGenKey)
	}

	// Catch all other routes and send back an error
	router.NoRoute(func(c *gin.Context) {
		c.HTML(http.StatusNotFound, "generic.tpl", environment(c, gin.H{
			"Title"			: http.StatusNotFound,
			"description"	: http.StatusText(http.StatusNotFound),
			"Text"			: "Command " + c.Request.URL.Path + " not found.",
		}))
	})
	router.NoMethod(func(c *gin.Context) {
		c.HTML(http.StatusMethodNotAllowed, "generic.tpl", environment(c, gin.H{
			"Title"			: http.StatusMethodNotAllowed,
			"description"	: http.StatusText(http.StatusMethodNotAllowed),
			"Text"			: "Method " + c.Request.Method + " not allowed.",
		}))
	})

	// this might require another layer to check for https
	log.Fatal(router.Run(host + serverPort))

}