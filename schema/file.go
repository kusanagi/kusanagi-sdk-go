// Go SDK for the KUSANAGI(tm) framework (http://kusanagi.io)
// Copyright (c) 2016-2019 KUSANAGI S.L. All rights reserved.
//
// Distributed under the MIT license.
//
// For the full copyright and license information, please view the LICENSE
// file that was distributed with this source code.

package schema

import "github.com/kusanagi/kusanagi-sdk-go/payload"

// MaxUint defines the maximum size for an unsigned integer
const MaxUint = ^uint(0)

// MinUint defines the minimum size for an unsigned integer
const MinUint = 0

// MaxInt defines the maximum size for an integer
const MaxInt = int(MaxUint >> 1)

// MinInt defines the minimum size for an integer
const MinInt = -MaxInt - 1

// NewFile creates a new file schema
func NewFile(name string, p *payload.Payload) *File {
	// When no payload is given use an empty payload
	if p == nil {
		p = payload.New()
	}

	return &File{name: name, payload: p}
}

// File represents a file parameter schema
type File struct {
	name    string
	payload *payload.Payload
}

// GetName gets the name of the file parameter
func (f File) GetName() string {
	return f.name
}

// GetMime gets the mime type of the file
func (f File) GetMime() string {
	return f.payload.GetDefault("mime", "text/plain").(string)
}

// IsRequired checks if the file parameter is required
func (f File) IsRequired() bool {
	return f.payload.GetBool("required")
}

// GetMax gets the maximum file size allowed
func (f File) GetMax() uint {
	var max uint

	v := f.payload.GetDefault("max", uint(MaxInt))
	switch v.(type) {
	case int:
		max = uint(v.(int))
	case uint64:
		max = uint(v.(uint64))
	case uint32:
		max = uint(v.(uint32))
	default:
		max = v.(uint)
	}
	return max
}

// IsExclusiveMax checks if maximum file size is inclusive
func (f File) IsExclusiveMax() bool {
	if !f.payload.Exists("max") {
		return false
	}
	return f.payload.GetBool("exclusive_max")
}

// GetMin gets the minimum file size allowed
func (f File) GetMin() uint {
	var min uint

	v := f.payload.GetDefault("min", 0)
	switch v.(type) {
	case int:
		min = uint(v.(int))
	case uint64:
		min = uint(v.(uint64))
	case uint32:
		min = uint(v.(uint32))
	default:
		min = v.(uint)
	}
	return min
}

// IsExclusiveMin checks if minimum file size is inclusive
func (f File) IsExclusiveMin() bool {
	if !f.payload.Exists("min") {
		return false
	}
	return f.payload.GetBool("exclusive_min")
}

// GetHTTPSchema gets HTTP file parameter schema
func (f File) GetHTTPSchema() *HTTPFile {
	p := payload.New()

	// Get HTTP schema data if it exists
	if v := f.payload.GetMap("http"); v != nil {
		p.Data = v
	}
	return NewHTTPFile(f.GetName(), p)
}

// NewHTTPFile creates a new HTTP file schema
func NewHTTPFile(name string, p *payload.Payload) *HTTPFile {
	// When no payload is given use an empty payload
	if p == nil {
		p = payload.New()
	}

	return &HTTPFile{name: name, payload: p}
}

// HTTPFile represents the HTTP semantics of a file parameter
type HTTPFile struct {
	name    string
	payload *payload.Payload
}

// IsAccessible checks if the Gateway has access to the file parameter
func (hf HTTPFile) IsAccessible() bool {
	return hf.payload.GetDefault("gateway", true).(bool)
}

// GetParam gets name for the HTTP param
func (hf HTTPFile) GetParam() string {
	return hf.payload.GetDefault("param", hf.name).(string)
}
