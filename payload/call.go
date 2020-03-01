// Go SDK for the KUSANAGI(tm) framework (http://kusanagi.io)
// Copyright (c) 2016-2020 KUSANAGI S.L. All rights reserved.
//
// Distributed under the MIT license.
//
// For the full copyright and license information, please view the LICENSE
// file that was distributed with this source code.

package payload

// NewEmptyCall creates a new empty request call payload
func NewEmptyCall() *Call {
	return &Call{Payload: NewNamespaced("call")}
}

// NewCallFromMap creates a new service call from a map.
func NewCallFromMap(m map[string]interface{}) *Call {
	c := NewEmptyCall()
	c.Data = m
	return c
}

// NewCall creates a new service call payload
func NewCall(service, version, action string) *Call {
	c := NewEmptyCall()
	c.SetService(service)
	c.SetVersion(version)
	c.SetAction(action)
	return c
}

// Call defines a service call payload
type Call struct {
	*Payload
}

// GetService gets the name of the Service
func (c Call) GetService() string {
	return c.GetString("service")
}

// SetName sets the name of the Service
func (c *Call) SetService(value string) error {
	return c.Set("service", value)
}

// GetVersion gets the version of the Service
func (c Call) GetVersion() string {
	return c.GetString("version")
}

// SetVersion sets the version of the Service
func (c *Call) SetVersion(value string) error {
	return c.Set("version", value)
}

// GetAction gets the action name of the Service
func (c Call) GetAction() string {
	return c.GetString("action")
}

// SetAction sets the action name of the Service
func (c *Call) SetAction(value string) error {
	return c.Set("action", value)
}

// GetParams gets the additinal parameters to pass to the action.
func (c Call) GetParams() (ps []*Param) {
	values := c.GetSlice("params")
	if len(values) == 0 {
		return nil
	}

	for _, v := range values {
		if m, ok := v.(map[string]interface{}); ok {
			ps = append(ps, NewParamFromMap(m))
		}
	}
	return ps
}

// SetParams sets the additinal parameters to pass to the action.
func (c *Call) SetParams(ps []*Param) error {
	var pps []Data
	for _, p := range ps {
		pps = append(pps, p.Data)
	}
	return c.Set("params", pps)
}
