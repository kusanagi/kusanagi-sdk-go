// Go SDK for the KUSANAGI(tm) framework (http://kusanagi.io)
// Copyright (c) 2016-2020 KUSANAGI S.L. All rights reserved.
//
// Distributed under the MIT license.
//
// For the full copyright and license information, please view the LICENSE
// file that was distributed with this source code.

package lib

import (
	"encoding/json"
	"fmt"
)

// StructToEntity converts a struct to an entity.
// The conversion is done using the json tags.
func StructToEntity(value interface{}) (map[string]interface{}, error) {
	// Serialize the value to JSON
	data, err := json.Marshal(value)
	if err != nil {
		return nil, err
	}

	// Restore the value into a dictionary
	result := make(map[string]interface{})
	if err := json.Unmarshal(data, &result); err != nil {
		return nil, fmt.Errorf("Failed to convert struct to entity: %v", err)
	}
	return result, nil
}

// SliceToCollection converts a slice to a collection.
// The conversion is done using the json tags.
func SliceToCollection(values interface{}) ([]map[string]interface{}, error) {
	// Serialize the values to JSON
	data, err := json.Marshal(values)
	if err != nil {
		return nil, err
	}

	// Restore the value into a dictionary
	var result []map[string]interface{}
	if err := json.Unmarshal(data, &result); err != nil {
		return nil, fmt.Errorf("Failed to convert slice to collection: %v", err)
	}
	return result, nil
}
