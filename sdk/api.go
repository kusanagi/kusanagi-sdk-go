// Go SDK for the KUSANAGI(tm) framework (http://kusanagi.io)
// Copyright (c) 2016-2020 KUSANAGI S.L. All rights reserved.
//
// Distributed under the MIT license.
//
// For the full copyright and license information, please view the LICENSE
// file that was distributed with this source code.

package sdk

import (
	"errors"
	"fmt"
	"strings"

	"github.com/kusanagi/kusanagi-sdk-go/logging"
	"github.com/kusanagi/kusanagi-sdk-go/payload"
	"github.com/kusanagi/kusanagi-sdk-go/schema"
	"github.com/kusanagi/kusanagi-sdk-go/version"
)

// ErrSchemaResolve is used when the information for a service is not found inside the current mappings schema.
type ErrSchemaResolve struct {
	Service string
	Version string
}

func (e ErrSchemaResolve) Error() string {
	return fmt.Sprintf("Cannot resolve schema for Service: \"%v\" (%v)", e.Service, e.Version)
}

type Variables map[string]string

type Api interface {
	// IsDebug checks if component is running in debug mode.
	IsDebug() bool

	// GetFrameworkVersion gets KUSANAGI framework version.
	GetFrameworkVersion() string

	// GetPath gets source file path.
	GetPath() string

	// GetName gets component name.
	GetName() string

	// GetVersion gets component version.
	GetVersion() string

	// GetVariables gets all component variables.
	GetVariables() Variables

	// HasVariable checks if a component variable exists.
	HasVariable(name string) bool

	// GetVariable gets a single component variable value.
	GetVariable(name string) string

	// HasResource checks if a resource exists.
	HasResource(name string) bool

	// GetResource gets a resource.
	GetResource(name string) interface{}

	// GetServices gets service names and versions.
	GetServices() (services []map[string]string)

	// GetServiceSchema gets the schema for a service version.
	//
	// Service version string may contain many `*` that will be
	// resolved to the higher version available. For example: `1.*.*`.
	GetServiceSchema(name, ver string) (*schema.Service, error)

	// Log writes a value to KUSANAGI logs.
	//
	// Given value is converted to string before being logged.
	// Output is truncated to have a maximum of 100000 characters.
	Log(value interface{}, level int)

	// Done is a dummy method that only returns an error.
	//
	// It is implemented to comply with KUSANAGI SDK specificacions.
	Done() error
}

type Resourcer interface {
	HasResource(name string) bool
	GetResource(name string) (interface{}, error)
}

func newApi(c Resourcer, path, name, version, frameworkVersion string) *api {
	return &api{
		component:        c,
		path:             path,
		name:             name,
		version:          version,
		frameworkVersion: frameworkVersion,
		registry:         schema.GetRegistry(),
	}
}

type api struct {
	component        Resourcer
	path             string
	name             string
	version          string
	frameworkVersion string
	registry         *schema.Registry
	variables        map[string]string
	debug            bool
}

func (a api) IsDebug() bool {
	return a.debug
}

func (a api) GetFrameworkVersion() string {
	return a.frameworkVersion
}

func (a api) GetPath() string {
	return a.path
}

func (a api) GetName() string {
	return a.name
}

func (a api) GetVersion() string {
	return a.version
}

func (a api) GetVariables() Variables {
	return a.variables
}

func (a api) HasVariable(name string) bool {
	_, ok := a.variables[name]
	return ok
}

func (a api) GetVariable(name string) string {
	return a.variables[name]
}

func (a api) HasResource(name string) bool {
	return a.component.HasResource(name)
}

func (a api) GetResource(name string) interface{} {
	v, _ := a.component.GetResource(name)
	return v
}

func (a api) GetServices() (services []map[string]string) {
	for _, n := range a.registry.ServiceNames() {
		for _, v := range a.registry.ServiceVersions(n) {
			services = append(services, map[string]string{
				"name":    n,
				"version": v,
			})
		}
	}
	return services
}

func (a api) GetServiceSchema(name, ver string) (*schema.Service, error) {
	// Resolve service version when wildcards are used
	if strings.Index(ver, "*") != -1 {
		v, err := version.New(ver)
		if err != nil {
			return nil, err
		}

		ver, err = v.Resolve(a.registry.ServiceVersions(name))
		if err != nil {
			return nil, err
		}
	}

	data := a.registry.Schema(name, ver)
	if data == nil {
		return nil, ErrSchemaResolve{name, ver}
	}
	p := payload.New()
	p.Data = data
	return schema.NewService(name, ver, p), nil
}

func (a api) Log(value interface{}, level int) {
	if a.IsDebug() {
		if err := logging.DebugValue(level, value); err != nil {
			logging.Errorf("value logging failed: %v", err)
		}
	}
}

func (a api) Done() error {
	return errors.New("SDK does not support async call to end action: Api.done()")
}
