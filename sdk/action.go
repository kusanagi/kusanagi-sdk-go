// Go SDK for the KUSANAGI(tm) framework (http://kusanagi.io)
// Copyright (c) 2016-2019 KUSANAGI S.L. All rights reserved.
//
// Distributed under the MIT license.
//
// For the full copyright and license information, please view the LICENSE
// file that was distributed with this source code.

package sdk

import (
	"errors"
	"fmt"
	"strings"

	"github.com/kusanagi/kusanagi-sdk-go/format"
	"github.com/kusanagi/kusanagi-sdk-go/logging"
	"github.com/kusanagi/kusanagi-sdk-go/payload"
	"github.com/kusanagi/kusanagi-sdk-go/schema"
)

// CallTimeout defines the default timeout in milliseconds for service calls.
var CallTimeout = uint64(10000)

// Entity defines a generic entity.
type Entity map[string]interface{}

// Collection defines a collection of entities.
type Collection []Entity

func (c Collection) SliceOfMap() (s []map[string]interface{}) {
	for _, e := range c {
		s = append(s, e)
	}
	return s
}

// ReturnValue stores a value returned by a service action.
//
// The value can have any type. The supported types are:
//  - boolean
//  - integer
//  - float
//  - string
//  - binary ([]byte)
//  - array  (slice)
//  - object (map)
type ReturnValue struct {
	empty bool
	value interface{}
}

// IsEmpty checks if a return value was assigned.
func (r ReturnValue) IsEmpty() bool {
	return r.empty
}

// Set a return value.
func (r *ReturnValue) Set(value interface{}) {
	r.value = value
	r.empty = false
}

// Get the return value.
func (r ReturnValue) Get() interface{} {
	return r.value
}

// Default sets the default value for a return type.
func (r *ReturnValue) Default(valueType string) bool {
	switch valueType {
	case "boolean":
		r.Set(false)
	case "integer":
		r.Set(0)
	case "float":
		r.Set(0.0)
	case "string":
		r.Set("")
	case "binary":
		r.Set([]byte(""))
	case "array":
		r.Set(make([]interface{}, 0))
	case "object":
		r.Set(make(map[string]interface{}))
	default:
		return false
	}
	return true
}

// Defines a service action.
type Action interface {
	Api

	// GetActionName gets the name of the action.
	GetActionName() string

	// IsOrigin checks if the current service is the origin of the request.
	IsOrigin() bool

	// SetProperty sets a userland property in the transport.
	SetProperty(name, value string) error

	// HasParam checks if a parameter exists.
	HasParam(name string) bool

	// GetParam gets an action parameter.
	GetParam(name string) *Param

	// GetParams gets all action parameters.
	GetParams(name string) []*Param

	// NewParam creates a new action parameter.
	//
	// If the data type is not defined then "string" is assumed.
	//
	// Valid data types: "null", "boolean", "integer", "float", "string", "array" and "object".
	NewParam(name string, value interface{}, paramType string) *Param

	// HasFile checks if a file parameter exists.
	HasFile(name string) bool

	// GetFile gets an action file parameter.
	GetFile(name string) *File

	// GetFiles gets all action file parameters.
	GetFiles(name string) []*File

	// NewFile creates a new action file parameter.
	NewFile(name, path, mime string) (*File, error)

	// SetDownload sets a file as download content.
	SetDownload(f *File) error

	// SetReturn sets the value to be returned by the service.
	SetReturn(value interface{}) error

	// SetEntity sets an entity to be returned by the action.
	SetEntity(e Entity)

	// SetCollection sets a collection of entities to be returned by the action.
	SetCollection(c Collection)

	// RelateOne creates a "one to one" relation between two entities.
	RelateOne(primaryKey, service, foreignKey string)

	// RelateMany creates a "one to many" relation between two entities.
	RelateMany(primaryKey, service string, foreignKeys []string)

	// RelateOneRemote creates a "one to one" relation between two entities in different realms.
	RelateOneRemote(primaryKey, address, service, foreignKey string)

	// RelateMany creates a "one to many" relation between two entities in different realms.
	RelateManyRemote(primaryKey, address, service string, foreignKeys []string)

	// SetLink sets a link to a URI.
	SetLink(link, uri string)

	// Commit registers a transaction to be called when the request succeeds.
	Commit(action string, ps []*Param)

	// Rollback registers a transaction to be called when the request fails.
	Rollback(action string, ps []*Param)

	// Complete registers a transaction to be called when the request finishes.
	// This transaction is ALWAYS executed, it doesn't matter if request
	// fails or succeeds.
	Complete(action string, ps []*Param)

	// Call a service using a run-time call.
	// The result is the value returned by the called service.
	Call(service, version, action string, ps []*Param, fs []*File, timeout int) (interface{}, error)

	// DeferCall registers a deferred call to a service within the same realm.
	DeferCall(service, version, action string, ps []*Param, fs []*File) error

	// RemoteCall registers a remote call to a service within a different realm.
	RemoteCall(address, service, version, action string, ps []*Param, fs []*File, timeout uint64) error

	// Error adds an error for the current service to the transport.
	// "500 Internal Server Error" is assumed when status is empty.
	Error(message string, code int, status string) error
}

func newAction(
	api *api,
	name string,
	tp *payload.Transport,
	rv *ReturnValue,
) *action {
	meta := tp.GetMeta()
	a := action{
		api:         api,
		name:        name,
		params:      make(map[string]*Param),
		files:       make(map[string]*File),
		meta:        meta,
		transport:   tp,
		gateway:     meta.GetGateway(),
		returnValue: rv,
	}

	// Add the file parameters
	files := tp.GetActionFiles(a.gateway.Public, a.GetName(), a.GetVersion(), a.GetActionName())
	for name, f := range files {
		file, err := NewFile(name, f.GetPath(), f.GetMime(), f.GetFilename(), f.GetSize(), f.GetToken())
		if err != nil {
			logging.Errorf("failed to add file parameter: \"%s\"", name)
			continue
		}
		a.files[name] = file
	}

	// Get the schema for current service and action
	var err error
	a.service, err = a.GetServiceSchema(a.GetName(), a.GetVersion())
	if err != nil {
		logging.Errorf("failed to get the service schema: %v", err)
	} else if a.service != nil {
		a.action, err = a.service.GetActionSchema(a.GetActionName())
		if err != nil {
			logging.Errorf("failed to get the action schema: %v", err)
		}
	}

	// Add a default return value if it is supported
	if a.action == nil {
		// When the action schema is not available set nil as default.
		// This is required when the action is executed from the CLI,
		// because at that point there are no mappings.
		a.returnValue.Set(nil)
	} else if a.action.HasReturn() {
		// Set the default value for current return type
		a.returnValue.Default(a.action.GetReturnType())
	}
	return &a
}

type action struct {
	*api

	name        string
	params      map[string]*Param
	files       map[string]*File
	returnValue *ReturnValue
	meta        *payload.TransportMeta
	transport   *payload.Transport
	gateway     *payload.GatewayAddr
	service     *schema.Service
	action      *schema.Action
}

func (a *action) setParams(params []map[string]interface{}) {
	// Add the service request parameters
	for _, m := range params {
		p := payload.NewParamFromMap(m)
		name := p.GetName()
		a.params[name] = newParam(name, p.GetValue(), p.GetType(), true)
	}
}

func (a *action) GetActionName() string {
	return a.name
}

func (a *action) IsOrigin() bool {
	origin := a.meta.GetOrigin()
	if len(origin) != 3 {
		return false
	}
	return (a.GetName() == origin[0] && a.GetVersion() == origin[1] && a.GetActionName() == origin[2])
}

func (a *action) SetProperty(name, value string) error {
	return a.meta.SetProperty(name, value)
}

func (a *action) HasParam(name string) bool {
	_, ok := a.params[name]
	return ok
}

func (a *action) GetParam(name string) *Param {
	return a.params[name]
}

func (a *action) GetParams(name string) []*Param {
	params := []*Param{}
	for _, p := range a.params {
		params = append(params, p)
	}
	return params
}

func (a *action) NewParam(name string, value interface{}, paramType string) *Param {
	return newParam(name, value, paramType, false)
}

func (a *action) HasFile(name string) bool {
	_, ok := a.files[name]
	return ok
}

func (a *action) GetFile(name string) *File {
	return a.files[name]
}

func (a *action) GetFiles(name string) []*File {
	files := []*File{}
	for _, f := range a.files {
		files = append(files, f)
	}
	return files
}

func (a *action) NewFile(name, path, mime string) (*File, error) {
	return NewFile(name, path, mime, "", 0, "")
}

func (a *action) SetDownload(f *File) error {
	// TODO: Validate that there is a file server enabled for current service (see specs)
	return a.transport.SetBody(FileToPayload(f))
}

func (a *action) SetReturn(value interface{}) error {
	// Check that return value is supported
	if a.action != nil && !a.action.HasReturn() {
		return newErrUndefinedReturnValue(a.GetName(), a.GetVersion(), a.GetActionName())
	}

	// When running from the CLI allow any return value otherwise check the type
	if a.action != nil {
		// TODO: Validate return value type
	}

	a.returnValue.Set(value)
	return nil
}

func (a *action) SetEntity(e Entity) {
	a.transport.SetDataEntity(a.gateway.Public, a.GetName(), a.GetVersion(), a.GetActionName(), e)
}

func (a *action) SetCollection(c Collection) {
	a.transport.SetDataCollection(a.gateway.Public, a.GetName(), a.GetVersion(), a.GetActionName(), c.SliceOfMap())
}

func (a *action) RelateOne(primaryKey, service, foreignKey string) {
	a.transport.AddRelation(a.gateway.Public, a.GetName(), primaryKey, a.gateway.Public, service, foreignKey)
}

func (a *action) RelateMany(primaryKey, service string, foreignKeys []string) {
	a.transport.AddRelation(a.gateway.Public, a.GetName(), primaryKey, a.gateway.Public, service, foreignKeys)
}

func (a *action) RelateOneRemote(primaryKey, address, service, foreignKey string) {
	a.transport.AddRelation(a.gateway.Public, a.GetName(), primaryKey, address, service, foreignKey)
}

func (a *action) RelateManyRemote(primaryKey, address, service string, foreignKeys []string) {
	a.transport.AddRelation(a.gateway.Public, a.GetName(), primaryKey, address, service, foreignKeys)
}

func (a *action) SetLink(link, uri string) {
	a.transport.AddLink(a.gateway.Public, a.GetName(), link, uri)
}

func (a *action) Commit(action string, ps []*Param) {
	a.transport.AddTransaction("commit", a.GetName(), a.GetVersion(), action, a.GetActionName(), paramsToPayloads(ps))
}

func (a *action) Rollback(action string, ps []*Param) {
	a.transport.AddTransaction("rollback", a.GetName(), a.GetVersion(), action, a.GetActionName(), paramsToPayloads(ps))
}

func (a *action) Complete(action string, ps []*Param) {
	a.transport.AddTransaction("complete", a.GetName(), a.GetVersion(), action, a.GetActionName(), paramsToPayloads(ps))
}

func (a *action) Call(service, version, action string, ps []*Param, fs []*File, timeout int) (interface{}, error) {
	// TODO: implement runtime call support for actions
	return nil, errors.New("not implemented")
}

func (a *action) DeferCall(service, version, action string, ps []*Param, fs []*File) error {
	if len(fs) > 0 {
		// When there is no file server check that theare are no local files
		if a.service != nil && !a.service.HasFileServer() {
			for _, f := range fs {
				if f.IsLocal() {
					return newErrNoFileServer(a.GetName(), a.GetVersion())
				}
			}
		}
		// When files are added to the call update the transport files
		a.transport.SetFiles(a.gateway.Public, service, version, action, filesToPayloads(fs))
	}

	// Create the service call payload and add it to the transport
	c := payload.NewEmptyServiceCall()
	c.SetName(service)
	c.SetVersion(version)
	c.SetAction(action)
	c.SetCaller(a.GetActionName())
	if len(ps) > 0 {
		c.SetParams(paramsToPayloads(ps))
	}
	return a.transport.AddCall(a.GetName(), a.GetVersion(), c)
}

func (a *action) RemoteCall(address, service, version, action string, ps []*Param, fs []*File, timeout uint64) error {
	// Make sure that the remote call is using the KTP protocol
	if !strings.HasPrefix(address, "ktp://") {
		return fmt.Errorf("remote KTP addresses must start with \"ktp://\": \"%s\"", address)
	}

	// When files are added to the call update the transport files
	if len(fs) > 0 {
		a.transport.SetFiles(a.gateway.Public, service, version, action, filesToPayloads(fs))
	}
	// Use default timeout when value is 0
	if timeout == 0 {
		timeout = CallTimeout
	}
	// Create the service call payload and add it to the transport
	c := payload.NewEmptyServiceCall()
	c.SetGateway(address)
	c.SetName(service)
	c.SetVersion(version)
	c.SetAction(action)
	c.SetCaller(a.GetActionName())
	c.SetTimeout(timeout)
	if len(ps) > 0 {
		c.SetParams(paramsToPayloads(ps))
	}
	return a.transport.AddCall(a.GetName(), a.GetVersion(), c)
}

func (a *action) Error(message string, code int, status string) error {
	if status == "" {
		status = "500 Internal Server Error"
	}
	err := payload.NewError()
	err.SetMessage(message)
	err.SetCode(code)
	err.SetStatus(status)
	return a.transport.AddError(a.gateway.Public, a.GetName(), a.GetVersion(), err)
}

func paramsToPayloads(ps []*Param) []*payload.Param {
	var pps []*payload.Param
	for _, p := range ps {
		pps = append(pps, paramToPayload(p))
	}
	return pps
}

func filesToPayloads(fs []*File) []*payload.File {
	var pfs []*payload.File
	for _, f := range fs {
		pfs = append(pfs, FileToPayload(f))
	}
	return pfs
}

func newErrNoFileServer(service, version string) error {
	return fmt.Errorf("File server not configured: %s", format.ServiceString(service, version, ""))
}

func newErrUndefinedReturnValue(service, version, action string) error {
	return fmt.Errorf(`Cannot set a return value in %s for action: "%s"`, format.ServiceString(service, version, ""), action)
}
