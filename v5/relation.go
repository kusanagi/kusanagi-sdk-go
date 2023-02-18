// Go SDK for the KUSANAGI(tm) framework (http://kusanagi.io)
// Copyright (c) 2016-2023 KUSANAGI S.L. All rights reserved.
//
// Distributed under the MIT license.
//
// For the full copyright and license information, please view the LICENSE
// file that was distributed with this source code.

package kusanagi

// RelationTypeOne defines a one to one relation between two services.
const RelationTypeOne = "one"

// RelationTypeMany defines a one to many relation between two services.
const RelationTypeMany = "many"

// Relation represents a relation between two services.
type Relation struct {
	address string
	service string
	pk      string
	foreign map[string]map[string]interface{}
}

// GetAddress returns the gateway address for the service.
func (r Relation) GetAddress() string {
	return r.address
}

// GetName returns the name of the service.
func (r Relation) GetName() string {
	return r.service
}

// GetPrimaryKey returns the value for the primary key of the relation.
func (r Relation) GetPrimaryKey() string {
	return r.pk
}

// GetForeignRelations returns the relation data for the foreign services.
func (r Relation) GetForeignRelations() (relations []ForeignRelation) {
	// Get the remote gateway address and the foreign relations
	for address, services := range r.foreign {
		// Each relation belongs to a service in the remote gateway
		for service, fk := range services {
			relations = append(relations, ForeignRelation{address, service, fk})
		}
	}
	return relations
}

// ForeignRelation represent a foreign relation bwtewwn two services.
type ForeignRelation struct {
	address string
	service string
	fk      interface{}
}

// GetAddress returns the gateway address for the foreign service.
func (r ForeignRelation) GetAddress() string {
	return r.address
}

// GetName returns the name of the foreign service.
func (r ForeignRelation) GetName() string {
	return r.service
}

// GetType returns the type of the relation.
//
// Relation type can be either "one" or "many".
func (r ForeignRelation) GetType() string {
	if _, ok := r.fk.(string); ok {
		return RelationTypeOne
	}
	return RelationTypeMany
}

// GetForeignKeys returns the foreign key value(s) of the relation.
func (r ForeignRelation) GetForeignKeys() (fks []string) {
	if r.GetType() == RelationTypeOne {
		if fk, ok := r.fk.(string); ok {
			fks = append(fks, fk)
		}
	} else if items, ok := r.fk.([]interface{}); ok {
		for _, item := range items {
			if fk, ok := item.(string); ok {
				fks = append(fks, fk)
			}
		}
	}
	return fks
}
