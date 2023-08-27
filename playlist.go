// All playlist-related functions.
// This approaches quasi-modular form.
// A playlist is essentially a list of gowalk.Dirents, but because those lack a few extra fields,
// these are added here.
//
// (c) 2023 by Gwyneth Llewelyn and released under a [MIT License](https://gwyneth-llewelyn.mit-license.org/).
//
package main

import (
	"github.com/karrick/godirwalk"
)

const validExtensions = ".mp3.m4a.aac"	// valid audio extensions, add more if needed.

// Represents a playlist item, including image, checkbox status etc.
type PlayListItem struct {
	de godirwalk.Dirent	// directory entry data retrieved from godirwalk.

	fullPath string	`validate:"filepath"`			// full path for the directory where this file is.
	cover string	`validate:"filepath,omitempty"`	// path to image for this file.
	checked bool		// eventually this will add the file to the playlist.
}

// Given a godirwalk.Dirent, tries to assembly a valid playlist item.
func NewPlayListItem(dirEntry godirwalk.Dirent, path string, coverPath string, checkForStreaming bool) (*PlayListItem) {
	return &PlayListItem{
		de:			dirEntry,
		fullPath:	path,
		cover:		coverPath,
		checked:	checkForStreaming,
	}
}

// Returns the playlist item name, including the full path.
func (p PlayListItem) Name() string {
	return p.fullPath
}

// Item checked for streaming.
func (p PlayListItem) Checked() bool {
	return p.checked
}

// interface to conform to the String() convention.
func (p PlayListItem) String() string {
	return p.fullPath
}

// Album cover image file.
func (p PlayListItem) Cover() string {
	return p.fullPath
}

// A series of valid playlist entries (only audio files).
// This is allegedly constructed once, by the ui/stream handler, and reused here.
// TODO(gwyneth): Or maybe put inside a Gin context?
var playlist []PlayListItem
