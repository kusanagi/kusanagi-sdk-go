// Go SDK for the KUSANAGI(tm) framework (http://kusanagi.io)
// Copyright (c) 2016-2019 KUSANAGI S.L. All rights reserved.
//
// Distributed under the MIT license.
//
// For the full copyright and license information, please view the LICENSE
// file that was distributed with this source code.

package payload

type Transformable interface {
	GetData() Data
}

// ToError converts a standard payload to an error payload
func ToError(t Transformable) *Error {
	np := NewError()
	np.Data = t.GetData()
	if np.IsError() {
		np.UndoEntity()
	}
	return np
}

// ToCall converts a standard payload to a call payload
func ToCall(t Transformable) *Call {
	np := NewEmptyCall()
	np.Data = t.GetData()
	if np.IsCall() {
		np.UndoEntity()
	}
	return np
}

// ToCommand converts a standard payload to a command payload
func ToCommand(t Transformable) *Command {
	np := NewEmptyCommand()
	np.Data = t.GetData()
	if np.IsCommand() {
		np.UndoEntity()
	}
	return np
}

// ToCommandReply converts a standard payload to a command reply payload
func ToCommandReply(t Transformable) *CommandReply {
	np := NewEmptyCommandReply()
	np.Data = t.GetData()
	if np.IsCommandReply() {
		np.UndoEntity()
	}
	return np
}

// ToHttpResponse converts a standard payload to an HTTP response payload.
func ToHttpResponse(t Transformable) *HttpResponse {
	np := NewEmptyHttpResponse()
	np.Data = t.GetData()
	// NOTE: The semantic of a response can be different than HTTP so this
	// condition can be true for other protocols too.
	if np.IsResponse() {
		np.UndoEntity()
	}
	return np
}

// ToTransport converts a standard payload to a transport payload
func ToTransport(t Transformable) *Transport {
	np := NewEmptyTransport()
	np.Data = t.GetData()
	if np.IsTransport() {
		np.UndoEntity()
	}
	return np
}
