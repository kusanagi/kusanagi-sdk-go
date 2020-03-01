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

// NewParam creates a new param schema
func NewParam(name string, p *payload.Payload) *Param {
	// When no payload is given use an empty payload
	if p == nil {
		p = payload.New()
	}

	return &Param{name: name, payload: p}
}

// Param represents a parameter schema
type Param struct {
	name    string
	payload *payload.Payload
}

// GetName gets the name of the parameter
func (p Param) GetName() string {
	return p.name
}

// GetType gets the parameter value type
func (p Param) GetType() string {
	return p.payload.GetDefault("type", "string").(string)
}

// GetFormat gets the parameter value format
func (p Param) GetFormat() string {
	return p.payload.GetString("format")
}

// GetArrayFormat gets format for the parameter if the type property is set to "array"
//
// Formats:
//   - "csv" for comma separated values (default)
//   - "ssv" for space separated values
//   - "tsv" for tab separated values
//   - "pipes" for pipe separated values
//   - "multi" for multiple parameter instances instead of multiple values for a single instance
func (p Param) GetArrayFormat() string {
	return p.payload.GetDefault("array_format", "csv").(string)
}

// GetPattern gets the ECMA 262 compliant regular expression to validate the parameter
func (p Param) GetPattern() string {
	return p.payload.GetString("pattern")
}

// AllowEmpty checks if the parameter allows an empty value
func (p Param) AllowEmpty() bool {
	return p.payload.GetBool("allow_empty")
}

// HasDefault checks if the parameter has a default value defined
func (p Param) HasDefault() bool {
	return p.payload.Exists("default_value")
}

// GetDefaultValue gets the default value for a parameter
func (p Param) GetDefaultValue() string {
	return p.payload.GetString("default_value")
}

// IsRequired checks if the parameter is required
func (p Param) IsRequired() bool {
	return p.payload.GetBool("required")
}

// GetItems gets a JSON schema with item object definitions.
//
// An empty map is returned when parameter is not an "array",
// otherwise the result contains a map with a JSON schema definition.
func (p Param) GetItems() map[string]interface{} {
	if p.GetType() != "array" {
		return make(map[string]interface{})
	}

	if v := p.payload.GetMap("items"); v != nil {
		return v
	}
	return make(map[string]interface{})
}

// GetMax gets the maximum value for parameter
func (p Param) GetMax() uint {
	var max uint

	v := p.payload.GetDefault("max", uint(MaxInt))
	switch v.(type) {
	case int:
		max = uint(v.(int))
	case uint64:
		max = uint(v.(uint64))
	case uint32:
		max = uint(v.(uint32))
	default:
		max = v.(uint)
	}
	return max
}

// IsExclusiveMax checks if maximum value is inclusive
func (p Param) IsExclusiveMax() bool {
	if !p.payload.Exists("max") {
		return false
	}
	return p.payload.GetBool("exclusive_max")
}

// GetMin gets the minimum value allowed
func (p Param) GetMin() int {
	var min int

	v := p.payload.GetDefault("min", MinInt)
	switch v.(type) {
	case int64:
		min = int(v.(int64))
	case int32:
		min = int(v.(int32))
	default:
		min = v.(int)
	}
	return min
}

// IsExclusiveMin checks if minimum value is inclusive
func (p Param) IsExclusiveMin() bool {
	if !p.payload.Exists("min") {
		return false
	}
	return p.payload.GetBool("exclusive_min")
}

// GetMaxLength gets the maximum length defined for the parameter
func (p Param) GetMaxLength() int {
	var max int

	v := p.payload.GetDefault("max_length", -1)
	switch v.(type) {
	case int64:
		max = int(v.(int64))
	case int32:
		max = int(v.(int32))
	default:
		max = v.(int)
	}
	return max
}

// GetMinLength gets the minimum length defined for the parameter
func (p Param) GetMinLength() int {
	var min int

	v := p.payload.GetDefault("min_length", -1)
	switch v.(type) {
	case int64:
		min = int(v.(int64))
	case int32:
		min = int(v.(int32))
	default:
		min = v.(int)
	}
	return min
}

// GetMaxItems gets the maximum number of items allowed for the parameter.
// Result is -1 when type is not "array" or values is not defined.
func (p Param) GetMaxItems() int {
	if p.GetType() != "array" {
		return -1
	}
	return p.payload.GetDefault("max_items", -1).(int)
}

// GetMinItems gets the minimum number of items allowed for the parameter.
// Result is -1 when type is not "array" or values is not defined.
func (p Param) GetMinItems() int {
	if p.GetType() != "array" {
		return -1
	}
	return p.payload.GetDefault("min_items", -1).(int)
}

// HasUniqueItems checks if param must contain a set of unique items.
func (p Param) HasUniqueItems() bool {
	return p.payload.GetBool("unique_items")
}

// GetEnum gets the set of unique values that the parameter allows
func (p Param) GetEnum() []interface{} {
	if v := p.payload.GetSlice("enum"); len(v) > 0 {
		return v
	}
	return make([]interface{}, 0)
}

// GetMultipleOf gets the value that parameter must be divisible by.
// Result is -1 when type is not "array" or values is not defined.
func (p Param) GetMultipleOf() int {
	return p.payload.GetDefault("multiple_of", -1).(int)
}

// GetHTTPSchema gets HTTP parameter schema
func (p Param) GetHTTPSchema() *HTTPParam {
	// Get HTTP schema data if it exists
	v := payload.New()
	if m := p.payload.GetMap("http"); m != nil {
		v.Data = m
	}
	return NewHTTPParam(p.GetName(), v)
}

// NewHTTPParam creates a new HTTP param schema
func NewHTTPParam(name string, p *payload.Payload) *HTTPParam {
	// When no payload is given use an empty payload
	if p == nil {
		p = payload.New()
	}
	return &HTTPParam{name: name, payload: p}
}

// HTTPParam represents the HTTP semantics of a parameter
type HTTPParam struct {
	name    string
	payload *payload.Payload
}

// IsAccessible checks if the Gateway has access to the parameter
func (p HTTPParam) IsAccessible() bool {
	return p.payload.GetDefault("gateway", true).(bool)
}

// GetInput gets the input location for the HTTP parameter
func (p HTTPParam) GetInput() string {
	return p.payload.GetDefault("input", p.name).(string)
}

// GetParam gets name for the HTTP parameter
func (p HTTPParam) GetParam() string {
	return p.payload.GetDefault("param", p.name).(string)
}

func getNumericValidationSchema(p *Param) map[string]interface{} {
	s := map[string]interface{}{}
	s["minimum"] = p.GetMin()
	s["exclusiveMinimum"] = p.IsExclusiveMin()
	s["maximum"] = p.GetMax()
	s["exclusiveMaximum"] = p.IsExclusiveMax()

	if v := p.GetMultipleOf(); v > 0 {
		s["multipleOf"] = v
	}

	if v := p.GetFormat(); v != "" {
		s["format"] = v
	}

	return s
}

func getStringValidationSchema(p *Param) map[string]interface{} {
	s := map[string]interface{}{}

	if v := p.GetMinLength(); v != -1 {
		s["minLength"] = v
	}

	if v := p.GetMaxLength(); v != -1 {
		s["maxLength"] = v
	}

	if v := p.GetPattern(); v != "" {
		s["pattern"] = v
	}

	if v := p.GetFormat(); v != "" {
		s["format"] = v
	}

	return s
}

func getArrayValidationSchema(p *Param) map[string]interface{} {
	s := map[string]interface{}{}

	if v := p.GetMinItems(); v >= 0 {
		s["minItems"] = v
	}

	if v := p.GetMaxItems(); v >= 0 {
		s["maxItems"] = v
	}

	if p.HasUniqueItems() {
		s["uniqueItems"] = true
	}

	return s
}

// ValidateParamValue validates a value for a parameter schema
func ValidateParamValue(p *Param, value interface{}) error {
	// Some types don't require validation
	pType := p.GetType()
	if pType == "null" || pType == "boolean" || pType == "object" {
		return nil
	}

	var s map[string]interface{}
	switch pType {
	case "integer", "float":
		s = getNumericValidationSchema(p)
	case "string":
		s = getStringValidationSchema(p)
	case "array":
		s = getArrayValidationSchema(p)
	default:
		// Don't validate when data type is not valid
		return fmt.Errorf("Unknown parameter type: %v", pType)
	}

	// KUSANAGI don't use number, intead uses integer and float.
	// JSONSCHEMA validation is not familiar with "float".
	if pType == "float" {
		pType = "number"
	}

	// Add parameter value type to schema
	s["type"] = pType

	// Add enum constraints to schema
	// WARNING: For arrays each item in enum MUST be an array,
	//          so enum value will be an array of arrays.
	if v := p.GetEnum(); len(v) > 0 {
		s["enum"] = v
	}

	if err := Validate(s, value); err != nil {
		return fmt.Errorf("parameter \"%s\" validation failed: %v", p.GetName(), err)
	}
	return nil
}

// IsEmptyParamValue checks if a parameter value is empty.
// Only "string", "array" and "object" types are checked.
func IsEmptyParamValue(v interface{}) bool {
	if v == nil {
		return true
	}

	switch v.(type) {
	case string:
		return len(v.(string)) == 0
	case []interface{}:
		return len(v.([]interface{})) == 0
	case map[string]interface{}:
		return len(v.(map[string]interface{})) == 0
		// TODO: Log unknown types
	}
	return false
}
