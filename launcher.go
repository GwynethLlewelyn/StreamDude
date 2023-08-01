// Deals with the API to launch ffmpeg
// For now, it works with just a single file
package main
import (
	//	"log"
	"fmt"
	"net/http"
	"os/exec"
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

	cmd := exec.Command(ffmpegPath, "-i", filename, "")
	stdoutStderr, err := cmd.CombinedOutput()
	if err != nil {
		logme.Printf("error while running %q: %q \n", ffmpegPath, err)
		return err
	}
	logme.Printf("✅ %s\n", stdoutStderr)
	return nil
}

// checks if we have received a valid JSON token
func payloadValidation(c *gin.Context, command *Command) {
	checkErrJSON(c, http.StatusBadRequest, "invalid request, no JSON found",
		c.ShouldBind(&command))

	logme.Debugf("Command to parse: %#v (should be JSON-ish)\n", command)

	// do some sanitation
	// returns nil or ValidationErrors ( []FieldError )
	if err := validate.Struct(command); err != nil {
		// this check is only needed when your code could produce
		// an invalid value for validation such as interface with nil
		// value most including myself do not usually have code like this.
		checkErrJSON(c, http.StatusBadRequest, "invalid request; could not validate JSON",
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
	}

	checkErrJSON(c, http.StatusNotFound, fmt.Sprintf("filename %q for streaming not found", command.Filename),
		streamFile(command.Filename))
}

// Handles /auth, gets the object PIN and returns a token.
// TODO(gwyneth): It's all fake for now.
func apiSimpleAuthGenKey(c *gin.Context) {
	var command Command

	payloadValidation(c, &command)

	pin, err := strconv.Atoi(command.ObjectPIN)
	logme.Debugf("Got PIN: %#v\n", pin)
	checkErrJSON(c, http.StatusBadRequest, "invalid request: invalid or empty PIN", err)
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
