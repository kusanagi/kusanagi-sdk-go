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

	"github.com/kusanagi/kusanagi-sdk-go/v2/lib"
	"github.com/kusanagi/kusanagi-sdk-go/v2/lib/payload"
)

// Flags used in multipart responses.
var serviceCallFlag = []byte("\x01")
var filesFlag = []byte("\x02")
var transactionsFlag = []byte("\x03")
var downloadFlag = []byte("\x04")

// Handles errors returned by middleware userland callbacks.
func handleMiddlewareUserlandError(middleware *Middleware, state *state, err error) *Response {
	state.logger.Errorf("Callback error: %v", err)

	// Call the userland error handler
	body := err.Error()
	if err := middleware.events.onError(err); err != nil {
		// When the error handler fails use the result as the error
		state.logger.Criticalf("Error callback failed: %v", err)
		body = err.Error()
	}

	// Create a new response with the error as body contents
	response := newResponse(middleware, state)
	httpResponse := response.GetHTTPResponse()
	httpResponse.SetStatus(500, "Internal Server Error")
	httpResponse.SetBody([]byte(body))
	return response
}

// Execute a response middleware userland callback.
func executeResponseMiddleware(middleware *Middleware, state *state) *Response {
	state.reply = payload.NewResponseReply(&state.command)
	callback := middleware.callbacks["response"].(ResponseCallback)
	response, err := callback(newResponse(middleware, state))
	if err != nil {
		response = handleMiddlewareUserlandError(middleware, state, err)
	}
	return response
}

// Execute a request middleware userland callback.
func executeRequestMiddleware(middleware *Middleware, state *state) interface{} {
	state.reply = payload.NewRequestReply(&state.command)
	callback := middleware.callbacks["request"].(RequestCallback)
	result, err := callback(newRequest(middleware, state))
	if err != nil {
		result = handleMiddlewareUserlandError(middleware, state, err)
	}
	return result
}

// Processor for middleware requests.
func middlewareRequestProcessor(c Component, state *state, out chan<- requestOutput) {
	var result interface{}

	// Execute the userland callback
	middleware := c.(*Middleware)
	if state.action == "request" {
		result = executeRequestMiddleware(middleware, state)
	} else {
		result = executeResponseMiddleware(middleware, state)
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
	message, err := lib.Pack(reply)
	if err != nil {
		output.err = fmt.Errorf("Failed to serialize the response: %v", err)
	} else {
		output.response = responseMsg{[]byte(state.id), emptyFrame, message}
	}

	out <- output
}

// Handle error when executing userland service callbacks.
func handleServiceUserlandError(service *Service, state *state, err error) *Action {
	state.logger.Errorf("Callback error: %v", err)

	// Call the userland error handler
	message := err.Error()
	if err := service.events.onError(err); err != nil {
		// When the error handler fails use the result as the error
		state.logger.Criticalf("Error callback failed: %v", err)
		message = err.Error()
	}

	// Create a new action with an error
	action := newAction(service, state)
	action.Error(message, 500, "Internal Server Error")
	return action
}

// Processor for service requests.
func serviceRequestProcessor(c Component, state *state, out chan<- requestOutput) {
	// Execute the userland callback
	service := c.(*Service)
	state.reply = payload.NewActionReply(&state.command)
	callback := service.callbacks[state.action].(ActionCallback)
	action, err := callback(newAction(service, state))
	if err != nil {
		action = handleServiceUserlandError(service, state, err)
	}

	// Inspect the transport to set the flags for the response
	var flags []byte
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

	// Serialize the payload
	output := requestOutput{state: state}
	message, err := lib.Pack(state.reply)
	if err != nil {
		output.err = fmt.Errorf("Failed to serialize the response: %v", err)
	} else {
		output.response = responseMsg{[]byte(state.id), flags, message}
	}

	out <- output
}