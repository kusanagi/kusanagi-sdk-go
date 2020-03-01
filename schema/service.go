// Go SDK for the KUSANAGI(tm) framework (http://kusanagi.io)
// Copyright (c) 2016-2020 KUSANAGI S.L. All rights reserved.
//
// Distributed under the MIT license.
//
// For the full copyright and license information, please view the LICENSE
// file that was distributed with this source code.

package schema

import (
	"fmt"

	"github.com/kusanagi/kusanagi-sdk-go/payload"
)

// NewService creates a new Service schema
func NewService(name, version string, p *payload.Payload) *Service {
	var a map[string]interface{}

	if p == nil {
		// When no payload is given use an empty payload
		p = payload.New()
	} else {
		// Get actions from payload
		a = p.GetMap("actions")
	}

	// Create an empty map when no actions exists
	if a == nil {
		a = make(map[string]interface{})
	}
	return &Service{name: name, version: version, payload: p, actions: a}
}

// Service defines a Service schema
type Service struct {
	name,
	version string
	payload *payload.Payload
	actions map[string]interface{}
}

// GetName gets the name of the Service
func (s Service) GetName() string {
	return s.name
}

// GetVersion gets the version of the Service
func (s Service) GetVersion() string {
	return s.version
}

// HasFileServer checks if Service has a file server
func (s Service) HasFileServer() bool {
	return s.payload.GetBool("files")
}

// GetActions gets Service action names
func (s Service) GetActions() []string {
	names := []string{}
	for k := range s.actions {
		names = append(names, k)
	}
	return names
}

// HasAction checks if an action exists for current Service schema
func (s Service) HasAction(name string) bool {
	_, exists := s.actions[name]
	return exists
}

// GetActionSchema gets the schema for an action
func (s Service) GetActionSchema(name string) (*Action, error) {
	if !s.HasAction(name) {
		return nil, fmt.Errorf("Can't resolve schema for action: %v", name)
	}

	// Create a payload with action schema data
	p := payload.New()
	p.Data = s.actions[name].(map[string]interface{})

	return NewAction(name, p), nil
}

// GetHTTPSchema gets HTTP Service schema
func (s Service) GetHTTPSchema() *HTTPService {
	// Get HTTP schema data if it exists
	p := payload.New()
	if v := s.payload.GetMap("http"); v != nil {
		p.Data = v
	}
	return NewHTTPService(p)
}

// NewHTTPService creates a new HTTP Service schema
func NewHTTPService(p *payload.Payload) *HTTPService {
	// When no payload is given use an empty payload
	if p == nil {
		p = payload.New()
	}
	return &HTTPService{payload: p}
}

// HTTPService represents the HTTP semantics of a Service
type HTTPService struct {
	payload *payload.Payload
}

// IsAccessible checks if the Gateway has access to the Service
func (hs HTTPService) IsAccessible() bool {
	return hs.payload.GetDefault("gateway", true).(bool)
}

// GetBasePath gets the base HTTP path for the Service
func (hs HTTPService) GetBasePath() string {
	return hs.payload.GetString("base_path")
}
