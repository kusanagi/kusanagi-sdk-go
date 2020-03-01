// Go SDK for the KUSANAGI(tm) framework (http://kusanagi.io)
// Copyright (c) 2016-2020 KUSANAGI S.L. All rights reserved.
//
// Distributed under the MIT license.
//
// For the full copyright and license information, please view the LICENSE
// file that was distributed with this source code.

package sdk

import (
	"errors"
	"fmt"
	"reflect"

	"github.com/kusanagi/kusanagi-sdk-go/payload"
	"github.com/kusanagi/kusanagi-sdk-go/traverse"
)

func newErrUnexpectedType(path string, value interface{}) error {
	return fmt.Errorf(`unexpected type in transport path "%s": %T`, path, value)
}

func newTransport(tp *payload.Transport) *Transport {
	meta := tp.GetMeta()
	return &Transport{
		meta:       meta,
		transport:  tp,
		properties: meta.GetProperties(),
	}
}

// Transport allows read-only access to the transport payload.
type Transport struct {
	meta       *payload.TransportMeta
	transport  *payload.Transport
	properties map[string]string
}

// GetRequestId gets the UUID of the request.
func (t Transport) GetRequestId() string {
	return t.meta.GetID()
}

// GetRequestTimestamp gets the timestamp with the request creation date.
func (t Transport) GetRequestTimestamp() string {
	return t.meta.GetDatetime()
}

// GetOriginService gets the origin of the request.
// The result contains the name, version and action of the
// service that was the origin of the request.
func (t Transport) GetOriginService() []string {
	return t.meta.GetOrigin()
}

// GetDuration gets the service execution time in milliseconds.
// The duration is the time spent processing the request by the service
// that was the origin of the request.
func (t Transport) GetOriginDuration() int64 {
	return t.meta.GetDuration()
}

// GetProperty gets a userland property value.
// The name of the property is case sensitive.
func (t Transport) GetProperty(name, defaultValue string) string {
	if v, ok := t.properties[name]; ok {
		return v
	}
	return defaultValue
}

// GetProperties gets all the userland properties.
func (t Transport) GetProperties() map[string]string {
	return t.properties
}

// HasDownload checks if a file download has been registered for the response.
func (t Transport) HasDownload() bool {
	return t.transport.Exists("body")
}

// GetDownload gets the file download registered for the response.
func (t Transport) GetDownload() *File {
	if fp := t.transport.GetBody(); fp != nil {
		return PayloadToFile(fp)
	}
	return nil
}

// GetData gets the transport data.
func (t Transport) GetData() ([]ServiceData, error) {
	var data []ServiceData

	for addr, v := range t.transport.GetData() {
		services, ok := v.(map[string]interface{})
		if !ok {
			return nil, newErrUnexpectedType("data.services", v)
		}

		for name, v := range services {
			versions, ok := v.(map[string]interface{})
			if !ok {
				return nil, newErrUnexpectedType("data.services.versions", v)
			}

			for version, v := range versions {
				actions, ok := v.(map[string]interface{})
				if !ok {
					return nil, newErrUnexpectedType("data.services.versions.actions", v)
				}

				data = append(data, ServiceData{addr, name, version, actions})
			}
		}
	}
	return data, nil
}

// GetRelations gets the service relations.
func (t Transport) GetRelations() ([]Relation, error) {
	var relations []Relation

	for addr, v := range t.transport.GetAllRelations() {
		services, ok := v.(map[string]interface{})
		if !ok {
			return nil, newErrUnexpectedType("relations.services", v)
		}

		for name, v := range services {
			pks, ok := v.(map[string]interface{})
			if !ok {
				return nil, newErrUnexpectedType("relations.services.pk", v)
			}

			for pk, v := range pks {
				foreign, ok := v.(map[string]interface{})
				if !ok {
					return nil, newErrUnexpectedType("relations.services.pk.foreign", v)
				}

				relations = append(relations, Relation{addr, name, pk, foreign})
			}
		}
	}
	return relations, nil
}

// GetLinks gets the service links.
func (t Transport) GetLinks() ([]Link, error) {
	var links []Link

	for addr, v := range t.transport.GetLinks() {
		services, ok := v.(map[string]interface{})
		if !ok {
			return nil, newErrUnexpectedType("links.services", v)
		}

		for name, v := range services {
			references, ok := v.(map[string]interface{})
			if !ok {
				return nil, newErrUnexpectedType("links.services.references", v)
			}

			for ref, v := range references {
				uri, ok := v.(string)
				if !ok {
					return nil, fmt.Errorf("transport link URI must be a string, got: %T", v)
				}

				links = append(links, Link{addr, name, ref, uri})
			}
		}
	}
	return links, nil
}

// GetCalls gets the service calls.
func (t Transport) GetCalls() ([]Caller, error) {
	var calls []Caller

	for service, v := range t.transport.GetCalls() {
		versions, ok := v.(map[string]interface{})
		if !ok {
			return nil, newErrUnexpectedType("calls.versions", v)
		}

		for version, v := range versions {
			callers, ok := v.([]interface{})
			if !ok {
				return nil, newErrUnexpectedType("calls.versions.callers", v)
			}

			for i, v := range callers {
				data, ok := v.(map[string]interface{})
				if !ok {
					return nil, newErrUnexpectedType(fmt.Sprintf("calls.versions.callers[%d]", i), v)
				}

				action, _ := traverse.Get(data, "caller", "", &payload.Aliases)
				calls = append(calls, Caller{service, version, action.(string), data})
			}
		}
	}
	return calls, nil
}

// GetTransactions gets the transactions.
// The transaction command is case sensitive, and supports "commit", "rollback" or "complete" as value.
func (t Transport) GetTransactions(command string) ([]Transaction, error) {
	if command != "commit" && command != "rollback" && command != "complete" {
		return nil, fmt.Errorf(`unknown transaction type provided: "%s"`, command)
	}

	var trs []Transaction
	transactions := t.transport.GetTransactions()
	for _, a := range transactions.GetActions(command) {
		trs = append(trs, Transaction{command, a})
	}
	return trs, nil
}

// GetErrors gets the service errors.
func (t Transport) GetErrors() ([]Error, error) {
	var errors []Error

	for addr, v := range t.transport.GetErrors() {
		services, ok := v.(map[string]interface{})
		if !ok {
			return nil, newErrUnexpectedType("errors.services", v)
		}

		for name, v := range services {
			versions, ok := v.(map[string]interface{})
			if !ok {
				return nil, newErrUnexpectedType("errors.services.versions", v)
			}

			for version, v := range versions {
				errs, ok := v.([]interface{})
				if !ok {
					return nil, newErrUnexpectedType("errors.services.versions.errors", v)
				}

				for i, v := range errs {
					m, ok := v.(map[string]interface{})
					if !ok {
						return nil, newErrUnexpectedType(fmt.Sprintf("errors.services.versions.errors[%d]", i), v)
					}

					errors = append(errors, Error{addr, name, version, payload.NewErrorFromMap(m)})
				}
			}
		}
	}
	return errors, nil
}

// ServiceData represents the data stored in the transport by a service.
type ServiceData struct {
	address string
	service string
	version string
	actions map[string]interface{}
}

// GetAddress gets the gateway address for the service.
func (s ServiceData) GetAddress() string {
	return s.address
}

// GetName gets the service name.
func (s ServiceData) GetName() string {
	return s.service
}

// GetVersion gets the service version.
func (s ServiceData) GetVersion() string {
	return s.version
}

// Get the list of action data items for current service.
// Each item represents an action on the Service for which data exists.
func (s ServiceData) GetActions() (actions []ActionData) {
	for name, v := range s.actions {
		if data, ok := v.([]interface{}); ok {
			actions = append(actions, ActionData{name, data})
		}
	}
	return actions
}

// ActionData represents the action data stored in the transport by a service.
type ActionData struct {
	name string
	data []interface{}
}

// GetName gets the name of the service action that returned the data.
func (a ActionData) GetName() string {
	return a.name
}

// IsCollection checks if the data for this action is a collection.
func (a ActionData) IsCollection() (bool, error) {
	if len(a.data) == 0 {
		return false, errors.New("the transport contains an empty action data slice")
	}
	return reflect.ValueOf(a.data[0]).Kind() == reflect.Slice, nil
}

// GetData gets the transport data for the service action.
// Each item in the list represents a call that included data
// in the transport, where each item may be a slice or a map,
// depending on whether the data is a collection or not.
func (a ActionData) GetData() []interface{} {
	return a.data
}

// Relation represents a service relation.
type Relation struct {
	address string
	service string
	pk      string
	foreign map[string]interface{}
}

// GetAddress gets the gateway address for the service.
func (r Relation) GetAddress() string {
	return r.address
}

// GetName gets the name of the service.
func (r Relation) GetName() string {
	return r.service
}

// GetPrimaryKey gets the value for the primary key of the relation.
func (r Relation) GetPrimaryKey() string {
	return r.pk
}

// GetForeignRelations gets the relation data for the foreign services.
func (r Relation) GetForeignRelations() (data []ForeignRelation) {
	for addr, v := range r.foreign {
		services, ok := v.(map[string]interface{})
		if !ok {
			break
		}

		for service, fks := range services {
			data = append(data, ForeignRelation{addr, service, fks})
		}
	}
	return data
}

// ForeignRelation represent a foreign relation.
type ForeignRelation struct {
	address string
	service string
	fks     interface{}
}

// GetAddress gets the gateway address for the foreign service.
func (r ForeignRelation) GetAddress() string {
	return r.address
}

// GetName gets the name of the foreign service.
func (r ForeignRelation) GetName() string {
	return r.service
}

// GetType gst the type of the relation.
// Relation type can be either "one" or "many".
func (r ForeignRelation) GetType() (string, error) {
	if r.fks == nil {
		return "", errors.New("the transport contains empty foreign relation data")
	}

	if _, ok := r.fks.(string); ok {
		return "one", nil
	}
	return "many", nil
}

// GetForeignKeys gets the foreign key value(s) of the relation.
func (r ForeignRelation) GetForeignKeys() (fks []string, err error) {
	t, err := r.GetType()
	if err != nil {
		return nil, err
	}

	if t == "one" {
		if fk, ok := r.fks.(string); ok {
			fks = append(fks, fk)
		} else {
			return nil, fmt.Errorf("foreign key value must be a string, got: %T", r.fks)
		}
	} else {
		values, ok := r.fks.([]interface{})
		if !ok {
			return nil, fmt.Errorf("invalid foreign keys type: %T", r.fks)
		}

		for _, v := range values {
			if fk, ok := v.(string); ok {
				fks = append(fks, fk)
			} else {
				return nil, fmt.Errorf("foreign key value must be a string, got: %T", r.fks)
			}
		}
	}
	return fks, nil
}

// Link represents a link to a service.
type Link struct {
	address   string
	service   string
	reference string
	uri       string
}

// GetAddress gets the gateway address for the service.
func (l Link) GetAddress() string {
	return l.address
}

// GetName gets the service name.
func (l Link) GetName() string {
	return l.service
}

// GetLink gets the link reference.
func (l Link) GetLink() string {
	return l.reference
}

// GetUri gets the link URI.
func (l Link) GetUri() string {
	return l.uri
}

// Caller represents a service which registered call.
type Caller struct {
	service string
	version string
	action  string
	callee  map[string]interface{}
}

// GetName gets the service name.
func (c Caller) GetName() string {
	return c.service
}

// GetVersion gets the service version.
func (c Caller) GetVersion() string {
	return c.version
}

// GetAction gets the name of the service acton that is making the call.
func (c Caller) GetAction() string {
	return c.action
}

// GetCallee gets the info of the service being called.
func (c Caller) GetCallee() Callee {
	return Callee{payload.NewServiceCallFromMap(c.callee)}
}

// Callee represents a service being called by another service.
type Callee struct {
	data *payload.ServiceCall
}

// GetDuration gets the duration of the call in milliseconds.
func (c Callee) GetDuration() uint64 {
	return c.data.GetDuration()
}

// IsRemote checks if is a call to a service in another realm.
func (c Callee) IsRemote() bool {
	return c.data.IsRemote()
}

// GetAddress gets the public gateway address for calls to another realm.
func (c Callee) GetAddress() string {
	return c.data.GetGateway()
}

// GetTimeout gets the timeout in milliseconds for a call to a service in another realm.
func (c Callee) GetTimeout() uint64 {
	return c.data.GetTimeout()
}

// GetName gets the name of the service being called.
func (c Callee) GetName() string {
	return c.data.GetName()
}

// GetVersion gets the version of the service being called.
func (c Callee) GetVersion() string {
	return c.data.GetVersion()
}

// GetAction gets the name of the service action being called.
func (c Callee) GetAction() string {
	return c.data.GetAction()
}

// GetParams gets the call parameters.
func (c Callee) GetParams() []*Param {
	var params []*Param

	for _, p := range c.data.GetParams() {
		params = append(params, payloadToParam(p))
	}
	return params
}

// Transaction represents a single transaction.
type Transaction struct {
	command string
	action  *payload.TransactionAction
}

// GetType gets the transaction type.
func (t Transaction) GetType() string {
	return t.command
}

// GetVersion gets the name of the service that registered the transaction.
func (t Transaction) GetName() string {
	return t.action.Name()
}

// GetVersion gets the version of the service that registered the transaction.
func (t Transaction) GetVersion() string {
	return t.action.Version()
}

// GetCallerAction gets the name of the action that registered the transaction.
func (t Transaction) GetCallerAction() string {
	return t.action.Caller()
}

// GetCalleeAction gets the name of the action to be called by the transaction.
func (t Transaction) GetCalleeAction() string {
	return t.action.Action()
}

// GetParams gets the transaction parameters.
func (t Transaction) GetParams() (params []*Param) {
	for _, p := range t.action.Params() {
		params = append(params, payloadToParam(p))
	}
	return params
}

// Error represents an error for a service call.
type Error struct {
	address string
	service string
	version string
	err     *payload.Error
}

// GetAddress gets the gateway address for the service.
func (e Error) GetAddress() string {
	return e.address
}

// GetName gets the service name.
func (e Error) GetName() string {
	return e.service
}

// GetVersion gets the service version.
func (e Error) GetVersion() string {
	return e.version
}

// GetMessage gets the error message.
func (e Error) GetMessage() string {
	return e.err.GetMessage()
}

// GetCode gets the error code.
func (e Error) GetCode() int {
	return e.err.GetCode()
}

// GetStatus gets the status message.
func (e Error) GetStatus() string {
	return e.err.GetStatus()
}
