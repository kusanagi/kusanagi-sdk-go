// Go SDK for the KUSANAGI(tm) framework (http://kusanagi.io)
// Copyright (c) 2016-2020 KUSANAGI S.L. All rights reserved.
//
// Distributed under the MIT license.
//
// For the full copyright and license information, please view the LICENSE
// file that was distributed with this source code.

package payload

// NewEmptyCommand creates a new empty command payload
func NewEmptyCommand() *Command {
	return &Command{Payload: NewNamespaced("command")}
}

// NewCommand creates a new command payload
func NewCommand(name, scope string) *Command {
	c := NewEmptyCommand()
	c.SetName(name)
	c.SetScope(scope)
	return c
}

// Command defines a command payload
type Command struct {
	*Payload
}

// GetName gets command name
func (c Command) GetName() string {
	return c.GetString("command/name")
}

// SetName sets command name
func (c *Command) SetName(value string) error {
	if value == "" {
		return nil
	}
	return c.Set("command/name", value)
}

// GetScope gets command scope
func (c Command) GetScope() string {
	return c.GetString("meta/scope")
}

// SetScope sets command scope
func (c *Command) SetScope(value string) error {
	if value == "" {
		return nil
	}
	return c.Set("meta/scope", value)
}

// GetArgs gets command arguments
func (c Command) GetArgs() map[string]interface{} {
	if value := c.GetMap("command/arguments"); value != nil {
		return value
	}
	return nil
}

// SetArgs sets command arguments
func (c *Command) SetArgs(value map[string]interface{}) error {
	return c.Set("command/arguments", value)
}

// NewEmptyCommandReply creates a new empty command reply payload
func NewEmptyCommandReply() *CommandReply {
	return &CommandReply{Payload: NewNamespaced("command_reply")}
}

// NewCommandReply creates a new command reply payload
func NewCommandReply(name string, result interface{}) *CommandReply {
	cr := NewEmptyCommandReply()
	cr.Set("name", name)
	cr.Set("result", result)
	return cr
}

// NewCommandReplyFromMap creates a new command reply payload from a map
func NewCommandReplyFromMap(data map[string]interface{}) *CommandReply {
	cr := NewEmptyCommandReply()
	cr.Data = data
	return cr
}

// CommandReply defines a command reply payload
type CommandReply struct {
	*Payload
}

// GetName gets command name
func (cr CommandReply) GetName() string {
	return cr.GetString("name")
}

// SetName sets command name
func (cr *CommandReply) SetName(value string) error {
	return cr.Set("name", value)
}

// GetResult gets command reply result
func (cr CommandReply) GetResult() interface{} {
	if value := cr.GetDefault("result", nil); value != nil {
		return value
	}
	return nil
}

// SetResult sets command reply result
func (cr *CommandReply) SetResult(value interface{}) error {
	return cr.Set("result", value)
}
