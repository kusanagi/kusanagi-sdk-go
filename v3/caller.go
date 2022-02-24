// Go SDK for the KUSANAGI(tm) framework (http://kusanagi.io)
// Copyright (c) 2016-2022 KUSANAGI S.L. All rights reserved.
//
// Distributed under the MIT license.
//
// For the full copyright and license information, please view the LICENSE
// file that was distributed with this source code.

package kusanagi

// Caller represents a service which registered call in the transport.
type Caller struct {
	service string
	version string
	action  string
	callee  Callee
}

// GetName returns the service name.
func (c Caller) GetName() string {
	return c.service
}

// GetVersion returns the service version.
func (c Caller) GetVersion() string {
	return c.version
}

// GetAction returns the name of the service action that is making the call.
func (c Caller) GetAction() string {
	return c.action
}

// GetCallee returns the callee info for the service being called.
func (c Caller) GetCallee() Callee {
	return c.callee
}
