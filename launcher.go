// Deals with the API to launch ffmpeg
// For now, it works with just a single file at a time,
// hopefully, playlists will come soon!
//
// © 2023 by Gwyneth Llewelyn. All rights reserved.
// Licensed under a MIT License (see https://gwyneth-llewelyn.mit-license.org/).
package main

import (
	//	"log"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
	// "strings"

	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	// "google.golang.org/genproto/googleapis/devtools/resultstore/v2"
	// "github.com/go-playground/validator/v10"
	// "github.com/sirupsen/logrus"
)

// Command JSON type
type Command struct {
	// UUID for the avatar making the request
	AvatarKey string	`validate:"omitempty,uuid" xml:"avatarKey" json:"avatarKey" form:"avatarKey" binding:"-"`
	// Avatar name making the streaming request
	AvatarName string	`validate:"omitempty,alphanum" xml:"avatarName" json:"avatarName" form:"avatarName" binding:"-"`
	// 4-digit PIN from in-world object
	ObjectPIN string	`validate:"omitempty,number" xml:"objectPIN" json:"objectPIN" form:"objectPIN" binding:"-"`
	// UUID of the Second Life®/Opensimulator object making the request (comes from the headers)
	ObjectKey string	`validate:"omitempty,uuid" xml:"objectKey" json:"objectKey" form:"objectKey" binding:"-"`
	// Name of the object making the request
	ObjectName string	`validate:"omitempty,alphanum" xml:"objectName" json:"objectName" form:"objectName" binding:"-"`
	// Made-up token for whatever reason.
	Token string		`validate:"omitempty,base64" xml:"token" json:"token" form:"token" binding:"-"`
	// ID returned on valid transaction
	SessionID string	`validate:"omitempty,hexadecimal" xml:"sessionID" json:"sessionID" form:"sessionID" binding:"-"`
	// Filename to stream (must be a locally-existing file).
	Filename string		`validate:"omitempty,filepath" xml:"filename" json:"filename" form:"filename" binding:"-"`
	// LAL Master Key
	MasterKey string	`validate:"omitempty,alphanum" xml:"masterKey" json:"masterKey" form:"masterKey" binding:"-"`
}

// Helper function to actually play a file via ffmpeg
func streamFile(filename string) error {
	logme.Debugf("Filename to stream: %q; Master key: %q\n", filename, obfuscate(lalMasterKey))

	// ffmpeg params
	/*
	-re -stream_loop -1 -i /var/www/clients/client6/web14/home/betafiles/data/beta-technologies/Universidade de Aveiro/LOCUS Project in Amiais/Panels SL/Painel_Preparativos/Preparativos.mp4 -acodec copy -vcodec copy -f rtsp -muxdelay 0.1 -rtsp_transport tcp rtsp://127.0.0.1:5544/Preparativos.mp4?lal_secret=0126471190816174f602a1e4b3cbd7b6
	*/

	// for lal server: calculate the simple hash allowing execution.
	// TODO(gwyneth): deal with the way it works for other streaming services,
	// where it is more customary to send login/password OOB. (gwyneth 20230803)
	basename := filepath.Base(filename)
	calcHash := getMD5Hash(lalMasterKey + basename)
	cmdURL, err := url.JoinPath(streamerURL, basename)
	if err != nil {
		logme.Errorf("❌ Could not create a proper URL from %q: %q\n", filename, err)
		return err
	}
	cmdURL += "?lal_secret=" + calcHash
	logme.Debugf("conjoined URL for streaming is: %q\n", cmdURL)

	// Since ffmpeg may be running for a while, let's start this in a goroutine
	// and wait there, while relinquishing resources and allowing other things to run.
	// (gwyneth 20230803)
	go func() {
		runtime.LockOSThread()	// lock to safely execute programs.
		defer runtime.UnlockOSThread()

		cmd := exec.Command(ffmpegPath, "-re", "-i", filename, "-acodec", "copy", "-vcodec", "copy",
			"-f", "rtsp", "-muxdelay", "0.1", "-tune", "zerolatency", "-rtsp_transport", "tcp", cmdURL)
		logme.Debugf("command to be executed: %s\n", cmd.String())
		// launch ffmpeg, but don't wait for it.
		err := cmd.Start()
		if err != nil {
			logme.Errorf("❌ could not start %s, error was: %s\n", ffmpegPath, err)
			runtime.Goexit()
		}
		logme.Infof("waiting for command to finish...")
		runtime.Gosched()	// we are waiting, so we yield CPU to other goroutines.
		err = cmd.Wait()
		if err != nil {
			logme.Errorf("❌ command finished with error: %v\n", err)
			runtime.Goexit()
		}
		logme.Infof("✅ %s %s terminated with success\n", ffmpegPath, filename)
	}()

	return nil
}


/*
 *  Router functions
 */

// Handles /play, body contains JSON-encoded filename etc.
// Plays a single file only.
func apiStreamFile(c *gin.Context) {
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
		checkErrReply(c, http.StatusInternalServerError, "play", err)
		return
	}

	logme.Debugf("Bound command: %+v\n", command)

	if command.Token == "" {
		checkErrReply(c, http.StatusUnauthorized, "play", fmt.Errorf("no valid token sent"))
		return
	}

	if command.Filename == "" {
		checkErrReply(c, http.StatusBadRequest, "play", fmt.Errorf("empty filename, cannot proceed"))
		return
	}
	// attempt to expand tilde (~) to user's home directory
	if command.Filename, err = expandPath(command.Filename); err != nil {
		checkErrReply(c, http.StatusBadRequest, "play: filename with ~ not properly expanded to existing file", err)
		return
	}
	// does the file exist?
	if _, err := os.Stat(command.Filename); err != nil {
		checkErrReply(c, http.StatusNotFound, fmt.Sprintf("play: filename %q for streaming not found", command.Filename),
		err)
	}
	// we should be good to go now!
	resultError := streamFile(command.Filename)

	checkErrReply(c, http.StatusNotFound, fmt.Sprintf("could not play %q", command.Filename), resultError)
	if resultError != nil {
		switch responseContent {
			case binding.MIMEJSON:
				c.JSON(http.StatusBadRequest, gin.H{
					"status": "error",
					"message": "Error playing " + command.Filename + ": " + err.Error(),
				})
			case binding.MIMEHTML, binding.MIMEPOSTForm, binding.MIMEMultipartPOSTForm:
				c.HTML(http.StatusBadRequest, "generic.tpl", environment(c, gin.H{
					"Title"			: "Error during play",
					"description"	: "The file failed to stream",
					"Text"			: "Error playing " + command.Filename + ": " + err.Error(),
				}))
			case binding.MIMEXML, "application/soap+xml", binding.MIMEXML2:
				c.XML(http.StatusBadRequest, gin.H{
						"status": "error",
						"message": "Error playing " + command.Filename + ": " + err.Error(),
					})
			case binding.MIMEPlain:
				fallthrough
			default:
				// minimalistic output, good for embedding
				c.String(http.StatusBadRequest, command.Filename + " successfully played")
		}
		return
	}

	switch responseContent {
		case binding.MIMEJSON:
			c.JSON(http.StatusOK, gin.H{
				"status": "ok",
				"message": command.Filename + " successfully played",
			})
		case binding.MIMEHTML, binding.MIMEPOSTForm, binding.MIMEMultipartPOSTForm:
			c.HTML(http.StatusOK, "generic.tpl", environment(c, gin.H{
				"Title"			: "File successfully played!",
				"description"	: "The file has been successfully played",
				"Text"			: command.Filename + " was successfully played!",
			}))
		case binding.MIMEXML, "application/soap+xml", binding.MIMEXML2:
			c.XML(http.StatusOK, gin.H{
					"status": "ok",
					"message": command.Filename + " successfully played",
				})
		case binding.MIMEPlain:
			fallthrough
		default:
			// minimalistic output, good for embedding
			c.String(http.StatusOK, command.Filename + " successfully played")
	}
}

// Handles /auth, gets the object PIN and returns a token.
// TODO(gwyneth): It's all fake for now.
func apiSimpleAuthGenKey(c *gin.Context) {
	var command Command
	responseContent := getContentType(c)

	// add headers from Second Life®/OpenSimulator:
	command.AvatarKey 	= c.GetHeader("X-SecondLife-Avatar-Key")	// owner, not toucher
	command.AvatarName	= c.GetHeader("X-SecondLife-Avatar-Name")	// will be overwritten with toucher
	command.ObjectKey	= c.GetHeader("X-SecondLife-Object-Key")
	command.ObjectName	= c.GetHeader("X-SecondLife-Object-Name")

	if err := c.ShouldBind(&command); err != nil {
		logme.Warningf("could not bind form using ShouldBind(&command); error was: %q\n;", err)

		checkErrReply(c, http.StatusInternalServerError, "auth: could not get input data", err)
		return
	}

	logme.Debugf("Bound command: %+v\n", command)

	pin, err := strconv.Atoi(command.ObjectPIN)
	checkErrReply(c, http.StatusBadRequest, "auth: invalid request: invalid or empty PIN", err)
	// TODO(gwyneth): obviously, check if this is a valid PIN...
	if err != nil {
		return
	}
	// if PIN was correct, save new master key (if it wasn't empty)
	if command.MasterKey != "" {
		lalMasterKey = command.MasterKey
	}

	logme.Debugf("Got PIN: %v\nGot LAL Master Key: %q\n", pin, obfuscate(command.MasterKey))

	// generate a random token, to be used for future authentication requests
	token := randomBase64String(32)
	// TODO: save the token on persistent storage somewhere, e.g. Redis or other KV store.
	logme.Debugln("Generated token:")

	// For now, we just return the bare-bones token, after checking *how* to
	// return it, depending on the Content-Type of the request:
	switch responseContent {
		case binding.MIMEJSON:
			c.JSON(http.StatusOK, gin.H{
				"status": "ok",
				"message": "PIN accepted, token follows",
				"token": token,
			})
		case binding.MIMEHTML, binding.MIMEPOSTForm, binding.MIMEMultipartPOSTForm:
			c.HTML(http.StatusOK, "generic.tpl", environment(c, gin.H{
				"Title"			: "PIN Accepted!",
				"description"	: "Returns a token",
				"Text"			: "Your token is: " + token,
			}))
		case binding.MIMEXML, "application/soap+xml", binding.MIMEXML2:
			c.XML(http.StatusOK, gin.H{
					"status": "ok",
					"message": "PIN accepted, token follows",
					"token": token,
				})
		case binding.MIMEPlain:
			fallthrough
		default:
			// minimalistic output, good for embedding in LSL
			c.String(http.StatusOK, token)
	}
}

// Handles /delete, body contains JSON-encoded token to be deleted.
func apiDeleteToken(c *gin.Context) {
	var command Command
	// var err error	// for scope issues on calls with multiple return params.
	responseContent := getContentType(c)

	// add headers from Second Life®/OpenSimulator:
	command.AvatarKey 	= c.GetHeader("X-SecondLife-Avatar-Key")	// owner, not touchee
	command.AvatarName	= c.GetHeader("X-SecondLife-Avatar-Name")	// will be overwriten with touchee
	command.ObjectKey	= c.GetHeader("X-SecondLife-Object-Key")
	command.ObjectName	= c.GetHeader("X-SecondLife-Object-Name")

	// we should now be able to do some validation on those
	if err := c.ShouldBind(&command); err != nil {
		logme.Warningf("delete: could not bind form using ShouldBind(&command); error was: %q\n;", err)

		checkErrReply(c, http.StatusInternalServerError, "delete: could not get input data", err)
		return
	}

	logme.Debugf("Bound command: %+v\n", command)

	if command.Token == "" {
		checkErrReply(c, http.StatusUnauthorized, "delete", fmt.Errorf("no valid token sent"))
		return
	}

	// TODO(gwyneth): no-op for now. In the future, the token shall be removed from the KV store.
	logme.Infoln("Token", command.Token, "deleted successfully.")

	switch responseContent {
		case binding.MIMEJSON:
			c.JSON(http.StatusOK, gin.H{
				"status": "ok",
				"message": "Token " + command.Token + " deleted",
			})
		case binding.MIMEHTML, binding.MIMEPOSTForm, binding.MIMEMultipartPOSTForm:
			c.HTML(http.StatusOK, "generic.tpl", environment(c, gin.H{
				"Title"			: "Token deleted!",
				"description"	: "Deletes a token",
				"Text"			: "Successfully deleted token: " + command.Token,
			}))
		case binding.MIMEXML, "application/soap+xml", binding.MIMEXML2:
			c.XML(http.StatusOK, gin.H{
					"status": "ok",
					"message": "Token " + command.Token + " deleted",
				})
		case binding.MIMEPlain:
			fallthrough
		default:
			c.String(http.StatusOK, "DELETED: " + command.Token)
	}
}