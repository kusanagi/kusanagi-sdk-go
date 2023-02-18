// Go SDK for the KUSANAGI(tm) framework (http://kusanagi.io)
// Copyright (c) 2016-2023 KUSANAGI S.L. All rights reserved.
//
// Distributed under the MIT license.
//
// For the full copyright and license information, please view the LICENSE
// file that was distributed with this source code.

package payload

// File represents a file parameter.
type File struct {
	Name     string `json:"n"`
	Path     string `json:"p"`
	Mime     string `json:"m"`
	Filename string `json:"f"`
	Size     uint   `json:"s"`
	Token    string `json:"t,omitempty"`
}

// GetMime returns the mime type of the file.
func (f File) GetMime() string {
	if f.Mime == "" {
		return "text/plain"
	}
	return f.Mime
}
