// Go SDK for the KUSANAGI(tm) framework (http://kusanagi.io)
// Copyright (c) 2016-2021 KUSANAGI S.L. All rights reserved.
//
// Distributed under the MIT license.
//
// For the full copyright and license information, please view the LICENSE
// file that was distributed with this source code.

package payload

import (
	"fmt"

	"github.com/kusanagi/kusanagi-sdk-go/v2/lib/semver"
)

// Mapping contains the schemas for the different services.
type Mapping map[string]map[string]Schema

// GetServices returns the name and version of all the services in the mapping.
func (m Mapping) GetServices() (services []ServiceVersion) {
	for name, versions := range m {
		for version := range versions {
			services = append(services, ServiceVersion{name, version})
		}
	}
	return services
}

// GetVersions returns the versions for a services that are available in the mappings.
//
// name: The name of the service.
func (m Mapping) GetVersions(name string) (versions []string) {
	for version := range m[name] {
		versions = append(versions, version)
	}
	return versions
}

// GetSchema returns a schema for a service.
// The version can be either a fixed version or a pattern that uses "*"
// and resolves to the higher version available that matches.
//
// name: The name of the service.
// version: The version of the service.
func (m Mapping) GetSchema(name, version string) (*Schema, error) {
	if versions, ok := m[name]; ok {
		schema, exists := versions[version]

		// When the version doesn't exist try to resolve the version pattern and get the closest
		// highest version from the ones registered in the mapping for the current service.
		if !exists {
			if resolved := semver.New(version).Resolve(m.GetVersions(name)); resolved != "" {
				schema = versions[resolved]
				exists = true
			}
		}

		// Assign the name and version and return the schema
		if exists {
			return &schema, nil
		}
	}
	return nil, fmt.Errorf(`cannot resolve schema for service: "%s" (%s)`, name, version)
}

// ServiceVersion contains the name and version of a service.
type ServiceVersion struct {
	Name    string
	Version string
}

// Schema contains the schema definitions for a service.
type Schema struct {
	Address string                  `json:"a"`
	Files   *bool                   `json:"f,omitempty"`
	HTTP    HTTPSchema              `json:"h"`
	Actions map[string]ActionSchema `json:"ac"`
}

// GetAddress returns the internal address of the host.
func (s Schema) GetAddress() string {
	return s.Address
}

// GetFiles returns the file server availability.
func (s Schema) GetFiles() bool {
	if s.Files != nil {
		return *s.Files
	}
	return false
}

type HTTPSchema struct {
	Gateway  *bool  `json:"g,omitempty"`
	BasePath string `json:"b"`
}

// GetGateway returns the service availability thought the gateway.
func (h HTTPSchema) GetGateway() bool {
	if h.Gateway != nil {
		return *h.Gateway
	}
	return true
}

// GetBasePath returns the base path for the service.
func (h HTTPSchema) GetBasePath() string {
	return h.BasePath
}

// ActionSchema contains the schema definition for an action.
type ActionSchema struct {
	Timeout       uint                   `json:"x"`
	EntityPath    string                 `json:"e"`
	PathDelimiter string                 `json:"d"`
	Collection    *bool                  `json:"c,omitempty"`
	Calls         [][]string             `json:"C,omitempty"`
	DeferredCalls [][]string             `json:"dc,omitempty"`
	RemoteCalls   [][]string             `json:"rc,omitempty"`
	Fallback      *FallbackSchema        `json:"F,omitempty"`
	Deprecated    *bool                  `json:"D,omitempty"`
	HTTP          HTTPActionSchema       `json:"h,omitempty"`
	Params        map[string]ParamSchema `json:"p,omitempty"`
	Files         map[string]FileSchema  `json:"f,omitempty"`
	Entity        *EntitySchema          `json:"R,omitempty"`
	Relations     []RelationSchema       `json:"r,omitempty"`
	Return        *ReturnSchema          `json:"rv,omitempty"`
	Tags          []string               `json:"t,omitempty"`
}

// FallbackSchema contains the schema definition for the transport fallback.
type FallbackSchema struct {
	Properties map[string]string  `json:"p,omitempty"`
	Data       []FallbackObject   `json:"d,omitempty"`
	Relations  []FallbackRelation `json:"r,omitempty"`
	Links      map[string]string  `json:"l,omitempty"`
	Errors     []FallbackError    `json:"e,omitempty"`
}

// FallbackData contains the fallback data objects
type FallbackObject map[string]FallbackValue

// FallbackValue contains the value(s) and type of a fallback object.
//
// "Value" contains the value when the type is not array,
// otherwise the values are stored in the "Items" property.
type FallbackValue struct {
	Type  string          `json:"t,omitempty"`
	Value interface{}     `json:"v,omitempty"`
	Items []FallbackValue `json:"i,omitempty"`
}

// GetType returns the value type.
func (v FallbackValue) GetType() string {
	if v.Type == "" {
		return TypeString
	}
	return v.Type
}

// IsArray checks that the value type is an array.
func (v FallbackValue) IsArray() bool {
	return v.Type == TypeArray
}

// GetValue returns the fallback value when the type is not array.
func (v FallbackValue) GetValue() (interface{}, bool) {
	if !v.IsArray() {
		return v.Value, true
	}
	return nil, false
}

// GetItems returns the fallback values when the type is array.
func (v FallbackValue) GetItems() ([]FallbackValue, bool) {
	if v.IsArray() {
		return v.Items, true
	}
	return nil, false
}

// FallbackRelation contains the fallback relations.
type FallbackRelation []interface{}

// GetPrimaryKey returns the value for the primary key.
func (r FallbackRelation) GetPrimaryKey() (string, bool) {
	if len(r) == 0 {
		return "", false

	}

	v, ok := r[0].(string)
	return v, ok
}

// GetRemoteService returns the name of the remote service.
func (r FallbackRelation) GetRemoteService() (string, bool) {
	if len(r) < 2 {
		return "", false

	}

	v, ok := r[1].(string)
	return v, ok
}

// IsOneToMany checks if the relation is a "one-to-many" relation.
func (r FallbackRelation) IsOneToMany() bool {
	if len(r) < 3 {
		return false
	}

	_, IsOneToOne := r[2].(string)
	return !IsOneToOne
}

// GetForeignKey returns the foreign key value for a "one-to-one" relation.
func (r FallbackRelation) GetForeignKey() (string, bool) {
	if r.IsOneToMany() {
		return "", false

	}

	v, ok := r[2].(string)
	return v, ok
}

// GetForeignKeys returns the foreign key values for a "one-to-many" relation.
func (r FallbackRelation) GetForeignKeys() ([]string, bool) {
	if !r.IsOneToMany() {
		return nil, false

	}

	// Cast the slice to a slice of strings
	if items, _ := r[2].([]interface{}); items != nil {
		keys := []string{}
		for _, v := range items {
			if fk, ok := v.(string); ok {
				keys = append(keys, fk)
			} else {
				// If any of the keys is not a string fail
				return nil, false
			}
		}
		return keys, true
	}
	return nil, false
}

// FallbackError contains a fallback error.
type FallbackError []interface{}

// GetMessage returns the error message.
func (e FallbackError) GetMessage() (string, bool) {
	if len(e) == 0 {
		return "", false
	}

	v, ok := e[0].(string)
	return v, ok
}

// GetCode returns the error code.
func (e FallbackError) GetCode() (int, bool) {
	if len(e) < 2 {
		return 0, false
	}

	// JSON decodes the numbers as float
	v, ok := e[1].(float64)
	return int(v), ok
}

// GetStatus returns the status message of the errror.
func (e FallbackError) GetStatus() (string, bool) {
	if len(e) < 3 {
		return "", false
	}

	v, ok := e[2].(string)
	return v, ok
}

// HTTPActionSchema contains the HTTP schema definition for the action.
type HTTPActionSchema struct {
	Gateway *bool    `json:"g,omitempty"`
	Path    string   `json:"p,omitempty"`
	Method  string   `json:"m,omitempty"`
	Input   string   `json:"i,omitempty"`
	Body    []string `json:"b,omitempty"`
}

// ParamSchema contains the schema definition of action parameters.
type ParamSchema struct {
	Name         string                 `json:"n"`
	Type         string                 `json:"t,omitempty"`
	Format       string                 `json:"f,omitempty"`
	ArrayFormat  string                 `json:"af,omitempty"`
	Pattern      string                 `json:"p,omitempty"`
	AllowEmpty   bool                   `json:"e,omitempty"`
	DefaultValue interface{}            `json:"d,omitempty"`
	Required     bool                   `json:"r,omitempty"`
	Items        map[string]interface{} `json:"i,omitempty"`
	Max          *int                   `json:"mx,omitempty"`
	ExclusiveMax bool                   `json:"ex,omitempty"`
	Min          *int                   `json:"mn,omitempty"`
	ExclusiveMin bool                   `json:"en,omitempty"`
	MaxItems     int                    `json:"xi,omitempty"`
	MinItems     *int                   `json:"ni,omitempty"`
	UniqueItems  bool                   `json:"ui,omitempty"`
	Enum         []interface{}          `json:"em,omitempty"`
	MultipleOf   int                    `json:"mo,omitempty"`
	HTTP         HTTPParamSchema        `json:"h,omitempty"`
}

// HTTPParamSchema contains the HTTP schema definition for a parameter.
type HTTPParamSchema struct {
	Gateway *bool  `json:"g,omitempty"`
	Input   string `json:"i,omitempty"`
	Param   string `json:"p,omitempty"`
}

// FileSchema contains the schema definition of action file.
type FileSchema struct {
	Mime         string         `json:"m,omitempty"`
	Required     bool           `json:"r,omitempty"`
	Max          uint           `json:"mx,omitempty"`
	ExclusiveMax bool           `json:"ex,omitempty"`
	Min          uint           `json:"mn,omitempty"`
	ExclusiveMin bool           `json:"en,omitempty"`
	HTTP         HTTPFileSchema `json:"h,omitempty"`
}

// HTTPFileSchema contains the HTTP schema definition for a file.
type HTTPFileSchema struct {
	Gateway *bool  `json:"g,omitempty"`
	Param   string `json:"p,omitempty"`
}

// EntitySchema contains the schema for an entity definition.
type EntitySchema struct {
	Field      []FieldSchema       `json:"f,omitempty"`
	Fields     []ObjectFieldSchema `json:"F,omitempty"`
	Name       string              `json:"n,omitempty"`
	Validate   bool                `json:"V,omitempty"`
	Primarykey string              `json:"k,omitempty"`
}

// FieldSchema contains the schema for an entity field.
type FieldSchema struct {
	Name     string `json:"n"`
	Type     string `json:"t"`
	Optional bool   `json:"o,omitempty"`
}

// ObjectFieldSchema contains the schema for an entity object field.
type ObjectFieldSchema struct {
	Name     string              `json:"n"`
	Field    []FieldSchema       `json:"f,omitempty"`
	Fields   []ObjectFieldSchema `json:"F,omitempty"`
	Optional bool                `json:"o,omitempty"`
}

// RelationSchema contains the schema for a relation.
type RelationSchema struct {
	Name string `json:"n"`
	Type string `json:"t,omitempty"`
}

// ReturnSchema contains the schema for the return value.
type ReturnSchema struct {
	Type       string `json:"t"`
	AllowEmpty bool   `json:"e,omitempty"`
}
