// All playlist-related functions.
// This approaches quasi-modular form.
// A playlist is essentially a list of gowalk.Dirents, but because those lack a few extra fields,
// these are added here.
//
// (c) 2023 by Gwyneth Llewelyn and released under a [MIT License](https://gwyneth-llewelyn.mit-license.org/).
package main

import (
//	"io/fs"
	"os"
	"time"

	"github.com/karrick/godirwalk"
)

const validExtensions		= ".mp3.m4a.aac"				// valid audio extensions, add more if needed.
const validCoverExtensions	= ".jpg.jpeg.png.gif.heic.webp"	// valid image extensions for album cover, add more if needed.

// Represents a playlist item, including image, checkbox status etc.
type PlayListItem struct {
	de godirwalk.Dirent	// directory entry data retrieved from godirwalk.

	fullPath string		`validate:"filepath"`			// full path for the directory where this file is.
	cover string		`validate:"filepath,omitempty"`	// path to album cover for this file.
	modTime time.Time	`validate:"datetime"`			// last modified date (at least on Unix-like systems).
	size int64			// filesize in bytes, as reported by the system.
	checked bool		// file checkbox enabled; eventually this will add the file to the playlist.
}

// Given a godirwalk.Dirent, tries to assembly a valid playlist item.
func NewPlayListItem(dirEntry godirwalk.Dirent, path string, coverPath string, lastModTime time.Time, fileSize int64, checkedForStreaming bool) (*PlayListItem) {
	return &PlayListItem{
		de:			dirEntry,
		fullPath:	path,
		cover:		coverPath,
		modTime:	lastModTime,
		size:		fileSize,
		checked:	checkedForStreaming,
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
	return p.cover
}

// File size in bytes, as returned by the underlying OS.
func (p PlayListItem) Size() int64 {
	return p.size
}

// Last modified date (at least on Unix-like systems)
func (p PlayListItem) ModTime() time.Time {
	return p.modTime
}

// IsDir returns true if and only if the Dirent represents a file system
// directory.  Note that on some operating systems, more than one file mode bit
// may be set for a node.  For instance, on Windows, a symbolic link that points
// to a directory will have both the directory and the symbolic link bits set.
func (p PlayListItem) IsDir() bool {
	return p.de.ModeType() & os.ModeDir != 0
}

// IsRegular returns true if and only if the Dirent represents a regular file.
// That is, it ensures that no mode type bits are set.
func (p PlayListItem) IsRegular() bool {
	return p.de.ModeType() & os.ModeType == 0
}

// IsSymlink returns true if and only if the Dirent represents a file system
// symbolic link.  Note that on some operating systems, more than one file mode
// bit may be set for a node.  For instance, on Windows, a symbolic link that
// points to a directory will have both the directory and the symbolic link bits
// set.
func (p PlayListItem) IsSymlink() bool {
	return p.de.ModeType() & os.ModeSymlink != 0
}

// IsDevice returns true if and only if the Dirent represents a device file.
func (p PlayListItem) IsDevice() bool {
	return p.de.ModeType() & os.ModeDevice != 0
}

// ModeType returns the mode bits that specify the file system node type.  We
// could make our own enum-like data type for encoding the file type, but Go's
// runtime already gives us architecture independent file modes, as discussed in
// `os/types.go`:
//
//    Go's runtime FileMode type has same definition on all systems, so that
//    information about files can be moved from one system to another portably.
func (p PlayListItem) ModeType() os.FileMode {
	return p.de.ModeType()
}

// reset releases memory held by most of the struct (except the Dirent).
func (p *PlayListItem) reset() {
	// p.de.reset()	// no way to free memory from the Dirent!
	p.fullPath = ""
	p.cover = ""
	p.modTime = time.Now()
	p.size = 0
	p.checked = false
}


// A series of valid playlist entries (only audio files).
// This is allegedly constructed once, by the ui/stream handler, and reused here.
// TODO(gwyneth): Or maybe put inside a Gin context?
var playlist []PlayListItem
