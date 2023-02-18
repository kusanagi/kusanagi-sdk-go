// Go SDK for the KUSANAGI(tm) framework (http://kusanagi.io)
// Copyright (c) 2016-2023 KUSANAGI S.L. All rights reserved.
//
// Distributed under the MIT license.
//
// For the full copyright and license information, please view the LICENSE
// file that was distributed with this source code.

package payload

import "errors"

// TransactionCommit defines the command type for commit transactions.
const TransactionCommit = "commit"

// TransactionRollback defines the command type for rollback transactions.
const TransactionRollback = "rollback"

// TransactionComplete defines the command type for complete transactions.
const TransactionComplete = "complete"

// Get the key to use in the transport payload for different transaction commands.
func transactionKey(command string) string {
	switch command {
	case TransactionCommit:
		return "c"
	case TransactionRollback:
		return "r"
	case TransactionComplete:
		return "C"
	}
	return ""
}

func mergeRuntimeCallTransportData(source, target *Transport) {
	if target.Data == nil {
		target.Data = ServiceData{}
	}

	target.Data.merge(source.Data)
}

func mergeRuntimeCallTransportRelations(source, target *Transport) {
	if target.Relations == nil {
		target.Relations = Relations{}
	}

	target.Relations.merge(source.Relations)
}

func mergeRuntimeCallTransportLinks(source, target *Transport) {
	if target.Links == nil {
		target.Links = Links{}
	}
	target.Links.merge(source.Links)
}

func mergeRuntimeCallTransportCalls(source, target *Transport) {
	if target.Calls == nil {
		target.Calls = Calls{}
	}
	target.Calls.merge(source.Calls)
}

func mergeRuntimeCallTransportTransactions(source, target *Transport) {
	if target.Transactions == nil {
		target.Transactions = Transactions{}
	}
	target.Transactions.merge(source.Transactions)
}

func mergeRuntimeCallTransportErrors(source, target *Transport) {
	if target.Errors == nil {
		target.Errors = Errors{}
	}
	target.Errors.merge(source.Errors)
}

func mergeRuntimeCallTransportFiles(source, target *Transport) {
	if target.Files == nil {
		target.Files = Files{}
	}
	target.Files.merge(source.Files)
}

// Merge a transport returned from a run-time call into another transport.
//
// source: The transport payload to merge.
// target: The target transport payload where to merge.
func mergeRuntimeCallTransport(source, target *Transport) {
	target.Meta.merge(source.Meta)

	if source.Data != nil {
		mergeRuntimeCallTransportData(source, target)
	}

	if source.Relations != nil {
		mergeRuntimeCallTransportRelations(source, target)
	}

	if source.Links != nil {
		mergeRuntimeCallTransportLinks(source, target)
	}

	if source.Calls != nil {
		mergeRuntimeCallTransportCalls(source, target)
	}

	if source.Transactions != nil {
		mergeRuntimeCallTransportTransactions(source, target)
	}

	if source.Errors != nil {
		mergeRuntimeCallTransportErrors(source, target)
	}

	if source.Files != nil {
		mergeRuntimeCallTransportFiles(source, target)
	}

	if source.Body == nil && target.Body != nil {
		source.Body = target.Body
	}
}

// Transport contains the transport payload data.
type Transport struct {
	reply        *Reply
	Meta         TransportMeta `json:"m"`
	Body         *File         `json:"b,omitempty"`
	Files        Files         `json:"f,omitempty"`
	Data         ServiceData   `json:"d,omitempty"`
	Relations    Relations     `json:"r,omitempty"`
	Links        Links         `json:"l,omitempty"`
	Transactions Transactions  `json:"t,omitempty"`
	Calls        Calls         `json:"C,omitempty"`
	Errors       Errors        `json:"e,omitempty"`
}

// Append files to the transport.
func (t *Transport) appendFiles(address, service, version, action string, files ...File) {
	if t.Files == nil {
		t.Files = Files{}
	}

	t.Files.append(address, service, version, action, files...)
}

// Add a relation to the transport.
func (t *Transport) setRelation(address, service, pk, remoteAddress, remoteService string, foreignKey interface{}) {
	if t.Relations == nil {
		t.Relations = Relations{}
	}

	t.Relations.add(address, service, pk, remoteAddress, remoteService, foreignKey)
}

// Append calls to the transport.
func (t *Transport) appendCalls(service, version string, calls ...Call) {
	if t.Calls == nil {
		t.Calls = Calls{}
	}

	t.Calls.append(service, version, calls...)
}

// Clone creates a clone of the transport.
//
// The returned transport won't keep references to the original transport values.
func (t *Transport) Clone() *Transport {
	transport := Transport{Meta: t.Meta}

	if t.Body != nil {
		body := *t.Body
		transport.Body = &body
	}

	if t.Files != nil {
		transport.Files = t.Files.clone()
	}

	if t.Data != nil {
		transport.Data = t.Data.clone()
	}

	if t.Relations != nil {
		transport.Relations = t.Relations.clone()
	}

	if t.Links != nil {
		transport.Links = t.Links.clone()
	}

	if t.Transactions != nil {
		transport.Transactions = t.Transactions.clone()
	}

	if t.Calls != nil {
		transport.Calls = t.Calls.clone()
	}

	if t.Errors != nil {
		transport.Errors = t.Errors.clone()
	}

	return &transport
}

// GetGateway returns the gateway addresses.
//
// The result contains two items, where the first item is the internal
// address and the second is the public address.
func (t *Transport) GetGateway() []string {
	if len(t.Meta.Gateway) == 0 {
		return []string{"", ""}
	}
	return t.Meta.Gateway
}

// GetOrigin returns the origin service.
//
// The result contains three items, where the first item is service name,
// the second is the version and the third is the action name.
func (t *Transport) GetOrigin() []string {
	if len(t.Meta.Origin) == 0 {
		return []string{"", "", ""}
	}
	return t.Meta.Origin
}

// GetLevel returns the depth of service requests.
func (t *Transport) GetLevel() uint {
	if t.Meta.Level == 0 {
		return 1
	}
	return t.Meta.Level
}

// SetReply assigns the the reply payload.
//
// reply: The reply payload.
func (t *Transport) SetReply(r *Reply) *Transport {
	t.reply = r
	return t
}

// SetDownload assigns a file to the body.
//
// file: The file to use as download content.
func (t *Transport) SetDownload(f *File) bool {
	t.Body = f
	return true
}

// SetReturn assigns the return value of an action.
//
// value: The value to use as return value in the payload.
func (t *Transport) SetReturn(value interface{}) bool {
	if t.reply != nil {
		t.reply.Command.Result.Return = value

		return true
	}

	return false
}

// SetData add data from a call to the transport payload.
//
// When there is existing data in the payload it is not removed. The new data
// is appended to the existing data in that case.
//
// name: The name of the Service.
// version: The version of the Service.
// action: The name of the action.
// data: The data to add.
func (t *Transport) SetData(name, version, action string, data interface{}) {
	if t.reply != nil {
		t.reply.Command.Result.Transport.SetData(name, version, action, data)
	}

	if t.Data == nil {
		t.Data = ServiceData{}
	}

	t.Data.append(t.GetGateway()[1], name, version, action, data)
}

// SetRelateOne adds a "one-to-one" relation.
//
// service: The name of the local service.
// pk: The primary key of the local entity.
// remote: The name of the remote service.
// fk: The primary key of the remote entity.
func (t *Transport) SetRelateOne(service, pk, remote, fk string) {
	if t.reply != nil {
		t.reply.Command.Result.Transport.SetRelateOne(service, pk, remote, fk)
	}

	gateway := t.GetGateway()[1]

	t.setRelation(gateway, service, pk, gateway, remote, fk)
}

// SetRelateMany adds a "many-to-many" relation.
//
// service: The name of the local service.
// pk: The primary key of the local entity.
// remote: The name of the remote service.
// fks: The primary keys of the remote entity.
func (t *Transport) SetRelateMany(service, pk, remote string, fks []string) {
	if t.reply != nil {
		t.reply.Command.Result.Transport.SetRelateMany(service, pk, remote, fks)
	}

	gateway := t.GetGateway()[1]

	t.setRelation(gateway, service, pk, gateway, remote, fks)
}

// SetRelateOneRemote adds a remote "one-to-one" relation.
//
// service: The name of the local service.
// pk: The primary key of the local entity.
// address: The address of the remote gateway.
// remote: The name of the remote service.
// fk: The primary key of the remote entity.
func (t *Transport) SetRelateOneRemote(service, pk, address, remote, fk string) {
	if t.reply != nil {
		t.reply.Command.Result.Transport.SetRelateOneRemote(service, pk, address, remote, fk)
	}

	t.setRelation(t.GetGateway()[1], service, pk, address, remote, fk)
}

// SetRelateManyRemote adds a remote "many-to-many" relation.
//
// service: The name of the local service.
// pk: The primary key of the local entity.
// address: The address of the remote gateway.
// remote: The name of the remote service.
// fks: The primary keys of the remote entity.
func (t *Transport) SetRelateManyRemote(service, pk, address, remote string, fks []string) {
	if t.reply != nil {
		t.reply.Command.Result.Transport.SetRelateManyRemote(service, pk, address, remote, fks)
	}

	t.setRelation(t.GetGateway()[1], service, pk, address, remote, fks)
}

// SetLink adds a link.
//
// service: The name of the Service.
// link: The link name.
// uri: The URI for the link.
func (t *Transport) SetLink(service, link, uri string) {
	if t.reply != nil {
		t.reply.Command.Result.Transport.SetLink(service, link, uri)
	}

	if t.Links == nil {
		t.Links = Links{}
	}

	t.Links.add(t.GetGateway()[1], service, link, uri)
}

// SetTransaction adds a transaction to be called when the request succeeds.
//
// command: The type of transaction.
// service: The name of the Service.
// version: The version of the Service.
// action: The name of the origin action.
// target: The name of the target action.
// params: Optional parameters for the transaction.
func (t *Transport) SetTransaction(command, service, version, action, target string, params []Param) {
	if t.reply != nil {
		t.reply.Command.Result.Transport.SetTransaction(command, service, version, action, target, params)
	}

	if t.Transactions == nil {
		t.Transactions = Transactions{}
	}

	t.Transactions.append(command, Transaction{
		Name:    service,
		Version: version,
		Action:  target,
		Caller:  action,
		Params:  params,
	})
}

// SetCall adds a run-time call.
//
// Current transport payload is used when the optional transport is not given.
//
// service: The name of the Service.
// version: The version of the Service.
// action: The name of the action making the call.
// callee_service: The called service.
// callee_version: The called version.
// callee_action: The called action.
// duration: The call duration.
// params: Optional parameters to send.
// files: Optional files to send.
// timeout: Optional timeout for the call.
// transport: Optional transport payload.
func (t *Transport) SetCall(
	service string,
	version string,
	action string,
	calleeService string,
	calleeVersion string,
	calleeAction string,
	duration uint,
	params []Param,
	files []File,
	timeout uint,
	transport *Transport,
) error {
	if duration == 0 {
		return errors.New("duration is required when adding run-time calls to transport")
	}

	if t.reply != nil {
		t.reply.Command.Result.Transport.SetCall(
			service,
			version,
			action,
			calleeService,
			calleeVersion,
			calleeAction,
			duration,
			params,
			files,
			timeout,
			transport,
		)
	}

	call := Call{
		Name:     calleeService,
		Version:  calleeVersion,
		Action:   calleeAction,
		Caller:   action,
		Duration: duration,
		Timeout:  timeout,
		Params:   params,
		Files:    files,
	}
	if transport != nil {
		// When a transport is present add the call to it and then merge it into the current transport
		transport.appendCalls(service, version, call)
		mergeRuntimeCallTransport(transport, t)
		// Update the transport in the reply payload with the runtime transport
		if t.reply != nil {
			t.reply.Command.Result.Transport = t.Clone()
		}
	} else {
		// When there is no transport just add the call to current transport
		t.appendCalls(service, version, call)
	}
	return nil
}

// SetDeferCall adds a deferred call.
//
// service: The name of the Service.
// version: The version of the Service.
// action: The name of the action making the call.
// callee_service: The called service.
// callee_version: The called version.
// callee_action: The called action.
// params: Optional parameters to send.
// files: Optional files to send.
func (t *Transport) SetDeferCall(
	service string,
	version string,
	action string,
	calleeService string,
	calleeVersion string,
	calleeAction string,
	params []Param,
	files []File,
) {
	if t.reply != nil {
		t.reply.Command.Result.Transport.SetDeferCall(
			service,
			version,
			action,
			calleeService,
			calleeVersion,
			calleeAction,
			params,
			files,
		)
	}

	t.appendCalls(service, version, Call{
		Name:    calleeService,
		Version: calleeVersion,
		Action:  calleeAction,
		Caller:  action,
		Params:  params,
		Files:   files,
	})
	//When there are files included in the call add them to the transport payload
	if len(files) > 0 {
		t.appendFiles(t.GetGateway()[1], calleeService, calleeVersion, calleeAction, files...)
	}
}

// SetRemoteCall adds a run-time call.
//
// Current transport payload is used when the optional transport is not given.
//
// address: The address of the remote Gateway.
// service: The name of the Service.
// version: The version of the Service.
// action: The name of the action making the call.
// callee_service: The called service.
// callee_version: The called version.
// callee_action: The called action.
// timeout: Optional timeout for the call.
// params: Optional parameters to send.
// files: Optional files to send.
func (t *Transport) SetRemoteCall(
	address string,
	service string,
	version string,
	action string,
	calleeService string,
	calleeVersion string,
	calleeAction string,
	timeout uint,
	params []Param,
	files []File,
) {
	if t.reply != nil {
		t.reply.Command.Result.Transport.SetRemoteCall(
			address,
			service,
			version,
			action,
			calleeService,
			calleeVersion,
			calleeAction,
			timeout,
			params,
			files,
		)
	}

	t.appendCalls(service, version, Call{
		Gateway: address,
		Name:    calleeService,
		Version: calleeVersion,
		Action:  calleeAction,
		Caller:  action,
		Timeout: timeout,
		Params:  params,
		Files:   files,
	})
	//When there are files included in the call add them to the transport payload
	if len(files) > 0 {
		t.appendFiles(t.GetGateway()[1], calleeService, calleeVersion, calleeAction, files...)
	}
}

// SetError adds a service error.
//
// service: The name of the Service.
// version: The version of the Service.
// message: The error message.
// code: The error code.
// status: The status message for the protocol.
func (t *Transport) SetError(service, version, message string, code int, status string) {
	if t.reply != nil {
		t.reply.Command.Result.Transport.SetError(service, version, message, code, status)
	}

	if t.Errors == nil {
		t.Errors = Errors{}
	}

	t.Errors.append(t.GetGateway()[1], service, version, Error{
		Message: message,
		Code:    code,
		Status:  status,
	})
}

// HasCalls checks if there are any type of calls registered for a Service.
//
// service: The name of the Service.
// version: The version of the Service.
func (t *Transport) HasCalls(service, version string) bool {
	if t.Calls == nil {
		return false
	}

	for _, call := range t.Calls.get(service, version) {
		// When duration is zero it means the call was not executed
		// so is safe to assume a call that has to be executed was found.
		if call.Duration == 0 {
			return true
		}
	}
	return false
}

// TransportMeta contains the metadata of the transport.
type TransportMeta struct {
	ID         string            `json:"i"`
	Version    string            `json:"v"`
	Datetime   string            `json:"d"`
	StartTime  string            `json:"s"`
	EndTime    string            `json:"e"`
	Duration   uint              `json:"D,omitempty"`
	Gateway    []string          `json:"g"`
	Origin     []string          `json:"o"`
	Level      uint              `json:"l"`
	Properties map[string]string `json:"p,omitempty"`
	Fallbacks  []Fallback        `json:"F,omitempty"`
}

func (t *TransportMeta) merge(meta TransportMeta) {
	// TODO: See how to merge fallbacks
	// t.Fallbacks.merge(meta.Fallbacks)

	// When there are properties to merge make sure the target meta is initialized
	if t.Properties == nil && meta.Properties != nil {
		t.Properties = make(map[string]string)
	}

	// Assign the properties to the target
	for name, value := range meta.Properties {
		// Don't overwrite existing properties
		if _, ok := t.Properties[name]; !ok {
			t.Properties[name] = value
		}
	}
}

// Fallback contains the triggered fallbacks.
type Fallback []interface{}

// GetName returns the service name.
func (f Fallback) GetName() string {
	if len(f) == 0 {
		return ""
	}

	name, _ := f[0].(string)
	return name
}

// GetVersion returns the service version.
func (f Fallback) GetVersion() string {
	if len(f) < 2 {
		return ""
	}

	version, _ := f[1].(string)
	return version
}

// GetActionNames returns the list of action names where fallbacks were triggered.
func (f Fallback) GetActionNames() (actions []string) {
	if len(f) < 3 {
		return nil
	}

	if names, ok := f[2].([]interface{}); ok {
		for _, v := range names {
			if action := v.(string); action != "" {
				actions = append(actions, action)
			}
		}
	}
	return actions
}

// Files contains the transport files.
type Files map[string]map[string]map[string]map[string][]File

func (f Files) clone() Files {
	clone := Files{}

	for address, services := range f {
		clone[address] = make(map[string]map[string]map[string][]File)

		for service, versions := range services {
			clone[address][service] = make(map[string]map[string][]File)

			for version, actions := range versions {
				clone[address][service][version] = make(map[string][]File)

				for action, files := range actions {
					clone[address][service][version][action] = append([]File{}, files...)
				}
			}
		}
	}

	return clone
}

func (f Files) append(address, service, version, action string, files ...File) {
	if v := f[address]; v == nil {
		f[address] = make(map[string]map[string]map[string][]File)
	}
	if v := f[address][service]; v == nil {
		f[address][service] = make(map[string]map[string][]File)
	}
	if v := f[address][service][version]; v == nil {
		f[address][service][version] = make(map[string][]File)
	}
	f[address][service][version][action] = append(f[address][service][version][action], files...)
}

func (f Files) merge(source Files) {
	for address, services := range source {
		if _, ok := f[address]; !ok {
			f[address] = services
			continue
		}

		for service, versions := range services {
			if _, ok := f[address][service]; !ok {
				f[address][service] = versions
				continue
			}

			for version, actions := range versions {
				if _, ok := f[address][service][version]; !ok {
					f[address][service][version] = actions
					continue
				}

				for action, files := range actions {
					currentFiles, ok := f[address][service][version][action]
					if !ok {
						f[address][service][version][action] = files
						continue
					}

					f[address][service][version][action] = append(currentFiles, files...)
				}
			}
		}
	}
}

// Get returns the files for the given service action.
//
// address: The gateway address.
// name: The service name.
// version: The service version.
// action: The action name.
func (f Files) Get(address, name, version, action string) []File {
	if services, ok := f[address]; ok {
		if versions, ok := services[name]; ok {
			if actions, ok := versions[version]; ok {
				return actions[action]
			}
		}
	}
	return nil
}

// ServiceData contains the transport data of the called services.
type ServiceData map[string]map[string]map[string]map[string][]interface{}

func (s ServiceData) clone() ServiceData {
	clone := ServiceData{}

	for address, services := range s {
		clone[address] = make(map[string]map[string]map[string][]interface{})

		for service, versions := range services {
			clone[address][service] = make(map[string]map[string][]interface{})

			for version, actions := range versions {
				clone[address][service][version] = make(map[string][]interface{})

				for action, data := range actions {
					clone[address][service][version][action] = append([]interface{}{}, data...)
				}
			}
		}
	}

	return clone
}

func (s ServiceData) append(address, service, version, action string, data ...interface{}) {
	if v := s[address]; v == nil {
		s[address] = make(map[string]map[string]map[string][]interface{})
	}
	if v := s[address][service]; v == nil {
		s[address][service] = make(map[string]map[string][]interface{})
	}
	if v := s[address][service][version]; v == nil {
		s[address][service][version] = make(map[string][]interface{})
	}
	s[address][service][version][action] = append(s[address][service][version][action], data...)
}

func (s ServiceData) merge(source ServiceData) {
	for address, services := range source {
		if _, ok := s[address]; !ok {
			s[address] = services
			continue
		}

		for service, versions := range services {
			if _, ok := s[address][service]; !ok {
				s[address][service] = versions
				continue
			}

			for version, actions := range versions {
				if _, ok := s[address][service][version]; !ok {
					s[address][service][version] = actions
					continue
				}

				for action, data := range actions {
					currentData, ok := s[address][service][version][action]
					if !ok {
						s[address][service][version][action] = data
						continue
					}

					s[address][service][version][action] = append(currentData, data...)
				}
			}
		}
	}
}

// Relations contains the transport relations.
type Relations map[string]map[string]map[string]map[string]map[string]interface{}

func (r Relations) clone() Relations {
	clone := Relations{}

	for address, services := range r {
		clone[address] = make(map[string]map[string]map[string]map[string]interface{})

		for service, pks := range services {
			clone[address][service] = make(map[string]map[string]map[string]interface{})

			for pk, remoteAddresses := range pks {
				clone[address][service][pk] = make(map[string]map[string]interface{})

				for remoteAddress, remoteServices := range remoteAddresses {
					clone[address][service][pk][remoteAddress] = make(map[string]interface{})

					for remoteService, foreignKey := range remoteServices {
						clone[address][service][pk][remoteAddress][remoteService] = foreignKey
					}
				}
			}
		}
	}

	return clone
}

func (r Relations) add(address, service, pk, remoteAddress, remoteService string, foreignKey interface{}) {
	if v := r[address]; v == nil {
		r[address] = make(map[string]map[string]map[string]map[string]interface{})
	}
	if v := r[address][service]; v == nil {
		r[address][service] = make(map[string]map[string]map[string]interface{})
	}
	if v := r[address][service][pk]; v == nil {
		r[address][service][pk] = make(map[string]map[string]interface{})
	}
	if v := r[address][service][pk][remoteAddress]; v == nil {
		r[address][service][pk][remoteAddress] = make(map[string]interface{})
	}
	r[address][service][pk][remoteAddress][remoteService] = foreignKey
}

func (r Relations) merge(source Relations) {
	for address, services := range source {
		if _, ok := r[address]; !ok {
			r[address] = services
			continue
		}

		for service, pks := range services {
			if _, ok := r[address][service]; !ok {
				r[address][service] = pks
				continue
			}

			for pk, remoteAddresses := range pks {
				if _, ok := r[address][service][pk]; !ok {
					r[address][service][pk] = remoteAddresses
					continue
				}

				for remoteAddress, remoteServices := range remoteAddresses {
					if _, ok := r[address][service][pk][remoteAddress]; !ok {
						r[address][service][pk][remoteAddress] = remoteServices
						continue
					}

					for remoteService, foreignKey := range remoteServices {
						// Add the foreign key(s) only when the relation doesn't exist
						if _, ok := r[address][service][pk][remoteAddress][remoteService]; !ok {
							r[address][service][pk][remoteAddress][remoteService] = foreignKey
						}
					}
				}
			}
		}
	}
}

// Links contains the transport links.
type Links map[string]map[string]map[string]string

func (l Links) clone() Links {
	clone := Links{}

	for address, services := range l {
		clone[address] = make(map[string]map[string]string)

		for service, links := range services {
			clone[address][service] = make(map[string]string)

			for link, uri := range links {
				clone[address][service][link] = uri
			}
		}
	}

	return clone
}

func (l Links) add(address, service, link, uri string) {
	if v := l[address]; v == nil {
		l[address] = make(map[string]map[string]string)
	}
	if v := l[address][service]; v == nil {
		l[address][service] = make(map[string]string)
	}
	l[address][service][link] = uri
}

func (l Links) merge(source Links) {
	for address, services := range source {
		if _, ok := l[address]; !ok {
			l[address] = services
			continue
		}

		for service, links := range services {
			if _, ok := l[address][service]; !ok {
				l[address][service] = links
				continue
			}

			for link, uri := range links {
				if _, ok := l[address][service][link]; !ok {
					l[address][service][link] = uri
				}
			}
		}
	}
}

// Transactions contains the transport transactions.
type Transactions map[string][]Transaction

// Get the transactions for a comman type.
func (t Transactions) Get(command string) (trx []Transaction) {
	if key := transactionKey(command); key != "" {
		trx = t[key]
	}
	return trx
}

func (t Transactions) clone() Transactions {
	clone := Transactions{}

	for key, trxs := range t {
		clone[key] = append(clone[key], trxs...)
	}

	return clone
}

func (t Transactions) append(command string, trxs ...Transaction) {
	// Append the transaction to the list of transactions for the current type
	if key := transactionKey(command); key != "" {
		t[key] = append(t[key], trxs...)
	}
}

func (t Transactions) merge(source Transactions) {
	for command, transactions := range source {
		currentTransactions, ok := t[command]
		if !ok {
			t[command] = transactions
			continue
		}

		t[command] = append(currentTransactions, transactions...)
	}
}

// Transaction represents a transaction object.
type Transaction struct {
	Name    string  `json:"n"`
	Version string  `json:"v"`
	Action  string  `json:"a"`
	Caller  string  `json:"C"`
	Params  []Param `json:"p,omitempty"`
}

// Calls contains the transport calls.
type Calls map[string]map[string][]Call

func (c Calls) get(service, version string) []Call {
	if versions, ok := c[service]; ok {
		return versions[version]
	}
	return nil
}

func (c Calls) clone() Calls {
	clone := Calls{}

	for service, versions := range c {
		clone[service] = versions

		for version, calls := range versions {
			clone[service][version] = append(clone[service][version], calls...)
		}
	}

	return clone
}

func (c Calls) append(service, version string, calls ...Call) {
	if v := c[service]; v == nil {
		c[service] = make(map[string][]Call)
	}
	c[service][version] = append(c[service][version], calls...)
}

func (c Calls) merge(source Calls) {
	for service, versions := range source {
		if _, ok := c[service]; !ok {
			c[service] = versions
			continue
		}

		for version, calls := range versions {
			currentCalls, ok := c[service][version]
			if !ok {
				c[service][version] = calls
				continue
			}

			c[service][version] = append(currentCalls, calls...)
		}
	}
}

// Call represents a call to a service.
type Call struct {
	Name     string  `json:"n"`
	Version  string  `json:"v"`
	Action   string  `json:"a"`
	Caller   string  `json:"C"`
	Duration uint    `json:"D"`
	Gateway  string  `json:"g,omitempty"`
	Timeout  uint    `json:"x,omitempty"`
	Params   []Param `json:"p,omitempty"`
	Files    []File  `json:"f,omitempty"`
}

// Errors contains the transport errors.
type Errors map[string]map[string]map[string][]Error

func (e Errors) clone() Errors {
	clone := Errors{}

	for address, services := range e {
		clone[address] = services

		for service, versions := range services {
			clone[address][service] = versions

			for version, errors := range versions {
				clone[address][service][version] = append(clone[address][service][version], errors...)
			}
		}
	}

	return clone
}

func (e Errors) append(address, service, version string, errors ...Error) {
	if v := e[address]; v == nil {
		e[address] = make(map[string]map[string][]Error)
	}
	if v := e[address][service]; v == nil {
		e[address][service] = make(map[string][]Error)
	}
	e[address][service][version] = append(e[address][service][version], errors...)
}

func (e Errors) merge(source Errors) {
	for address, services := range source {
		if _, ok := e[address]; !ok {
			e[address] = services
			continue
		}

		for service, versions := range services {
			if _, ok := e[address][service]; !ok {
				e[address][service] = versions
				continue
			}

			for version, errors := range versions {
				currentErrors, ok := e[address][service][version]
				if !ok {
					e[address][service][version] = errors
					continue
				}

				e[address][service][version] = append(currentErrors, errors...)
			}
		}
	}
}
