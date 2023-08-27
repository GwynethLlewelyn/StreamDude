// Invoking the VLC library to create a list of files to stream

package main

import (
	"fmt"
	"html/template"
//	"io/fs"
	"net/http"

	//	"os"
//	"path/filepath"
	//	"strings"

	vlc "github.com/adrg/libvlc-go/v3"
	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
// "github.com/karrick/godirwalk"
)

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
	logme.Infof("streaming from playlist: %v\n", playlist)
	logme.Debugf("Bound command: %+v\n", command)

	// Error related to streaming via VLC.
	var resultError error
	// We don't want to stream media if the playlist is empty.
	if len(playlist) == 0 {
		resultError = fmt.Errorf("empty playlist passed")
	} else {
		// run this in a separate goroutine, since it might take a LONG time to play!
		go func() {
			resultError = streamMedia(playlist)
		}()
	}

	checkErrReply(c, http.StatusNotFound, "could not stream from " + mediaDirectory, resultError)
	if resultError != nil {
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

	switch responseContent {
		case binding.MIMEJSON:
			c.JSON(http.StatusOK, gin.H{
				"status": "ok",
				"message": "successfully streaming from " + mediaDirectory,
			})
		case binding.MIMEHTML, binding.MIMEPOSTForm, binding.MIMEMultipartPOSTForm:
			c.HTML(http.StatusOK, "streamdir.tpl", environment(c, gin.H{
				"Title"			 : template.HTML("<i class=\"bi bi-music-note-beamed\" aria-hidden=\"true\"></i><i class=\"bi bi-music-note-beamed\" aria-hidden=\"true\"></i>&nbsp;Stream from media directory<br><code>" + mediaDirectory + "</code>"),
				"description"	 : "Successfully streaming from " + mediaDirectory,
				"Text"			 : "👍🆗✅ Successfully streaming (in the background) from " + mediaDirectory,
				"hasDirList"	 : true,
				"setBanner"		 : true,
				"mediaDirectory" : mediaDirectory,
				"playlist"		 : playlist,
			}))
		case binding.MIMEXML, "application/soap+xml", binding.MIMEXML2:
			c.XML(http.StatusOK, gin.H{
					"status": "ok",
					"message": "successfully streaming from " + mediaDirectory,
			})
		case binding.MIMEPlain:
			fallthrough
		default:
			// minimalistic output, good for embedding
			c.String(http.StatusOK, "successfully streaming from " + mediaDirectory)
	}
}

// Internal function to stream media via VLC, based on a playlist we got earlier.
func streamMedia(myPlayList []PlayListItem) error {
	// Make sure we got *something*!
	if len(myPlayList) == 0 {
		return fmt.Errorf("streamMedia() got an empty playlist for media dir: %q", mediaDirectory)
	}

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

	// Now loop through the whole playlist and count the entries.
	var total, checked int
	for _, entry := range myPlayList {
		// add only if this entry is in fact checked to play.
		if entry.Checked() {
			err = list.AddMediaFromPath(entry.Name())
			checked++
		}
//		err = list.AddMediaFromPath(filepath.Join(mediaDirectory, entry.Name()))
		if err != nil {
			return err
		}
		total++
	}
	logme.Infof("%d/%d checked entries from playlist added to streamer\n", total, checked)
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
		// TODO(gwyneth): *theoretically* we should return *something* to the browser. But how? (gwyneth 20230827)
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

	// should we have a timeout here? (gwyneth 20230827)
	<-quit

	return nil
}