// Web user interface
//
// © 2023 by Gwyneth Llewelyn. All rights reserved.
// Licensed under a MIT License (see https://gwyneth-llewelyn.mit-license.org/).
package main

import (
	"fmt"
//	"io/fs"
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
	// For type PlayListItem, see playlist.go

	var err error	// for scope issues on calls with multiple return params
	responseContent := getContentType(c)

	logme.Infoln("streaming from directory:", mediaDirectory)

	playlist = nil	// clear the last playlist and start from scratch.
	var lastCoverPath string	// 'cache' of the cover art for this directory (= album),

	err = godirwalk.Walk(mediaDirectory,
		&godirwalk.Options{
			FollowSymbolicLinks: true,
			Callback: func(osPathname string, de *godirwalk.Dirent) error {
				// go one level deeper
				isDir, dirErr := de.IsDirOrSymlinkToDir();
				if isDir {
					if dirErr == nil {
						logme.Debugf("entering %q (base name: %q)...\n", osPathname, de.Name())
						// Check for a `Folder.jpg` file
						coverFile := filepath.Join(osPathname, "Folder.jpg")
						if potentialCover, err := os.Stat(coverFile); err == nil {
							logme.Debugf("stat() found an album cover file for %q named %q\n", osPathname, potentialCover.Name())
							lastCoverPath = filepath.Join(urlPathPrefix, coverFile)
						} else {
							logme.Debugf("`Folder.jpg` not found on album at %q; no cover set\n", osPathname)
						}
						return nil
					}
					logme.Errorf("error while trying to access directory/symlink %q: %s",
						osPathname, dirErr)
						return nil 	// or should we return godirwalk.SkipThis?
				}

				// FileInfo for the file currently being considered.
				// We need it here because of scope issues. (gwyneth 20230828)
				// var fiThis fs.FileInfo	// not needed any longer, actually, due to code refactoring.

				// Check if this is a valid audio file, a possible album cover, or none of those.
				// First, take a look at the extension. We need to make sure we actually get anything,
				// since an empty extension "" will match *any* file, which is NOT what we want here!
				fileExtension := strings.ToLower(filepath.Ext(osPathname))

				// TODO(gwyneth): beyond checking the file extension, we should check for its MIME type!
				if fileExtension != "" {
					if strings.Contains(validExtensions, fileExtension) {
						// Ok, this is a valid audio file, so get the fileinfo for this entry:
						fiThis, err := os.Stat(osPathname)
						if err != nil {
							logme.Errorf("stat() failed on file %s: %s\n", osPathname, err)
							return err
						}
						// Add another file to the list...
						// Note: we will make all checkboxes true for now, to simplify testing; later,
						// they will be correctly set.
						temp := NewPlayListItem(*de, filepath.Join(urlPathPrefix, osPathname), lastCoverPath, fiThis.ModTime(), fiThis.Size(), true)
						playlist = append(playlist, *temp)
						// All clear, let's move on!
						return nil
					} else if strings.Contains(validCoverExtensions, fileExtension) {
						// this is a potential album cover image; we save it and skip to the next.
						// Note: multiple files are possible, the (alphabetically) last one will be used.
						logme.Debugf("potential cover found: %q (unused)\n", lastCoverPath)
						// Ok, no more processing on this file, we can skip the entry.
						return godirwalk.SkipThis
					}
					// skip this file if not a valid audio file, nor a cover image:
					logme.Debugf("skipping %q (extension found: %q)...\n", de.Name(), fileExtension)
					return godirwalk.SkipThis
				} else {
					// Rare case where a file hasn't got an extension, so we cannot figure out what its type is.
					logme.Debugf("empty file extension for %q, skipping...\n", de.Name())
					return godirwalk.SkipThis
				}
			},	// ends Callback
			ErrorCallback: func(osPathname string, err error) godirwalk.ErrorAction {
				logme.Errorf("on file %s: %s\n", osPathname, err)
				return godirwalk.SkipNode
			},
			// Called at the end of every directory, after all the children have been invoked.
			PostChildrenCallback: func(osPathName string, de *godirwalk.Dirent) error {
				logme.Debugf("finished with directory %q: emptying album cover path (%s)\n", osPathName, lastCoverPath)
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
				logme.Debugf("%d: %#v\n", i, dirEntry)
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
