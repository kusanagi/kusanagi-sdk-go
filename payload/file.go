// Go SDK for the KUSANAGI(tm) framework (http://kusanagi.io)
// Copyright (c) 2016-2020 KUSANAGI S.L. All rights reserved.
//
// Distributed under the MIT license.
//
// For the full copyright and license information, please view the LICENSE
// file that was distributed with this source code.

package payload

func NewEmptyFile() *File {
	return &File{Payload: New()}
}

// NewFile creates a new file payload
func NewFile(name, filename, path string) *File {
	f := NewEmptyFile()
	f.SetName(name)
	f.SetFilename(filename)
	f.SetPath(path)
	return f
}

// NewFileFromMap creates a new file payload from a map
func NewFileFromMap(data map[string]interface{}) *File {
	f := NewEmptyFile()
	f.Data = data
	return f
}

type File struct {
	*Payload
}

func (f *File) GetName() string {
	return f.GetString("name")
}

func (f *File) SetName(name string) error {
	return f.Set("name", name)
}

func (f *File) GetFilename() string {
	return f.GetString("filename")
}

func (f *File) SetFilename(filename string) error {
	return f.Set("filename", filename)
}

func (f *File) GetPath() string {
	return f.GetString("path")
}

func (f *File) SetPath(path string) error {
	return f.Set("path", path)
}

func (f *File) GetMime() string {
	if v := f.GetString("mime"); v != "" {
		return v
	}
	return "text/plain"
}

func (f *File) SetMime(mime string) error {
	return f.Set("mime", mime)
}

func (f *File) GetSize() uint64 {
	return f.GetUint64("size")
}

func (f *File) SetSize(size uint64) error {
	return f.Set("size", size)
}

func (f *File) GetToken() string {
	return f.GetString("token")
}

func (f *File) SetToken(token string) error {
	return f.Set("token", token)
}
