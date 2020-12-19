// Go SDK for the KUSANAGI(tm) framework (http://kusanagi.io)
// Copyright (c) 2016-2020 KUSANAGI S.L. All rights reserved.
//
// Distributed under the MIT license.
//
// For the full copyright and license information, please view the LICENSE
// file that was distributed with this source code.

package payload

const DefaultErrorStatus = "500 Internal Server Error"
const DefaultErrorMessage = "Unknown Error"

// Error represents a reply that is returned when there is an error during command execution.
type Error struct {
	Message string `json:"m"`
	Code    int    `json:"c"`
	Status  string `json:"s"`
}

// GetMessage returns the error message.
func (e Error) GetMessage() string {
	if e.Message == "" {
		return DefaultErrorMessage
	}
	return e.Message
}

// GetMessage returns the error message.
func (e Error) GetCode() int {
	return e.Code
}

// GetStatus returns the status message of the error.
func (e Error) GetStatus() string {
	if e.Status == "" {
		return DefaultErrorStatus
	}
	return e.Status
}
