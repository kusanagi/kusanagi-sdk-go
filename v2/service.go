// Go SDK for the KUSANAGI(tm) framework (http://kusanagi.io)
// Copyright (c) 2016-2020 KUSANAGI S.L. All rights reserved.
//
// Distributed under the MIT license.
//
// For the full copyright and license information, please view the LICENSE
// file that was distributed with this source code.

package kusanagi

// ActionCallback is called when a service request is received.
type ActionCallback func(*Action) (*Action, error)

// NewService creates a new Service component.
func NewService() *Service {
	return &Service{newComponent(serviceRequestProcessor)}
}

// Service component.
type Service struct {
	component
}

// Action assigns a callback to execute when a service action request is received.
func (s *Service) Action(name string, callback ActionCallback) *Service {
	s.callbacks[name] = callback
	return s
}
