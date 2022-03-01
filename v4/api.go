// Go SDK for the KUSANAGI(tm) framework (http://kusanagi.io)
// Copyright (c) 2016-2022 KUSANAGI S.L. All rights reserved.
//
// Distributed under the MIT license.
//
// For the full copyright and license information, please view the LICENSE
// file that was distributed with this source code.

package kusanagi

import (
	"path"

	"github.com/kusanagi/kusanagi-sdk-go/v4/lib/cli"
	"github.com/kusanagi/kusanagi-sdk-go/v4/lib/log"
	"github.com/kusanagi/kusanagi-sdk-go/v4/lib/payload"
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

// GetPath returns the source file path.
func (a *Api) GetPath() string {
	return path.Dir(a.input.GetPath())
}

// GetName returns the component name.
func (a *Api) GetName() string {
	return a.input.GetName()
}

// GetVersion returns the component version.
func (a *Api) GetVersion() string {
	return a.input.GetVersion()
}

// HasVariable checks if a variable exists.
//
// name: The name of the variable.
func (a *Api) HasVariable(name string) bool {
	return a.input.HasVariable(name)
}

// GetVariables returns all component variables.
func (a *Api) GetVariables() map[string]string {
	return a.input.GetVariables()
}

// GetVariable returns a single component variable.
//
// name: The name of the variable.
func (a *Api) GetVariable(name string) string {
	return a.input.GetVariable(name)
}

// HasResource checks if a resource exists.
//
// name: The name of the resource.
func (a *Api) HasResource(name string) bool {
	return a.component.HasResource(name)
}

// GetResource returns a resource.
//
// name: The name of the resource.
func (a *Api) GetResource(name string) (interface{}, error) {
	return a.component.GetResource(name)
}

// GetServices return service names and versions from the mapping schemas.
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

// Done returns a channel that signals the deadline or cancellation of the call.
func (a *Api) Done() <-chan struct{} {
	return a.state.ctx.Done()
}
