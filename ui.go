// Web user interface
//
// © 2023 by Gwyneth Llewelyn. All rights reserved.
// Licensed under a MIT License (see https://gwyneth-llewelyn.mit-license.org/).
package main

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
)


/*
 *  Some handlers called directly here
 */

// Homepage is the front-end's first page. It might get some authentication at sme point.
func homepage(c *gin.Context) {
	contentType := c.Copy().ContentType()

	// Default message for those who do NOT use application/html!
	homepageMessage := "It works. You should see it in HTML instead, it's so much nicer!"
	logme.Debugf("homepage: Content-Type: %q; Request method: %q\n", contentType, c.Request.Method)

	// GET method, by specs, should not have a Content-Type header,
	// but we need one for correctly replying! (gwyneth 20230807)
	if c.Request.Method == http.MethodGet {
		contentType = binding.MIMEHTML
		logme.Debugf("homepage: Get method detected; Content-Type is now %q\n", contentType)
	}

//	switch getContentType(c) {
	switch contentType {
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
	contentType := c.Copy().ContentType()

	// this will work even behind Cloudflare (gwyneth 20230804)
	payload := "pong back to " + c.ClientIP()
	logme.Debugf("Ping request (%s) from %q received with Content-Type: %q\n", c.Request.Method, payload, contentType)

	// GET method, by specs, should not have a Content-Type header,
	// but we need one for correctly replying! (gwyneth 20230807)
	if c.Request.Method == http.MethodGet {
		contentType = binding.MIMEHTML
	}

//		switch getContentType(c) {
	switch contentType {
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