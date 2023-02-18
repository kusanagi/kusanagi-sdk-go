// Go SDK for the KUSANAGI(tm) framework (http://kusanagi.io)
// Copyright (c) 2016-2023 KUSANAGI S.L. All rights reserved.
//
// Distributed under the MIT license.
//
// For the full copyright and license information, please view the LICENSE
// file that was distributed with this source code.

package kusanagi

// Error represents an error for a service call.
type Error struct {
	address string
	service string
	version string
	message string
	code    int
	status  string
}

// GetAddress returns the gateway address for the service.
func (e Error) GetAddress() string {
	return e.address
}

// GetName returns the service name.
func (e Error) GetName() string {
	return e.service
}

// GetVersion returns the service version.
func (e Error) GetVersion() string {
	return e.version
}

// GetMessage returns the error message.
func (e Error) GetMessage() string {
	return e.message
}

// GetCode returns the error code.
func (e Error) GetCode() int {
	return e.code
}

// GetStatus returns the status message.
func (e Error) GetStatus() string {
	return e.status
}
