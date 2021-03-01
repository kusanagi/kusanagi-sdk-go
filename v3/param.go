// Go SDK for the KUSANAGI(tm) framework (http://kusanagi.io)
// Copyright (c) 2016-2021 KUSANAGI S.L. All rights reserved.
//
// Distributed under the MIT license.
//
// For the full copyright and license information, please view the LICENSE
// file that was distributed with this source code.

package kusanagi

import (
	"fmt"

	"github.com/kusanagi/kusanagi-sdk-go/v3/lib/datatypes"
	"github.com/kusanagi/kusanagi-sdk-go/v3/lib/payload"
)

// Cast a value from one supported type to another
// TODO: Cast from type to type using strconv
func cast(value interface{}, valueType string) (v interface{}, ok bool) {
	// The following types are the only ones that can be used to cast other types.
	// Casting from other types to "array" or "object" is not supported.
	switch valueType {
	case datatypes.Null:
		v = nil
		ok = true
	case datatypes.String:
		v, ok = value.(string)
	case datatypes.Binary:
		v, ok = value.([]byte)
	case datatypes.Integer:
		v, ok = value.(int)
	case datatypes.Float:
		v, ok = value.(float64)
	case datatypes.Boolean:
		v, ok = value.(bool)
	}
	return v, ok
}

// Creates a new parameter.
//
// name: Name of the parameter.
// value: Optional value for the parameter.
// valueType: Optional type for the parameter value.
// exists: Optional flag to know if the parameter exists in the service call.
func newParam(name string, value interface{}, valueType string, exists bool) (*Param, error) {
	if valueType == "" {
		valueType = datatypes.String
	} else if !payload.IsValidType(valueType) {
		return nil, fmt.Errorf(`Invalid parameter type: "%s"`, valueType)
	}

	if t := datatypes.ResolveType(value); t != valueType {
		return nil, fmt.Errorf("Value must be %s", valueType)
	}

	return &Param{name, value, valueType, exists}, nil
}

// Creates a new empty parameter.
func newEmptyParam(name string) *Param {
	p, _ := newParam(name, "", "", false)
	return p
}

// Param represents an input parameter.
//
// Actions receive parameters thought calls to a service component.
type Param struct {
	name      string
	value     interface{}
	valueType string
	exists    bool
}

// GetName reads the name of the parameter.
func (p *Param) GetName() string {
	return p.name
}

// GetType reads the type of the parameter value.
func (p *Param) GetType() string {
	return p.valueType
}

// GetValue reads the value of the parameter.
func (p *Param) GetValue() interface{} {
	return p.value
}

// Exists checks if the parameter exists in the service call.
func (p *Param) Exists() bool {
	return p.exists
}

// CopyWithName creates a copy of the parameter with a different name.
//
// name: Name of the new parameter.
func (p *Param) CopyWithName(name string) *Param {
	return &Param{name, p.GetValue(), p.GetType(), p.Exists()}
}

// CopyWithValue creates a copy of the parameter with a different value.
//
// value: Value for the new parameter.
func (p *Param) CopyWithValue(value interface{}) *Param {
	return &Param{p.GetName(), value, p.GetType(), p.Exists()}
}

// CopyWithType creates a copy of the parameter with a different type.
//
// valueType: Value type for the new parameter.
func (p *Param) CopyWithType(valueType string) (*Param, error) {
	var value interface{} = p.GetValue()

	// When the parameter type is different cast the current value to the new type
	if valueType != p.GetType() {
		name := p.GetName()

		// Check that the type is supported
		if !payload.IsValidType(valueType) {
			return nil, fmt.Errorf(
				`Param "%s" copy failed because the type is invalid: "%s"`,
				name,
				valueType,
			)
		}

		// Cast the value to the new type
		if v, ok := cast(value, valueType); ok {
			value = v
		} else {
			return nil, fmt.Errorf(
				`Param "{%s}" copy failed: Type "{%s}" is not compatible with "{%s}"`,
				name,
				valueType,
				p.GetType(),
			)
		}
	}
	return &Param{p.GetName(), value, valueType, p.Exists()}, nil
}

// Converts a param to a param payload.
func paramToPayload(p *Param) payload.Param {
	return payload.Param{
		Name:  p.GetName(),
		Value: p.GetValue(),
		Type:  p.GetType(),
	}
}

// Converts a param payload to a param.
func payloadToParam(p payload.Param) *Param {
	return &Param{
		name:      p.Name,
		value:     p.Value,
		valueType: p.Type,
		exists:    true,
	}
}

// Converts a list params to a list of param payloads.
func paramsToPayload(ps []*Param) (params []payload.Param) {
	for _, p := range ps {
		params = append(params, paramToPayload(p))
	}
	return params
}

// Converts a list param payloads to a list of params.
func payloadToParams(ps []payload.Param) (params []*Param) {
	for _, p := range ps {
		params = append(params, payloadToParam(p))
	}
	return params
}
