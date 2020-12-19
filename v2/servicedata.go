// Go SDK for the KUSANAGI(tm) framework (http://kusanagi.io)
// Copyright (c) 2016-2020 KUSANAGI S.L. All rights reserved.
//
// Distributed under the MIT license.
//
// For the full copyright and license information, please view the LICENSE
// file that was distributed with this source code.

package kusanagi

// ServiceData represents a service which stored data in the transport.
type ServiceData struct {
	address string
	service string
	version string
	actions map[string][]interface{}
}

// GetAddress returns the gateway address for the service.
func (s ServiceData) GetAddress() string {
	return s.address
}

// GetName returns the service name.
func (s ServiceData) GetName() string {
	return s.service
}

// GetVersion returns the service version.
func (s ServiceData) GetVersion() string {
	return s.version
}

// GetActions returns the list of action data items for current service.
//
// Each item represents an action on the service for which data exists.
func (s ServiceData) GetActions() (actions []ActionData) {
	for name, data := range s.actions {
		actions = append(actions, ActionData{name, data})
	}
	return actions
}
