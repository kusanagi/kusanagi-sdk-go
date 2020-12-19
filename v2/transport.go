// Go SDK for the KUSANAGI(tm) framework (http://kusanagi.io)
// Copyright (c) 2016-2020 KUSANAGI S.L. All rights reserved.
//
// Distributed under the MIT license.
//
// For the full copyright and license information, please view the LICENSE
// file that was distributed with this source code.

package kusanagi

import (
	"fmt"

	"github.com/kusanagi/kusanagi-sdk-go/v2/lib/payload"
)

// Transport encapsulates the transport object.
type Transport struct {
	payload *payload.Transport
}

// GetRequestID returns the UUID of the request.
func (t Transport) GetRequestID() string {
	return t.payload.Meta.ID
}

// GetRequestTimestamp returns the request creation timestamp.
func (t Transport) GetRequestTimestamp() string {
	return t.payload.Meta.Datetime
}

// GetOriginService returns the origin of the request.
//
// Result is an array containing name, version and action
// of the service that was the origin of the request.
func (t Transport) GetOriginService() []string {
	return t.payload.Meta.Origin
}

// GetDuration returns the service execution time in milliseconds.
//
// This time is the number of milliseconds spent by the service that was the origin of the request.
func (t Transport) GetOriginDuration() uint {
	return t.payload.Meta.Duration
}

// GetProperty returns a userland property value.
//
// The name of the property is case sensitive.
//
// An empty string is returned when a property with the specified
// name does not exist, and no default value is provided.
//
// name: The name of the property.
// preset: The default value to use when the property doesn't exist.
func (t Transport) GetProperty(name, preset string) string {
	if properties := t.payload.Meta.Properties; properties != nil {
		if value, ok := properties[name]; ok {
			return value
		}
	}
	return preset
}

// GetProperties returns all the userland properties.
func (t Transport) GetProperties() (properties map[string]string) {
	if t.payload.Meta.Properties != nil {
		// Create a new map with the transport meta properties
		properties = make(map[string]string)
		for name, value := range t.payload.Meta.Properties {
			properties[name] = value
		}
	}
	return properties
}

// HasDownload checks if a file download has been registered for the response.
func (t Transport) HasDownload() bool {
	return t.payload.Body != nil
}

// GetDownload returns the file download registered for the response.
func (t Transport) GetDownload() (f *File) {
	if t.payload.Body != nil {
		*f = payloadToFile(*t.payload.Body)
	}
	return f
}

// GetData returns the transport data.
func (t Transport) GetData() (data []ServiceData) {
	for address, services := range *t.payload.Data {
		for service, versions := range services {
			for version, actions := range versions {
				data = append(data, ServiceData{address, service, version, actions})
			}
		}
	}
	return data
}

// GetRelations returns the service relations.
func (t Transport) GetRelations() (relations []Relation) {
	for address, services := range *t.payload.Relations {
		for service, pks := range services {
			for pk, foreign := range pks {
				relations = append(relations, Relation{address, service, pk, foreign})
			}
		}
	}
	return relations
}

// GetLinks returns the service links.
func (t Transport) GetLinks() (links []Link) {
	for address, services := range *t.payload.Links {
		for service, references := range services {
			for ref, uri := range references {
				links = append(links, Link{address, service, ref, uri})
			}
		}
	}
	return links
}

// GetCalls returns the service calls.
func (t Transport) GetCalls() (callers []Caller) {
	for service, versions := range *t.payload.Calls {
		for version, calls := range versions {
			for _, call := range calls {
				callee := Callee{
					gateway:  call.Gateway,
					name:     call.Name,
					version:  call.Version,
					action:   call.Action,
					duration: call.Duration,
					timeout:  call.Timeout,
					params:   payloadToParams(call.Params),
				}
				action := call.Caller
				callers = append(callers, Caller{service, version, action, callee})
			}
		}
	}
	return callers
}

// GetTransactions returns the transactions for a specific type.
//
// The transaction type is case sensitive, and supports "commit", "rollback" or "complete" as value.
//
// command: The transaction command.
func (t Transport) GetTransactions(command string) ([]Transaction, error) {
	if command != Commit && command != Rollback && command != Complete {
		return nil, fmt.Errorf(`invalid transaction command: "%s"`, command)
	}

	var transactions []Transaction
	for _, trx := range t.payload.Transactions.Get(command) {
		transactions = append(transactions, Transaction{
			command: command,
			name:    trx.Name,
			version: trx.Version,
			action:  trx.Action,
			caller:  trx.Caller,
			params:  payloadToParams(trx.Params),
		})
	}
	return transactions, nil
}

// GetErrors returns the transport errors.
func (t Transport) GetErrors() (errors []Error) {
	for address, services := range *t.payload.Errors {
		for service, versions := range services {
			for version, errors := range versions {
				for _, err := range errors {
					errors = append(errors, Error{
						address: address,
						service: service,
						version: version,
						message: err.GetMessage(),
						code:    err.GetCode(),
						status:  err.GetStatus(),
					})
				}
			}
		}
	}
	return errors
}
