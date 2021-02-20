// Go SDK for the KUSANAGI(tm) framework (http://kusanagi.io)
// Copyright (c) 2016-2021 KUSANAGI S.L. All rights reserved.
//
// Distributed under the MIT license.
//
// For the full copyright and license information, please view the LICENSE
// file that was distributed with this source code.

package kusanagi

import (
	"reflect"
)

// ActionData represents the action data stored in the transport by a service.
type ActionData struct {
	name string
	data []interface{}
}

// GetName returns the name of the service action that returned the data.
func (a ActionData) GetName() string {
	return a.name
}

// IsCollection checks if the data for this action is a collection.
func (a ActionData) IsCollection() bool {
	if len(a.data) == 0 {
		return false
	}
	return reflect.ValueOf(a.data[0]).Kind() == reflect.Slice
}

// GetData returns the transport data for the service action.
//
// Each item in the list represents a call that included data in the transport, where
// each item may be a slice or a map, depending on whether the data is a collection or not.
func (a ActionData) GetData() []interface{} {
	// Return a new copy of the list
	data := make([]interface{}, len(a.data))
	copy(data, a.data)
	return data
}
