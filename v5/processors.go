// Go SDK for the KUSANAGI(tm) framework (http://kusanagi.io)
// Copyright (c) 2016-2022 KUSANAGI S.L. All rights reserved.
//
// Distributed under the MIT license.
//
// For the full copyright and license information, please view the LICENSE
// file that was distributed with this source code.

package kusanagi

import (
	"fmt"
	"runtime/debug"

	"github.com/kusanagi/kusanagi-sdk-go/v5/lib/msgpack"
	"github.com/kusanagi/kusanagi-sdk-go/v5/lib/payload"
)

// Flags used in multipart responses.
var serviceCallFlag = []byte("\x01")
var filesFlag = []byte("\x02")
var transactionsFlag = []byte("\x03")
var downloadFlag = []byte("\x04")

func buildErrorResponse(m *Middleware, s *state, err error) *Response {
	s.logger.Errorf("Callback error: %v", err)

	// Call the userland error handler
	m.events.error(err)

	// Create a new response with the error as body contents
	r := newResponse(m, s)

	hr := r.GetHTTPResponse()
	hr.SetStatus(500, "Internal Server Error")
	hr.SetBody([]byte(err.Error()))

	return r
}

// Execute a response middleware userland callback.
func executeResponseMiddleware(m *Middleware, s *state) *Response {
	s.reply = payload.NewResponseReply(&s.command)
	callback := m.callbacks["response"].(ResponseCallback)

	r, err := callback(newResponse(m, s))
	if err != nil {
		r = buildErrorResponse(m, s, err)
	}

	return r
}

// Execute a request middleware userland callback.
func executeRequestMiddleware(m *Middleware, s *state) interface{} {
	s.reply = payload.NewRequestReply(&s.command)
	callback := m.callbacks["request"].(RequestCallback)

	r, err := callback(newRequest(m, s))
	if err != nil {
		r = buildErrorResponse(m, s, err)
	}

	return r
}

// Processor for middleware requests.
func middlewareRequestProcessor(c Component, state *state, out chan<- requestOutput) {
	defer close(out)

	defer func() {
		// Handle panics gracefully
		if err := recover(); err != nil {
			state.logger.Criticalf("Panic: %v\n%s", err, debug.Stack())

			out <- requestOutput{state: state, err: fmt.Errorf("Panic: %v", err)}
		}
	}()

	var result interface{}

	// Execute the userland callback
	m := c.(*Middleware)
	if state.action == "request" {
		result = executeRequestMiddleware(m, state)
	} else {
		result = executeResponseMiddleware(m, state)
	}

	var reply payload.Reply

	// Get the payload for the response
	if _, ok := result.(*Request); ok {
		reply = state.reply.ForRequest()
	} else {
		reply = state.reply.ForResponse()
	}

	// Serialize the payload
	output := requestOutput{state: state}
	message, err := msgpack.Encode(reply)
	if err != nil {
		output.err = fmt.Errorf("Failed to serialize the response: %v", err)
	} else {
		output.response = responseMsg{emptyFrame, message}
	}

	out <- output
}

// Processor for service requests.
func serviceRequestProcessor(c Component, state *state, out chan<- requestOutput) {
	defer close(out)

	defer func() {
		// Handle panics gracefully
		if err := recover(); err != nil {
			state.logger.Criticalf("Panic: %v\n%s", err, debug.Stack())

			out <- requestOutput{state: state, err: fmt.Errorf("Panic: %v", err)}
		}
	}()

	// Execute the userland callback
	service := c.(*Service)
	callback := service.callbacks[state.action].(ActionCallback)
	state.reply = payload.NewActionReply(&state.command)

	action, err := callback(newAction(service, state))
	if action == nil {
		panic(fmt.Sprintf("callback returned a nil action: %s", state.action))
	} else if err != nil {
		state.logger.Errorf("Callback error: %v", err)

		// Call the userland error handler
		service.events.error(err)

		// Add the error to the action to it is saved in the transport
		action.Error(err.Error(), 0, "500 Internal Server Error")
	}

	var flags []byte

	// Inspect the transport to set the flags for the response
	if t := state.reply.GetTransport(); t != nil {
		if t.HasCalls(action.GetName(), action.GetVersion()) {
			flags = append(flags, serviceCallFlag...)
		}

		if t.Files != nil {
			flags = append(flags, filesFlag...)
		}

		if t.Transactions != nil {
			flags = append(flags, transactionsFlag...)
		}

		if t.Body != nil {
			flags = append(flags, downloadFlag...)
		}
	}

	if flags == nil {
		flags = emptyFrame
	}

	output := requestOutput{state: state}

	// Serialize the payload
	message, err := msgpack.Encode(state.reply)
	if err != nil {
		output.err = fmt.Errorf("Failed to serialize the response: %v", err)
	} else {
		output.response = responseMsg{flags, message}
	}

	out <- output
}
