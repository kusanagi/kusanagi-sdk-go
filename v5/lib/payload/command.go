// Go SDK for the KUSANAGI(tm) framework (http://kusanagi.io)
// Copyright (c) 2016-2023 KUSANAGI S.L. All rights reserved.
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
	return c.Command.Arguments.GetAttributes()
}

// GetCall returns the information of the service to contact.
func (c Command) GetCall() *CallInfo {
	return c.Command.Arguments.GetCall()
}

// GetTransport returns the transport payload.
func (c Command) GetTransport() *Transport {
	return c.Command.Arguments.Transport
}

// GetResponse returns the HTTP response payload.
func (c Command) GetResponse() *HTTPResponse {
	return c.Command.Arguments.Response
}

// CommandInfo contains the semantics of the command.
type CommandInfo struct {
	Name      string            `json:"n"`
	Arguments *CommandArguments `json:"a,omitempty"`
}

// ActionParams contains all the parameters sent to an action during a runtime call.
type ActionParams []Param

// ActionFiles contains all the files sent to an action during a runtime call.
type ActionFiles []File

// CommandArguments contains the arguments of the command.
// TODO: The specs need to fix the name clashes for "c" and "a" in the arguments.
type CommandArguments struct {
	// NOTE: "A" might contain the action name (string) or attributes (map[string]string)
	A interface{} `json:"a,omitempty"`

	// NOTE: "C" might contain the callee ([]string) or call info (map[string]interface{})
	C interface{} `json:"c,omitempty"`

	Meta      Meta          `json:"m,omitempty"`
	Request   *HTTPRequest  `json:"r,omitempty"`
	Response  *HTTPResponse `json:"R,omitempty"`
	Transport *Transport    `json:"T,omitempty"`
	Params    ActionParams  `json:"p,omitempty"` // TODO: The specs seem to be wrong here
	Files     ActionFiles   `json:"f,omitempty"`
	Return    interface{}   `json:"rv,omitempty"`
}

// GetCall returns the info for the call.
func (a *CommandArguments) GetCall() *CallInfo {
	if data, ok := a.C.(map[string]interface{}); ok {
		return mapToCallInfo(data)
	}
	return nil
}

// GetAttributes returns the attributes for the command.
func (a *CommandArguments) GetAttributes() map[string]string {
	v, _ := a.A.(map[string]string)
	return v
}

// GetAction returns the action name for the call.
func (a *CommandArguments) GetAction() string {
	v, _ := a.A.(string)
	return v
}

// SetAction sets the name of the action for the call.
func (a *CommandArguments) SetAction(name string) {
	a.A = name
}

// GetCallee returns the callee service information.
func (a *CommandArguments) GetCallee() (callee []string) {
	if values, ok := a.C.([]interface{}); ok {
		// Cast the values in the slice to string
		for _, v := range values {
			callee = append(callee, v.(string))
		}
	}
	return callee
}

// SetCallee sets the calle service information.
func (a *CommandArguments) SetCallee(callee []string) {
	a.C = callee
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
