// Assorted functions that one day really, really should make it into
// its own package.
//
//
// © 2023 by Gwyneth Llewelyn. All rights reserved.
// Licensed under a MIT License (see https://gwyneth-llewelyn.mit-license.org/).
//
package main

import (
	"crypto/md5"
	"crypto/rand"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"html/template"
	"math"
	"net/http"
	"os/user"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/dchest/uniuri"
//	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
//	"github.com/sirupsen/logrus"
)

// funcName is @Sonia's solution to get the name of the function that Go is currently running.
//  This will be extensively used to deal with figuring out where in the code the errors are!
//  Source: https://stackoverflow.com/a/10743805/1035977 (gwyneth 20170708)
func funcName() string {
	pc, _, _, _ := runtime.Caller(1)
	return runtime.FuncForPC(pc).Name()
}

// getContentType parses the Accept and Content-Type headers to figure out
// what content type to use.
func getContentType(c *gin.Context) string {
	contextCopy := c.Copy()
	// As per specs, check Accept header first:
	acceptHeader, _, _ := strings.Cut(contextCopy.GetHeader("Accept"), ",")
	contentType := contextCopy.ContentType()
	responseContent := contentType	// may be empty, but it's our fallack scenario

	logme.Debugf("For %q, full range of accepted headers is: %+v (total entries: %d)\n", contextCopy.FullPath(), contextCopy.GetHeader("Accept"), len(contextCopy.GetHeader("Accept")))

	if len(acceptHeader) != 0 {
		responseContent = acceptHeader
	}

	// No Accept, or Accept set to */*? (e.g. GET). Then check the Content-Type header,
	// using Gin's built-in function.
	// This _should_ give us a chance to return the same content type as what the
	// caller requested (e.g. browsers will send )
	if responseContent == "*/*" {
		responseContent = contentType
		// Note that Gin only returns the content type and not the charset etc.
		// This might require in the future to use strings.Cut(). (gwyneth 20230803)
	}
	logme.Debugf("first acceptHeader: %v responseContent set to: %v\n", acceptHeader, responseContent)

	return responseContent
}

// Note: all the error codes need to be rewritten... it's getting unmanageable this way. (gwyneth 20220328)
// Some ideas are presented here, by the maintainer of Gin: https://github.com/gin-gonic/gin/issues/274
// These suggest creating middleware to collect error messages and spew them out on demand. It looks pretty simple.
// Some rethinking might be needed to make sure we get the necessary runtime information.

// checkErrFatal logs a fatal error and does whatever log.Fatal() is supposed to do.
func checkErrFatal(err error) {
	// Note: logrus should already print file and line

	if err != nil {
		pc, file, line, ok := runtime.Caller(1)
		// logme.Fatal(filepath.Base(file), ":", line, ":", pc, ok, " - panic:", err)
		logme.Fatalf("%s:%d [PC: %v] (%t) - %s ▶ %s\n", filepath.Base(file), line, pc, ok, runtime.FuncForPC(pc).Name(), err)
	}
}

// checkErrPanic logs a fatal error and panics.
// Note: logrus should already print file and line
func checkErrPanic(err error) {
	if err != nil {
		pc, file, line, ok := runtime.Caller(1)
		// logme.Panic(filepath.Base(file), ":", line, ":", pc, ok, " - panic:", err)
		logme.Panicf("%s:%d [PC: %v] (%t) - %s ▶ %s\n", filepath.Base(file), line, pc, ok, runtime.FuncForPC(pc).Name(), err)
	}
}

// checkErr checks if there is an error, and if yes, it logs it out and continues.
//  this is for 'normal' situations when we want to get a log if something goes wrong but do not need to panic.
// Note: logrus should already print file and line
func checkErr(err error) {
	if err != nil {
		pc, file, line, ok := runtime.Caller(1)
//		logme.Error(filepath.Base(file), ":", line, ":", pc, ok, " - error:", err)
		logme.Errorf("%s:%d [PC: %v] (%t) - %s ▶ %s\n", filepath.Base(file), line, pc, ok, runtime.FuncForPC(pc).Name(), err)
	}
}

// Auxiliary functions for HTTP handling, adapted to Gin Gonic. (gwyneth 20220328)

// checkErrHTTP returns an error via HTTP and also logs the error.
//
// Deprecated: use checkErrReply instead, which formats everything according to the
// content type of the current request. (gwyneth 20230801)
func checkErrHTTP(c *gin.Context, httpStatus int, errorMessage string, err error) {
	if err != nil {	// save a function call, if there is no error
		checkErrReply(c, httpStatus, errorMessage, err)
	}
}

// checkErrPanicHTTP returns an error via HTTP and logs the error with a panic.
//
// Deprecated: use checkErrReply instead, which formats everything according to the
// content type of the current request, and add a panic to the end. (gwyneth 20230801)
func checkErrPanicHTTP(c *gin.Context, httpStatus int, errorMessage string, err error) {
	if err != nil {
		checkErrReply(c, httpStatus, errorMessage, err)
		panic(err)
	}
}

// Same as checkErrHTTP, but errors are returned.
//
// Deprecated: use checkErrReply instead, which formats everything according to the
// content type of the current request. (gwyneth 20230801)
func checkErrJSON(c *gin.Context, httpStatus int, errorMessage string, err error) {
	if err != nil {
		checkErrReply(c, httpStatus, errorMessage, err)
	}
}

// Universal check for errors and reply using the correct content type.
func checkErrReply(c *gin.Context, httpStatus int, errorMessage string, err error) {
	if err != nil {
		contentType := getContentType(c)	// maybe this wasn't a good idea after all...
//		contentType := c.Copy().ContentType()
		switch contentType {
			case binding.MIMEJSON:
				c.JSON(httpStatus,
					gin.H{"status":" error", "code": httpStatus, "message": fmt.Sprintf("%q: \"%v\"", errorMessage, err)})
			case binding.MIMEHTML, binding.MIMEPOSTForm, binding.MIMEMultipartPOSTForm:
				statusMessage := fmt.Sprintf("%d %s", httpStatus, http.StatusText(httpStatus))
				c.HTML(httpStatus, "generic.tpl", environment(c, gin.H{
					"Title"			: statusMessage,
					"description"	: statusMessage,
					"Text"			: errorMessage + ": " + err.Error(),
				}))
			case binding.MIMEXML, "application/soap+xml", binding.MIMEXML2:
				c.XML(httpStatus, gin.H{"status": "error", "code": httpStatus, "message": fmt.Sprintf("%q: \"%v\"", errorMessage, err)})
			case binding.MIMEPlain:
				fallthrough
			default:
				c.String(httpStatus, errorMessage + ": " + err.Error())
		}

		pc, file, line, ok := runtime.Caller(1)

		logme.Errorf("‹%s› (error %s) on %s:%d [PC: %v] (%t) - %s ▶ %s ▶ %s\n", contentType, http.StatusText(httpStatus), filepath.Base(file), line, pc, ok, runtime.FuncForPC(pc).Name(), errorMessage, err)
		if contentType == binding.MIMEJSON {
			c.AbortWithStatusJSON(httpStatus, gin.H{"status": "error", "code": httpStatus, "message": err.Error()})
		} else {
			if ginErr := c.AbortWithError(httpStatus, err); ginErr != nil {
				logme.Errorln("Additionally, AbortWithError failed with error:", ginErr)
			}
		}
	}
}

// expandPath expands the tilde as the user's home directory.
//  found at http://stackoverflow.com/a/43578461/1035977
func expandPath(path string) (string, error) {
	if len(path) == 0 || path[0] != '~' {
		return path, nil
	}

	// we have two cases. The simplest one is the current user:
	if path[1] == '/' {
		usr, err := user.Current()
		if err != nil {
			return "", err
		}
		return filepath.Join(usr.HomeDir, path[1:]), nil
	}
	// cut everything to the first slash
	username, restOfPath, _ := strings.Cut(path[1:], "/")
	usr, err := user.Lookup(username)
	if err != nil {
		return "", err
	}
	return filepath.Join(usr.HomeDir, restOfPath), nil
}

/**
*	Cryptographic helper functions.
**/

// getMD5Hash calculates the MD5 hash of any string. See aviv's solution on SO: https://stackoverflow.com/a/25286918/1035977.
func getMD5Hash(text string) string {
	hash := md5.Sum([]byte(text))
	return hex.EncodeToString(hash[:])
}

// randomBase64String is Steven Soroka's simple solution to generate a cryptographically secure random string with base64 encoding (see https://stackoverflow.com/a/55860599/1035977) (gwyneth 20200706).
func randomBase64String(l int) string {
	buff := make([]byte, int(math.Round(float64(l)/float64(1.33333333333))))
	rand.Read(buff)
	str := base64.RawURLEncoding.EncodeToString(buff)
	return str[:l] // strip 1 extra character we get from odd length results
}

// generatePIN with `nr` digits (0-9).
func generatePIN(nr int) string {
	const digits = "0123456789"

	return uniuri.NewLenChars(nr, []byte(digits))
}

// MergeMaps adds lots of map[string]interface{} together, returning the merged map[string]interface{}.
// It overwrites duplicate keys, maps to the right overwriting whatever keys are on the left.
// This allows for setting 'default' arguments later below, which can be overriden.
// See https://play.golang.org/p/8a9cXdSL_o3 as well as https://stackoverflow.com/a/39406305/1035977.
func MergeMaps(maps ...map[string]interface{}) map[string]interface{} {
	result := make(map[string]interface{})
	for _, m := range maps {
		for k, v := range m {
			result[k] = v
		}
	}
	return result
}

// Functions used inside templates. (some are probably not needed - gwyneth 20210821)

// bitTest applies a mask to a flag and returns true if the bit is set in the mask, false otherwise.
func bitTest(flag int, mask int) bool {
	return (flag & mask) != 0
}

// formatAsDate is a function for the templating system.
func formatAsDate(t time.Time) string {
	year, month, day := t.Date()
	return fmt.Sprintf("%d/%02d/%02d", year, month, day)
}

// formatAsYear is another function for the templating system.
func formatAsYear(t time.Time) string {
	year, _, _ := t.Date()
	return fmt.Sprintf("%d", year)
}

// Allows passing strings that should not be escaped, such as _comments_.
// (gwyneth 20220321)
// Note: This used to be built-in into `html/template` in the olden days.
// see	https://forum.golangbridge.org/t/unescaping-html-in-template-html/7085/7
// 		https://groups.google.com/g/golang-nuts/c/8y6by6SERyU/m/XQRnbw3aBhwJ
func skipescape(str string) template.HTML {
	return template.HTML(str)
}


/**
 * Auxiliary functions for the Gin Gonic environment.
 **/

// environment pushes a lot of stuff into the common environment.
func environment(c *gin.Context, env gin.H) gin.H {
	// session := sessions.Default(c)

	// host should not be empty (we've validated that in `main.go`) but also any other thing that
	// resolves to localhost should be set to, well, localhost. (gwyneth 20220321)
	tplHost := host
	if host == "" || host == "127.0.0.1" || host == "[::1]" {
		tplHost = "localhost"
	}
	// Update 2: actually, yes, it's ok to be empty. Stupid mistake upstream! (gwyneth 20220321)
	// Update: this message is wrong: 'it's ok if it's empty!'
	// No, it's _not_ ok :-P (gwyneth 20220321)
	// But we should simply go ahead with a reasonable default.
	//	serverPort := ServerPort
	// if serverPort == "" {
	// 	serverPort = ":3012"
	// }

	// Check if we have http or https; this is just to allow correctly parsed URLs on templates.
	// (gwyneth 20220320)
	// scheme := "http://";
	// if tlsCRT != "" && tlsKEY != "" {
	// 	scheme = "https://";
	// }

	// Check if we have a (configured) frontend, and, if so, adjust templates.
	if frontEnd == "nginx" {
		serverPort = externalPort	// should also be fine if it's empty!
		if externalHost == "" || externalHost == "127.0.0.1" || externalHost == "[::1]" || externalHost == "localhost" {
			tplHost = "localhost"
		} else {
			tplHost = externalHost
		}
	}

	// data is what gets sent to the underlying template engine as variables to fill in placeholders.
	var data = gin.H{
		/* common environment */
		"now"			: formatAsYear(time.Now()),
		"titleCommon"	: "StreamDude",
		"description"	: "",	// No description by default; this will be shown on the header title.
		"LSLSignaturePIN" :  lslSignaturePIN,
		"URLPathPrefix"	: urlPathPrefix,
		"Host"			: template.URL(tplHost),			// this gets adjusted depending on having a reverse proxy or not, (gwyneth 20220112)
		"ServerPort"	: template.URL(serverPort),		//  template.URL() allows hostnames/ports not to be parsed
//		"scheme"		: template.URL(scheme),			// either http:// or https://; see above. (gwyneth 20220320)

		/* session data — not implemented yet! (gwyneth 20220112) */
		// "Username"		: session.Get("Username"),
		// "UUID"			: session.Get("UUID"),
		// "Libravatar"	: session.Get("Libravatar"),
		// "Token"			: session.Get("Token"),
		// "Email"			: session.Get("Email"),
		// "RememberMe"	: session.Get("RememberMe"),

		"cacheBuster"	: generatePIN(64),				// just a random number for cache-busting. (gwyneth 20220408)
	}

	retMap := MergeMaps(data, env)

	// if *config["ginMode"] == "debug" && retMap["Username"] != nil && retMap["Username"] != "" {
	// 	logme.Debugf("environment(): All messages for user %q: %+v\n", retMap["Username"], retMap["Messages"])
	// }

	c.Header("X-Clacks-Overhead", "GNU Terry Pratchett")	// fans will know what this is for (gwyneth 20211115)

	return retMap
}

// simple string obfuscator.
func obfuscate(s string) string {
	slice := "********"
	if len(s) > 8 {
		slice += s[len(s)-8:]
	}
	return slice
}

// ISO-3166-1 two-letter country codes to-emoji.
func getFlag(countryCode string) string {
	// 0x1F1E6 - REGIONAL INDICATOR SYMBOL LETTER A
	// 0x1F1FF - REGIONAL INDICATOR SYMBOL LETTER Z
	// 0x0041 — LATIN CAPITAL LETTER A
	// 0x005A — LATIN CAPITAL LETTER Z

	return string(rune(countryCode[0])+127397) + string(rune(countryCode[1])+127397)
}
