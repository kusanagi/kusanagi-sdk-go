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

// Multipart message parts
const (
	msgIdentityPart = iota
	msgForwardIdentityPart
	msgEmptyPart
	msgRequestIDPart
	msgActionPart
	msgSchemasPart
	msgPayloadPart
)

// Response message contains the frames for a ZMQ multipart response.
type responseMsg [][]byte

// Request message contains the frames for a ZMQ multipart request.
type requestMsg [][]byte

// Validates that the multipart message has the right format.
func (m requestMsg) check() error {
	if length := len(m); length != 7 {
		return fmt.Errorf("Invalid multipart request length: %d", length)
	}
	return nil
}

// Get the ID for the current request.
func (m requestMsg) getRequestID() string {
	return string(m[msgRequestIDPart])
}

// Get the name of the component action to process.
func (m requestMsg) getAction() string {
	return string(m[msgActionPart])
}

// Get the mapping schemas stream.
func (m requestMsg) getSchemas() []byte {
	return m[msgSchemasPart]
}

// Get the command payload stream.
func (m requestMsg) getPayload() []byte {
	return m[msgPayloadPart]
}

// Create the multipart response for the request message.
func (m requestMsg) makeResponseMessage(parts ...[]byte) responseMsg {
	// Add ZMQ message prefix to the response
	response := responseMsg{
		m[msgIdentityPart],
		m[msgForwardIdentityPart],
		m[msgEmptyPart],
		m[msgRequestIDPart],
	}
	return append(response, parts...)
}
