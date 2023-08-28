// Web user interface
//
// © 2023 by Gwyneth Llewelyn. All rights reserved.
// Licensed under a MIT License (see https://gwyneth-llewelyn.mit-license.org/).
package main

import (
	"fmt"
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
	var lastCoverPath string	// 'cache' of the civer art for this directory (= album),

	err = godirwalk.Walk(mediaDirectory,
		&godirwalk.Options{
			FollowSymbolicLinks: true,
			Callback: func(osPathname string, de *godirwalk.Dirent) error {
				// skip directories/symlinks to directories (if not in recursive mode)
				isDir, dirErr := de.IsDirOrSymlinkToDir();
				if isDir {
					if dirErr == nil {
						// return godirwalk.SkipThis
						logme.Debugf("entering %q...\n", de.Name())
						return nil
					}
					logme.Errorf("error when trying to access directory/symlink %q: %s",
						osPathname, dirErr)
						return nil
				}
				// check if this IS a valid audio file or not.
				// First, take a look at the extension. We need to make sure we actually get anything,
				// since an empty extension "" will match *any* file, which is NOT what we want here!
				fileExtension := strings.ToLower(filepath.Ext(de.Name()))

				// TODO(gwyneth): beyond checking the file extension, we should check for its MIME type

				if fileExtension != "" && !strings.Contains(validExtensions, fileExtension) {
					// skip this file if not a valid audio file
					logme.Debugf("Skipping %q (extension found: %q)...\n", de.Name(), fileExtension)
					return godirwalk.SkipThis
				}
				// ok, get the fileinfo for this entry:
				fiThis, err := os.Stat(osPathname)
				if err != nil {
					logme.Errorf("stat() failed on file %s: %s\n", osPathname, err)
					return err
				}

				// Check for album cover. To save resources, we sort of cache it.
				if lastCoverPath == "" {
					// does a file named "Folder.jpg" exist in the same folder? If so, use it!
					// Note: "Folder.jpg" seems to be some sort of convention; we might get anything which is an image instead...(gwyneth 20230827)
					potentialCoverPath := filepath.Join(filepath.Dir(osPathname), "Folder.jpg")

					if _, err := os.Stat(potentialCoverPath); err == nil {
						lastCoverPath = potentialCoverPath
					}
					logme.Debugf("Potential cover found: %q; last cover path is set to %q\n", potentialCoverPath, lastCoverPath)
				}

				// add another file to the list...
				// note: we will make all checkboxes true for now, to simplify testing; later,
				// they will be correctly set.
				temp := NewPlayListItem(*de, osPathname, lastCoverPath, fiThis.ModTime(), fiThis.Size(), true)
				playlist = append(playlist, *temp)
				// all clear, let's move on!
				return nil
			},	// ends Callback
			ErrorCallback: func(osPathname string, err error) godirwalk.ErrorAction {
				logme.Errorf("on file %s: %s\n", osPathname, err)
				return godirwalk.SkipNode
			},
			// Called at the end of every directory, after all the children have been invoked.
			PostChildrenCallback: func(osPathName string, de *godirwalk.Dirent) error {
				logme.Debugf("at directory: %s; emptying album cover path for this directory\n", osPathName)
				lastCoverPath = ""
				return nil
			},
			// Unsorted: false, // (optional) set true for faster yet non-deterministic enumeration (see godoc)
	})	// end options for dirwalk
	if err != nil {
		logme.Errorf("sorry, walking through %q got error: %s\n", mediaDirectory, err)
	}
	// no need to tranverse everything if we're not in debug mode!
	if (debug) {
		logme.Debugln("Walkthrough finished; let's see what we've got:")
		// index.
		var i = 0
		if len(playlist) != 0 {
			for _, dirEntry := range playlist {
				logme.Debugf("%d: %+v\n", i, dirEntry)
				i++
			}
		}
		logme.Debugf("%d entries found; Go reports %d elements \n", i, len(playlist))
//		logme.Debugf("Currently, error is %v and responseContent is %q\n", err, responseContent)
	}
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
		"Title"			 : skipescape("<i class=\"bi bi-music-note-beamed\" aria-hidden=\"true\"></i><i class=\"bi bi-music-note-beamed\" aria-hidden=\"true\"></i>&nbsp;Stream from media directory"),
		"description"	 : "Streaming from " + mediaDirectory,
		"Text"			 : fmt.Sprintf("Ready to start streaming from %q with %d entries...", mediaDirectory, len(playlist)),
		"hasDirList"	 : true,
		"mediaDirectory" : mediaDirectory,
		"playlist"		 : playlist,
	}))
}
