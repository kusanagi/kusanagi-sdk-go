// Go SDK for the KUSANAGI(tm) framework (http://kusanagi.io)
// Copyright (c) 2016-2022 KUSANAGI S.L. All rights reserved.
//
// Distributed under the MIT license.
//
// For the full copyright and license information, please view the LICENSE
// file that was distributed with this source code.

package kusanagi

import (
	"fmt"
	"strings"

	"github.com/kusanagi/kusanagi-sdk-go/v3/lib/datatypes"
	"github.com/kusanagi/kusanagi-sdk-go/v3/lib/payload"
	"github.com/kusanagi/kusanagi-sdk-go/v3/lib/semver"
)

// ExecutionTimeout defines the number of milliseconds to wait by default when an action is executed.
const ExecutionTimeout = 30000

// ServiceSchema contains the schema definition for a service of a specific version.
type ServiceSchema struct {
	name    string
	version string
	payload payload.Schema
}

// GetName returns the service name.
func (s ServiceSchema) GetName() string {
	return s.name
}

// GetVersion returns the service version.
func (s ServiceSchema) GetVersion() string {
	return s.version
}

// GetAddress returns the network address of the service.
func (s ServiceSchema) GetAddress() string {
	return s.payload.Address
}

// HasFileServer checks that the service has a files server enabled.
func (s ServiceSchema) HasFileServer() bool {
	return s.payload.Files != nil && *s.payload.Files
}

// GetActionNames returns the names of the service actions.
func (s ServiceSchema) GetActionNames() (actions []string) {
	for name := range s.payload.Actions {
		actions = append(actions, name)
	}
	return actions
}

// HasAction checks that an action exists in the schema.
//
// name: The action name.
func (s ServiceSchema) HasAction(name string) bool {
	if s.payload.Actions != nil {
		_, exists := s.payload.Actions[name]
		return exists
	}
	return false
}

// GetActionSchema returns the schema for an action.
//
// name: The action name.
func (s ServiceSchema) GetActionSchema(name string) (*ActionSchema, error) {
	if schema, ok := s.payload.Actions[name]; ok {
		return &ActionSchema{name, schema}, nil
	}
	return nil, fmt.Errorf(`Cannot resolve schema for "%s" (%s) action: %s`, s.GetName(), s.GetVersion(), name)
}

// GetHTTPSchema returns the HTTP schema.
func (s ServiceSchema) GetHTTPSchema() *HTTPServiceSchema {
	return &HTTPServiceSchema{s.payload.HTTP}
}

// HTTPServiceSchema contains the HTTP schema definition for the service.
type HTTPServiceSchema struct {
	payload payload.HTTPSchema
}

// IsAccesible checks that the gateway has access to the service.
func (s HTTPServiceSchema) IsAccesible() bool {
	if s.payload.Gateway == nil {
		return true
	}
	return *s.payload.Gateway
}

// GetBasePath returns the base HTTP path for the service.
func (s ServiceSchema) GetBasePath() string {
	return s.payload.HTTP.BasePath
}

// ActionSchema contains the schema definition for an action.
type ActionSchema struct {
	name    string
	payload payload.ActionSchema
}

// GetTimeout returns the maximum execution time defined in milliseconds for the action.
func (s ActionSchema) GetTimeout() uint {
	if s.payload.Timeout == 0 {
		return ExecutionTimeout
	}
	return s.payload.Timeout
}

// IsDeprecated checks if the action has been deprecated.
func (s ActionSchema) IsDeprecated() bool {
	if s.payload.Deprecated == nil {
		return false
	}
	return *s.payload.Deprecated
}

// IsCollection checks if the action returns a collection of entities.
func (s ActionSchema) IsCollection() bool {
	if s.payload.Collection == nil {
		return false
	}
	return *s.payload.Collection
}

// GetName returns the action name.
func (s ActionSchema) GetName() string {
	return s.name
}

// GetEntityPath returns the path to the entity.
func (s ActionSchema) GetEntityPath() string {
	return s.payload.EntityPath
}

// GetPathDelimiter returns the delimiter to use for the entity path.
func (s ActionSchema) GetPathDelimiter() string {
	if s.payload.PathDelimiter == "" {
		return "/"
	}
	return s.payload.PathDelimiter
}

// ResolveEntity returns the entity extracted from the data.
//
// Get the entity part, based upon the `entity-path` and `path-delimiter`
// properties in the action configuration.
//
// data: The object to get entity from.
func (s ActionSchema) ResolveEntity(data map[string]interface{}) (map[string]interface{}, error) {
	// The data is traversed only when there is a path, otherwise data is returned as is
	if p := s.GetEntityPath(); p != "" {
		delimiter := s.GetPathDelimiter()
		path := strings.Split(p, delimiter)
		last := len(path) - 1
		for i, name := range path {
			// Get the value for the current path element
			v := data[name]
			data, _ = v.(map[string]interface{})
			// The value of data must be a map, unless the last path element
			// is being traversed, in which case data can be nil.
			if data == nil && i != last {
				return nil, fmt.Errorf("Cannot resolve entity for action: %s", s.GetName())
			}
		}
	}
	return data, nil
}

// HasEntity checks if an entity definition exists for the action.
func (s ActionSchema) HasEntity() bool {
	return s.payload.Entity != nil
}

// GetEntity returns the entity definition.
func (s ActionSchema) GetEntity() *Entity {
	entity := &Entity{Primarykey: "id"}
	if schema := s.payload.Entity; schema != nil {
		entity.Validate = schema.Validate
		entity.Field = copyFields(schema.Field)
		entity.Fields = copyObjectFields(schema.Fields)
	}
	return entity
}

// HasRelations checks if any relations exists for the action.
func (s ActionSchema) HasRelations() bool {
	return len(s.payload.Relations) > 0
}

// GetRelations return the relations.
//
// Each item is an array contains the relation type and the service name.
func (s ActionSchema) GetRelations() (relations [][]string) {
	for _, relation := range s.payload.Relations {
		// By default the relation type is "one"
		t := relation.Type
		if t == "" {
			t = "one"
		}

		relations = append(relations, []string{t, relation.Name})
	}
	return relations
}

// GetCalls returns the run-time service calls.
//
// Each call item is a list containing the service name, the service version and the action name.
func (s ActionSchema) GetCalls() (calls [][]string) {
	for _, c := range s.payload.Calls {
		call := make([]string, len(c))
		copy(c, call)
		calls = append(calls, call)
	}
	return calls
}

// HasCall checks if a run-time call exists for a service.
//
// name: The service name.
// version: Optional service version.
// action: Optional action name.
func (s ActionSchema) HasCall(name, version, action string) bool {
	for _, call := range s.payload.Calls {
		if len(call) != 3 {
			continue
		}

		if call[0] != "*" && call[0] != name {
			continue
		}

		if version != "" && call[1] != "*" && call[1] != version && !semver.New(version).Match(call[1]) {
			continue
		}

		if action != "" && call[2] != "*" && call[2] != action {
			continue
		}

		return true
	}
	return false
}

// HasCalls checks if any run-time call exists for the action.
func (s ActionSchema) HasCalls() bool {
	return len(s.payload.Calls) > 0
}

// GetDeferCalls returns the deferred service calls.
//
// Each call item is a list containing the service name, the service version and the action name.
func (s ActionSchema) GetDeferCalls() (calls [][]string) {
	for _, c := range s.payload.DeferredCalls {
		call := make([]string, len(c))
		copy(c, call)
		calls = append(calls, call)
	}
	return calls
}

// HasDeferCall checks if a deferred call exists for a service.
//
// name: The service name.
// version: Optional service version.
// action: Optional action name.
func (s ActionSchema) HasDeferCall(name, version, action string) bool {
	for _, call := range s.payload.DeferredCalls {
		if len(call) != 3 {
			continue
		}

		if call[0] != "*" && call[0] != name {
			continue
		}

		if version != "" && call[1] != "*" && call[1] != version && !semver.New(version).Match(call[1]) {
			continue
		}

		if action != "" && call[2] != "*" && call[2] != action {
			continue
		}

		return true
	}
	return false
}

// HasDeferCalls checks if any deferred call exists for the action.
func (s ActionSchema) HasDeferCalls() bool {
	return len(s.payload.DeferredCalls) > 0
}

// GetRemoteCalls returns the remote service calls.
//
// Each call item is a list containing the public address of the gateway,
// the service name, the service version and the action name.
func (s ActionSchema) GetRemoteCalls() (calls [][]string) {
	for _, c := range s.payload.RemoteCalls {
		call := make([]string, len(c))
		copy(c, call)
		calls = append(calls, call)
	}
	return calls
}

// HasRemoteCall checks if a remote call exists for a service.
//
// address: Gateway address.
// name: Optional service name.
// version: Optional service version.
// action: Optional action name.
func (s ActionSchema) HasRemoteCall(address, name, version, action string) bool {
	for _, call := range s.payload.RemoteCalls {
		if len(call) != 4 {
			continue
		}

		if call[0] != "*" && call[0] != address {
			continue
		}

		if name != "" && call[1] != "*" && call[1] != name {
			continue
		}

		if version != "" && call[2] != "*" && call[2] != version && !semver.New(version).Match(call[2]) {
			continue
		}

		if action != "" && call[3] != "*" && call[3] != action {
			continue
		}

		return true
	}
	return false
}

// HasRemoteCalls checks if any remote call exists for the action.
func (s ActionSchema) HasRemoteCalls() bool {
	return len(s.payload.RemoteCalls) > 0
}

// HasReturn checks if a return value is defined for the action.
func (s ActionSchema) HasReturn() bool {
	return s.payload.Return != nil
}

// GetReturnType returns the data type of the returned action value.
func (s ActionSchema) GetReturnType() (string, error) {
	if s.payload.Return == nil || s.payload.Return.Type == "" {
		return "", fmt.Errorf("Return value not defined for action: %s", s.GetName())
	}
	return s.payload.Return.Type, nil
}

// GetParams returns the parameter names defined for the action.
func (s ActionSchema) GetParams() (params []string) {
	for name := range s.payload.Params {
		params = append(params, name)
	}
	return params
}

// HasParam checks that a schema for a parameter exists.
//
// name: The parameter name.
func (s ActionSchema) HasParam(name string) bool {
	_, exists := s.payload.Params[name]
	return exists
}

// GetParamSchema returns the schema for a parameter.
//
// name: The parameter name.
func (s ActionSchema) GetParamSchema(name string) (*ParamSchema, error) {
	schema, ok := s.payload.Params[name]
	if !ok {
		return nil, fmt.Errorf(`Cannot resolve schema for parameter: "%s"`, name)
	}
	return &ParamSchema{schema}, nil
}

// GetFiles returns the file parameter names defined for the action.
func (s ActionSchema) GetFiles() (files []string) {
	for name := range s.payload.Files {
		files = append(files, name)
	}
	return files
}

// HasFile checks that a schema for a file parameter exists.
//
// name: The file parameter name.
func (s ActionSchema) HasFile(name string) bool {
	_, exists := s.payload.Files[name]
	return exists
}

// GetFileSchema returns the schema for a file parameter.
//
// name: The file parameter name.
func (s ActionSchema) GetFileSchema(name string) (*FileSchema, error) {
	schema, ok := s.payload.Files[name]
	if !ok {
		return nil, fmt.Errorf(`Cannot resolve schema for file parameter: "%s"`, name)
	}
	return &FileSchema{name, schema}, nil
}

// GetTags returns the tags defined for the action.
func (s ActionSchema) GetTags() (tags []string) {
	for _, tag := range s.payload.Tags {
		tags = append(tags, tag)
	}
	return tags
}

// HasTag checks if a tag is defined for the action.
//
// The tag name is case sensitive.
//
// name: The tag name.
func (s ActionSchema) HasTag(name string) bool {
	for _, tag := range s.payload.Tags {
		if tag == name {
			return true
		}
	}
	return false
}

// GetHTTPSchema returns the HTTP schema.
func (s ActionSchema) GetHTTPSchema() *HTTPActionSchema {
	return &HTTPActionSchema{s.payload.HTTP}
}

// HTTPActionSchema contains the HTTP schema definition for the action.
type HTTPActionSchema struct {
	payload payload.HTTPActionSchema
}

// IsAccesible checks that the action is accesible by HTTP requests via the gateway.
func (s HTTPActionSchema) IsAccesible() bool {
	if s.payload.Gateway == nil {
		return true
	}
	return *s.payload.Gateway
}

// GetPath returns the path to resolve to the action.
func (s HTTPActionSchema) GetPath() string {
	if s.payload.Path == "" {
		return ""
	}
	return s.payload.Path
}

// GetMethod returns the HTTP method expected for the request to the gateway.
func (s HTTPActionSchema) GetMethod() string {
	if s.payload.Method == "" {
		return "GET"
	}
	return s.payload.Method
}

// GetInput returns the default HTTP parameter location.
func (s HTTPActionSchema) GetInput() string {
	if s.payload.Method == "" {
		return "query"
	}
	return s.payload.Method
}

// GetBody returns the expected MIME type of the HTTP request body
// for methods other than "get", "options" and "head".
func (s HTTPActionSchema) GetBody() string {
	if len(s.payload.Body) == 0 {
		return "text/plain"
	}
	return strings.Join(s.payload.Body, ",")
}

func copyFields(schemas []payload.FieldSchema) (fields []Field) {
	for _, schema := range schemas {
		fields = append(fields, Field{
			Name:     schema.Name,
			Type:     schema.Type,
			Optional: schema.Optional,
		})
	}
	return fields
}

func copyObjectFields(schemas []payload.ObjectFieldSchema) (fields []ObjectField) {
	for _, schema := range schemas {
		fields = append(fields, ObjectField{
			Name:     schema.Name,
			Optional: schema.Optional,
			Field:    copyFields(schema.Field),
			Fields:   copyObjectFields(schema.Fields),
		})
	}
	return fields
}

// Entity definition.
type Entity struct {
	Field      []Field
	Fields     []ObjectField
	Name       string
	Validate   bool
	Primarykey string
}

// IsEmpty checks if the entity is empty.
func (e Entity) IsEmpty() bool {
	return len(e.Field) == 0 && len(e.Fields) == 0
}

// Field defines an entity field.
type Field struct {
	Name     string
	Type     string
	Optional bool
}

// ObjectField defines an entity object field.
type ObjectField struct {
	Name     string
	Field    []Field
	Fields   []ObjectField
	Optional bool
}

// ArrayFormatCSV defines the array parameter format for comma separated values.
const ArrayFormatCSV = "csv"

// ArrayFormatSSV defines the array parameter format for space separated values.
const ArrayFormatSSV = "ssv"

// ArrayFormatTSV defines the array parameter format for tab separated values.
const ArrayFormatTSV = "tsv"

// ArrayFormatPipe defines the array parameter format for pipe ("|") separated values.
const ArrayFormatPipe = "pipe"

// ArrayFormatMulti defines the array format for multiple parameter arguments instead
// of a single string argument containing all the values.
const ArrayFormatMulti = "multi"

// ParamSchema contains the schema definition of action parameters.
type ParamSchema struct {
	payload payload.ParamSchema
}

// GetName returns the parameter name.
func (s ParamSchema) GetName() string {
	return s.payload.Name
}

// GetType returns the parameter value type.
func (s ParamSchema) GetType() string {
	if s.payload.Type == "" {
		return datatypes.String
	}
	return s.payload.Type
}

// GetFormat returns the parameter value format.
func (s ParamSchema) GetFormat() string {
	return s.payload.Format
}

// GetArrayFormat returns the format for the parameter if the type property is set to "array".
//
// Formats:
//   - "csv" for comma separated values (default)
//   - "ssv" for space separated values
//   - "tsv" for tab separated values
//   - "pipes" for pipe separated values
//   - "multi" for multiple parameter instances instead of multiple values for a single instance.
func (s ParamSchema) GetArrayFormat() string {
	if s.payload.ArrayFormat == "" {
		return ArrayFormatCSV
	}
	return s.payload.ArrayFormat
}

// GetPattern returns the ECMA 262 compliant regular expression to validate the parameter.
func (s ParamSchema) GetPattern() string {
	return s.payload.Pattern
}

// AllowEmpty checks if the parameter allows an empty value.
func (s ParamSchema) AllowEmpty() bool {
	return s.payload.AllowEmpty
}

// HasDefaultValue checks if parameter has a default value defined.
func (s ParamSchema) HasDefaultValue() bool {
	return s.payload.DefaultValue != nil
}

// GetDefaultValue returns the default value for parameter.
func (s ParamSchema) GetDefaultValue() interface{} {
	return s.payload.DefaultValue
}

// IsRequired checks if the parameter is required.
func (s ParamSchema) IsRequired() bool {
	return s.payload.Required
}

// GetItems returns the JSON schema with items object definition.
func (s ParamSchema) GetItems() map[string]interface{} {
	if s.payload.Type != datatypes.Array {
		return make(map[string]interface{})
	}
	return s.payload.Items
}

// GetMax returns the maximum value for parameter.
func (s ParamSchema) GetMax() int {
	if s.payload.Max == nil {
		return datatypes.MaxInt
	}
	return *s.payload.Max
}

// IsExclusiveMax chechs that the maximum value is inclusive.
func (s ParamSchema) IsExclusiveMax() bool {
	return s.payload.ExclusiveMax
}

// GetMin returns the minimum value for parameter.
func (s ParamSchema) GetMin() int {
	if s.payload.Min == nil {
		return datatypes.MinInt
	}
	return *s.payload.Min
}

// IsExclusiveMin checks that minimum value is inclusive.
func (s ParamSchema) IsExclusiveMin() bool {
	return s.payload.ExclusiveMin
}

// GetMaxItems returns the maximum number of items allowed for the parameter.
func (s ParamSchema) GetMaxItems() int {
	if s.payload.MaxItems == 0 || s.payload.Type != datatypes.Array {
		return -1
	}
	return int(s.payload.MaxItems)
}

// GetMinItems returns the minimum number of items allowed for the parameter.
func (s ParamSchema) GetMinItems() int {
	if s.payload.MinItems == nil || s.payload.Type != datatypes.Array {
		return -1
	}
	return int(*s.payload.MinItems)
}

// HasUniqueItems checks that the param must contain a set of unique items.
func (s ParamSchema) HasUniqueItems() bool {
	return s.payload.UniqueItems
}

// GetEnum returns the set of unique values that parameter allows.
func (s ParamSchema) GetEnum() []interface{} {
	return s.payload.Enum
}

// GetMultipleOf returns the value that parameter must be divisible by.
func (s ParamSchema) GetMultipleOf() int {
	if s.payload.MultipleOf == 0 {
		return -1
	}
	return s.payload.MultipleOf
}

// GetHTTPSchema returns the HTTP schema.
func (s ParamSchema) GetHTTPSchema() *HTTPParamSchema {
	return &HTTPParamSchema{s.payload.HTTP}
}

// HTTPParamSchema contains the HTTP schema definition for a parameter.
type HTTPParamSchema struct {
	payload payload.HTTPParamSchema
}

// IsAccesible checks that the parameter is writable by HTTP request via the gateway.
func (s HTTPParamSchema) IsAccesible() bool {
	if s.payload.Gateway == nil {
		return true
	}
	return *s.payload.Gateway
}

// GetInput returns the location of the parameter in the HTTP request.
func (s HTTPParamSchema) GetInput() string {
	if s.payload.Input == "" {
		return "query"
	}
	return s.payload.Input
}

// GetParam returns the name of the parameter in the HTTP request.
//
// This name can be different than the parameter name, and it is used
// in the HTTP request to submit the parameter value.
func (s HTTPParamSchema) GetParam() string {
	return s.payload.Param
}

// FileSchema contains the schema definition of action file.
type FileSchema struct {
	name    string
	payload payload.FileSchema
}

// GetName returns the file parameter name.
func (s FileSchema) GetName() string {
	return s.name
}

// GetMime returns the mime type.
func (s FileSchema) GetMime() string {
	return s.payload.Mime
}

// IsRequired checks if the file parameter is required.
func (s FileSchema) IsRequired() bool {
	return s.payload.Required
}

// GetMax returns the maximum file size allowed for the parameter.
func (s FileSchema) GetMax() uint {
	return s.payload.Max
}

// IsExclusiveMax checks if the maximum size is inclusive.
func (s FileSchema) IsExclusiveMax() bool {
	return s.payload.ExclusiveMax
}

// GetMin returns the minimum file size allowed for the parameter.
func (s FileSchema) GetMin() uint {
	return s.payload.Min
}

// IsExclusiveMin checks if the minimum size is inclusive.
func (s FileSchema) IsExclusiveMin() bool {
	return s.payload.ExclusiveMin
}

// GetHTTPSchema returns the HTTP schema.
func (s FileSchema) GetHTTPSchema() *HTTPFileSchema {
	return &HTTPFileSchema{s.payload.HTTP}
}

// HTTPFileSchema contains the HTTP schema definition for a file.
type HTTPFileSchema struct {
	payload payload.HTTPFileSchema
}

// IsAccesible checks that the file is writable by HTTP request via the gateway.
func (s HTTPFileSchema) IsAccesible() bool {
	if s.payload.Gateway == nil {
		return true
	}
	return *s.payload.Gateway
}

// GetParam returns the name of the file parameter in the HTTP request.
//
// This name can be different than the file name, and it is used
// in the HTTP request to submit the file contents.
func (s HTTPFileSchema) GetParam() string {
	return s.payload.Param
}
