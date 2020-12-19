// Go SDK for the KUSANAGI(tm) framework (http://kusanagi.io)
// Copyright (c) 2016-2020 KUSANAGI S.L. All rights reserved.
//
// Distributed under the MIT license.
//
// For the full copyright and license information, please view the LICENSE
// file that was distributed with this source code.

package json

import (
	"bytes"
	"encoding/json"
)

// Serialize a value to a JSON representation.
func Serialize(value interface{}, pretty bool) (string, error) {
	buffer := new(bytes.Buffer)
	encoder := json.NewEncoder(buffer)
	if pretty {
		encoder.SetIndent("", "  ")
	}
	if err := encoder.Encode(value); err != nil {
		return "", err
	}
	return buffer.String(), nil
}

// Deserialize a value to a JSON representation.
func Deserialize(data string) (value interface{}, err error) {
	if err = json.Unmarshal([]byte(data), &value); err != nil {
		return nil, err
	}
	return value, nil
}
