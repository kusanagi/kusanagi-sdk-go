// Go SDK for the KUSANAGI(tm) framework (http://kusanagi.io)
// Copyright (c) 2016-2020 KUSANAGI S.L. All rights reserved.
//
// Distributed under the MIT license.
//
// For the full copyright and license information, please view the LICENSE
// file that was distributed with this source code.

package schema

import (
	"errors"
	"sync"

	"github.com/kusanagi/kusanagi-sdk-go/logging"
	"github.com/kusanagi/kusanagi-sdk-go/payload"
	"github.com/kusanagi/kusanagi-sdk-go/transform"
	"github.com/kusanagi/kusanagi-sdk-go/traverse"
)

var registry *Registry

// GetRegistry gets the global schema registry.
func GetRegistry() *Registry {
	if registry == nil {
		logging.Debug("creating a new schema registry...")
		registry = NewRegistry()
	}
	return registry
}

// NewRegistry creates a new schema mappings registry.
func NewRegistry() *Registry {
	return &Registry{mappings: payload.New()}
}

// Registry defines a registry for service schema mappings.
type Registry struct {
	sync.RWMutex

	mappings *payload.Payload
}

// HasMappings checks if the registry contains schema mappings.
func (r *Registry) HasMappings() bool {
	r.RLock()
	defer r.RUnlock()

	return !r.mappings.IsEmpty()
}

// Update the registry with the latest schema mappings.
func (r *Registry) Update(data []byte) error {
	r.Lock()
	defer r.Unlock()

	if err := transform.Unpack(data, &(r.mappings.Data)); err != nil {
		return err
	} else if r.mappings.IsEmpty() {
		return errors.New("the mappings are empty")
	}
	return nil
}

// Schema gets a schema for a service.
func (r *Registry) Schema(service, version string) map[string]interface{} {
	r.RLock()
	defer r.RUnlock()

	p := traverse.NewSpacedPath(service, version)
	return r.mappings.PgetMap(p.String(), p.Sep)
}

// ServiceNames gets the list of the service names registered with the discovery component.
func (r *Registry) ServiceNames() (names []string) {
	r.RLock()
	defer r.RUnlock()

	// Get all the service names
	for n, _ := range r.mappings.Data {
		names = append(names, n)
	}
	return names
}

// ServiceVersions gets the list of versions for a service registered with the discovery component.
func (r *Registry) ServiceVersions(service string) (versions []string) {
	r.RLock()
	defer r.RUnlock()

	// Get the versions available for a service name
	for v, _ := range r.mappings.GetMap(service) {
		versions = append(versions, v)
	}
	return versions
}
