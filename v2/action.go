// Go SDK for the KUSANAGI(tm) framework (http://kusanagi.io)
// Copyright (c) 2016-2021 KUSANAGI S.L. All rights reserved.
//
// Distributed under the MIT license.
//
// For the full copyright and license information, please view the LICENSE
// file that was distributed with this source code.

package kusanagi

import (
	"errors"
	"fmt"
	"reflect"
	"time"

	"github.com/kusanagi/kusanagi-sdk-go/v2/lib/datatypes"
	"github.com/kusanagi/kusanagi-sdk-go/v2/lib/payload"
)

// Default action return values by type
var defaultReturnValues = map[string]interface{}{
	datatypes.Null:    nil,
	datatypes.Boolean: false,
	datatypes.Integer: 0,
	datatypes.Float:   0.0,
	datatypes.String:  "",
	datatypes.Binary:  []byte(""),
	datatypes.Array:   make([]interface{}, 0),
	datatypes.Object:  make(map[string]interface{}),
}

func newAction(c Component, s *state) *Action {
	api := newApi(c, s)

	// Copy the command transport without keeping references to avoid changing
	// the transport data in the command when the action changes the transport.
	// This leaves a "vanilla" transport inside the command, to be used as base
	// transport for the runtime calls.
	transport := api.command.Command.Arguments.Transport.Clone()
	transport.SetReply(api.reply)

	// Index the files for the current action by name
	gateway := transport.GetGateway()[1]
	service := api.GetName()
	version := api.GetVersion()
	files := make(map[string]payload.File)
	if transport.Files != nil {
		for _, f := range transport.Files.Get(gateway, service, version, s.action) {
			files[f.Name] = f
		}
	}

	// Index parameters by name
	params := make(map[string]payload.Param)
	for _, p := range api.command.Command.Arguments.Params {
		params[p.Name] = p
	}

	// Set a default return value for the action when there are schemas
	if api.schemas != nil {
		if schema, _ := api.GetServiceSchema(service, version); schema != nil {
			if s, err := schema.GetActionSchema(s.action); err == nil && s.HasReturn() {
				if rtype, err := s.GetReturnType(); err == nil {
					transport.SetReturn(defaultReturnValues[rtype])
				}
			}
		}
	}

	return &Action{api, transport, params, files}
}

// Action API type for the service component.
type Action struct {
	*Api

	transport *payload.Transport
	params    map[string]payload.Param
	files     map[string]payload.File
}

func (a *Action) warnWhenSchemaIsMissing(service, version, action string) {
	schema, err := a.GetServiceSchema(service, version)
	if err != nil {
		a.logger.Warning(err)
	} else if _, err := schema.GetActionSchema(action); err != nil {
		a.logger.Warning(err)
	}
}

func (a *Action) checkFiles(schema *ServiceSchema, files []File) error {
	// Check that the file server is enabled when one of the files is local
	for _, f := range files {
		if f.IsLocal() {
			// Stop checking when one local file is found and the file server is enabled
			if schema.HasFileServer() {
				return nil
			}

			return errors.New("File server not configured")
		}
	}
	return nil
}

// IsOrigin checks if the current service is the origin of the request.
func (a *Action) IsOrigin() bool {
	o := a.reply.Command.Result.Transport.Meta.Origin
	return o[0] == a.GetName() && o[1] == a.GetVersion() && o[2] == a.GetActionName()
}

// GetActionName returns the name of the action.
func (a *Action) GetActionName() string {
	return a.state.action
}

// SetProperty sets a userland property in the transport.
//
// name: The property name.
// value: The property value.
func (a *Action) SetProperty(name, value string) *Action {
	a.reply.Command.Result.Transport.Meta.Properties[name] = value
	return a
}

// HasParam checks if a parameter exists.
//
// name: The name of the parameter.
func (a *Action) HasParam(name string) bool {
	_, exists := a.params[name]
	return exists
}

// GetParam returns an action parameter.
//
// name: The name of the parameter.
func (a *Action) GetParam(name string) *Param {
	if p, exists := a.params[name]; exists {
		return payloadToParam(p)
	}
	return newEmptyParam(name)
}

// GetParams returns all the action's parameters.
func (a *Action) GetParams() (params []*Param) {
	for _, p := range a.params {
		params = append(params, payloadToParam(p))
	}
	return params
}

// NewParam creates a new parameter.
//
// Creates an instance of Param with the given name, and optionally the value and data type.
// When the value is not provided then an empty string is assumed.
// If the data type is not defined then "string" is assumed.
//
// name: The parameter name.
// value: The parameter value.
// dataType: The data type of the value.
func (a *Action) NewParam(name string, value interface{}, dataType string) (*Param, error) {
	return newParam(name, value, dataType, true)
}

// HasFile checks if a file was provided for the action.
//
// name: The name of the file parameter.
func (a *Action) HasFile(name string) bool {
	_, exists := a.files[name]
	return exists
}

// GetFile returns an uploaded file.
//
// name: The name of the file parameter.
func (a *Action) GetFile(name string) File {
	if f, exists := a.files[name]; exists {
		return payloadToFile(f)
	}
	return File{name: name}
}

// GetFiles returns all the uploaded files.
func (a *Action) GetFiles() (files []File) {
	for _, f := range a.files {
		files = append(files, payloadToFile(f))
	}
	return files
}

// NewFile creates a new file.
//
// name: Name of the file parameter.
// path: Optional path to the file.
// mimeType: Optional MIME type of the file contents.
func (a *Action) NewFile(name, path, mimeType string) (*File, error) {
	return NewFile(name, path, mimeType, "", 0, "")
}

// SetDownload sets a file as the download.
//
// file: The file.
func (a *Action) SetDownload(f File) (*Action, error) {
	// Check that files server is enabled when the file is a local file
	if f.IsLocal() {
		name := a.GetName()
		version := a.GetVersion()
		schema, err := a.GetServiceSchema(name, version)
		if err != nil {
			return nil, err
		} else if !schema.HasFileServer() {
			return nil, fmt.Errorf(`File server not configured: "%s" (%s)`, name, version)
		}
	}
	p := fileToPayload(f)
	a.transport.SetDownload(&p)
	return a, nil
}

// SetReturn sets the value to be returned by the action.
//
// value: The action's return value.
func (a *Action) SetReturn(value interface{}) (*Action, error) {
	if a.schemas != nil {
		name := a.GetName()
		version := a.GetVersion()
		// Check that the schema for the current action is available
		schema, err := a.GetServiceSchema(name, version)
		if err != nil {
			return nil, err
		}

		action := a.GetActionName()
		actionSchema, err := schema.GetActionSchema(action)
		if err != nil {
			return nil, err
		}

		if !actionSchema.HasReturn() {
			return nil, fmt.Errorf(`Cannot set a return value in "%s" (%s) for action: "%s"`, name, version, action)
		}

		// Validate that the return value has the type defined in the config
		rtype, err := actionSchema.GetReturnType()
		if err != nil {
			return nil, err
		} else if datatypes.ResolveType(value) != rtype {
			return nil, fmt.Errorf(`Invalid return type given in "%s" (%s) for action: "%s"`, name, version, action)
		}
	} else {
		// When running the action from the CLI there is no schema available, but the
		// setting of return values must be allowed without restrictions in this case.
		a.logger.Warning("Return value set without discovery mapping available")
	}
	a.transport.SetReturn(value)
	return a, nil
}

// SetEntity set the entity data.
//
// Sets an object as the entity to be returned by the action.
//
// The entity can only be a struct or a map.
//
// Entity is validated when validation is enabled for an entity in the service config file.
//
// entity: The entity.
func (a *Action) SetEntity(entity interface{}) (*Action, error) {
	// Check that the entity type is valid
	t := reflect.TypeOf(entity)
	if k := t.Kind(); k != reflect.Struct && k != reflect.Map {
		return nil, fmt.Errorf("Entity type must be struct or map, got %s", k)
	}

	// Add the entity to the transport
	a.transport.SetData(a.GetName(), a.GetVersion(), a.GetActionName(), entity)
	return a, nil
}

// SetCollection sets the collection data.
//
// The collection can only be a slice that contains either struct or a map types.
//
// Collection is validated when validation is enabled for an entity in the service config file.
//
// collection: The collection.
func (a *Action) SetCollection(collection interface{}) (*Action, error) {
	// Check that the collection and item types are valid
	t := reflect.TypeOf(collection)
	if k := t.Kind(); k != reflect.Slice {
		return nil, fmt.Errorf("Collections must be of type slice, got %s", k)
	} else if k := t.Elem().Kind(); k != reflect.Struct && k != reflect.Map {
		return nil, fmt.Errorf("Collections must contain struct or map types, got %s", k)
	}

	// Add the collection to the transport
	a.transport.SetData(a.GetName(), a.GetVersion(), a.GetActionName(), collection)
	return a, nil
}

// RelateOne creates a "one-to-one" relation between entities.
//
// Creates a "one-to-one" relation between the entity's primary key and service with the foreign key.
//
// pk: The primary key of the local entity.
// remote: The name of the remote service.
// fk: The primary key of the remote entity.
func (a *Action) RelateOne(pk, service, fk string) (*Action, error) {
	if pk == "" {
		return nil, fmt.Errorf("The primary key is empty")
	} else if service == "" {
		return nil, fmt.Errorf("The foreign service name is empty")
	} else if fk == "" {
		return nil, fmt.Errorf("The foreign key is empty")
	}

	a.transport.SetRelateOne(a.GetName(), pk, service, fk)
	return a, nil
}

// RelateMany creates a "one-to-many" relation between entities.
//
// Creates a "one-to-many" relation between the entity's primary key and service with the foreign keys.
//
// pk: The primary key.
// service: The foreign service.
// fks: The foreign keys.
func (a *Action) RelateMany(pk, service string, fks []string) (*Action, error) {
	if pk == "" {
		return nil, fmt.Errorf("The primary key is empty")
	} else if service == "" {
		return nil, fmt.Errorf("The foreign service name is empty")
	} else if len(fks) == 0 {
		return nil, fmt.Errorf("The foreign keys are empty")
	}

	a.transport.SetRelateMany(a.GetName(), pk, service, fks)
	return a, nil
}

// RelateOneRemote creates a "one-to-one" relation between two entities.
//
// Creates a "one-to-one" relation between the entity's primary key and service with the foreign key.
//
// This type of relation is done between entities in different realms.
//
// pk: The primary key.
// address: Foreign service public address.
// service: The foreign service.
// fk: The foreign key.
func (a *Action) RelateOneRemote(pk, address, service, fk string) (*Action, error) {
	if pk == "" {
		return nil, fmt.Errorf("The primary key is empty")
	} else if address == "" {
		return nil, fmt.Errorf("The foreign service address is empty")
	} else if service == "" {
		return nil, fmt.Errorf("The foreign service name is empty")
	} else if fk == "" {
		return nil, fmt.Errorf("The foreign key is empty")
	}

	a.transport.SetRelateOneRemote(a.GetName(), pk, address, service, fk)
	return a, nil
}

// RelateManyRemote creates a "one-to-many" relation between entities.
//
// Creates a "one-to-many" relation between the entity's primary key and service with the foreign keys.
//
// This type of relation is done between entities in different realms.
//
// pk: The primary key.
// address: Foreign service public address.
// service: The foreign service.
// fks: The foreign keys.
func (a *Action) RelateManyRemote(pk, address, service string, fks []string) (*Action, error) {
	if pk == "" {
		return nil, fmt.Errorf("The primary key is empty")
	} else if address == "" {
		return nil, fmt.Errorf("The foreign service address is empty")
	} else if service == "" {
		return nil, fmt.Errorf("The foreign service name is empty")
	} else if len(fks) == 0 {
		return nil, fmt.Errorf("The foreign keys are empty")
	}

	a.transport.SetRelateManyRemote(a.GetName(), pk, address, service, fks)
	return a, nil
}

// SetLink sets a link for the given URI.
//
// link: The link name.
// uri: The link URI.
func (a *Action) SetLink(link, uri string) (*Action, error) {
	if link == "" {
		return nil, fmt.Errorf("The link is empty")
	} else if uri == "" {
		return nil, fmt.Errorf("The URI is empty")
	}

	a.transport.SetLink(a.GetName(), link, uri)
	return a, nil
}

// Commit registers a transaction to be called when request succeeds.
//
// action: The action name.
// params: Optional list of parameters.
func (a *Action) Commit(action string, params []*Param) (*Action, error) {
	if action == "" {
		return nil, fmt.Errorf("The action name is empty")
	}

	a.transport.SetTransaction(
		payload.TransactionCommit,
		a.GetName(),
		a.GetVersion(),
		a.GetActionName(),
		action,
		paramsToPayload(params),
	)
	return a, nil
}

// Rollback registers a transaction to be called when request fails.
//
// action: The action name.
// params: Optional list of parameters.
func (a *Action) Rollback(action string, params []*Param) (*Action, error) {
	if action == "" {
		return nil, fmt.Errorf("The action name is empty")
	}

	a.transport.SetTransaction(
		payload.TransactionRollback,
		a.GetName(),
		a.GetVersion(),
		a.GetActionName(),
		action,
		paramsToPayload(params),
	)
	return a, nil
}

// Complete registers a transaction to be called when request finishes.
//
// action: The action name.
// params: Optional list of parameters.
func (a *Action) Complete(action string, params []*Param) (*Action, error) {
	if action == "" {
		return nil, fmt.Errorf("The action name is empty")
	}

	a.transport.SetTransaction(
		payload.TransactionComplete,
		a.GetName(),
		a.GetVersion(),
		a.GetActionName(),
		action,
		paramsToPayload(params),
	)
	return a, nil
}

// Call performs a run-time call to a service.
//
// The result of this call is the return value from the remote action.
//
// service: The service name.
// version: The service version.
// action: The action name.
// params: Optional list of Param objects.
// files: Optional list of File objects.
// timeout: Optional timeout in milliseconds.
func (a *Action) Call(
	service string,
	version string,
	action string,
	params []*Param,
	files []File,
	timeout uint,
) (returnValue interface{}, err error) {
	// Check that the call exists in the config
	title := fmt.Sprintf(`"%s" (%s)`, service, version)
	schema, err := a.GetServiceSchema(a.GetName(), a.GetVersion())
	if err != nil {
		return nil, err
	}

	actionSchema, err := schema.GetActionSchema(a.GetActionName())
	if err != nil {
		return nil, err
	} else if !actionSchema.HasCall(service, version, action) {
		return nil, fmt.Errorf(`Call not configured, connection to action on %s aborted: "%s"`, title, action)
	}

	// Check that the remote action exists and can return a value, and if it doesn't issue a warning
	remoteSchema, err := a.GetServiceSchema(service, version)
	if err != nil {
		a.logger.Warning(err)
	}

	remoteActionSchema, err := remoteSchema.GetActionSchema(action)
	if err != nil {
		a.logger.Warning(err)
	} else if remoteActionSchema.HasReturn() {
		return nil, fmt.Errorf(`Cannot return value from %s for action: "%s"`, title, action)
	}

	// Check that the file server is enabled when one of the files is local
	for _, file := range files {
		if file.IsLocal() {
			// Stop checking when one local file is found and the file server is enables
			if schema.HasFileServer() {
				break
			}
			return nil, fmt.Errorf("File server not configured: %s", title)
		}
	}

	if timeout == 0 {
		timeout = ExecutionTimeout
	}

	var transport *payload.Transport
	var duration time.Duration

	// Make sure the action's transport always contains the call info
	// TODO: Check that duration and transport are set correctly after the runtime call
	defer func() {
		a.transport.SetCall(
			a.GetName(),
			a.GetVersion(),
			a.GetActionName(),
			service,
			version,
			action,
			uint(duration*time.Millisecond),
			paramsToPayload(params),
			filesToPayload(files),
			timeout,
			transport,
		)
	}()

	// Make the runtime call
	callee := []string{service, version, action}
	c, err := call(
		a.GetContext(),
		schema.GetAddress(),
		a.GetActionName(),
		callee,
		a.command.GetTransport().Clone(),
		params,
		files,
		a.input.IsTCPEnabled(),
		timeout,
	)
	if err != nil {
		return nil, fmt.Errorf("Run-time call failed: %v", err)
	}

	// Wait for the runtime response
	result := <-c
	if err := result.Error; err != nil {
		return nil, fmt.Errorf("Run-time call failed: %v", err)
	}

	// When the call succeeds update the transport and duration
	duration = result.Duration
	transport = result.Transport
	return result.ReturnValue, nil
}

// DeferCall registera a deferred call to a service.
//
// service: The service name.
// version: The service version.
// action: The action name.
// params: Optional list of parameters.
// files: Optional list of files.
func (a *Action) DeferCall(service, version, action string, params []*Param, files []File) (*Action, error) {
	// Check that the deferred call exists in the config
	schema, err := a.GetServiceSchema(a.GetName(), a.GetVersion())
	if err != nil {
		return nil, err
	}

	actionSchema, err := schema.GetActionSchema(a.GetActionName())
	if err != nil {
		return nil, err
	}

	if !actionSchema.HasDeferCall(service, version, action) {
		return nil, fmt.Errorf(
			`Deferred call not configured, connection to action on "%s" (%s) aborted: "%s"`,
			service,
			version,
			action,
		)
	}

	// Check that the remote action exists and if it doesn't issue a warning
	a.warnWhenSchemaIsMissing(service, version, action)

	// Check that the file server is enabled when one of the files is local
	if err := a.checkFiles(schema, files); err != nil {
		return nil, fmt.Errorf(`%v: "%s" (%s)`, err, service, version)
	}

	a.transport.SetDeferCall(
		a.GetName(),
		a.GetVersion(),
		a.GetActionName(),
		service,
		version,
		action,
		paramsToPayload(params),
		filesToPayload(files),
	)
	return a, nil
}

// RemoteCall registers a call to a remote service in another realm.
//
// These types of calls are done using KTP (KUSANAGI transport protocol).
//
// address: Public address of a gateway from another realm.
// service: The service name.
// version: The service version.
// action: The action name.
// params: Optional list of parameters.
// files: Optional list of files.
// timeout: Optional call timeout in milliseconds.
func (a *Action) RemoteCall(
	address string,
	service string,
	version string,
	action string,
	params []*Param,
	files []File,
	timeout uint,
) (*Action, error) {
	if len(address) < 6 || address[:6] == "ktp://" {
		return nil, fmt.Errorf(`The address must start with "ktp://": %s`, address)
	}

	if timeout == 0 {
		timeout = ExecutionTimeout
	}

	// Check that the deferred call exists in the config
	schema, err := a.GetServiceSchema(a.GetName(), a.GetVersion())
	if err != nil {
		return nil, err
	}

	actionSchema, err := schema.GetActionSchema(a.GetActionName())
	if err != nil {
		return nil, err
	}

	if !actionSchema.HasRemoteCall(address, service, version, action) {
		return nil, fmt.Errorf(
			`Remote call not configured, connection to action on [%s] "%s" (%s) aborted: "%s"`,
			address,
			service,
			version,
			action,
		)
	}

	// Check that the remote action exists and if it doesn't issue a warning
	a.warnWhenSchemaIsMissing(service, version, action)

	// Check that the file server is enabled when one of the files is local
	if err := a.checkFiles(schema, files); err != nil {
		return nil, fmt.Errorf(`%v: [%s] "%s" (%s)`, err, address, service, version)
	}

	a.transport.SetRemoteCall(
		address,
		a.GetName(),
		a.GetVersion(),
		a.GetActionName(),
		service,
		version,
		action,
		timeout,
		paramsToPayload(params),
		filesToPayload(files),
	)
	return a, nil
}

// Error adds an error for the current service.
//
// Adds an error object to the Transport with the specified message.
//
// message: The error message.
// code: The error code.
// status: The HTTP status message.
func (a *Action) Error(message string, code int, status string) *Action {
	if status == "" {
		status = payload.DefaultErrorStatus
	}
	a.transport.SetError(a.GetName(), a.GetVersion(), message, code, status)
	return a
}
