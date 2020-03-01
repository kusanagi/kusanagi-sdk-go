// Go SDK for the KUSANAGI(tm) framework (http://kusanagi.io)
// Copyright (c) 2016-2020 KUSANAGI S.L. All rights reserved.
//
// Distributed under the MIT license.
//
// For the full copyright and license information, please view the LICENSE
// file that was distributed with this source code.

package payload

// TODO: Rewrite all payload modules to use traverse.Path objects.

import (
	"fmt"

	"github.com/kusanagi/kusanagi-sdk-go/format"
	"github.com/kusanagi/kusanagi-sdk-go/traverse"
)

// NewEmptyServiceCall creates a new empty service call payload
func NewEmptyServiceCall() *ServiceCall {
	return &ServiceCall{Payload: New()}
}

// NewServiceCallFromMap creates a new service call from a map.
func NewServiceCallFromMap(m map[string]interface{}) *ServiceCall {
	c := NewEmptyServiceCall()
	c.Data = m
	return c
}

// ServiceCall defines a service call payload
type ServiceCall struct {
	*Payload
}

// GetName gets the name of the Service to call
func (c ServiceCall) GetName() string {
	return c.GetString("name")
}

// SetName sets the name of the Service to call
func (c *ServiceCall) SetName(value string) error {
	return c.Set("name", value)
}

// GetVersion gets the version of the Service to call
func (c ServiceCall) GetVersion() string {
	return c.GetString("version")
}

// SetVersion sets the version of the Service to call
func (c *ServiceCall) SetVersion(value string) error {
	return c.Set("version", value)
}

// GetAction gets the action name of the Service to call
func (c ServiceCall) GetAction() string {
	return c.GetString("action")
}

// SetAction sets the action name of the Service to call
func (c *ServiceCall) SetAction(value string) error {
	return c.Set("action", value)
}

// GetParams gets the parameters to pass to the action.
func (c ServiceCall) GetParams() (ps []*Param) {
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

// SetParams sets the parameters to pass to the action.
func (c *ServiceCall) SetParams(ps []*Param) error {
	var ds []Data
	for _, p := range ps {
		ds = append(ds, p.Data)
	}
	return c.Set("params", ds)
}

// GetCaller gets the name of the Service action that registers the call.
func (c ServiceCall) GetCaller() string {
	return c.GetString("caller")
}

// SetCaller sets the name of the service action that registers the call.
func (c *ServiceCall) SetCaller(name string) error {
	return c.Set("caller", name)
}

// GetDuration gets the duration of the Service call in milliseconds.
// Duration is 0 until the call is processed.
func (c ServiceCall) GetDuration() uint64 {
	return c.GetUint64("duration")
}

// SetDuration sets the duration of the Service call in milliseconds.
func (c *ServiceCall) SetDuration(value uint64) error {
	return c.Set("duration", value)
}

// GetGateway gets the public address of a Gateway for Service calls to other realms.
func (c ServiceCall) GetGateway() string {
	return c.GetString("gateway")
}

// SetGateway sets the public address of a Gateway for Service calls to other realms.
func (c *ServiceCall) SetGateway(address string) error {
	return c.Set("gateway", address)
}

// GetTimeout gets the timeout in milliseconds to be used for calls to other realms.
func (c ServiceCall) GetTimeout() uint64 {
	if t := c.GetUint64("timeout"); t != 0 {
		return t
	}
	return 10000 // milliseconds
}

// SetTimeout sets the timeout in milliseconds to be used for calls to other realms.
func (c *ServiceCall) SetTimeout(timeout uint64) error {
	if timeout == 0 {
		timeout = 10000 // milliseconds
	}
	return c.Set("timeout", timeout)
}

// IsRemote checks if is a remote call.
func (c ServiceCall) IsRemote() bool {
	return c.Exists("gateway")
}

// NewEmptyTransport creates a new empty transport payload
func NewEmptyTransport() *Transport {
	return &Transport{Payload: NewNamespaced("transport")}
}

// NewTransport creates a new transport payload
func NewTransport(tm *TransportMeta) *Transport {
	t := NewEmptyTransport()
	t.SetMeta(tm)
	return t
}

// NewTransportFromMap creates a new transport payload from a map
func NewTransportFromMap(m map[string]interface{}) *Transport {
	t := NewEmptyTransport()
	t.Data = m
	return t
}

// Transport defines a transport payload
type Transport struct {
	*Payload
}

func (t Transport) getObjectProperty(name string) map[string]interface{} {
	if value := t.GetMap(name); value != nil {
		return value
	}
	return nil
}

// GetMeta gets transport meta as a map
func (t Transport) GetMeta() *TransportMeta {
	if m := t.getObjectProperty("meta"); m != nil {
		return NewTransportMetaFromMap(m)
	}
	return nil
}

// SetMeta sets transport meta values
func (t *Transport) SetMeta(tm *TransportMeta) error {
	meta := make(map[string]interface{})
	// Note: Inner tm slices or maps are referenced
	for k, v := range tm.Data {
		meta[k] = v
	}
	return t.Set("meta", meta)
}

// GetBody gets the semantics of the file to download in the response
func (t Transport) GetBody() *File {
	if m := t.getObjectProperty("body"); m != nil {
		return NewFileFromMap(m)
	}
	return nil
}

// SetBody sets the semantics of the file to download in the response
func (t *Transport) SetBody(f *File) error {
	return t.Set("body", f.Data)
}

// GetFiles gets files uploaded to Gateway or sent by a Service
func (t Transport) GetFiles() map[string]interface{} {
	return t.getObjectProperty("files")
}

// GetActionFiles gets files sent by a Service action
func (t Transport) GetActionFiles(address, service, version, action string) map[string]*File {
	p := traverse.NewSpacedPath("files", address, traverse.N(service), traverse.N(version), traverse.N(action))
	if s := t.PgetSlice(p.String(), p.Sep); s != nil {
		files := make(map[string]*File)
		for _, d := range s {
			data, ok := d.(map[string]interface{})
			if !ok {
				continue
			}
			f := NewFileFromMap(data)
			files[f.GetName()] = f
		}
		return files
	}
	return nil
}

// SetFiles sets a list of files.
// This call overwrites any files that already exists for the action.
func (t *Transport) SetFiles(address, service, version, action string, fs []*File) error {
	// Get the data for each file payload
	var files []Data
	for _, f := range fs {
		files = append(files, f.Data)
	}
	// Add the files data to the transport
	p := traverse.NewSpacedPath(
		"files",
		address,
		traverse.N(service),
		traverse.N(version),
		traverse.N(action),
	)
	return t.Pset(p.String(), files, p.Sep)
}

// AddFile adds a file.
func (t *Transport) AddFile(address, service, version, action string, f *File) error {
	p := traverse.NewSpacedPath(
		"files",
		address,
		traverse.N(service),
		traverse.N(version),
		traverse.N(action),
	)
	return t.Ppush(p.String(), f.Data, p.Sep)
}

// GetData gets data stored by each Service for the response
func (t Transport) GetData() map[string]interface{} {
	return t.getObjectProperty("data")
}

// SetData sets data stored by each Service for the response
func (t *Transport) SetData(value map[string]interface{}) error {
	return t.Set("data", value)
}

// GetRelations gets relationships for a service
func (t Transport) GetRelations(address, service string) map[string]interface{} {
	p := traverse.NewSpacedPath("relations", address, traverse.N(service))
	return t.PgetMap(p.String(), p.Sep)
}

// GetAllRelations gets all relationships.
func (t Transport) GetAllRelations() map[string]interface{} {
	return t.GetMap("relations")
}

// AddRelation adds a new relation between two entities.
func (t *Transport) AddRelation(address, service, pk, relAddress, relService string, fk interface{}) error {
	p := traverse.NewSpacedPath(
		"relations",
		address,
		traverse.N(service),
		traverse.N(pk),
		relAddress,
		traverse.N(relService),
	)
	return t.Pset(p.String(), fk, p.Sep)
}

// GetLinks gets hyperlinks defined by each Service
func (t Transport) GetLinks() map[string]interface{} {
	return t.getObjectProperty("links")
}

// AddLink adds a new service hyperlink.
func (t *Transport) AddLink(address, service, link, uri string) error {
	p := traverse.NewSpacedPath(
		"links",
		address,
		traverse.N(service),
		traverse.N(link),
	)
	return t.Pset(p.String(), uri, p.Sep)
}

// GetCalls gets calls to other Services within current request
func (t Transport) GetCalls() map[string]interface{} {
	return t.getObjectProperty("calls")
}

// GetServiceCalls gets the calls registered by a Service
func (t Transport) GetServiceCalls(name, version string) []*ServiceCall {
	p := traverse.NewSpacedPath("calls", traverse.N(name), traverse.N(version))
	if info := t.PgetSlice(p.String(), p.Sep); len(info) > 0 {
		// Calls info is a slice of maps
		calls := []*ServiceCall{}
		for _, d := range info {
			data, ok := d.(map[string]interface{})
			if !ok {
				continue
			}
			calls = append(calls, NewServiceCallFromMap(data))
		}
		return calls
	}
	return nil
}

// AddCall add a call to another service action.
func (t *Transport) AddCall(service, version string, c *ServiceCall) error {
	p := traverse.NewSpacedPath(
		"calls",
		traverse.N(service),
		traverse.N(version),
	)
	return t.Pset(p.String(), c.Data, p.Sep)
}

// GetTransactions gets transactions registered by each Service
func (t Transport) GetTransactions() *Transactions {
	if m := t.getObjectProperty("transactions"); m != nil {
		return NewTransactionsFromMap(m)
	}
	return nil
}

// AddTransaction adds a transaction for a service action.
func (t *Transport) AddTransaction(name, service, version, action, caller string, ps []*Param) error {
	if name != "commit" && name != "rollback" && name != "complete" {
		return fmt.Errorf("invalid transaction command: \"%s\"", name)
	}

	p := New()
	p.Set("name", service)
	p.Set("version", version)
	p.Set("action", action)
	p.Set("caller", caller)
	if len(ps) > 0 {
		p.Set("params", ps)
	}
	return t.Ppush(fmt.Sprintf("transactions/%s", name), p.Data, "/")
}

func (t Transport) HasTransactions() bool {
	return t.Exists("transactions")
}

// GetErrors gets errors returned by each Service
func (t Transport) GetErrors() map[string]interface{} {
	return t.getObjectProperty("errors")
}

// AddError adds a service error.
func (t *Transport) AddError(address, service, version string, err *Error) error {
	p := traverse.NewSpacedPath(
		"errors",
		address,
		traverse.N(service),
		traverse.N(version),
	)
	return t.Ppush(p.String(), err.Data, p.Sep)
}

func (t Transport) HasErrors() bool {
	return t.Exists("errors")
}

func (t *Transport) ResolveError() *Error {
	if !t.HasErrors() {
		return nil
	}

	// Get error for origin service if there is one
	m := t.GetMeta()
	if m == nil {
		return nil
	}
	g := m.GetGateway()
	origin := m.GetOrigin()
	if len(origin) == 0 || g == nil {
		return nil
	}
	name := origin[0]
	version := origin[1]

	errors := New()
	errors.Data = t.GetErrors()
	p := traverse.NewSpacedPath(g.Public, traverse.N(name), traverse.N(version))
	if errs := errors.PgetSlice(p.String(), p.Sep); len(errs) > 0 {
		// There is at least one error for the origin service. Get first error.
		return NewErrorFromMap(errs[0].(map[string]interface{}))
	}

	// Get the first error that is found searching by the registered services call order
	p = traverse.NewSpacedPath("calls", traverse.N(name), traverse.N(version))
	if calls := t.PgetSlice(p.String(), p.Sep); len(calls) > 0 {
		for _, c := range calls {
			call := New()
			call.Data = c.(map[string]interface{})
			p := traverse.NewSpacedPath(
				call.GetString("gateway"),
				traverse.N(call.GetString("name")),
				traverse.N(call.GetString("version")),
			)
			if errs := errors.PgetSlice(p.String(), p.Sep); len(errs) > 0 {
				// Get the first error that is found
				return NewErrorFromMap(errs[0].(map[string]interface{}))
			}
		}
	}

	// When non of the above methods finds an error get the first error that is found
	for _, svc := range errors.Data {
		for _, vers := range svc.(map[string]interface{}) {
			for _, errs := range vers.(map[string]interface{}) {
				e := errs.([]interface{})[0]
				return NewErrorFromMap(e.(map[string]interface{}))
			}
		}
	}
	return nil
}

func (t *Transport) SetDataEntity(address, service, version, action string, value map[string]interface{}) error {
	p := traverse.NewSpacedPath(
		"data",
		address,
		traverse.N(service),
		traverse.N(version),
		traverse.N(action),
	)
	return t.Pset(p.String(), []interface{}{value}, p.Sep)
}

func (t Transport) GetDataEntity(address, service, version, action string) (map[string]interface{}, error) {
	p := traverse.NewSpacedPath(
		"data",
		address,
		traverse.N(service),
		traverse.N(version),
		traverse.N(action),
	)
	// Get the transport data that contains all the entities for the calls
	// to current service, version an action.
	// Data is a list of entities when action response is not a collection.
	data := t.PgetSlice(p.String(), p.Sep)
	if len(data) == 0 {
		// There is no transport data
		return nil, nil
	}
	// Get the first entity element in data.
	// Multiple request to the same service can add more entities.
	if entity, ok := data[0].(map[string]interface{}); ok {
		return entity, nil
	}
	return nil, fmt.Errorf(`transport data for service %s action "%s" is not a single entity`, // CODE
		format.ServiceString(address, service, version),
		action,
	)
}

func (t *Transport) SetDataCollection(address, service, version, action string, values []map[string]interface{}) error {
	p := traverse.NewSpacedPath(
		"data",
		address,
		traverse.N(service),
		traverse.N(version),
		traverse.N(action),
	)
	return t.Pset(p.String(), values, p.Sep)
}

func (t Transport) GetDataCollection(address, service, version, action string) ([]map[string]interface{}, error) {
	p := traverse.NewSpacedPath(
		"data",
		address,
		traverse.N(service),
		traverse.N(version),
		traverse.N(action),
	)
	// Get the transport data that contains all the collections for the calls
	// to current service, version an action.
	// Data is a list of entity collections when action response is a collection.
	data := t.PgetSlice(p.String(), p.Sep)
	if len(data) == 0 {
		// There is no transport data
		return nil, nil
	}
	// Get the first collection element in data.
	// Multiple request to the same service can add more collections.
	if entities, ok := data[0].([]interface{}); ok {
		collection := make([]map[string]interface{}, 0)
		for _, e := range entities {
			// Convert each entity to a map
			if entity, ok := e.(map[string]interface{}); ok {
				collection = append(collection, entity)
			}
		}
		return collection, nil
	}
	return nil, fmt.Errorf(`transport data for service %s action "%s" is not a collection`, // CODE
		format.ServiceString(address, service, version),
		action,
	)
}
