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
	"strings"

	"github.com/kusanagi/kusanagi-sdk-go/payload"
	"github.com/kusanagi/kusanagi-sdk-go/traverse"
)

type ErrSchemaParam struct {
	Name string
}

func (e ErrSchemaParam) Error() string {
	return fmt.Sprintf("Cannot resolve schema for parameter: %v", e.Name)
}

type ErrSchemaFile struct {
	Name string
}

func (e ErrSchemaFile) Error() string {
	return fmt.Sprintf("Cannot resolve schema for file parameter: %v", e.Name)
}

// NewAction creates a new Service action schema
func NewAction(name string, p *payload.Payload) *Action {
	var params map[string]interface{}
	var files map[string]interface{}

	if p == nil {
		// When no payload is given use an empty payload
		p = payload.New()
	} else {
		// Get action parameters from payload
		params = p.GetMap("params")
		// Get file parameters from payload
		files = p.GetMap("files")
	}

	// Create an empty map when no parameter exists
	if params == nil {
		params = make(map[string]interface{})
	}

	// Create an empty map when no files exists
	if files == nil {
		files = make(map[string]interface{})
	}

	return &Action{name: name, payload: p, params: params, files: files}
}

// Action defines an action schema
type Action struct {
	name    string
	payload *payload.Payload
	params  map[string]interface{}
	files   map[string]interface{}
}

// IsDeprecated checks if action has been deprecated
func (a Action) IsDeprecated() bool {
	return a.payload.GetBool("deprecated")
}

// IsCollection checks if action returns a collection of entities
func (a Action) IsCollection() bool {
	return a.payload.GetBool("collection")
}

// GetName gets the name of the action
func (a Action) GetName() string {
	return a.name
}

// GetEntityPath gets the path to the entity
func (a Action) GetEntityPath() string {
	return a.payload.GetString("entity_path")
}

// GetPathDelimiter gets the path to the entity
func (a Action) GetPathDelimiter() string {
	return a.payload.GetDefault("path_delimiter", "/").(string)
}

// GetPrimaryKey gets the primary key field's name
// Gets the name of the property in the entity which
// contains the primary key.
func (a Action) GetPrimaryKey() string {
	return a.payload.GetDefault("primary_key", "id").(string)
}

// ResolveEntity gets and entity from data
//
// Get the entity part, based upon the `entity-path` and `path-delimiter`
// properties in the action configuration.
func (a Action) ResolveEntity(d map[string]interface{}) (map[string]interface{}, error) {
	path := a.GetEntityPath()
	if path == "" {
		// No traversing is done when there is no path
		return d, nil
	}

	v, err := traverse.Get(d, path, a.GetPathDelimiter(), nil)
	if err != nil {
		return nil, fmt.Errorf("Can't resolve entity for action: %v", a.GetName())
	}

	entity, ok := v.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("Invalid entity type for path: %v", path)
	}
	return entity, nil
}

// HasEntityDefinition checks if an entity definition exists for the action
func (a Action) HasEntityDefinition() bool {
	return a.payload.Exists("entity")
}

// GetEntity gets the entity definition
func (a Action) GetEntity() Entity {
	if m := a.payload.GetMap("entity"); m != nil {
		p := payload.New()
		p.Data = m
		return payloadToEntity(p, nil)
	}
	return make(map[string]interface{})
}

// HasRelations checks if any relation exists for the action
func (a Action) HasRelations() bool {
	return a.payload.Exists("relations")
}

// GetRelations gets action relations
func (a Action) GetRelations() [][]string {
	relations := [][]string{}
	if v := a.payload.GetSlice("relations"); v != nil {
		for _, d := range v {
			p := payload.New()
			p.Data = d.(map[string]interface{})
			// Create the tuple with the relation type and name
			rel := []string{}
			rel = append(rel, p.GetDefault("type", "one").(string))
			rel = append(rel, p.GetString("name"))
			relations = append(relations, rel)
		}
	}
	return relations
}

// HasCall checks if a run-time call exists for a Service
func (a Action) HasCall(name, version, action string) bool {
	for _, call := range a.GetCalls() {
		if call[0] != name {
			continue
		}

		if version != "" && call[1] != version {
			continue
		}

		if action != "" && call[2] != action {
			continue
		}

		return true
	}

	return false
}

// HasCalls checks if any run-time call exists for the action
func (a Action) HasCalls() bool {
	return a.payload.Exists("calls")
}

// Get Service run-time calls.
//
// Each call items is a list containing the Service name,
// the Service version and the action name.
func (a Action) GetCalls() [][]string {
	if v := a.payload.GetSlice("calls"); len(v) > 0 {
		return formatCalls(v)
	}
	return [][]string{}
}

// HasDeferCall checks if a deferred call exists for a Service
func (a Action) HasDeferCall(name, version, action string) bool {
	for _, call := range a.GetDeferCalls() {
		if call[0] != name {
			continue
		}

		if version != "" && call[1] != version {
			continue
		}

		if action != "" && call[2] != action {
			continue
		}

		return true
	}

	return false
}

// HasDeferCalls checks if any deferred call exists for the action
func (a Action) HasDeferCalls() bool {
	return a.payload.Exists("deferred_calls")
}

// Get Service deferred calls.
//
// Each call items is a list containing the Service name,
// the Service version and the action name.
func (a Action) GetDeferCalls() [][]string {
	if v := a.payload.GetSlice("deferred_calls"); len(v) > 0 {
		return formatCalls(v)
	}
	return [][]string{}
}

// HasRemoteCall checks if a remote call exists for a Service
func (a Action) HasRemoteCall(address, name, version, action string) bool {
	for _, call := range a.GetRemoteCalls() {
		if call[0] != address {
			continue
		}

		if name != "" && call[1] != name {
			continue
		}

		if version != "" && call[2] != version {
			continue
		}

		if action != "" && call[3] != action {
			continue
		}

		return true
	}

	return false
}

// HasRemoteCalls checks if any remote call exists for the action
func (a Action) HasRemoteCalls() bool {
	return a.payload.Exists("remote_calls")
}

// Get Service remote calls.
//
// Each remote call items is a list containing the public address
// of the Gateway, the Service name, the Service version and the
// action name.
func (a Action) GetRemoteCalls() [][]string {
	if v := a.payload.GetSlice("remote_calls"); len(v) > 0 {
		return formatCalls(v)
	}
	return [][]string{}
}

// HasReturn checks if a return value is defined for the action
func (a Action) HasReturn() bool {
	return a.payload.Exists("return")
}

// GetReturnType gets the data type of the returned action value
func (a Action) GetReturnType() string {
	return a.payload.PgetDefault("return/type", "/", "").(string)
}

// GetParams gets the parameters names defined for the action
func (a Action) GetParams() []string {
	names := []string{}
	for name, _ := range a.params {
		names = append(names, name)
	}
	return names
}

// HasParam checks that a parameter schema exists
func (a Action) HasParam(name string) bool {
	_, exists := a.params[name]
	return exists
}

// GetParamSchema gets the schema for a parameter
func (a Action) GetParamSchema(name string) (*Param, error) {
	if !a.HasParam(name) {
		return nil, ErrSchemaParam{name}
	}

	data := a.params[name]
	p := payload.New()
	p.Data = data.(map[string]interface{})
	return NewParam(name, p), nil
}

// GetFiles gets the file parameter names defined for the action
func (a Action) GetFiles() []string {
	names := []string{}
	for name, _ := range a.files {
		names = append(names, name)
	}
	return names
}

// HasFile checks that a file parameter schema exists
func (a Action) HasFile(name string) bool {
	_, exists := a.files[name]
	return exists
}

// GetFileSchema gets the schema for a file parameter
func (a Action) GetFileSchema(name string) (*File, error) {
	if !a.HasFile(name) {
		return nil, ErrSchemaFile{name}
	}

	data := a.files[name]
	p := payload.New()
	p.Data = data.(map[string]interface{})
	return NewFile(name, p), nil
}

// GetHTTPSchema gets HTTP action schema
func (a Action) GetHTTPSchema() *HTTPAction {
	p := payload.New()

	// Get HTTP schema data if it exists
	if v := a.payload.GetMap("http"); v != nil {
		p.Data = v
	}
	return NewHTTPAction(p)
}

// NewHTTPAction creates a new HTTP action schema
func NewHTTPAction(p *payload.Payload) *HTTPAction {
	// When no payload is given use an empty payload
	if p == nil {
		p = payload.New()
	}

	return &HTTPAction{payload: p}
}

// HTTPAction represents the HTTP semantics of an action
type HTTPAction struct {
	payload *payload.Payload
}

// IsAccessible checks if the Gateway has access to the action
func (ha HTTPAction) IsAccessible() bool {
	return ha.payload.GetDefault("gateway", true).(bool)
}

// GetMethod gets the HTTP method for the action
func (ha HTTPAction) GetMethod() string {
	return strings.ToLower(ha.payload.GetDefault("method", "get").(string))
}

// GetPath gets the URL path for the action
func (ha HTTPAction) GetPath() string {
	return ha.payload.GetString("path")
}

// GetInput gets the default location of parameters for the action
func (ha HTTPAction) GetInput() string {
	return ha.payload.GetDefault("input", "query").(string)
}

// GetBody gets the expected MIME types of the HTTP request body
//
// Result may contain a comma separated list of MIME types.
func (ha HTTPAction) GetBody() string {
	mimes := []string{}
	data := ha.payload.GetDefault("body", []interface{}{"text/plain"})
	for _, v := range data.([]interface{}) {
		mimes = append(mimes, v.(string))
	}

	return strings.Join(mimes, ",")
}

// Cast call data to a non generic type
func formatCalls(data []interface{}) [][]string {
	calls := [][]string{}
	for _, elem := range data {
		call := []string{}
		for _, c := range elem.([]interface{}) {
			call = append(call, c.(string))
		}
		calls = append(calls, call)
	}
	return calls
}
