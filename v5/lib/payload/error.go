// Go SDK for the KUSANAGI(tm) framework (http://kusanagi.io)
// Copyright (c) 2016-2023 KUSANAGI S.L. All rights reserved.
//
// Distributed under the MIT license.
//
// For the full copyright and license information, please view the LICENSE
// file that was distributed with this source code.

package payload

// DefaultErrorStatus contains the default status to use for errors.
const DefaultErrorStatus = "500 Internal Server Error"

// DefaultErrorMessage contains the default message to use for errors.
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

// GetCode returns the error code.
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
