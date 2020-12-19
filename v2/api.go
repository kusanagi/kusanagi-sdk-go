// Go SDK for the KUSANAGI(tm) framework (http://kusanagi.io)
// Copyright (c) 2016-2020 KUSANAGI S.L. All rights reserved.
//
// Distributed under the MIT license.
//
// For the full copyright and license information, please view the LICENSE
// file that was distributed with this source code.

package kusanagi

import (
	"context"
	"path"

	"github.com/kusanagi/kusanagi-sdk-go/v2/lib/cli"
	"github.com/kusanagi/kusanagi-sdk-go/v2/lib/log"
	"github.com/kusanagi/kusanagi-sdk-go/v2/lib/payload"
)

func newApi(c Component, s *state) *Api {
	if s.schemas == nil {
		s.logger.Warning("Discovery mappings are not available")
	}

	return &Api{
		component: c,
		state:     s,
		logger:    s.logger,
		input:     s.input,
		schemas:   s.schemas,
		command:   s.command,
		reply:     s.reply,
	}
}

// Api type for SDK components.
type Api struct {
	component Component
	state     *state
	input     cli.Input
	schemas   *payload.Mapping
	logger    log.RequestLogger
	command   payload.Command
	reply     *payload.Reply
}

// IsDebug checks if the component is running in debug mode.
func (a *Api) IsDebug() bool {
	return a.input.IsDebugEnabled()
}

// GetFrameworkVersion gets the KUSANAGI framework version.
func (a *Api) GetFrameworkVersion() string {
	return a.input.GetFrameworkVersion()
}

// Get source file path.
func (a *Api) GetPath() string {
	return path.Dir(a.input.GetPath())
}

// Get component name.
func (a *Api) GetName() string {
	return a.input.GetName()
}

// Get component version.
func (a *Api) GetVersion() string {
	return a.input.GetVersion()
}

// Checks if a variable exists.
//
// name: The name of the variable.
func (a *Api) HasVariable(name string) bool {
	return a.input.HasVariable(name)
}

// Gets all component variables.
func (a *Api) GetVariables() map[string]string {
	return a.input.GetVariables()
}

// Get a single component variable.
//
// name: The name of the variable.
func (a *Api) GetVariable(name string) string {
	return a.input.GetVariable(name)
}

// Checks if a resource exists.
//
// name: The name of the resource.
func (a *Api) HasResource(name string) bool {
	return a.component.HasResource(name)
}

// Get a resource.
//
// name: The name of the resource.
func (a *Api) GetResource(name string) (interface{}, error) {
	return a.component.GetResource(name)
}

// Get service names and versions from the mapping schemas.
func (a *Api) GetServices() []payload.ServiceVersion {
	return a.schemas.GetServices()
}

// GetServiceSchema returns a schema for a service.
//
// The version can be either a fixed version or a pattern that uses "*"
// and resolves to the higher version available that matches.
//
// name: The name of the service.
// version: The version of the service.
func (a *Api) GetServiceSchema(name, version string) (*ServiceSchema, error) {
	payload, err := a.schemas.GetSchema(name, version)
	if err != nil {
		return nil, err
	}
	schema := ServiceSchema{name, version, *payload}
	return &schema, nil
}

// Log writes a value to the KUSANAGI logs.
//
// Given value is converted to string before being logged.
//
// Output is truncated to have a maximum of 100000 characters.
//
// value: The value to log.
// level: An optional log level to use for the log message.
func (a *Api) Log(value interface{}, level int) (*Api, error) {
	s, err := log.ValueToLogString(value)
	if err != nil {
		return nil, err
	}
	a.logger.Log(level, s)
	return a, nil
}

// GetAsyncContext return the context for the current request.
func (a *Api) GetAsyncContext() context.Context {
	return a.state.context
}

// Done is a dummy method to comply with KUSANAGI SDK specifications.
func (a *Api) Done() bool {
	panic("SDK does not support async call to end action: Api.done()")
}
