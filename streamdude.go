// StreamDude
//
// Main is here.
//
// Environment variables used when launching:
//
// `LAL_MASTER_KEY` - because it's too dangerous to keep it in code and/or files
// `STREAMER_URL` - another way to override the streamer URL; may be useful in scripts
//
// © 2023 by Gwyneth Llewelyn. All rights reserved.
// Licensed under a MIT License (see https://gwyneth-llewelyn.mit-license.org/).
//
package main

import (
	//	"log"
	"fmt"
	"html/template"
	"net/http"
	"os"
	"os/signal"
	"path"
	"path/filepath"
	"syscall"

	"github.com/coreos/go-systemd/v22/daemon"

	"github.com/gin-gonic/gin"
//	"github.com/gin-gonic/gin/binding"
	"github.com/go-playground/validator/v10"

	//	"github.com/google/martian/log"
	flag "github.com/karrick/golf" // flag replacement library
	"github.com/mattn/go-isatty"
	"github.com/sirupsen/logrus"
)

// Globals

var (
	help bool					// if set, show usage
	ffmpegPath string			// path to ffmpeg executable
	ginMode *string				// ginMode is `debug` for development, `release` for production.
	host string					// this host — where StreamDude is running.
	serverPort string			// port where StreamDude server is listening
	frontEnd string				// FrontEnd is usually nginx but will probably be ignored later on.
	externalPort string			// external port if using a reverse proxy
	externalHost string			// external hostname if using a reverse proxy
	templatePath string			// where templates are held
	pathToStaticFiles string	// where static assets are stored
	workingDirectory string		// workingDirectory is the result of os.Getwd() or "." if that fails.
	mediaDirectory string		// where media can be found on this server.
	urlPathPrefix string		// URL path prefix
	lslSignaturePIN string		// what we send from LSL
	debug bool					// set to debug level
	activeSystemd bool	= true	// if set, systemd is available (checked on start)

	// use a single instance of Validate, it caches struct info
	validate *validator.Validate

	// Global logger using logrus.
	logme = logrus.New()

	// Stuff for the lal streaming server
	streamerURL string		// RTSP streaming URL for lal
	lalMasterKey string		// too dangerous to show, put into LAL_MASTER_KEY environment
)

/*
 *  Ye Olde Maine Starts Here!
 *  Here Be Dragons
 *  And Unicorns
 */
func main() {
	// talk to systemd, inform that we're reloading
	b, err := daemon.SdNotify(false, daemon.SdNotifyReloading)
	// (false, nil) - notification not supported (i.e. NOTIFY_SOCKET is unset)
	// (false, err) - notification supported, but failure happened (e.g. error connecting to NOTIFY_SOCKET or while sending data)
	// (true, nil) - notification supported, data has been sentif
	switch {
		case !b && err == nil:
			// the logging system is not available, either, so we just print out
			fmt.Println("[WARN] systemd not available")
			activeSystemd = false
		case !b && err != nil:
			fmt.Println("[WARN] systemd answered with error:", err)
		case b && err == nil:
			fmt.Println("[INFO] systemd was succesfully notified that we're starting")
		default:
			fmt.Println("[WARN] unknown/confused systemd status, ignoring")
	}

	// Note: postponing some of the error logging until we know if the terminal is set and
	// handle colours etc. The above are "emergency" messages that need to be sent out before
	// the logging system is configured (namely, the first action should be) (gwyneth 20230807)

	// get the hostname, which is just used once, though (can be overriden with command-line arguments)
	hostname, err := os.Hostname()
	if err != nil {
		// who cares what the error was... in any case, we don't have the logging system yet:
		fmt.Printf("[WARN] system hostname not found (%s), using localhost instead\n", err)
		hostname = "localhost"
	}

	// Extract things from command line
	flag.BoolVarP(&help,			'h', "help",			false, 			"show command usage")
	flag.StringVarP(&ffmpegPath,	'm', "ffmpeg",			"/usr/local/bin/ffmpeg", "path to ffmpeg executable")
	flag.StringVarP(&host,			'j', "host",			"localhost", 	"server host where we're running")
	flag.StringVarP(&serverPort,	'p', "port", 			":3554", 		"port where StreamDude server is listening")
	flag.StringVarP(&frontEnd,		'f', "frontend", 		"nginx", 		"type of frontend/reverse proxy")
	flag.StringVarP(&externalPort,	'P', "externalport",	":80",			"external port if using a reverse proxy")
	flag.StringVarP(&externalHost,	'x', "externalhost",	hostname,		"external hostname if using a reverse proxy")
	flag.StringVarP(&templatePath,	't', "templatepath",	"./templates",	"where templates are held")
	flag.StringVarP(&pathToStaticFiles, 's', "staticpath",	"./assets",		"where static assets are stored")
	flag.StringVarP(&mediaDirectory, 'g', "mediapath",		"/tmp",			"absolute path where media files can be found")
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
	// router.SetTrustedProxies(nil)	// as per https://pkg.go.dev/github.com/gin-gonic/gin#readme-don-t-trust-all-proxies (gwyneth 20220111).
	router.SetTrustedProxies([]string{"127.0.0.1"})	// apparently we should at least trust "our" proxy
	router.TrustedPlatform = gin.PlatformCloudflare	// we're running behind Cloudflare CDN, this will retrieve the correct IP address. Hopefully.
	router.SetFuncMap(template.FuncMap{
		"bitTest": bitTest,
		"formatAsDate": formatAsDate,
	})

	// Configure logrus.
	//	log.SetFlags(log.LstdFlags | log.Lshortfile)

	logme.Formatter = new(logrus.TextFormatter)
	logme.Formatter.(*logrus.TextFormatter).ForceColors = activeSystemd	// if systemd is active, force colours on log
	logme.Formatter.(*logrus.TextFormatter).DisableColors = false		// keep colors
	logme.Formatter.(*logrus.TextFormatter).DisableTimestamp = false	// keep timestamp

	// set debug level, depending on the argument value
	if (debug) {
		logme.SetLevel(logrus.DebugLevel)
	}

	// respect CLICOLOR_FORCE and NO_COLOR in Gin (logrus is already compliant)
	// Figure out if we're running in a terminal, and, if so, apply all the relevant commands
	// See https://bixense.com/clicolors/ and https://no-color.org/ (gwyneth 20230731)
	// Note: when we're sending logs via journald, we force coloured output, because
	// journald supports it even if it's not a TTY.

	var isTerm = true	// are we logging to a tty?

	logme.Debugf("terminal type: %q activeSystemd: %t NO_COLOR: %q CLICOLOR_FORCE: %q\n",
		os.Getenv("TERM"), activeSystemd, os.Getenv("NO_COLOR"), os.Getenv("CLICOLOR_FORCE"))

	// for the weird type casting, see https://github.com/mattn/go-isatty/issues/80#issuecomment-1470096598 (gwyneth 20230801)
	if (os.Getenv("TERM") == "dumb" || os.Getenv("TERM") == "") && !activeSystemd ||
		(!isatty.IsTerminal(logme.Out.(*os.File).Fd()) && !isatty.IsCygwinTerminal(logme.Out.(*os.File).Fd())) {
			isTerm = false
	}
	if _, ok := os.LookupEnv("NO_COLOR"); ok {
		gin.DisableConsoleColor()
	} else if _, ok := os.LookupEnv("CLICOLOR_FORCE"); (ok && isTerm) || activeSystemd {
		gin.ForceConsoleColor()
	}
	logme.Debugf("debugging to terminal? - %t\n", isTerm)
	// at this stage we have the logging output well configured
	logme.Debugf("Logging debug level set to %q\n", logme.GetLevel().String())


	// Override lal master key from environment.
	if temp := os.Getenv("LAL_MASTER_KEY"); temp != "" {
		lalMasterKey = temp
	}
	if lalMasterKey == "" {
		logme.Warningln("lal master key not found or empty; streaming will probably not work.")
	} else {
		logme.Debugf("lal key (obfuscated): %q\n", obfuscate(lalMasterKey))
	}

	// Override streamer, if env exists.
	if temp := os.Getenv("STREAMER_URL"); temp != "" {
		streamerURL = temp
	}
	// Validate that the streamer has a valid URL (either from command-line or env var).
	if err := validate.Var(streamerURL, "required,url"); err != nil {
		logme.Fatalf("invalid streamer URL: %q, aborting\n", streamerURL)
	}
	logme.Infof("remote streamer URL set to: %q\n", streamerURL)

	if err := validate.Var(externalHost, "hostname_rfc1123,omitempty"); err != nil {
		logme.Errorf("invalid external host name: %q, reverting to empty string\n", externalHost)
		externalHost = ""
	}
	logme.Infof("external hostname set to: %q (empty is ok)\n", externalHost)

	// Validate path to media files. /tmp is perfectly accetable and valid.
	if err := validate.Var(mediaDirectory, "filepath"); err != nil {
		if fsInfo, err := os.Stat(mediaDirectory); err != nil {
			if !fsInfo.IsDir() {
				logme.Warnf("%q exists, but is not a valid directory; setting to /tmp\n", mediaDirectory)
				mediaDirectory = "/tmp"
			} else {
				logme.Infof("valid media directory found at %q (default should be /tmp which is ok)\n", mediaDirectory)

			}
		} else {
			logme.Warnf("cannot stat %q, error was: %v\n", mediaDirectory, err)
		}
	} else {
		logme.Warnf("invalid directory path %q, error was: %v\n", mediaDirectory, err)
	}

 	// Setup templating system.

	// var err error	// needed for scope issues
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

	// Generic funcionality

	// Ping handler (who knows, it might be useful in some contexts... such as Let's Encrypt certificates
	router.Any(path.Join(urlPathPrefix, "/ping"),			uiPing)

	// Main website, as far as we can call it a "website".
	router.GET(path.Join(urlPathPrefix, "/home"), 			homepage)
	router.GET(urlPathPrefix + string(os.PathSeparator),	homepage)

	// Shows the credits page.
	router.GET(path.Join(urlPathPrefix, "/credits"),		uiCredits)

	// Lower-leval API for calling things (mostly non-tty low-level calls)
	apiRoutes := router.Group(path.Join(urlPathPrefix, "/api"))
	{		// base page for complex scripts.
		apiRoutes.POST("/play",	apiStreamFile)
		apiRoutes.POST("/auth",	apiSimpleAuthGenKey)
		apiRoutes.POST("/delete", apiDeleteToken)
		apiRoutes.POST("/stream", apiStreamPath)
	}

	// Specific routes just for the user interface
	uiRoutes := router.Group(path.Join(urlPathPrefix, "/ui"))
	{
		uiRoutes.GET("/auth", func(c *gin.Context) {
			// not much to pass really
			c.HTML(http.StatusOK, "form-auth.tpl", environment(c, gin.H{
			}))
		})
		uiRoutes.GET("/play", func(c *gin.Context) {
			// not much to pass really
			c.HTML(http.StatusOK, "form-play.tpl", environment(c, gin.H{
			}))
		})
		uiRoutes.GET("/stream", uiStream)
	}

	// Catch all other routes and send back an error
	router.NoRoute(func(c *gin.Context) {
		errorMessage := "Command " + c.Request.URL.Path + " not found."
		checkErrReply(c, http.StatusNotFound, errorMessage, fmt.Errorf("(routing error)"))
	})

	router.NoMethod(func(c *gin.Context) {
		errorMessage := "Method " + c.Request.Method + " not allowed"
		checkErrReply(c, http.StatusMethodNotAllowed, errorMessage, fmt.Errorf("(unsupported method)"))
	})

	/*
	 *  Deal with Unix system signals (at least those we can catch)
	 */

	// prepares a special (buffered) channel to look for termination signals.
	sigs := make(chan os.Signal, 1)
	// signal.Notify(sigs)	// Note: this should catch all catchable signals!
	signal.Notify(sigs, syscall.SIGHUP, syscall.SIGUSR1, syscall.SIGUSR2, syscall.SIGCONT)

	// goroutine which listens to signals
	// Will handle re-configurations in the future.
	// For now, it exists mostly to signal systemd
	go func() {
		for {
			sig := <-sigs
			switch sig {
				case syscall.SIGUSR1:
					daemon.SdNotify(false, daemon.SdNotifyReloading + "\nSTATUS=reloading request received, but config reload is not supported yet")
					logme.Infoln("SIGUSR1 received, ignoring; we might reload config one day")
					daemon.SdNotify(false, daemon.SdNotifyReady)
				case syscall.SIGUSR2:
					logme.Infoln("SIGUSR2 received, ignoring")
				case syscall.SIGHUP:
					// Note: we *might* interpret this to suspend the processing and/or reload config (gwyneth 202230804)
					logme.Infoln("SIGHUP received (possibly from systemd): hanging up!")
					// if we were called by systemd, then notify it that we're done.
					// if not, just exit normally.
					daemon.SdNotify(true, daemon.SdNotifyStopping)
					os.Exit(129)
				case syscall.SIGCONT:
					logme.Infoln("SIGCONT received, ignoring")
				default:
					// should never happen...?
					logme.Warning("Unknown UNIX signal", sig, "caught, ignoring")
			}
		}
	}()

	// attempt to talk to systemd to notify we're now ready
	b, err = daemon.SdNotify(false, daemon.SdNotifyReady)
	switch {
		case !b && err == nil:
			// the logging system is not available, either, so we just print out
			logme.Warningln("systemd not available")
			activeSystemd = false
		case !b && err != nil:
			logme.Warningln("systemd answered with error:", err)
		case b && err == nil:
			logme.Infoln("systemd was succesfully notified that we're ready")
		default:
			logme.Warningln("unknown/confused systemd status, ignoring")
	}

	/*
	 *  Launch the server (finally) and log an error if it crashes.
	 */

	// this might require another layer to check for https
	errGin := router.Run(host + serverPort)

	// Notify systemd that we're peacefully stopping
	b, err = daemon.SdNotify(true, daemon.SdNotifyStopping  + "\nEXIT_STATUS=126")
	switch {
		case !b && err == nil:
			// the logging system is not available, either, so we just print out
			logme.Warningln("systemd not available")
			activeSystemd = false
		case !b && err != nil:
			logme.Warningln("systemd answered with error:", err)
		case b && err == nil:
			logme.Infoln("systemd was succesfully notified that we're stopping")
		default:
			logme.Warningln("unknown/confused systemd status, ignoring")
	}
	if errGin != nil {
		logme.Errorln("Gin aborted with", errGin)
	} else {
		logme.Errorln("Unexpected error, Gin terminated abruptly without error code")
	}
	os.Exit(126)
}