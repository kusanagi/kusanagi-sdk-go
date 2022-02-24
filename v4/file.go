// Go SDK for the KUSANAGI(tm) framework (http://kusanagi.io)
// Copyright (c) 2016-2022 KUSANAGI S.L. All rights reserved.
//
// Distributed under the MIT license.
//
// For the full copyright and license information, please view the LICENSE
// file that was distributed with this source code.

package kusanagi

import (
	"errors"
	"fmt"
	"io/ioutil"
	"mime"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/kusanagi/kusanagi-sdk-go/v4/lib/payload"
)

func checkLocalFileExist(path string) error {
	// Remove the schema from the path
	if path[:7] == "file://" {
		path = path[7:]
	}

	if _, err := os.Stat(path); os.IsNotExist(err) {
		return fmt.Errorf(`file doesn't exist: "%s"`, path)
	}
	return nil
}

func extractLocalFileMimeType(path string) string {
	mimeType := mime.TypeByExtension(filepath.Ext(path))
	if mimeType != "" {
		return mimeType
	}
	// Use a default mime type when it can't be guessed
	return "text/plain"
}

func extractLocalFileSize(path string) uint {
	if info, err := os.Stat(path[7:]); err == nil {
		return uint(info.Size())
	}
	return 0
}

// NewFile creates a new file object.
//
// When the path is local it can start with "file://" or be a path to a local file,
// otherwise it means is a remote file and it must start with "http://".
//
// name: Name of the file parameter.
// path: Optional path to the file.
// mimeType: Optional MIME type of the file contents.
// filename: Optional file name and extension.
// size: Optional file size in bytes.
// token: Optional file server security token to access the file.
func NewFile(name, path, mimeType, filename string, size uint, token string) (*File, error) {
	length := len(path)
	if length > 7 && path[:7] == "http://" {
		if strings.TrimSpace(mimeType) == "" {
			return nil, errors.New("file missing MIME type")
		} else if strings.TrimSpace(filename) == "" {
			return nil, errors.New("file missing file name")
		} else if strings.TrimSpace(token) == "" {
			return nil, errors.New("file missing token")
		}
	} else if length > 0 {
		// Token must be used for remote files only
		if strings.TrimSpace(token) == "" {
			return nil, errors.New("unexpected file token")
		}

		// Check that the local file exists
		if err := checkLocalFileExist(path); err != nil {
			return nil, err
		}

		// When mime type is not given get it from the file path
		if strings.TrimSpace(mimeType) == "" {
			mimeType = extractLocalFileMimeType(path)
		}

		// When file name is not given get it from the last path element
		if strings.TrimSpace(filename) == "" {
			filename = filepath.Base(path)
		}

		// When file size is not given get it from the file but only when it is a local one
		if size == 0 {
			size = extractLocalFileSize(path)
		}

		if path[:7] != "file://" {
			path = fmt.Sprintf("file://%s", path)
		}
	}

	f := File{
		name:     strings.TrimSpace(name),
		path:     path,
		mime:     mimeType,
		filename: filename,
		size:     size,
		token:    token,
	}
	return &f, nil
}

// File parameter.
//
// Actions receive files thought calls to a service component.
// Files can also be returned from the service actions.
type File struct {
	name     string
	path     string
	mime     string
	filename string
	size     uint
	token    string
}

// GetName returns the name of the file parameter.
func (f File) GetName() string {
	return f.name
}

// GetPath returns the file path.
func (f File) GetPath() string {
	return f.path
}

// GetMime returns the MIME type of the file contents.
func (f File) GetMime() string {
	return f.mime
}

// GetFilename returns the file name.
func (f File) GetFilename() string {
	return f.filename
}

// GetSize returns the file size in bytes.
func (f File) GetSize() uint {
	return f.size
}

// GetToken returns the file server security token to access the file.
func (f File) GetToken() string {
	return f.token
}

// Exists checks if file exists.
func (f File) Exists() bool {
	return f.path != "" && f.path[:7] != "file://"
}

// IsLocal checks if file is located in the local file system.
func (f File) IsLocal() bool {
	return f.path[:7] == "file://"
}

// Read file's contents.
func (f File) Read() (contents []byte, err error) {
	if f.IsLocal() {
		// Read local file contents
		if contents, err = ioutil.ReadFile(f.path[7:]); err != nil {
			return nil, fmt.Errorf(`failed to read local file "%s": %v`, f.path, err)
		}
	} else {
		req, err := http.NewRequest("GET", f.path, nil)
		if err != nil {
			return nil, fmt.Errorf(`failed to read file "%s": %v`, f.path, err)
		}
		req.Header.Add("X-Token", f.token)

		// Make a request to read the remote file
		// TODO: We should add the timeout to read remote files in the specs
		client := &http.Client{}
		resp, err := client.Do(req)
		if err != nil {
			return nil, fmt.Errorf(`failed to read file "%s": %v`, f.path, err)
		}
		defer resp.Body.Close()
		if contents, err = ioutil.ReadAll(resp.Body); err != nil {
			return nil, fmt.Errorf(`failed to read file "%s": %v`, f.path, err)
		}
	}
	return contents, nil
}

// CopyWithName creates a new file parameter with a new name.
//
// name: Name of the new file parameter.
func (f File) CopyWithName(name string) *File {
	file, _ := NewFile(name, f.GetPath(), f.GetMime(), f.GetFilename(), f.GetSize(), f.GetToken())
	return file
}

// CopyWithMime creates a new file parameter with a new MIME type.
//
// mime: MIME type of the new file parameter.
func (f File) CopyWithMime(mimeType string) *File {
	file, _ := NewFile(f.GetName(), f.GetPath(), mimeType, f.GetFilename(), f.GetSize(), f.GetToken())
	return file
}

// Converts a file to a file payload.
func fileToPayload(f File) payload.File {
	return payload.File{
		Name:     f.GetName(),
		Path:     f.GetPath(),
		Mime:     f.GetMime(),
		Filename: f.GetFilename(),
		Size:     f.GetSize(),
		Token:    f.GetToken(),
	}
}

// Converts a file payload to a file.
func payloadToFile(f payload.File) File {
	return File{
		name:     f.Name,
		path:     f.Path,
		mime:     f.GetMime(),
		filename: f.Filename,
		size:     f.Size,
		token:    f.Token,
	}
}

// Converts a list files to a list of file payloads.
func filesToPayload(fs []File) (files []payload.File) {
	for _, f := range fs {
		files = append(files, fileToPayload(f))
	}
	return files
}
