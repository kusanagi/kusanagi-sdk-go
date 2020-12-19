// Go SDK for the KUSANAGI(tm) framework (http://kusanagi.io)
// Copyright (c) 2016-2020 KUSANAGI S.L. All rights reserved.
//
// Distributed under the MIT license.
//
// For the full copyright and license information, please view the LICENSE
// file that was distributed with this source code.

package payload

// NewCommand creates a new command payload.
func NewCommand(name, scope string) Command {
	return Command{
		Command: CommandInfo{
			Name: name,
		},
		Meta: CommandMeta{
			Scope: scope,
		},
	}
}

// Command represents a framework command payload.
type Command struct {
	Command CommandInfo `json:"c"`
	Meta    CommandMeta `json:"m"`
}

// GetName returns the name of the command.
func (c Command) GetName() string {
	return c.Command.Name
}

// GetRequestID returns the ID of the current request.
func (c Command) GetRequestID() string {
	// The ID is available for request, response and action commands.
	// Request and response payloads have a meta argument with the ID.
	// Action have the ID in the transport meta instead.
	if c.Command.Arguments.Meta.ID != "" {
		return c.Command.Arguments.Meta.ID
	} else if c.Command.Arguments.Transport != nil {
		if t := c.Command.Arguments.Transport; t.Meta.ID != "" {
			return t.Meta.ID
		}
	}
	return ""
}

// GetAttributes returns the command attributes.
func (c Command) GetAttributes() map[string]string {
	return c.Command.Arguments.Attributes
}

// GetCall returns the information of the service to contact.
func (c Command) GetCall() *CallInfo {
	return c.Command.Arguments.Call
}

// GetTransport returns the transport payload.
func (c Command) GetTransport() *Transport {
	return c.Command.Arguments.Transport
}

// GetResponse returns the HTTP response payload.
func (c Command) GetResponse() *HTTPResponse {
	return c.Command.Arguments.Response
}

// CommandMeta contains the semantics of the command.
type CommandInfo struct {
	Name      string            `json:"n"`
	Arguments *CommandArguments `json:"a,omitempty"`
}

// ActionParams contains all the parameters sent to the action.
type ActionParams []Param

// CommandArguments contains the arguments of the command.
type CommandArguments struct {
	// NOTE: Attributes must also exist here according to the specs
	Attributes map[string]string `json:"a,omitempty"`

	Meta      Meta          `json:"m,omitempty"`
	Call      *CallInfo     `json:"c,omitempty"`
	Request   *HTTPRequest  `json:"r,omitempty"`
	Response  *HTTPResponse `json:"R,omitempty"`
	Transport *Transport    `json:"T,omitempty"`
	Params    ActionParams  `json:"p,omitempty"` // TODO: The specs seems to be wrong here
	Return    interface{}   `json:"rv,omitempty"`
}

// CommandMeta contains the meta-data associated with the command.
type CommandMeta struct {
	Scope string `json:"s"`
}

// Meta contains the meta-data associated with the payload.
type Meta struct {
	Version    string            `json:"v"`
	ID         string            `json:"i"`
	Datetime   string            `json:"d"`
	Type       uint              `json:"t"`
	Protocol   string            `json:"p"`
	Gateway    []string          `json:"g"`
	Client     string            `json:"c"`
	Attributes map[string]string `json:"a,omitempty"`
}

// GetGateway returns the gateway addresses.
//
// The result contains two items, where the first item is the internal
// address and the second is the public address.
func (m Meta) GetGateway() []string {
	if len(m.Gateway) == 0 {
		return []string{"", ""}
	}
	return m.Gateway
}
