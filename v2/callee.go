// Go SDK for the KUSANAGI(tm) framework (http://kusanagi.io)
// Copyright (c) 2016-2021 KUSANAGI S.L. All rights reserved.
//
// Distributed under the MIT license.
//
// For the full copyright and license information, please view the LICENSE
// file that was distributed with this source code.

package kusanagi

// Callee represents a service being called by another service.
type Callee struct {
	gateway  string
	name     string
	version  string
	action   string
	duration uint
	timeout  uint
	params   []*Param
}

// GetDuration returns the duration of the call in milliseconds.
func (c Callee) GetDuration() uint {
	return c.duration
}

// IsRemote checks if the call is to a service in another Realm.
func (c Callee) IsRemote() bool {
	return c.gateway != ""
}

// GetAddress returns the public gateway address for calls to another realm.
func (c Callee) GetAddress() string {
	return c.gateway
}

// GetTimeout returns the timeout in milliseconds for a call to a service in another realm.
func (c Callee) GetTimeout() uint {
	return c.timeout
}

// GetName returns the name of the service being called.
func (c Callee) GetName() string {
	return c.name
}

// GetVersion returns the version of the service being called.
func (c Callee) GetVersion() string {
	return c.version
}

// GetAction returns the name of the service action being called.
func (c Callee) GetAction() string {
	return c.action
}

// GetParams returns the call parameters.
func (c Callee) GetParams() (params []*Param) {
	// Add the parameters to a new list
	for _, p := range c.params {
		params = append(params, p)
	}
	return params
}
