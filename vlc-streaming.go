// Invoking the VLC library to create a list of files to stream

package main

import (
	"io/fs"
	"html/template"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	vlc "github.com/adrg/libvlc-go/v3"
	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"github.com/karrick/godirwalk"
)

const validExtensions = ".mp3.m4a.aac"	// valid audio extensions, add more if needed.

// A series of valid playlist entries (only audio files).
var playlist []fs.FileInfo

// Gin handler to stream from a directory.
// Everything is pretty much embedded in the code for now, except the path, which is on mediaDirectory.
func apiStreamPath(c *gin.Context) {
	var command Command
	var err error	// for scope issues on calls with multiple return params
	responseContent := getContentType(c)

	// add headers from Second Life®/OpenSimulator:
	command.AvatarKey 	= c.GetHeader("X-SecondLife-Avatar-Key")	// owner, not toucher
	command.AvatarName	= c.GetHeader("X-SecondLife-Avatar-Name")	// will be overwriten with toucher
	command.ObjectKey	= c.GetHeader("X-SecondLife-Object-Key")
	command.ObjectName	= c.GetHeader("X-SecondLife-Object-Name")

	// we should now be able to do some validation on those
	if err = c.ShouldBind(&command); err != nil {
		checkErrReply(c, http.StatusInternalServerError, "stream", err)
		return
	}
	logme.Infoln("streaming from directory:", mediaDirectory)
	logme.Debugf("Bound command: %+v\n", command)

	playlist = nil	// boom?

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

	logme.Debugf("Directory retrieved:")

/*	resultError := streamMedia(mediaDirectory)

	checkErrReply(c, http.StatusNotFound, "could not stream from " + mediaDirectory, resultError)
	if resultError != nil {
		*/
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
		"Text"			 : "Streaming from " + mediaDirectory,
		"hasDirList"	 : true,
		"mediaDirectory" : mediaDirectory,
		"playlist"		 : playlist,
	}))
}

func streamMedia(mediaLibrary string) error {
	// Initialize libVLC. Additional command line arguments can be passed in
	// to libVLC by specifying them in the Init function.
	if err := vlc.Init("--no-video", "--quiet"); err != nil {
		return err
	}
	defer vlc.Release()

	// Create a new list player.
	player, err := vlc.NewListPlayer()
	if err != nil {
		return err
	}
	defer func() {
		player.Stop()
		player.Release()
	}()

	// Create a new media list.
	list, err := vlc.NewMediaList()
	if err != nil {
		return err
	}
	defer list.Release()

	err = list.AddMediaFromPath(mediaLibrary)
	if err != nil {
		return err
	}
/*
	err = list.AddMediaFromURL("http://stream-uk1.radioparadise.com/mp3-32")
	if err != nil {
		return err
	}
 */
	// Set player media list.
	if err = player.SetMediaList(list); err != nil {
		return err
	}

/* 	// Media files can be added to the list after the list has been added
	// to the player. The player will play these files as well.
	err = list.AddMediaFromPath("localpath/test2.mp3")
	if err != nil {
		return err
	}
*/
	// Retrieve player event manager.
	manager, err := player.EventManager()
	if err != nil {
		return err
	}

	// Register the media end reached event with the event manager.
	quit := make(chan struct{})
	eventCallback := func(event vlc.Event, userData interface{}) {
		close(quit)
	}

	eventID, err := manager.Attach(vlc.MediaListPlayerPlayed, eventCallback, nil)
	if err != nil {
		return err
	}
	defer manager.Detach(eventID)

	// Start playing the media list.
	if err = player.Play(); err != nil {
		return err
	}

	<-quit

	return nil
}