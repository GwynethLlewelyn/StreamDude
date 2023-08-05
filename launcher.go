// Deals with the API to launch ffmpeg
// For now, it works with just a single file at a time,
// hopefully, playlists will come soon!
//
// © 2023 by Gwyneth Llewelyn. All rights reserved.
// Licensed under a MIT License (see https://gwyneth-llewelyn.mit-license.org/).
//
package main

import (
	//	"log"
	"fmt"
	"path/filepath"
	"net/http"
	"net/url"
	"os/exec"
	"runtime"
	"strconv"

	"github.com/gin-gonic/gin"
//	"github.com/go-playground/validator/v10"
//	"github.com/sirupsen/logrus"
)

// Command JSON type
type Command struct {
	AvatarID string		`validate:"omitempty,uuid" xml:"avatarID" json:"avatarID"`
	AvatarName string	`validate:"omitempty,alphanum" xml:"avatarName" json:"avatarName"`
	ObjectPIN string	`validate:"omitempty,number" xml:"objectPIN" json:"objectPIN"`	// 4-digit PIN from in-world object
	Token string		`validate:"omitempty,base64" xml:"token" json:"token"`				// made-up token for whatever reason
	SessionID string	`validate:"omitempty,hexadecimal" xml:"sessionID" json:"sessionID"`		// returned on valid transaction
	Filename string		`validate:"omitempty,filepath" xml:"filename" json:"filename"`
}

// Helper function to actually play a file via ffmpeg
func streamFile(filename string) error {
	logme.Debugf("Filename to stream: %q\n", filename)

	// ffmpeg params
	/*
	-re -stream_loop -1 -i /var/www/clients/client6/web14/home/betafiles/data/beta-technologies/Universidade de Aveiro/LOCUS Project in Amiais/Panels SL/Painel_Preparativos/Preparativos.mp4 -acodec copy -vcodec copy -f rtsp -muxdelay 0.1 -rtsp_transport tcp rtsp://127.0.0.1:5544/Preparativos.mp4?lal_secret=0126471190816174f602a1e4b3cbd7b6
	*/

	if err := validate.Var(filename, "required,file"); err != nil {
		logme.Errorf("cannot find/open file at %q: %q\n", filename, err)
		return err
	}

	// for lal server: calculate the simple hash allowing execution.
	// TODO(gwyneth): deal with the way it works for other streaming services,
	// where it is more customary to send login/password OOB. (gwyneth 20230803)
	basename := filepath.Base(filename)
	calcHash := getMD5Hash(lalMasterKey + basename)
	cmdURL, err := url.JoinPath(streamerURL, basename)
	if err != nil {
		logme.Errorf("Could not create a proper URL from %q: %q\n", filename, err)
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
			logme.Errorf("could not start %s, error was: %s\n", ffmpegPath, err)
			runtime.Goexit()
		}
		logme.Infof("waiting for command to finish...")
		runtime.Gosched()	// we are waiting, so we yield CPU to other goroutines.
		err = cmd.Wait()
		if err != nil {
			logme.Errorf("command finished with error: %v\n", err)
			runtime.Goexit()
		}
		logme.Infof("✅ %s %s terminated with success\n", ffmpegPath, filename)
	}()

	return nil
}

// checks if we have received a valid JSON token
func payloadValidation(c *gin.Context, command *Command) {
	checkErrReply(c, http.StatusBadRequest, "invalid request, no valid body found",
		c.ShouldBind(command))

	logme.Debugf("Command to parse: %#v (should be JSON-ish)\n", command)

	// do some sanitation
	// returns nil or ValidationErrors ( []FieldError )
	if err := validate.Struct(command); err != nil {
		checkErrReply(c, http.StatusBadRequest, "invalid request; could not validate body",
			err)
	}
}

/*
 *  Router function
 */

// Handles /play, body contains JSON-encoded filename etc.
func apiStreamFile(c *gin.Context) {
	var command Command

	payloadValidation(c, &command)
	if command.Token == "" {
		c.JSON(http.StatusUnauthorized, gin.H{ "status": "error", "msg": "no valid token sent"})
		return
	}

	if command.Filename == "" {
		c.JSON(http.StatusBadRequest, gin.H{ "status": "error", "msg": "empty filename, cannot proceed"})
		return
	}

	checkErrReply(c, http.StatusNotFound, fmt.Sprintf("filename %q for streaming not found", command.Filename),
		streamFile(command.Filename))
}

// Handles /auth, gets the object PIN and returns a token.
// TODO(gwyneth): It's all fake for now.
func apiSimpleAuthGenKey(c *gin.Context) {
	var command Command

	payloadValidation(c, &command)

	pin, err := strconv.Atoi(command.ObjectPIN)
	logme.Debugf("Got PIN: %#v\n", pin)
	checkErrReply(c, http.StatusBadRequest, "invalid request: invalid or empty PIN", err)
	// TODO(gwyneth): obviously, check if this is a valid PIN...

	// generate a random token, to be used for future authentication requests
	token := randomBase64String(32)
	// TODO: save the token on persistent storage somewhere, e.g. Redis or other KV store.

	// For now, we just return the bare-bones token:
	c.JSON(http.StatusOK, gin.H{
		"status": "ok",
		"message": "PIN accepted, token follows",
		"token": token,
	})
}
