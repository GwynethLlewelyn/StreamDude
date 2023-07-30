// StreamDude

package main

import (
	"log"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/gin-gonic/gin"
	flag "github.com/karrick/golf" // flag replacement library
)

// Globals

var (
	help bool			// if set, show usage
	ffmpegPath string	// path to ffmpeg executable
	ginMode *string		// ginMode is `debug` for development, `release` for production.
	serverPort string	// port where StreamDude server is listening
	externalPort string	// external port if using a reverse proxy
	templatePath string	// where templates are held
	pathToStaticFiles string // where static assets are stored
	workingDirectory string // workingDirectory is the result of os.Getwd() or "." if that fails.
)

func streamFile(filename string) error {
	cmd := exec.Command(ffmpegPath, "-i", filename, "")
	stdoutStderr, err := cmd.CombinedOutput()
	if err != nil {
		log.Printf("error while running %q: %q \n", ffmpegPath, err)
		return err
	}
	log.Printf("✅ %s\n", stdoutStderr)
	return nil
}

func main() {
	// Extract things from command line
	flag.BoolVarP(&help,			'h', "help",			false, "show command usage")
	flag.StringVarP(&ffmpegPath,	'm', "ffmpeg",			"/usr/local/bin/ffmpeg", "path to ffmpeg executable")
	flag.StringVarP(&serverPort,	'p', "port", 			":3554", 		"port where StreamDude server is listening")
	flag.StringVarP(&externalPort,	'P', "external",		":80",			"external port if using a reverse proxy")
	flag.StringVarP(&templatePath,	't', "templatepath",	"./templates",	"where templates are held")
	flag.StringVarP(&pathToStaticFiles, 's', "staticpath",	"./assets",		"where static assets are stored")

	flag.Parse()

	if help {
		flag.Usage()
		os.Exit(0)
	}

	// setup default logger
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	/**
	 * Starting backend web server using Gin Gonic.
	 */
	router := gin.Default()
	router.Delims("{{", "}}") // stick to default delims for Go templates.
	router.SetTrustedProxies(nil)	// as per https://pkg.go.dev/github.com/gin-gonic/gin#readme-don-t-trust-all-proxies (gwyneth 20220111).
	router.TrustedPlatform = gin.PlatformCloudflare	// we're running behind Cloudflare CDN

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
}