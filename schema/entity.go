// Go SDK for the KUSANAGI(tm) framework (http://kusanagi.io)
// Copyright (c) 2016-2019 KUSANAGI S.L. All rights reserved.
//
// Distributed under the MIT license.
//
// For the full copyright and license information, please view the LICENSE
// file that was distributed with this source code.

package schema

import "github.com/kusanagi/kusanagi-sdk-go/payload"

// Entity represents an entity definition
type Entity map[string]interface{}

// Creates an entity definition from a payload
func payloadToEntity(p *payload.Payload, e Entity) Entity {
	if e == nil {
		e = Entity{}
		// Add validate only to top level fields
		e["validate"] = p.GetBool("validate")
	}
	// When payload is empty return the entity
	if p.IsEmpty() {
		return e
	}

	// Add single fields to entity
	if p.Exists("field") {
		fields := []map[string]interface{}{}

		// Get field definitions
		for _, f := range p.GetSlice("field") {
			// Create a payload for current field definition
			fp := payload.New()
			fp.Data = f.(map[string]interface{})
			// Add field to the list of entity fields
			fields = append(fields, map[string]interface{}{
				"name":     fp.GetString("name"),
				"type":     fp.GetDefault("type", "string"),
				"optional": fp.GetBool("optional"),
			})
		}

		if len(fields) > 0 {
			e["field"] = fields
		}
	}

	// Add field sets to entity
	if p.Exists("fields") {
		fields := []map[string]interface{}{}

		// Get field set definitions
		for _, f := range p.GetSlice("fields") {
			// Create a payload for current field set definition
			fp := payload.New()
			fp.Data = f.(map[string]interface{})
			// Create a field set
			fieldset := map[string]interface{}{
				"name":     fp.GetString("name"),
				"optional": fp.GetBool("optional"),
			}
			// When field set contains inner sets or fields parse it
			if fp.Exists("field") || fp.Exists("fields") {
				fieldset = payloadToEntity(fp, fieldset)
			}
			fields = append(fields, fieldset)
		}

		if len(fields) > 0 {
			e["fields"] = fields
		}
	}

	return e
}
