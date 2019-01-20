// Go SDK for the KUSANAGI(tm) framework (http://kusanagi.io)
// Copyright (c) 2016-2019 KUSANAGI S.L. All rights reserved.
//
// Distributed under the MIT license.
//
// For the full copyright and license information, please view the LICENSE
// file that was distributed with this source code.

package sdk

import (
	"reflect"

	"github.com/kusanagi/kusanagi-sdk-go/payload"
)

// Defines the list of valid types for a parameter.
var paramTypes = []string{
	"null",
	"boolean",
	"integer",
	"float",
	"string",
	"array",
	"object",
}

// Checks if a string is a valid type for a parameter.
func isValidParamType(paramType string) bool {
	for _, t := range paramTypes {
		if t == paramType {
			return true
		}
	}
	return false
}

// Resolves the type of a value to a KUSANAGI type name.
func resolveType(value interface{}) string {
	if value == nil {
		return "null"
	} else if _, ok := value.([]byte); ok {
		return "binary"
	}

	// By default data type is unknown. Unsupported types are treated as unknown.
	paramType := ""

	kind := reflect.ValueOf(value).Kind()
	if kind == reflect.String {
		paramType = "string"
	} else if kind == reflect.Bool {
		paramType = "boolean"
	} else if reflect.Int <= kind && kind <= reflect.Int64 {
		paramType = "integer"
	} else if kind == reflect.Float32 || kind == reflect.Float64 {
		paramType = "float"
	} else if kind == reflect.Slice {
		paramType = "array"
	} else if kind == reflect.Map {
		paramType = "object"
	}
	return paramType
}

// Creates a new parameter.
//
// The type is resolved from the value when the parameter type is empty.
//
// String is used as default parameter type when the type can't be resolved from the
// value or when the name of the type is invalid. In any of these cases the value is
// changed to an empty string.
// TODO: What if the type doesn't match the value type ?
func newParam(name string, value interface{}, paramType string, exists bool) *Param {
	// When there is no data type guess the type from the value
	if paramType == "" {
		if paramType = resolveType(value); paramType == "" {
			// When the data type is unknown use a string value
			paramType = "string"
			value = ""
		}
	} else if !isValidParamType(paramType) {
		paramType = "string"
		value = ""
	}

	return &Param{
		name:      name,
		value:     value,
		paramType: paramType,
		exists:    exists,
	}
}

func newEmptyParam(name string) *Param {
	return newParam(name, "", "", false)
}

// Param represents a parameter received for an action in a call to a service.
type Param struct {
	name      string
	value     interface{}
	paramType string
	exists    bool
}

// GetName reads the name of the parameter.
func (p *Param) GetName() string {
	return p.name
}

// GetType reads the type of the parameter.
func (p *Param) GetType() string {
	return p.paramType
}

// GetValue reads the value of the parameter.
func (p *Param) GetValue() interface{} {
	return p.value
}

// Exists checks if the parameter exists.
func (p *Param) Exists() bool {
	return p.exists
}

// CopyWithName creates a copy of the parameter using a different name.
func (p *Param) CopyWithName(name string) *Param {
	return &Param{name: name, value: p.value, paramType: p.paramType}
}

// CopyWithValue creates a copy of the parameter using a different value.
func (p *Param) CopyWithValue(value interface{}) *Param {
	return &Param{name: p.name, value: value, paramType: p.paramType}
}

// CopyWithType creates a copy of the parameter using a different type.
func (p *Param) CopyWithType(paramType string) *Param {
	return &Param{name: p.name, value: p.value, paramType: paramType}
}

// Converts a param object to a param payload.
func paramToPayload(p *Param) *payload.Param {
	pp := payload.NewParam(p.GetName(), p.GetType())
	pp.SetValue(p.GetValue())
	return pp
}

// Converts a param payload to a param object.
func payloadToParam(pp *payload.Param) *Param {
	return newParam(pp.GetName(), pp.GetValue(), pp.GetType(), true)
}
