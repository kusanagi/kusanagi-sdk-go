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

	"github.com/kusanagi/kusanagi-sdk-go/v3/lib/cli"
	"github.com/kusanagi/kusanagi-sdk-go/v3/lib/log"
)

// Component represents a KUSANAGI SDK generic component.
type Component interface {
	// HasResource checks if a resource name exist.
	//
	// name: Name of the resource.
	HasResource(name string) bool

	// SetResource stores a resource.
	//
	// name: Name of the resource.
	// factory: A callable that returns the resource value.
	SetResource(name string, factory ResourceFactory) error

	// GetResource returns a resource.
	//
	// name: Name of the resource.
	GetResource(name string) (interface{}, error)

	// Startup registers a callback to be called during component startup.
	//
	// callback: A callback to execute on startup.
	Startup(callback Callback) Component

	// Shutdown registers allback to be called during component shutdown.
	//
	// callback: A callback to execute on shutdown.
	Shutdown(callback Callback) Component

	// Error registers a callback to be called error.
	//
	// callback: A callback to execute when the component fails to handle a request.
	Error(callback ErrorCallback) Component

	// Log writes a value to KUSANAGI logs.
	//
	// Given value is converted to string before being logged.
	//
	// Output is truncated to have a maximum of 100000 characters.
	//
	// value: The value to log.
	// level: An optional log level to use for the log message.
	Log(value interface{}, level int) Component

	// Run the SDK component
	Run() bool
}

// ResourceFactory functions create resources to be stored in a component.
//
// The factory argument is the component that is running.
//
// It is possible to get the specific component by casting, for example:
//  middleware := component.(*Middleware)
// or for service components:
//  service := component.(*Service)
type ResourceFactory func(Component) (interface{}, error)

// ErrorCallback is called whenever an error is returned while processing a framework request in userland.
type ErrorCallback func(error) error

// Callback is called by components during startup and shutdown.
type Callback func(Component) error

// Event handler for components
type eventsHandler struct {
	onStartup  Callback
	onShutdown Callback
	onError    ErrorCallback
}

func (h eventsHandler) startup(c Component) bool {
	if h.onStartup != nil {
		log.Info("Running startup callback...")
		if err := h.onStartup(c); err != nil {
			log.Errorf("Startup callback failed: %v", err)
			return false
		}
	}
	return true
}

func (h eventsHandler) shutdown(c Component) bool {
	if h.onShutdown != nil {
		log.Info("Running shutdown callback...")
		if err := h.onShutdown(c); err != nil {
			log.Errorf("Shutdown callback failed: %v", err)
			return false
		}
	}
	return true
}

func (h eventsHandler) error(e error) bool {
	if h.onError != nil {
		log.Info("Running error callback...")
		if err := h.onError(e); err != nil {
			log.Errorf("Error callback failed: %v", err)
			return false
		}
	}
	return true
}

func newComponent(p requestProcessor) component {
	return component{
		events:    eventsHandler{},
		resources: make(map[string]interface{}),
		callbacks: make(map[string]interface{}),
		processor: p,
	}
}

type component struct {
	events    eventsHandler
	resources map[string]interface{}
	callbacks map[string]interface{}
	processor requestProcessor
}

func (c *component) hasCallback(name string) bool {
	_, ok := c.callbacks[name]
	return ok
}

func (c *component) HasResource(name string) bool {
	_, ok := c.resources[name]
	return ok
}

func (c *component) SetResource(name string, factory ResourceFactory) error {
	resource, err := factory(c)
	if err != nil {
		return err
	} else if resource == nil {
		return fmt.Errorf("invalid result value for resource: \"%s\"", name)
	}
	c.resources[name] = resource
	return nil
}

func (c *component) GetResource(name string) (interface{}, error) {
	if resource, ok := c.resources[name]; ok {
		return resource, nil
	}
	return nil, fmt.Errorf(`resource not found: "%s"`, name)
}

func (c *component) Startup(callback Callback) Component {
	c.events.onStartup = callback
	return c
}

func (c *component) Shutdown(callback Callback) Component {
	c.events.onShutdown = callback
	return c
}

func (c *component) Error(callback ErrorCallback) Component {
	c.events.onError = callback
	return c
}

func (c *component) Log(value interface{}, level int) Component {
	log.Log(level, value)
	return c
}

func (c *component) Run() bool {
	// Read CLI input values
	input, err := cli.Parse()
	if err != nil {
		log.Errorf("Component error: %v", err)
		return false
	}

	// Setup the log level before the server is created
	log.SetLevel(input.GetLogLevel())

	// Run the server and check that all callbacks are run successfully
	success := false
	if c.events.startup(c) {
		server := newServer(input, c, c.processor)
		if err := server.start(); err != nil {
			log.Errorf("Component error: %v", err)
		} else {
			success = true
		}
	}

	// Return false when shutdown fails, otherwise use the success value
	if c.events.shutdown(c) {
		return success
	}
	return false
}
