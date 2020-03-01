// Go SDK for the KUSANAGI(tm) framework (http://kusanagi.io)
// Copyright (c) 2016-2020 KUSANAGI S.L. All rights reserved.
//
// Distributed under the MIT license.
//
// For the full copyright and license information, please view the LICENSE
// file that was distributed with this source code.

package sdk

import (
	"errors"
	"fmt"
	"io/ioutil"
	"mime"
	"os"
	"path/filepath"
	"strings"

	"github.com/kusanagi/kusanagi-sdk-go/payload"
)

// ErrInvalidFileName is used to indicate an invalid file name.
var ErrInvalidFileName = errors.New("invalid file name")

// ErrTokenRequired is used to indicate that the remote file server token is missing for a file.
var ErrTokenRequired = errors.New("token is required for remote file paths")

// NewFile creates a new file object.
func NewFile(name, path, mimeType, filename string, size uint64, token string) (*File, error) {
	// File field name is required
	name = strings.TrimSpace(name)
	if name == "" {
		return nil, ErrInvalidFileName
	}

	// Get the protocol to get the file
	proto := ""
	path = strings.TrimSpace(path)
	if len(path) > 7 {
		proto = path[:7]
		if proto != "file://" && proto != "http://" {
			path = fmt.Sprintf("file://%s", path)
			proto = "file://"
		}
	}

	// Token is required for HTTP files
	if proto == "http://" && token == "" {
		return nil, ErrTokenRequired
	}

	// When mime type is not given get it from the file path
	if mimeType == "" {
		if mimeType = mime.TypeByExtension(filepath.Ext(path)); mimeType == "" {
			// Set a default mime type when it can't be guessed
			mimeType = "text/plain"
		}
	}

	// When file name is not given get it from the last path element
	if filename == "" {
		filename = filepath.Base(path)
	}

	// When file size is not given get it from the file but only when it is a local one
	if size == 0 && proto == "file://" {
		info, err := os.Stat(path[7:])
		if err != nil {
			return nil, fmt.Errorf("failed to get file size for \"%s\": %v", path, err)
		}
		size = uint64(info.Size())
	}

	f := File{
		name:     name,
		path:     path,
		mime:     mimeType,
		filename: filename,
		size:     size,
		token:    token,
	}
	return &f, nil
}

// Create a file that is empty.
func newEmptyFile(name string) *File {
	f, _ := NewFile(name, "", "", "", 0, "")
	return f
}

// File represents a files server file
type File struct {
	name     string
	path     string
	mime     string
	filename string
	size     uint64
	token    string
}

func (f File) protocol() string {
	// Return file's protocol prefix
	return f.path[:7]
}

// GetName gets file's parameter name.
func (f File) GetName() string {
	return f.name
}

// GetPath gets file's path.
func (f File) GetPath() string {
	return f.path
}

// GetMime gets file's mime type.
func (f File) GetMime() string {
	return f.mime
}

// GetFilename gets file's name.
func (f File) GetFilename() string {
	return f.filename
}

// GetSize gets file's size.
func (f File) GetSize() uint64 {
	return f.size
}

// GetToken gets the files server token where the file is hosted.
func (f File) GetToken() string {
	return f.token
}

// Exists checks if file exists.
// A request us made to check existence when file is located in a remote HTTP server.
func (f File) Exists() (bool, error) {
	if f.IsLocal() {
		// Get info for local file
		info, err := os.Stat(f.path[7:])
		if err != nil {
			if os.IsNotExist(err) {
				// The file doesn't exist
				return false, nil
			}
			// Otherwise an error happened trying to read the file
			return false, fmt.Errorf("failed to get file info for \"%s\": %v", f.path, err)
		} else if info.IsDir() {
			return false, fmt.Errorf("invalid file, it is a directory: \"%s\"", f.path)
		}
	} else {
		// TODO: Implement file exists for HTTP protocol
		panic("not implemented")
	}
	// By default file exists
	return true, nil
}

// IsLocal checks if file is a local one.
func (f File) IsLocal() bool {
	return f.protocol() == "file://"
}

// Read file's contents.
func (f File) Read() (contents []byte, err error) {
	if f.IsLocal() {
		// Check that the file exists locally
		if exists, err := f.Exists(); !exists {
			if err != nil {
				return nil, err
			}
			return nil, fmt.Errorf("file doesn't exist: \"%s\"", f.path)
		}

		// Read local file contents
		contents, err = ioutil.ReadFile(f.path[7:])
		if err != nil {
			return nil, fmt.Errorf("failed to read local file \"%s\": %v", f.path, err)
		}
	} else {
		// TODO: Implement file read for HTTP protocol
		panic("not implemented")
	}
	return contents, nil
}

// CopyWithName creates a new file object with a different name.
func (f File) CopyWithName(name string) *File {
	file, _ := NewFile(name, f.GetPath(), f.GetMime(), f.GetFilename(), f.GetSize(), f.GetToken())
	return file
}

// CopyWithMime creates a new file object with a different mime type.
func (f File) CopyWithMime(mimeType string) *File {
	file, _ := NewFile(f.GetName(), f.GetPath(), mimeType, f.GetFilename(), f.GetSize(), f.GetToken())
	return file
}

// FileToPayload converts a file object to a file payload.
func FileToPayload(f *File) *payload.File {
	fp := payload.NewFile(f.GetName(), f.GetFilename(), f.GetPath())
	fp.SetMime(f.GetMime())
	fp.SetSize(f.GetSize())
	fp.SetToken(f.GetToken())
	return fp
}

// PayloadToFile converts a file payload to a file object.
func PayloadToFile(fp *payload.File) *File {
	f, _ := NewFile(fp.GetName(), fp.GetPath(), fp.GetMime(), fp.GetFilename(), fp.GetSize(), fp.GetToken())
	return f
}
