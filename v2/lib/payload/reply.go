// Go SDK for the KUSANAGI(tm) framework (http://kusanagi.io)
// Copyright (c) 2016-2020 KUSANAGI S.L. All rights reserved.
//
// Distributed under the MIT license.
//
// For the full copyright and license information, please view the LICENSE
// file that was distributed with this source code.

package payload

import (
	"fmt"
	"net/http"
)

// NewErrorReply creates a new error reply payload.
func NewErrorReply() Reply {
	return Reply{
		Error: &Error{
			Status:  DefaultErrorStatus,
			Message: DefaultErrorMessage,
		},
	}
}

// NewRequestReply creates a new command reply for a request.
func NewRequestReply(c *Command) *Reply {
	call := c.GetCall()
	return &Reply{
		Command: &CommandReply{
			Name: c.GetName(),
			Result: CommandResult{
				Attributes: c.GetAttributes(),
				Call: &CallInfo{
					Service: call.Service,
					Version: call.Version,
					Action:  call.Action,
					Params:  call.Params,
				},
				Response: NewHTTPResponse(),
			},
		},
	}
}

// NewResponseReply creates a new command reply for a response.
func NewResponseReply(c *Command) *Reply {
	return &Reply{
		Command: &CommandReply{
			Name: c.GetName(),
			Result: CommandResult{
				Attributes: c.GetAttributes(),
				Response:   c.GetResponse(),
			},
		},
	}
}

// NewActionReply creates a new command reply for a service call.
func NewActionReply(c *Command) *Reply {
	return &Reply{
		Command: &CommandReply{
			Name: c.GetName(),
			Result: CommandResult{
				Transport: c.GetTransport(),
			},
		},
	}
}

// Reply represents a generic reply to a framework command.
type Reply struct {
	Error   *Error        `json:"E,omitempty"`
	Command *CommandReply `json:"cr,omitempty"`
}

// IsError checks if the reply is an error reply.
func (r *Reply) IsError() bool {
	return r.Error != nil
}

// IsCommand checks if the reply is a command reply.
func (r *Reply) IsCommand() bool {
	return r.Command != nil
}

// IsValid checks if the reply is a valid command reply.
func (r *Reply) IsValid() bool {
	return r.Command != nil || r.Error != nil
}

// GetTransport returns the transport for the reply.
func (r *Reply) GetTransport() *Transport {
	if r.Command != nil {
		return r.Command.Result.Transport
	}
	return nil
}

// GetReturnValue returns the return value for the reply.
func (r *Reply) GetReturnValue() interface{} {
	if r.Command != nil {
		return r.Command.Result.Return
	}
	return nil
}

// SetResponse sets a response in the payload.
//
// code: The HTTP status code for the response.
// text: The HTTP status text for the response.
func (r *Reply) SetResponse(code int, text string) Reply {
	r.Command.Result.Response = &HTTPResponse{
		Version: HTTPResponseVersion,
		Status:  fmt.Sprintf("%d %s", code, text),
		Headers: http.Header{},
		Body:    []byte(""),
	}
	return *r
}

// ForRequest prepares the reply for a request middleware.
func (r *Reply) ForRequest() Reply {
	r.Command.Result.Response = nil
	return *r
}

// ForResponse prepares the reply for a response middleware.
func (r *Reply) ForResponse() Reply {
	r.Command.Result.Transport = nil
	return *r
}

// CommandReply represents a successful reply to a command.
type CommandReply struct {
	Name   string        `json:"n"`
	Result CommandResult `json:"r"`
}

// IsRequest checks if the reply is a middleware request reply.
func (r CommandReply) IsRequest() bool {
	return r.Result.Response != nil && r.Result.Call != nil
}

// IsResponse checks if the reply is a middleware response reply.
func (r CommandReply) IsResponse() bool {
	return r.Result.Response != nil && r.Result.Call == nil
}

// IsAction checks if the reply is a service action reply.
func (r CommandReply) IsAction() bool {
	return r.Result.Transport != nil
}

// CommandResult contains the result values of a command reply.
type CommandResult struct {
	Attributes map[string]string `json:"a,omitempty"`
	Call       *CallInfo         `json:"c,omitempty"`
	Response   *HTTPResponse     `json:"R,omitempty"`
	Transport  *Transport        `json:"T,omitempty"`
	Return     interface{}       `json:"rv,omitempty"`
}

// Create a new CallInfo from a map.
// NOTE: This function is required because there is an issue with the command
// payload where the same short name "c" is used for "call" info and "callee".
func mapToCallInfo(data map[string]interface{}) *CallInfo {
	c := &CallInfo{
		Service: data["s"].(string),
		Version: data["v"].(string),
		Action:  data["a"].(string),
	}
	if params, ok := data["p"]; ok {
		for _, v := range params.([]interface{}) {
			p := v.(map[string]interface{})
			c.Params = append(c.Params, Param{
				Name:  p["n"].(string),
				Value: p["v"],
				Type:  p["t"].(string),
			})
		}
	}
	return c
}

// CallInfo contains the information of the service to contact.
type CallInfo struct {
	Service string  `json:"s"`
	Version string  `json:"v"`
	Action  string  `json:"a"`
	Params  []Param `json:"p,omitempty"`
}
