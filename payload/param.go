// Go SDK for the KUSANAGI(tm) framework (http://kusanagi.io)
// Copyright (c) 2016-2020 KUSANAGI S.L. All rights reserved.
//
// Distributed under the MIT license.
//
// For the full copyright and license information, please view the LICENSE
// file that was distributed with this source code.

package payload

// NewEmptyParam creates a new empty param payload
func NewEmptyParam() *Param {
	return &Param{Payload: New()}
}

// NewParamFromMap creates a new param payload from a map
func NewParamFromMap(m map[string]interface{}) *Param {
	p := NewEmptyParam()
	p.Data = m
	return p
}

// NewParam creates a new param payload
func NewParam(name, pType string) *Param {
	c := NewEmptyParam()
	c.SetName(name)
	c.SetType(pType)
	return c
}

// Param defines a param payload
type Param struct {
	*Payload
}

func (p *Param) GetName() string {
	return p.GetString("name")
}

func (p *Param) SetName(name string) error {
	return p.Set("name", name)
}

func (p *Param) GetType() string {
	return p.GetDefault("type", "string").(string)
}

func (p *Param) SetType(pType string) error {
	if pType == "" {
		pType = "string"
	}
	return p.Set("type", pType)
}

func (p *Param) GetValue() interface{} {
	return p.GetDefault("value", nil)
}

func (p *Param) SetValue(v interface{}) error {
	return p.Set("value", v)
}
