// Go SDK for the KUSANAGI(tm) framework (http://kusanagi.io)
// Copyright (c) 2016-2020 KUSANAGI S.L. All rights reserved.
//
// Distributed under the MIT license.
//
// For the full copyright and license information, please view the LICENSE
// file that was distributed with this source code.

package kusanagi

import (
	"fmt"
)

// Empty frame defines an empty frame for a multipart response.
var emptyFrame = []byte("\x00")

// Request message contains the frames for a ZMQ multipart request.
type requestMsg [][]byte

// Validates that the multipart message has the right format.
func (m requestMsg) check() error {
	if length := len(m); length != 4 {
		return fmt.Errorf("Invalid multipart request length: %d", length)
	}
	return nil
}

// Get the ID for the current request.
func (m requestMsg) getRequestID() string {
	return string(m[0])
}

// Get the name of the component action to process.
func (m requestMsg) getAction() string {
	return string(m[1])
}

// Get the mapping schemas stream.
func (m requestMsg) getSchemas() []byte {
	return m[2]
}

// Get the command payload stream.
func (m requestMsg) getPayload() []byte {
	return m[3]
}

// Response message contains the frames for a ZMQ multipart response.
type responseMsg [][]byte
