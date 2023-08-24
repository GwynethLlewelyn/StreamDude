// Invoking the VLC library to create a list of files to stream

package main

import (
	"net/http"

	vlc "github.com/adrg/libvlc-go/v3"
	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
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
	logme.Infoln("streaming from directory:", mediaDirectory)
	logme.Debugf("Bound command: %+v\n", command)

	resultError := streamMedia(mediaDirectory)

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