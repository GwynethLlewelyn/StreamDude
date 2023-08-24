// Web user interface
//
// © 2023 by Gwyneth Llewelyn. All rights reserved.
// Licensed under a MIT License (see https://gwyneth-llewelyn.mit-license.org/).
package main

import (
	"fmt"
	"html/template"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"github.com/karrick/godirwalk"
)

/*
 *  Some handlers called directly here
 */

// Homepage is the front-end's first page. It might get some authentication at sme point.
func homepage(c *gin.Context) {
	responseContent := getContentType(c)

	// Default message for those who do NOT use application/html!
	homepageMessage := "It works. You should see it in HTML instead, it's so much nicer!"
	logme.Debugf("homepage: response Content-Type: %q; Request method: %q\n", responseContent, c.Request.Method)

//	switch getContentType(c) {
	switch responseContent {
		case binding.MIMEJSON:
			c.JSON(http.StatusOK, gin.H{"status": "ok", "message": homepageMessage})
		case binding.MIMEHTML, binding.MIMEPOSTForm, binding.MIMEMultipartPOSTForm:
			c.HTML(http.StatusOK, "home.tpl", environment(c, gin.H{
				"Title"			: "Welcome!",
				"description"	: "StreamDude demo homepage",
				"Text"			: "This is StreamDude — nothing will work on the menus, except Ping.",
			}))
		case binding.MIMEXML, "application/soap+xml", binding.MIMEXML2:	// we'll probably ignore soap/xml
			c.XML(http.StatusOK, gin.H{"status": "ok", "message": homepageMessage})
		case binding.MIMEPlain:
			fallthrough
		default:
			c.String(http.StatusOK, homepageMessage)
	}
}

// uiPing is the all-purpose ping testing function. Works with HTML too.
func uiPing(c *gin.Context) {
	responseContent := getContentType(c)

	// this will work even behind Cloudflare (gwyneth 20230804)
	payload := "pong back to " + c.ClientIP()
	logme.Debugf("Ping request (%s) from %q received; replying with Content-Type: %q\n", c.Request.Method, payload, responseContent)

	switch responseContent {
		case binding.MIMEJSON:
			c.JSON(http.StatusOK, gin.H{"status": "ok", "message": payload})
		case binding.MIMEHTML, binding.MIMEPOSTForm, binding.MIMEMultipartPOSTForm:
			// if we're behind Cloudflare, we can get a cute emoji flag
			// telling us which country this ping came from! (gwyneth 20230804)
			cfIPCountry := c.GetHeader("CF-IPCountry")
			if cfIPCountry != "" {			// this will usually be set by Cloudflare, too
				payload += " " + getFlag(cfIPCountry)
			}
			c.HTML(http.StatusOK, "generic.tpl", environment(c, gin.H{
				"Title"			: "Ping results",
				"description"	: http.StatusText(http.StatusOK) + " " + payload,
				"Text"			: payload,
			}))
		case binding.MIMEXML, "application/soap+xml", binding.MIMEXML2:
			c.XML(http.StatusOK, gin.H{"status": "ok", "message": payload})
		case binding.MIMEPlain:
			fallthrough
		default:
			c.String(http.StatusOK, payload)
	}
}

// Displays credits for the software. Only configured for HTML outpit.
func uiCredits(c *gin.Context) {
	c.HTML(http.StatusOK, "generic.tpl", environment(c, gin.H{
		"Title"			: "Credits",
		"description"	: "Credits",
		"Text"			: "One day, we will credit here everybody."		}))
}

// Displays a page with the contents of the media directory. In the future, checkboxes & changing dir will work, too.
func uiStream(c *gin.Context) {
	var err error	// for scope issues on calls with multiple return params
	responseContent := getContentType(c)

	logme.Infoln("streaming from directory:", mediaDirectory)

	playlist = nil	// clean the last playlist and start from scratch.

	err = godirwalk.Walk(mediaDirectory,
		&godirwalk.Options{
			Callback: func(osPathname string, de *godirwalk.Dirent) error {
				// skip directories/symlinks to directories (if not in recursive mode)
				isDir, dirErr := de.IsDirOrSymlinkToDir();
				if isDir {
					if dirErr == nil {
						// return godirwalk.SkipThis
						return nil
					}
					logme.Errorf("error when trying to access directory/symlink %q: %s",
						osPathname, dirErr)
						return nil
				}
				// check if this IS a valid audio file or not.
				// a more stricter check should deal with
				if !strings.Contains(validExtensions, strings.ToLower(filepath.Ext(de.Name()))) {
					// skip this file if not
					return godirwalk.SkipThis
				}
				// ok, get the fileinfo for this entry
				st, err := os.Stat(osPathname)
				if err != nil {
					logme.Errorf("stat() failed on file %s: %s\n", osPathname, err)
					return err
				}
				// add another file to the list...
				playlist = append(playlist, st)
				// all clear, let's move on!
				return nil
			},	// ends Callback
			ErrorCallback: func(osPathname string, err error) godirwalk.ErrorAction {
				logme.Errorf("on file %s: %s\n", osPathname, err)
				return godirwalk.SkipNode
			},
			Unsorted: false, // (optional) set true for faster yet non-deterministic enumeration (see godoc)
	})	// end options for dirwalk
	if err != nil {
		logme.Errorf("sorry, walking through %q got error: %s\n", mediaDirectory, err)
	}
	// index.
	var i = 0
	if len(playlist) != 0 {
		for _, dirEntry := range playlist {
			logme.Debugf("%d: %+v\n", i, dirEntry)
		}
		i++
	}
	logme.Debugf("%d entries found; Go reports %d elements \n", i, len(playlist))

	if err != nil {
		switch responseContent {
			case binding.MIMEJSON:
				c.JSON(http.StatusBadRequest, gin.H{
					"status": "error",
					"message": "Error streaming from " + mediaDirectory + ": " + err.Error(),
				})
			case binding.MIMEHTML, binding.MIMEPOSTForm, binding.MIMEMultipartPOSTForm:
				c.HTML(http.StatusBadRequest, "generic.tpl", environment(c, gin.H{
					"Title"			: "Error during streaming",
					"description"	: "Failure to stream from " + mediaDirectory,
					"Text"			: "Error streaming from " + mediaDirectory + ": " + err.Error(),
				}))
			case binding.MIMEXML, "application/soap+xml", binding.MIMEXML2:
				c.XML(http.StatusBadRequest, gin.H{
						"status": "error",
						"message": "Error streaming from " + mediaDirectory + ": " + err.Error(),
					})
			case binding.MIMEPlain:
				fallthrough
			default:
				// minimalistic output, good for embedding
				c.String(http.StatusBadRequest, "successfully streamed from " + mediaDirectory)
		}
		return
	}

	c.HTML(http.StatusOK, "streamdir.tpl", environment(c, gin.H{
		"Title"			 : template.HTML("<i class=\"bi bi-music-note-beamed\" aria-hidden=\"true\"></i><i class=\"bi bi-music-note-beamed\" aria-hidden=\"true\"></i>&nbsp;Stream from media directory<br><code>" + mediaDirectory + "</code>"),
		"description"	 : "Streaming from " + mediaDirectory,
		"Text"			 : fmt.Sprintf("Streaming from %q with %d entries...", mediaDirectory, i),
		"hasDirList"	 : true,
		"mediaDirectory" : mediaDirectory,
		"playlist"		 : playlist,
	}))
}
