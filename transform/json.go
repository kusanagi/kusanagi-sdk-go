// Go SDK for the KUSANAGI(tm) framework (http://kusanagi.io)
// Copyright (c) 2016-2020 KUSANAGI S.L. All rights reserved.
//
// Distributed under the MIT license.
//
// For the full copyright and license information, please view the LICENSE
// file that was distributed with this source code.

package transform

import (
	"bytes"
	"encoding/json"
	"reflect"

	"github.com/ugorji/go/codec"
)

// TODO: Implement a custom encoder/decoder for dates/times/datetimes/decimal

// Serialize convert a value to a JSON string.
func Serialize(value interface{}, pretty bool) ([]byte, error) {
	var (
		stream []byte
		buffer bytes.Buffer
	)

	// Prepare JSON encoder
	handle := new(codec.JsonHandle)
	handle.Canonical = true // This allows sorting of keys
	encoder := codec.NewEncoderBytes(&stream, handle)

	// Encode value to a stream
	err := encoder.Encode(value)
	if err != nil {
		return nil, err
	}

	if pretty {
		// Format JSON output to be more readable
		err = json.Indent(&buffer, stream, "", "  ")
	} else {
		// Compact JSON output to use less bytes
		err = json.Compact(&buffer, stream)
	}
	if err != nil {
		return nil, err
	}
	return buffer.Bytes(), nil
}

// Deserialize converts a JSON value to its original type.
func Deserialize(stream []byte, value interface{}) error {
	// Prepare JSON decoder
	handle := new(codec.JsonHandle)
	handle.MapType = reflect.TypeOf(map[string]interface{}(nil))
	decoder := codec.NewDecoderBytes(stream, handle)

	// Decode the stream to a type
	if err := decoder.Decode(value); err != nil {
		return err
	}
	return nil
}

// AreEqual compares two types using JSON serialization.
func AreEqual(map1, map2 interface{}) bool {
	json1, err := Serialize(map1, false)
	if err != nil {
		return false
	}

	json2, err := Serialize(map2, false)
	if err != nil {
		return false
	}
	return string(json1) == string(json2)
}
