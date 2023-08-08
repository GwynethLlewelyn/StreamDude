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

	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	// "google.golang.org/genproto/googleapis/devtools/resultstore/v2"
	// "github.com/go-playground/validator/v10"
	// "github.com/sirupsen/logrus"
)

// Command JSON type
type Command struct {
	AvatarID string		`validate:"omitempty,uuid" xml:"avatarID" json:"avatarID" form:"avatarID" binding:"-"`
	AvatarName string	`validate:"omitempty,alphanum" xml:"avatarName" json:"avatarName" form:"avatarName" binding:"-"`
	ObjectPIN string	`validate:"omitempty,number" xml:"objectPIN" json:"objectPIN" form:"objectPIN" binding:"-"`	// 4-digit PIN from in-world object
	Token string		`validate:"omitempty,base64" xml:"token" json:"token" form:"token" binding:"-"`				// made-up token for whatever reason
	SessionID string	`validate:"omitempty,hexadecimal" xml:"sessionID" json:"sessionID" form:"sessionID" binding:"-"`		// returned on valid transaction
	Filename string		`validate:"omitempty,filepath" xml:"filename" json:"filename" form:"filename" binding:"-"`
	MasterKey string	`validate:"omitempty,alphanum" xml:"masterKey" json:"masterKey" form:"masterKey" binding:"-"`	// LAL Master Key
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
			"-f", "rtsp", "-muxdelay", "0.1", "-rtsp_transport", "tcp", cmdURL)
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
func apiStreamFile(c *gin.Context) {
	var command Command
	var err error	// for scope issues on calls with multiple return params
	contentType := c.Copy().ContentType()

	if err = c.ShouldBind(&command); err != nil {
		checkErrReply(c, http.StatusInternalServerError, "could not get input data", err)
		return
	}
	logme.Debugf("Bound command: %+v\n", command)

	if command.Token == "" {
		checkErrReply(c, http.StatusUnauthorized, "no valid token sent", fmt.Errorf("no valid token sent"))
		return
	}

	if command.Filename == "" {
		checkErrReply(c, http.StatusBadRequest, "no valid token sent", fmt.Errorf("empty filename, cannot proceed"))
		return
	}
	// attempt to expand tilde (~) to user's home directory
	if command.Filename, err = expandPath(command.Filename); err != nil {
		checkErrReply(c, http.StatusBadRequest, "filename with ~ not properly expanded to existing file", err)
		return
	}
	// does the file exist?
	if _, err := os.Stat(command.Filename); err != nil {
		checkErrReply(c, http.StatusNotFound, fmt.Sprintf("filename %q for streaming not found", command.Filename),
		err)
	}
	// we should be good to go now!
	resultError := streamFile(command.Filename)

	checkErrReply(c, http.StatusNotFound, fmt.Sprintf("could not stream %q", command.Filename), resultError)
	if resultError != nil {
		switch contentType {
			case binding.MIMEJSON:
				c.JSON(http.StatusBadRequest, gin.H{
					"status": "error",
					"message": "Error streaming " + command.Filename + ": " + err.Error(),
				})
			case binding.MIMEHTML, binding.MIMEPOSTForm, binding.MIMEMultipartPOSTForm:
				c.HTML(http.StatusBadRequest, "generic.tpl", environment(c, gin.H{
					"Title"			: "Error during stream",
					"description"	: "The file failed to stream",
					"Text"			: "Error streaming " + command.Filename + ": " + err.Error(),
				}))
			case binding.MIMEXML, "application/soap+xml", binding.MIMEXML2:
				c.XML(http.StatusBadRequest, gin.H{
						"status": "error",
						"message": "Error streaming " + command.Filename + ": " + err.Error(),
					})
			case binding.MIMEPlain:
				fallthrough
			default:
				// minimalistic output, good for embedding
				c.String(http.StatusBadRequest, command.Filename + " successfully streamed")
		}
		return
	}

	switch contentType {
		case binding.MIMEJSON:
			c.JSON(http.StatusOK, gin.H{
				"status": "ok",
				"message": command.Filename + " successfully streamed",
			})
		case binding.MIMEHTML, binding.MIMEPOSTForm, binding.MIMEMultipartPOSTForm:
			c.HTML(http.StatusOK, "generic.tpl", environment(c, gin.H{
				"Title"			: "File successfully streamed!",
				"description"	: "The file has been successfully streamed",
				"Text"			: command.Filename + " was successfully streamed!",
			}))
		case binding.MIMEXML, "application/soap+xml", binding.MIMEXML2:
			c.XML(http.StatusOK, gin.H{
					"status": "ok",
					"message": command.Filename + " successfully streamed",
				})
		case binding.MIMEPlain:
			fallthrough
		default:
			// minimalistic output, good for embedding
			c.String(http.StatusOK, command.Filename + " successfully streamed")
	}
}

// Handles /auth, gets the object PIN and returns a token.
// TODO(gwyneth): It's all fake for now.
func apiSimpleAuthGenKey(c *gin.Context) {
/* 	type CommandSimple struct {
		ObjectPIN string	`validate:"number" xml:"objectPIN" json:"objectPIN" form:"objectPIN" binding:"required"`	// 4-digit PIN from in-world object
		MasterKey string	`validate:"alphanum" xml:"masterKey" json:"masterKey" form:"masterKey" binding:"required"`	// LAL Master Key
	}

	var command CommandSimple */
	var command Command

	contentType := c.Copy().ContentType()

	// weirdly, forms are an exception?

	// probe into request, for debugging purposes
	logme.Debugf("Request received with Content-Type: %q\n", contentType)

	/* if contentType == binding.MIMEPOSTForm || contentType == binding.MIMEMultipartPOSTForm {
		// command.ObjectPIN = c.Copy().Request.FormValue("objectPIN")
		// command.MasterKey = c.Copy().Request.FormValue("masterKey")
		if err := c.ShouldBindWith(&command, binding.Query); err == nil {
			logme.Debugf("HTML form received (%q). Assigning objectID to %q and masterKey to %q\n",
				contentType, command.ObjectPIN, command.MasterKey)
		} else {
			logme.Warningf("could not bind form using ShouldBindWith(&command, binding.Query); error was: %q\n;", err)
		}
	} else { */
		if err := c.ShouldBind(&command); err != nil {
			logme.Warningf("could not bind form using ShouldBind(&command); error was: %q\n;", err)

			checkErrReply(c, http.StatusInternalServerError, "could not get input data", err)
			return
		}
	//}

	logme.Debugf("Bound command: %+v\n", command)


	pin, err := strconv.Atoi(command.ObjectPIN)
	checkErrReply(c, http.StatusBadRequest, "invalid request: invalid or empty PIN", err)
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

	// For now, we just return the bare-bones token, after checking *how* to
	// return it, depending on the Content-Type of the request:
	// contentType := getContentType(c)
	switch contentType {
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
			// minimalistic output, good for embedding
			c.String(http.StatusOK, token)
	}
}
