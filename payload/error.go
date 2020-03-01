// Go SDK for the KUSANAGI(tm) framework (http://kusanagi.io)
// Copyright (c) 2016-2020 KUSANAGI S.L. All rights reserved.
//
// Distributed under the MIT license.
//
// For the full copyright and license information, please view the LICENSE
// file that was distributed with this source code.

package payload

import "fmt"

// NewError creates a new error payload
func NewError() *Error {
	return &Error{Payload: NewNamespaced("error")}
}

// NewErrorFromObj creates a new error payload from an error object or string
func NewErrorFromObj(reason interface{}) *Error {
	p := NewError()
	p.SetMessage(fmt.Sprintf("%v", reason))
	return p
}

// NewErrorFromMap creates a new error payload from a map
func NewErrorFromMap(data map[string]interface{}) *Error {
	p := NewError()
	p.Data = data
	return p
}

// Error represents an error payload
type Error struct {
	*Payload
}

func (e Error) Error() string {
	return e.GetMessage()
}

// GetMessage gets the error message
func (e Error) GetMessage() string {
	return e.GetDefault("message", "Unknown error").(string)
}

// SetMessage sets the error message
func (e *Error) SetMessage(value string) error {
	return e.Set("message", value)
}

// GetCode gets the error code
func (e Error) GetCode() int {
	return e.GetDefault("code", 0).(int)
}

// SetCode sets the error code
func (e *Error) SetCode(value int) error {
	return e.Set("code", value)
}

// GetStatus gets the status message
func (e Error) GetStatus() string {
	return e.GetDefault("status", "500 Internal Server Error").(string)
}

// SetStatus sets the status message
func (e *Error) SetStatus(value string) error {
	return e.Set("status", value)
}
